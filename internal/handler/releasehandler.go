package handler

import (
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/rest/httpx"

	"shorebird-server/internal/logic"
	"shorebird-server/internal/svc"
	"shorebird-server/internal/types"
)

func GetReleasesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AppIDPathReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewReleaseLogic(r.Context(), svcCtx)
		resp, err := l.GetReleases(req.AppID)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func CreateReleaseHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateReleaseReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewReleaseLogic(r.Context(), svcCtx)
		resp, err := l.CreateRelease(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusCreated, resp)
		}
	}
}

func UpdateReleaseStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateReleaseStatusReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewReleaseLogic(r.Context(), svcCtx)
		if err := l.UpdateReleaseStatus(req.ReleaseID, req.Status); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func GetReleaseArtifactsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetReleaseArtifactsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewReleaseLogic(r.Context(), svcCtx)
		resp, err := l.GetReleaseArtifacts(req.ReleaseID, req.Arch, req.Platform)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func CreateReleaseArtifactHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pathReq types.ReleasePathReq
		if err := httpx.Parse(r, &pathReq); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		_ = r.ParseMultipartForm(32 << 20)
		arch := r.FormValue("arch")
		platform := r.FormValue("platform")
		hash := r.FormValue("hash")
		filename := r.FormValue("filename")
		podfileLockHash := r.FormValue("podfile_lock_hash")
		canSideload := r.FormValue("can_sideload") == "true"
		size, _ := strconv.ParseInt(r.FormValue("size"), 10, 64)

		l := logic.NewReleaseLogic(r.Context(), svcCtx)
		resp, err := l.CreateReleaseArtifact(pathReq.AppID, pathReq.ReleaseID, arch, platform, hash, filename, podfileLockHash, size, canSideload)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusCreated, resp)
		}
	}
}
