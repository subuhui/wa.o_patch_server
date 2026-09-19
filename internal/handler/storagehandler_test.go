package handler

import (
	"context"
	"github.com/zeromicro/go-zero/rest/router"
	"net/http"
	"net/http/httptest"
	"shorebird-server/internal/storage"
	"shorebird-server/internal/svc"
	"strings"
	"testing"
)

func TestGeneratedLocalDownloadURL(t *testing.T) {
	st, err := storage.NewLocalStorage(t.TempDir(), "http://localhost")
	if err != nil {
		t.Fatal(err)
	}
	uri, err := st.Upload(context.Background(), "patches/app/1/android arm64+#.diff", strings.NewReader("artifact"), 8, "")
	if err != nil {
		t.Fatal(err)
	}
	r := router.NewRouter()
	if err := r.Handle(http.MethodGet, "/api/v1/storage/download", DownloadHandler(&svc.ServiceContext{Storage: st})); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, uri, nil))
	if w.Code != 200 || w.Body.String() != "artifact" {
		t.Fatalf("status=%d body=%q", w.Code, w.Body.String())
	}
}
