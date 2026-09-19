package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"shorebird-server/internal/logic"
	"shorebird-server/internal/svc"
)

func GetCurrentUserHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewUserLogic(r.Context(), svcCtx)
		resp, err := l.GetCurrentUser()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

func GetPlanHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{
			"level": "enterprise",
		})
	}
}

func GetOrganizationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{
			"organizations": []map[string]interface{}{
				{
					"organization": map[string]interface{}{
						"id":                1,
						"name":              "Personal",
						"organization_type": "personal",
						"created_at":        "2026-01-01T00:00:00Z",
						"updated_at":        "2026-01-01T00:00:00Z",
					},
					"role": "owner",
				},
			},
		})
	}
}
