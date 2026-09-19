package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"shorebird-server/internal/logic"
	"shorebird-server/internal/svc"
)

func RenameAppHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			AppID string `path:"app_id"`
			Name  string `json:"name"`
		}
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if err := logic.NewAppLogic(r.Context(), svcCtx).RenameApp(req.AppID, req.Name); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteChannelHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			AppID     string `path:"app_id"`
			ChannelID string `path:"channel_id"`
		}
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		if err := logic.NewAppLogic(r.Context(), svcCtx).DeleteChannel(req.AppID, req.ChannelID); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func RollforwardPatchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			AppID     string `path:"app_id"`
			ReleaseID string `path:"release_id"`
			PatchID   string `path:"patch_id"`
		}
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		changed, err := logic.NewPatchLogic(r.Context(), svcCtx).RollforwardPatch(req.AppID, req.ReleaseID, req.PatchID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else if !changed {
			w.WriteHeader(http.StatusNotModified)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
