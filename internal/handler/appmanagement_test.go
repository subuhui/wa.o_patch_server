package handler

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/rest/router"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"shorebird-server/internal/db"
	"shorebird-server/internal/svc"
)

func TestAppManagementHTTP(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if err := database.AutoMigrate(&db.App{}, &db.Channel{}, &db.Release{}, &db.Patch{}); err != nil {
		t.Fatal(err)
	}
	channelID := uint(1)
	for _, record := range []any{
		&db.App{ID: "app", DisplayName: "old"},
		&db.Channel{ID: 1, AppID: "app", Name: "qa"},
		&db.Release{ID: 1, AppID: "app", Version: "1"},
		&db.Patch{ID: 1, ReleaseID: 1, ChannelID: &channelID, Number: 1, Status: "rolled_back"},
	} {
		if err := database.Create(record).Error; err != nil {
			t.Fatal(err)
		}
	}
	s := &svc.ServiceContext{DB: database}
	r := router.NewRouter()
	for _, route := range []struct {
		method, path string
		handler      http.HandlerFunc
	}{
		{http.MethodPatch, "/api/v1/apps/:app_id", RenameAppHandler(s)},
		{http.MethodDelete, "/api/v1/apps/:app_id/channels/:channel_id", DeleteChannelHandler(s)},
		{http.MethodPost, "/api/v1/apps/:app_id/releases/:release_id/patches/:patch_id/rollforward", RollforwardPatchHandler(s)},
	} {
		if err := r.Handle(route.method, route.path, route.handler); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{http.MethodPatch, "/api/v1/apps/app", `{"name":"新名称"}`, 204},
		{http.MethodPost, "/api/v1/apps/app/releases/1/patches/1/rollforward", "", 204},
		{http.MethodPost, "/api/v1/apps/app/releases/1/patches/1/rollforward", "", 304},
		{http.MethodDelete, "/api/v1/apps/app/channels/1", "", 204},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tc.status || w.Body.Len() != 0 {
			t.Fatalf("%s: status=%d body=%s", tc.path, w.Code, w.Body.String())
		}
	}
	var app db.App
	if err := database.First(&app, "id = ?", "app").Error; err != nil {
		t.Fatal(err)
	}
	if app.DisplayName != "新名称" {
		t.Fatal(app.DisplayName)
	}
}
