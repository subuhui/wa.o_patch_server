package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"

	"shorebird-server/internal/svc"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	// 1. Health check & root
	server.AddRoutes([]rest.Route{
		{
			Method: http.MethodGet,
			Path:   "/health",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"status":"ok"}`))
			},
		},
		{
			Method: http.MethodGet,
			Path:   "/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"service":"shorebird-private-server","framework":"go-zero","version":"1.0.0"}`))
			},
		},
	})

	// 2. Public client updater endpoints (no auth required)
	server.AddRoutes([]rest.Route{
		{
			Method:  http.MethodPost,
			Path:    "/api/v1/patches/check",
			Handler: CheckPatchHandler(serverCtx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/api/v1/patches/events",
			Handler: TrackEventHandler(serverCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/v1/storage/download",
			Handler: DownloadHandler(serverCtx),
		},
		{
			Method:  http.MethodGet,
			Path:    "/api/v1/storage/download/:key",
			Handler: DownloadHandler(serverCtx),
		},
	})

	// 3. Protected CLI endpoints (wrapped with serverCtx.AuthMiddleware)
	server.AddRoutes(
		rest.WithMiddleware(
			serverCtx.AuthMiddleware,
			// Current user & account
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/api/v1/users/me",
				Handler: GetCurrentUserHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/api/v1/plan",
				Handler: GetPlanHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/api/v1/organizations",
				Handler: GetOrganizationsHandler(serverCtx),
			},
			// Apps & channels
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/api/v1/apps",
				Handler: GetAppsHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/api/v1/apps",
				Handler: CreateAppHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodDelete,
				Path:    "/api/v1/apps/:app_id",
				Handler: DeleteAppHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/api/v1/apps/:app_id/channels",
				Handler: GetChannelsHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/api/v1/apps/:app_id/channels",
				Handler: CreateChannelHandler(serverCtx),
			},
			// Releases
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/api/v1/apps/:app_id/releases",
				Handler: GetReleasesHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/api/v1/apps/:app_id/releases",
				Handler: CreateReleaseHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPatch,
				Path:    "/api/v1/apps/:app_id/releases/:release_id",
				Handler: UpdateReleaseStatusHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/api/v1/apps/:app_id/releases/:release_id/artifacts",
				Handler: GetReleaseArtifactsHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/api/v1/apps/:app_id/releases/:release_id/artifacts",
				Handler: CreateReleaseArtifactHandler(serverCtx),
			},
			// Patches
			rest.Route{
				Method:  http.MethodGet,
				Path:    "/api/v1/apps/:app_id/releases/:release_id/patches",
				Handler: GetPatchesHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/api/v1/apps/:app_id/patches",
				Handler: CreatePatchHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPatch,
				Path:    "/api/v1/apps/:app_id/patches/:patch_id",
				Handler: UpdatePatchHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/api/v1/apps/:app_id/patches/promote",
				Handler: PromotePatchHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/api/v1/apps/:app_id/patches/:patch_id/promote",
				Handler: PromotePatchHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/api/v1/apps/:app_id/releases/:release_id/patches/:patch_id/rollback",
				Handler: RollbackPatchHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/api/v1/apps/:app_id/patches/:patch_id/rollback",
				Handler: RollbackPatchHandler(serverCtx),
			},
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/api/v1/apps/:app_id/patches/:patch_id/artifacts",
				Handler: CreatePatchArtifactHandler(serverCtx),
			},
			// Storage upload
			rest.Route{
				Method:  http.MethodPost,
				Path:    "/api/v1/storage/upload",
				Handler: UploadHandler(serverCtx),
			},
		),
	)
}
