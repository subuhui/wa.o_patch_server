package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/rest/httpx"
)

const diagnosticDownloadSize = 16000000
const diagnosticUploadSize = 5000000

type diagnosticsHandler struct {
	publicURL string
	key       [32]byte
	payload   [64000]byte
}

func newDiagnosticsHandler(publicURL string) *diagnosticsHandler {
	h := &diagnosticsHandler{publicURL: strings.TrimRight(publicURL, "/")}
	if _, err := rand.Read(h.key[:]); err != nil {
		panic(err)
	}
	if _, err := rand.Read(h.payload[:]); err != nil {
		panic(err)
	}
	return h
}

func (h *diagnosticsHandler) signature(direction, expires string) string {
	mac := hmac.New(sha256.New, h.key[:])
	_, _ = mac.Write([]byte(direction + ":" + expires))
	return hex.EncodeToString(mac.Sum(nil))
}

func (h *diagnosticsHandler) link(direction string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.publicURL == "" {
			httpx.WriteJson(w, http.StatusServiceUnavailable, map[string]string{"code": "unavailable", "message": "PublicURL is required for diagnostics"})
			return
		}
		expires := strconv.FormatInt(time.Now().Add(5*time.Minute).Unix(), 10)
		query := url.Values{"expires": {expires}, "signature": {h.signature(direction, expires)}}
		httpx.OkJson(w, map[string]string{direction + "_url": h.publicURL + "/api/v1/diagnostics/" + direction + "?" + query.Encode()})
	}
}

func (h *diagnosticsHandler) transfer(direction string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		expires := r.URL.Query().Get("expires")
		deadline, err := strconv.ParseInt(expires, 10, 64)
		if err != nil || deadline <= time.Now().Unix() ||
			!hmac.Equal([]byte(r.URL.Query().Get("signature")), []byte(h.signature(direction, expires))) {
			http.Error(w, "invalid or expired diagnostic URL", http.StatusForbidden)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		if direction == "download" {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Length", strconv.Itoa(diagnosticDownloadSize))
			for remaining := diagnosticDownloadSize; remaining > 0; {
				n := min(remaining, len(h.payload))
				if _, err := w.Write(h.payload[:n]); err != nil {
					return
				}
				remaining -= n
			}
			return
		}
		// Consume the multipart stream without persisting a test artifact.
		r.Body = http.MaxBytesReader(w, r.Body, diagnosticUploadSize+65536)
		reader, err := r.MultipartReader()
		if err != nil {
			http.Error(w, "multipart upload required", http.StatusBadRequest)
			return
		}
		found := false
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				http.Error(w, "invalid multipart upload", http.StatusBadRequest)
				return
			}
			size, err := io.Copy(io.Discard, part)
			if err != nil {
				http.Error(w, "invalid or oversized upload", http.StatusBadRequest)
				return
			}
			if part.FormName() == "file" {
				if found || size != diagnosticUploadSize {
					http.Error(w, fmt.Sprintf("expected one %d-byte file", diagnosticUploadSize), http.StatusBadRequest)
					return
				}
				found = true
			}
		}
		if !found {
			http.Error(w, "file is required", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func TransferAppHandler(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJson(w, http.StatusNotImplemented, map[string]string{
		"code":    "not_implemented",
		"message": "App transfer is not supported by this single-tenant server.",
	})
}
