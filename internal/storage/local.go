package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type LocalStorage struct {
	baseDir         string
	serverPublicURL string
}

func NewLocalStorage(baseDir, serverPublicURL string) (*LocalStorage, error) {
	if baseDir == "" {
		baseDir = "data/storage"
	}
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local storage directory: %w", err)
	}

	return &LocalStorage{
		baseDir:         baseDir,
		serverPublicURL: strings.TrimSuffix(serverPublicURL, "/"),
	}, nil
}

func (s *LocalStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	root, err := os.OpenRoot(s.baseDir)
	if err != nil {
		return "", err
	}
	defer root.Close()
	if !filepath.IsLocal(objectName) {
		return "", fmt.Errorf("invalid object name")
	}
	targetPath := objectName
	if err := root.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory for artifact: %w", err)
	}

	temporaryPath := filepath.Join(filepath.Dir(targetPath), ".upload-"+uuid.NewString())
	file, err := root.OpenFile(temporaryPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return "", fmt.Errorf("failed to create file on local storage: %w", err)
	}
	defer file.Close()
	defer root.Remove(temporaryPath)

	written, err := io.Copy(file, reader)
	if err != nil {
		return "", fmt.Errorf("failed to write file bytes: %w", err)
	}

	if written != size {
		return "", fmt.Errorf("artifact size mismatch: got %d, want %d", written, size)
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	if err := root.Rename(temporaryPath, targetPath); err != nil {
		return "", err
	}
	return s.ResolveDownloadURL(objectName), nil
}

func (s *LocalStorage) GetObject(ctx context.Context, objectName string) (io.ReadCloser, int64, error) {
	root, err := os.OpenRoot(s.baseDir)
	if err != nil {
		return nil, 0, err
	}
	defer root.Close()
	if !filepath.IsLocal(objectName) {
		return nil, 0, fmt.Errorf("invalid object name")
	}
	file, err := root.Open(objectName)
	if err != nil {
		return nil, 0, fmt.Errorf("artifact not found: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, fmt.Errorf("failed to stat artifact: %w", err)
	}

	return file, stat.Size(), nil
}

func (s *LocalStorage) ResolveDownloadURL(objectName string) string {
	return fmt.Sprintf("%s/api/v1/storage/download?key=%s", s.serverPublicURL, url.QueryEscape(objectName))
}
