package logic

import (
	"context"
	"errors"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"

	"shorebird-server/internal/db"
	"shorebird-server/internal/svc"
	"shorebird-server/internal/types"
)

type CheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckLogic {
	return &CheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckLogic) CheckPatch(req *types.PatchCheckReq) (*types.PatchCheckResp, error) {
	// 1. Locate the Release matching app_id and release_version
	var release db.Release
	err := l.svcCtx.DB.Where("app_id = ? AND version = ? AND status = 'active'", req.AppID, req.ReleaseVersion).First(&release).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &types.PatchCheckResp{PatchAvailable: false}, nil
	}
	if err != nil {
		return nil, err
	}

	// 2. Fetch all rolled back patch numbers for this release
	var rolledBackPatches []db.Patch
	var rolledBackNumbers []int
	if err := l.svcCtx.DB.Where("release_id = ? AND status = 'rolled_back'", release.ID).Find(&rolledBackPatches).Error; err != nil {
		return nil, err
	} else {
		for _, p := range rolledBackPatches {
			rolledBackNumbers = append(rolledBackNumbers, p.Number)
		}
	}

	// Select a published patch for this channel, platform and architecture.
	channel := req.Channel
	if channel == "" {
		channel = "stable"
	}
	var latestPatch db.Patch
	err = l.svcCtx.DB.Model(&db.Patch{}).
		Joins("JOIN channels ON channels.id = patches.channel_id").
		Where("patches.release_id = ? AND patches.status = ? AND channels.app_id = ? AND channels.name = ?", release.ID, "active", req.AppID, channel).
		Where("EXISTS (SELECT 1 FROM patch_artifacts WHERE patch_artifacts.patch_id = patches.id AND arch = ? AND platform = ?)", req.Arch, req.Platform).
		Order("patches.number DESC").First(&latestPatch).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err != nil {
		return &types.PatchCheckResp{
			PatchAvailable:         false,
			RolledBackPatchNumbers: rolledBackNumbers,
		}, nil
	}

	// 4. Check if client is already running this patch or higher
	clientInstalledPatch := 0
	if req.PatchNumber != nil {
		clientInstalledPatch = *req.PatchNumber
	}
	if req.CurrentPatchNumber != nil && *req.CurrentPatchNumber > clientInstalledPatch {
		clientInstalledPatch = *req.CurrentPatchNumber
	}

	if clientInstalledPatch >= latestPatch.Number {
		return &types.PatchCheckResp{
			PatchAvailable:         false,
			RolledBackPatchNumbers: rolledBackNumbers,
		}, nil
	}

	// 5. Find the artifact for this architecture and platform
	var artifact db.PatchArtifact
	err = l.svcCtx.DB.Where("patch_id = ? AND arch = ? AND platform = ?", latestPatch.ID, req.Arch, req.Platform).First(&artifact).Error
	if err != nil {
		return &types.PatchCheckResp{
			PatchAvailable:         false,
			RolledBackPatchNumbers: rolledBackNumbers,
		}, nil
	}

	downloadURL := l.svcCtx.Storage.ResolveDownloadURL(artifact.StoragePath)

	l.Infof("[Shorebird] Serving patch #%d to app %s (arch: %s, platform: %s)",
		latestPatch.Number, req.AppID, req.Arch, req.Platform)

	return &types.PatchCheckResp{
		PatchAvailable: true,
		Patch: &types.PatchCheckMetadata{
			Number:        latestPatch.Number,
			DownloadURL:   downloadURL,
			Hash:          artifact.Hash,
			HashSignature: artifact.HashSignature,
		},
		RolledBackPatchNumbers: rolledBackNumbers,
	}, nil
}

func (l *CheckLogic) TrackEvent(req *types.PatchEventReq) error {
	appID := req.AppID
	clientID := req.ClientID
	patchNumber := req.PatchNumber
	eventType := req.Type

	if req.Event != nil {
		if req.Event.AppID != "" {
			appID = req.Event.AppID
		}
		if req.Event.ClientID != "" {
			clientID = req.Event.ClientID
		}
		if req.Event.PatchNumber != 0 {
			patchNumber = req.Event.PatchNumber
		}
		if req.Event.Type != "" {
			eventType = req.Event.Type
		}
	}

	event := db.PatchEvent{
		AppID:       appID,
		ClientID:    clientID,
		PatchNumber: patchNumber,
		Type:        eventType,
		Status:      "recorded",
	}
	return l.svcCtx.DB.Create(&event).Error
}
