package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"

	"shorebird-server/internal/logic"
	"shorebird-server/internal/svc"
)

func UploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			httpx.WriteJson(w, http.StatusBadRequest, map[string]string{"message": "missing key query parameter"})
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.WriteJson(w, http.StatusBadRequest, map[string]string{"message": fmt.Sprintf("failed to get form file 'file': %v", err)})
			return
		}
		defer file.Close()

		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		l := logic.NewStorageLogic(r.Context(), svcCtx)
		resp, err := l.Upload(key, file, header.Size, contentType)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func DownloadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			path := r.URL.Path
			const prefix = "/api/v1/storage/download/"
			if strings.HasPrefix(path, prefix) {
				key, _ = url.PathUnescape(strings.TrimPrefix(path, prefix))
			}
		}

		if key == "" {
			http.Error(w, "missing key", http.StatusBadRequest)
			return
		}

		l := logic.NewStorageLogic(r.Context(), svcCtx)
		reader, size, err := l.GetObject(key)
		if err != nil {
			http.Error(w, fmt.Sprintf("artifact not found: %v", err), http.StatusNotFound)
			return
		}
		defer reader.Close()

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", key))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		w.WriteHeader(http.StatusOK)

		_, _ = io.Copy(w, reader)
	}
}
