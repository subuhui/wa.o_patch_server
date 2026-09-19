package logic

import (
	"context"
	"io"

	"github.com/zeromicro/go-zero/core/logx"

	"shorebird-server/internal/svc"
	"shorebird-server/internal/types"
)

type StorageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewStorageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StorageLogic {
	return &StorageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *StorageLogic) Upload(key string, reader io.Reader, size int64, contentType string) (*types.UploadResp, error) {
	downloadURL, err := l.svcCtx.Storage.Upload(l.ctx, key, reader, size, contentType)
	if err != nil {
		return nil, err
	}

	return &types.UploadResp{
		Message:     "success",
		Key:         key,
		DownloadURL: downloadURL,
	}, nil
}

func (l *StorageLogic) GetObject(key string) (io.ReadCloser, int64, error) {
	return l.svcCtx.Storage.GetObject(l.ctx, key)
}
