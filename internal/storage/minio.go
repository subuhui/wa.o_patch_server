package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"path"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"shorebird-server/internal/config"
)

type Storage interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error)
	GetObject(ctx context.Context, objectName string) (io.ReadCloser, int64, error)
	ResolveDownloadURL(objectName string) string
}

type MinIOStorage struct {
	client            *minio.Client
	bucket            string
	proxyDownload     bool
	serverPublicURL   string
	publicDownloadURL string
}

func NewMinIOStorage(cfg *config.Config) (*MinIOStorage, error) {
	mCfg := cfg.Storage.MinIO
	endpoint := strings.TrimPrefix(strings.TrimPrefix(mCfg.Endpoint, "http://"), "https://")

	creds := credentials.NewStaticV4(mCfg.AccessKeyID, mCfg.SecretAccessKey, "")


	client, err := minio.New(endpoint, &minio.Options{
		Creds:  creds,
		Secure: mCfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, mCfg.Bucket)
	if err != nil {
		log.Printf("Warning: failed to check MinIO bucket existence (network or auth issue): %v", err)
	} else if !exists {
		err = client.MakeBucket(ctx, mCfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			log.Printf("Warning: failed to auto-create bucket '%s': %v", mCfg.Bucket, err)
		} else {
			log.Printf("Created MinIO bucket '%s'", mCfg.Bucket)
		}
	}

	return &MinIOStorage{
		client:            client,
		bucket:            mCfg.Bucket,
		proxyDownload:     mCfg.ProxyDownload,
		serverPublicURL:   strings.TrimSuffix(cfg.PublicURL, "/"),
		publicDownloadURL: strings.TrimSuffix(mCfg.PublicDownloadURL, "/"),
	}, nil
}

// Upload uploads an object to MinIO and returns its download URL.
func (s *MinIOStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object to MinIO: %w", err)
	}

	return s.ResolveDownloadURL(objectName), nil
}

// GetObject retrieves an object stream from MinIO.
func (s *MinIOStorage) GetObject(ctx context.Context, objectName string) (io.ReadCloser, int64, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get object: %w", err)
	}
	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, 0, fmt.Errorf("failed to stat object: %w", err)
	}
	return obj, stat.Size, nil
}

// ResolveDownloadURL computes the accessible URL for clients to download this artifact.
func (s *MinIOStorage) ResolveDownloadURL(objectName string) string {
	if s.proxyDownload {
		// Server acts as download relay, so phone only needs to reach serverPublicURL
		return fmt.Sprintf("%s/api/v1/storage/download?key=%s", s.serverPublicURL, url.QueryEscape(objectName))
	}

	if s.publicDownloadURL != "" {
		return fmt.Sprintf("%s/%s", s.publicDownloadURL, path.Clean(objectName))
	}

	return fmt.Sprintf("%s/api/v1/storage/download?key=%s", s.serverPublicURL, url.QueryEscape(objectName))
}
