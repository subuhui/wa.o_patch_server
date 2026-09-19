package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDiagnostics(t *testing.T) {
	h := newDiagnosticsHandler("https://private.example")
	link := func(direction string) string {
		t.Helper()
		w := httptest.NewRecorder()
		h.link(direction)(w, httptest.NewRequest(http.MethodGet, "/", nil))
		var body map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		url := body[direction+"_url"]
		if !strings.HasPrefix(url, "https://private.example/") {
			t.Fatal(url)
		}
		return url
	}
	downloadURL := link("download")
	w := httptest.NewRecorder()
	h.transfer("download")(w, httptest.NewRequest(http.MethodGet, downloadURL, nil))
	if w.Code != 200 || w.Body.Len() != diagnosticDownloadSize {
		t.Fatalf("%d %d", w.Code, w.Body.Len())
	}
	uploadURL := link("upload")
	for _, size := range []int{diagnosticUploadSize, 1} {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "test")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(make([]byte, size)); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, uploadURL, &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w = httptest.NewRecorder()
		h.transfer("upload")(w, req)
		want := 204
		if size == 1 {
			want = 400
		}
		if w.Code != want {
			t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
		}
	}
	for _, url := range []string{
		"/?expires=1&signature=" + h.signature("download", "1"),
		downloadURL + "&signature=bad",
		uploadURL,
	} {
		req := httptest.NewRequest(http.MethodGet, url, nil)
		if strings.Contains(url, "signature=bad") {
			q := req.URL.Query()
			q.Set("signature", "bad")
			req.URL.RawQuery = q.Encode()
		}
		w = httptest.NewRecorder()
		h.transfer("download")(w, req)
		if w.Code != 403 {
			t.Fatalf("invalid token accepted: %d", w.Code)
		}
	}
	w = httptest.NewRecorder()
	TransferAppHandler(w, httptest.NewRequest(http.MethodPost, "/", nil))
	if w.Code != 501 || !strings.Contains(w.Body.String(), "not_implemented") {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
}
