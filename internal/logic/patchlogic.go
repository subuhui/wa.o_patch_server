package logic

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"shorebird-server/internal/crypto"
	"shorebird-server/internal/db"
	"shorebird-server/internal/svc"
	"shorebird-server/internal/types"
)

type PatchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPatchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PatchLogic {
	return &PatchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PatchLogic) GetPatches(releaseIDStr string) (*types.GetReleasePatchesResp, error) {
	releaseID, _ := strconv.ParseUint(releaseIDStr, 10, 32)

	var patches []db.Patch
	if err := l.svcCtx.DB.Where("release_id = ?", releaseID).Order("number DESC").Find(&patches).Error; err != nil {
		return nil, err
	}

	result := make([]types.ReleasePatchResp, 0, len(patches))
	for _, p := range patches {
		var artifacts []db.PatchArtifact
		_ = l.svcCtx.DB.Where("patch_id = ?", p.ID).Find(&artifacts)

		artResps := make([]types.PatchArtifactResp, 0, len(artifacts))
		for _, a := range artifacts {
			artResps = append(artResps, types.PatchArtifactResp{
				ID:        a.ID,
				PatchID:   a.PatchID,
				Arch:      a.Arch,
				Platform:  a.Platform,
				Hash:      a.Hash,
				Size:      a.Size,
				CreatedAt: a.CreatedAt.Format(time.RFC3339),
			})
		}

		var channel *string
		if p.ChannelID != nil {
			var ch db.Channel
			if err := l.svcCtx.DB.First(&ch, *p.ChannelID).Error; err != nil {
				return nil, err
			}
			channel = &ch.Name
		}
		result = append(result, types.ReleasePatchResp{
			ID:           p.ID,
			Number:       p.Number,
			Channel:      channel,
			Artifacts:    artResps,
			IsRolledBack: p.Status == "rolled_back",
		})
	}

	return &types.GetReleasePatchesResp{Patches: result}, nil
}

func (l *PatchLogic) CreatePatch(req *types.CreatePatchReq) (*types.CreatePatchResp, error) {
	var maxNumber int
	row := l.svcCtx.DB.Model(&db.Patch{}).Where("release_id = ?", req.ReleaseID).Select("COALESCE(MAX(number), 0)").Row()
	_ = row.Scan(&maxNumber)

	patch := db.Patch{
		ReleaseID: uint(req.ReleaseID),
		Number:    maxNumber + 1,
		Status:    "draft",
	}

	if err := l.svcCtx.DB.Create(&patch).Error; err != nil {
		return nil, err
	}

	return &types.CreatePatchResp{
		ID:     patch.ID,
		Number: patch.Number,
	}, nil
}

func (l *PatchLogic) UpdatePatch(patchIDStr, status string) error {
	if status == "active" || status == "" {
		var patch db.Patch
		if err := l.svcCtx.DB.First(&patch, "id = ?", patchIDStr).Error; err != nil {
			return err
		}
		if patch.ChannelID == nil {
			return fmt.Errorf("promote the patch to a channel before activating it")
		}
		var release db.Release
		if err := l.svcCtx.DB.First(&release, patch.ReleaseID).Error; err != nil {
			return err
		}
		return l.PromotePatch(release.AppID, patchIDStr, int(*patch.ChannelID))
	}
	if status != "rolled_back" {
		return fmt.Errorf("invalid patch status: %s", status)
	}
	return l.RollbackPatch(patchIDStr)
}

func (l *PatchLogic) PromotePatch(appID, patchIDStr string, channelID int) error {
	var patch db.Patch
	if err := l.svcCtx.DB.Joins("JOIN releases ON releases.id = patches.release_id").
		Where("patches.id = ? AND releases.app_id = ?", patchIDStr, appID).First(&patch).Error; err != nil {
		return err
	}
	var channel db.Channel
	if err := l.svcCtx.DB.Where("id = ? AND app_id = ?", channelID, appID).First(&channel).Error; err != nil {
		return err
	}
	var artifacts []db.PatchArtifact
	if err := l.svcCtx.DB.Where("patch_id = ?", patch.ID).Find(&artifacts).Error; err != nil {
		return err
	}
	if len(artifacts) == 0 {
		return fmt.Errorf("patch has no artifacts")
	}
	for _, artifact := range artifacts {
		reader, size, err := l.svcCtx.Storage.GetObject(l.ctx, artifact.StoragePath)
		if err != nil {
			return fmt.Errorf("patch artifact is not uploaded: %w", err)
		}
		if err := reader.Close(); err != nil {
			return err
		}
		if size != artifact.Size || size <= 0 {
			return fmt.Errorf("patch artifact size mismatch")
		}
	}
	return l.svcCtx.DB.Model(&patch).Updates(map[string]interface{}{"status": "active", "channel_id": channel.ID}).Error
}

func (l *PatchLogic) RollbackPatch(patchIDStr string) error {
	return l.svcCtx.DB.Model(&db.Patch{}).Where("id = ?", patchIDStr).Update("status", "rolled_back").Error
}

func (l *PatchLogic) CreatePatchArtifact(appID, patchIDStr, arch, platform, hash, hashSig, podfileLockHash string, size int64) (*types.CreatePatchArtifactResp, error) {
	patchID, err := strconv.ParseUint(patchIDStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid patch_id: %w", err)
	}

	var patch db.Patch
	if err := l.svcCtx.DB.Joins("JOIN releases ON releases.id = patches.release_id").Where("patches.id = ? AND releases.app_id = ?", patchID, appID).First(&patch).Error; err != nil {
		return nil, err
	}
	if patch.Status != "draft" {
		return nil, fmt.Errorf("only draft patch artifacts can be changed")
	}
	if arch == "" || platform == "" || hash == "" || size <= 0 {
		return nil, fmt.Errorf("arch, platform, hash and positive size are required")
	}

	// Auto-sign if hash is present and signature is empty
	if hashSig == "" && l.svcCtx.PrivateKey != nil && hash != "" {
		sig, err := crypto.SignPatchHash(hash, l.svcCtx.PrivateKey)
		if err == nil {
			hashSig = sig
		}
	}

	storageKey := fmt.Sprintf("patches/%s/%d/%s_%s.diff", appID, patchID, platform, arch)
	downloadURL := l.svcCtx.Storage.ResolveDownloadURL(storageKey)

	artifact := db.PatchArtifact{
		PatchID:         uint(patchID),
		Arch:            arch,
		Platform:        platform,
		Hash:            hash,
		HashSignature:   hashSig,
		Size:            size,
		URL:             downloadURL,
		StoragePath:     storageKey,
		PodfileLockHash: podfileLockHash,
	}

	var existing db.PatchArtifact
	if err := l.svcCtx.DB.Where("patch_id = ? AND arch = ? AND platform = ?", patchID, arch, platform).First(&existing).Error; err == nil {
		existing.Hash = hash
		existing.HashSignature = hashSig
		existing.Size = size
		existing.URL = downloadURL
		existing.StoragePath = storageKey
		existing.PodfileLockHash = podfileLockHash
		if err := l.svcCtx.DB.Save(&existing).Error; err != nil {
			return nil, err
		}
		artifact = existing
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := l.svcCtx.DB.Create(&artifact).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	serverURL := strings.TrimSuffix(l.svcCtx.Config.PublicURL, "/")
	uploadURL := fmt.Sprintf("%s/api/v1/storage/upload?key=%s", serverURL, url.QueryEscape(storageKey))

	return &types.CreatePatchArtifactResp{
		ID:           artifact.ID,
		PatchID:      artifact.PatchID,
		Arch:         artifact.Arch,
		Platform:     artifact.Platform,
		Hash:         artifact.Hash,
		Size:         artifact.Size,
		URL:          uploadURL,
		UploadMethod: "multipart",
	}, nil
}
