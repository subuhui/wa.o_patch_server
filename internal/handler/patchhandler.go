package handler

import (
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"

	"shorebird-server/internal/logic"
	"shorebird-server/internal/svc"
	"shorebird-server/internal/types"
)

func GetPatchesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReleasePathReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewPatchLogic(r.Context(), svcCtx)
		resp, err := l.GetPatches(req.ReleaseID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func CreatePatchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreatePatchReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewPatchLogic(r.Context(), svcCtx)
		resp, err := l.CreatePatch(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusCreated, resp)
		}
	}
}

func UpdatePatchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdatePatchReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewPatchLogic(r.Context(), svcCtx)
		if err := l.UpdatePatch(req.PatchID, req.Status); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func PromotePatchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PromotePatchReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		patchIDStr := ""
		if req.PatchID > 0 {
			patchIDStr = strconv.Itoa(req.PatchID)
		} else {
			patchIDStr = pathvar.Vars(r)["patch_id"]
		}

		l := logic.NewPatchLogic(r.Context(), svcCtx)
		if err := l.PromotePatch(req.AppID, patchIDStr, req.ChannelID); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func RollbackPatchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RollbackPatchReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		patchIDStr := req.PatchID
		if patchIDStr == "" {
			patchIDStr = pathvar.Vars(r)["patch_id"]
		}

		l := logic.NewPatchLogic(r.Context(), svcCtx)
		if err := l.RollbackPatch(patchIDStr); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, map[string]bool{"success": true})
		}
	}
}

func CreatePatchArtifactHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.PatchPathReq
		if err := httpx.Parse(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		_ = r.ParseMultipartForm(32 << 20)
		arch := r.FormValue("arch")
		platform := r.FormValue("platform")
		hash := r.FormValue("hash")
		hashSig := r.FormValue("hash_signature")
		podfileLockHash := r.FormValue("podfile_lock_hash")
		size, _ := strconv.ParseInt(r.FormValue("size"), 10, 64)

		l := logic.NewPatchLogic(r.Context(), svcCtx)
		resp, err := l.CreatePatchArtifact(pathReq.AppID, pathReq.PatchID, arch, platform, hash, hashSig, podfileLockHash, size)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusCreated, resp)
		}
	}
}
