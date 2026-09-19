package logic

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"

	"shorebird-server/internal/db"
	"shorebird-server/internal/svc"
	"shorebird-server/internal/types"
)

type AppLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppLogic {
	return &AppLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppLogic) GetApps() (*types.GetAppsResp, error) {
	var apps []db.App
	if err := l.svcCtx.DB.Find(&apps).Error; err != nil {
		return nil, err
	}

	result := make([]types.AppMetadataResp, 0, len(apps))
	for _, app := range apps {
		var latestRel db.Release
		var latestPatch db.Patch
		var latestVersion *string
		var latestPatchNum *int

		if err := l.svcCtx.DB.Where("app_id = ? AND status = 'active'", app.ID).Order("id DESC").First(&latestRel).Error; err == nil {
			latestVersion = &latestRel.Version
			if err := l.svcCtx.DB.Where("release_id = ? AND status = 'active'", latestRel.ID).Order("number DESC").First(&latestPatch).Error; err == nil {
				latestPatchNum = &latestPatch.Number
			}
		}

		result = append(result, types.AppMetadataResp{
			AppID:                app.ID,
			DisplayName:          app.DisplayName,
			LatestReleaseVersion: latestVersion,
			LatestPatchNumber:    latestPatchNum,
			CreatedAt:            app.CreatedAt.Format(time.RFC3339),
			UpdatedAt:            app.UpdatedAt.Format(time.RFC3339),
			Platforms:            []string{"android", "ios"},
		})
	}

	return &types.GetAppsResp{Apps: result}, nil
}

func (l *AppLogic) CreateApp(req *types.CreateAppReq) (*types.AppResp, error) {
	app := db.App{
		ID:          uuid.New().String(),
		DisplayName: req.DisplayName,
	}

	if err := l.svcCtx.DB.Create(&app).Error; err != nil {
		return nil, err
	}

	// Auto-create default 'stable' channel
	channel := db.Channel{
		AppID: app.ID,
		Name:  "stable",
	}
	_ = l.svcCtx.DB.Create(&channel)

	return &types.AppResp{
		ID:          app.ID,
		DisplayName: app.DisplayName,
	}, nil
}

func (l *AppLogic) DeleteApp(appID string) error {
	return l.svcCtx.DB.Where("id = ?", appID).Delete(&db.App{}).Error
}

func (l *AppLogic) GetChannels(appID string) ([]db.Channel, error) {
	var channels []db.Channel
	err := l.svcCtx.DB.Where("app_id = ?", appID).Find(&channels).Error
	return channels, err
}

func (l *AppLogic) CreateChannel(appID string, req *types.CreateChannelReq) (*db.Channel, error) {
	channel := db.Channel{
		AppID: appID,
		Name:  req.Channel,
	}
	if err := l.svcCtx.DB.Create(&channel).Error; err != nil {
		return nil, err
	}
	return &channel, nil
}
