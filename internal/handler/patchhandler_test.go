package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/zeromicro/go-zero/rest/router"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"shorebird-server/internal/db"
	"shorebird-server/internal/svc"
)

func TestRollbackPatchHTTP(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if err := database.AutoMigrate(&db.Patch{}); err != nil {
		t.Fatal(err)
	}
	for _, route := range []string{
		"/api/v1/apps/:app_id/releases/:release_id/patches/:patch_id/rollback",
		"/api/v1/apps/:app_id/patches/:patch_id/rollback",
	} {
		t.Run(route, func(t *testing.T) {
			p := db.Patch{ReleaseID: 1, Status: "active"}
			if err := database.Create(&p).Error; err != nil {
				t.Fatal(err)
			}
			r := router.NewRouter()
			if err := r.Handle(http.MethodPost, route, RollbackPatchHandler(&svc.ServiceContext{DB: database})); err != nil {
				t.Fatal(err)
			}
			url := fmt.Sprintf("/api/v1/apps/app/patches/%d/rollback", p.ID)
			if route == "/api/v1/apps/:app_id/releases/:release_id/patches/:patch_id/rollback" {
				url = fmt.Sprintf("/api/v1/apps/app/releases/1/patches/%d/rollback", p.ID)
			}
			for _, want := range []int{http.StatusOK, http.StatusNotModified} {
				w := httptest.NewRecorder()
				r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, url, nil))
				if w.Code != want {
					t.Fatalf("status=%d, want=%d, body=%s", w.Code, want, w.Body.String())
				}
				if want == http.StatusNotModified && w.Body.Len() != 0 {
					t.Fatalf("304 has a body: %s", w.Body.String())
				}
			}
			if err := database.First(&p, p.ID).Error; err != nil {
				t.Fatal(err)
			}
			if p.Status != "rolled_back" {
				t.Fatalf("status=%s", p.Status)
			}
			if err := database.Delete(&p).Error; err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, url, nil))
			if w.Code < 400 {
				t.Fatalf("missing patch returned status=%d", w.Code)
			}
		})
	}
}
