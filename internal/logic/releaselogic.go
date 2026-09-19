package logic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"shorebird-server/internal/db"
	"shorebird-server/internal/svc"
	"shorebird-server/internal/types"
)

type ReleaseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReleaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseLogic {
	return &ReleaseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReleaseLogic) formatRelease(r *db.Release, platforms []string) types.ReleaseResp {
	var flutterVersion *string
	if r.FlutterVersion != "" {
		flutterVersion = &r.FlutterVersion
	}
	statuses := make(map[string]string, len(platforms))
	for _, platform := range platforms {
		statuses[platform] = r.Status
	}
	return types.ReleaseResp{
		ID:               r.ID,
		AppID:            r.AppID,
		Version:          r.Version,
		FlutterRevision:  r.FlutterRevision,
		FlutterVersion:   flutterVersion,
		PlatformStatuses: statuses,
		CreatedAt:        r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        r.UpdatedAt.Format(time.RFC3339),
	}
}

func (l *ReleaseLogic) GetReleases(appID string) (*types.GetReleasesResp, error) {
	var releases []db.Release
	if err := l.svcCtx.DB.Where("app_id = ?", appID).Order("id DESC").Find(&releases).Error; err != nil {
		return nil, err
	}

	platforms := make(map[uint][]string)
	if len(releases) > 0 {
		ids := make([]uint, 0, len(releases))
		for _, release := range releases {
			ids = append(ids, release.ID)
		}
		var artifacts []db.ReleaseArtifact
		if err := l.svcCtx.DB.Select("DISTINCT release_id, platform").
			Where("release_id IN ?", ids).Find(&artifacts).Error; err != nil {
			return nil, err
		}
		for _, artifact := range artifacts {
			platforms[artifact.ReleaseID] = append(platforms[artifact.ReleaseID], artifact.Platform)
		}
	}

	result := make([]types.ReleaseResp, 0, len(releases))
	for _, r := range releases {
		result = append(result, l.formatRelease(&r, platforms[r.ID]))
	}
	return &types.GetReleasesResp{Releases: result}, nil
}

func (l *ReleaseLogic) CreateRelease(req *types.CreateReleaseReq) (*types.CreateReleaseResp, error) {
	var flutterVersion string
	if req.FlutterVersion != nil {
		flutterVersion = *req.FlutterVersion
	}
	release := db.Release{
		AppID:           req.AppID,
		Version:         req.Version,
		FlutterRevision: req.FlutterRevision,
		FlutterVersion:  flutterVersion,
		Status:          "active",
	}

	if err := l.svcCtx.DB.Create(&release).Error; err != nil {
		return nil, err
	}

	resp := l.formatRelease(&release, nil)
	return &types.CreateReleaseResp{Release: resp}, nil
}

func (l *ReleaseLogic) UpdateReleaseStatus(releaseIDStr, status string) error {
	if status == "" {
		status = "active"
	}
	return l.svcCtx.DB.Model(&db.Release{}).Where("id = ?", releaseIDStr).Update("status", status).Error
}

func (l *ReleaseLogic) GetReleaseArtifacts(releaseIDStr, arch, platform string) (*types.GetReleaseArtifactsResp, error) {
	query := l.svcCtx.DB.Where("release_id = ?", releaseIDStr)
	if arch != "" {
		query = query.Where("arch = ?", arch)
	}
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	var artifacts []db.ReleaseArtifact
	err := query.Find(&artifacts).Error
	if err != nil {
		return nil, err
	}

	result := make([]types.ReleaseArtifactResp, 0, len(artifacts))
	for _, a := range artifacts {
		result = append(result, types.ReleaseArtifactResp{
			ID:              a.ID,
			ReleaseID:       a.ReleaseID,
			Arch:            a.Arch,
			Platform:        a.Platform,
			Hash:            a.Hash,
			Size:            a.Size,
			URL:             a.URL,
			CanSideload:     a.CanSideload,
			PodfileLockHash: a.PodfileLockHash,
		})
	}
	return &types.GetReleaseArtifactsResp{Artifacts: result}, nil
}

func (l *ReleaseLogic) CreateReleaseArtifact(appID, releaseIDStr, arch, platform, hash, filename, podfileLockHash string, size int64, canSideload bool) (*types.CreateReleaseArtifactResp, error) {
	releaseID, err := strconv.ParseUint(releaseIDStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid release_id: %w", err)
	}

	if filename == "" {
		filename = "artifact"
	}

	storageKey := fmt.Sprintf("releases/%s/%d/%s_%s", appID, releaseID, arch, filename)
	downloadURL := l.svcCtx.Storage.ResolveDownloadURL(storageKey)

	artifact := db.ReleaseArtifact{
		ReleaseID:       uint(releaseID),
		Arch:            arch,
		Platform:        platform,
		Hash:            hash,
		Size:            size,
		URL:             downloadURL,
		StoragePath:     storageKey,
		CanSideload:     canSideload,
		Filename:        filename,
		PodfileLockHash: podfileLockHash,
	}

	var existing db.ReleaseArtifact
	if err := l.svcCtx.DB.Where("release_id = ? AND arch = ? AND platform = ?", releaseID, arch, platform).First(&existing).Error; err == nil {
		existing.Hash = hash
		existing.Size = size
		existing.URL = downloadURL
		existing.StoragePath = storageKey
		existing.CanSideload = canSideload
		existing.Filename = filename
		existing.PodfileLockHash = podfileLockHash
		if err := l.svcCtx.DB.Save(&existing).Error; err != nil {
			return nil, err
		}
		artifact = existing
	} else {
		if err := l.svcCtx.DB.Create(&artifact).Error; err != nil {
			return nil, err
		}
	}

	serverURL := strings.TrimSuffix(l.svcCtx.Config.PublicURL, "/")
	uploadURL := fmt.Sprintf("%s/api/v1/storage/upload?key=%s", serverURL, url.QueryEscape(storageKey))

	return &types.CreateReleaseArtifactResp{
		ID:           artifact.ID,
		ReleaseID:    artifact.ReleaseID,
		Arch:         artifact.Arch,
		Platform:     artifact.Platform,
		Hash:         artifact.Hash,
		Size:         artifact.Size,
		URL:          uploadURL,
		UploadMethod: "multipart",
	}, nil
}
