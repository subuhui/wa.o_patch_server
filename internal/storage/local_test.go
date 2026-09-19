package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStorageConfinesReadsAndWrites(t *testing.T) {
	parent := t.TempDir()
	base := filepath.Join(parent, "objects")
	s, err := NewLocalStorage(base, "http://localhost")
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(parent, "secret")
	if err := os.WriteFile(outside, []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(parent, filepath.Join(base, "escape")); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"../secret", outside, "escape/secret"} {
		t.Run(key, func(t *testing.T) {
			if f, _, err := s.GetObject(context.Background(), key); err == nil {
				f.Close()
				t.Fatal("read escaped root")
			}
			if _, err := s.Upload(context.Background(), key, strings.NewReader("overwrite"), 9, ""); err == nil {
				t.Fatal("write escaped root")
			}
		})
	}
	data, err := os.ReadFile(outside)
	if err != nil || string(data) != "secret" {
		t.Fatalf("outside file modified: %q %v", data, err)
	}
}

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("interrupted upload") }

func TestLocalStorageOnlyPublishesCompleteUploads(t *testing.T) {
	s, err := NewLocalStorage(t.TempDir(), "http://localhost")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	key := "patches/app/1/android_arm64.diff"
	if _, err := s.Upload(ctx, key, strings.NewReader("original"), 8, ""); err != nil {
		t.Fatal(err)
	}
	for _, reader := range []io.Reader{brokenReader{}, strings.NewReader("short")} {
		if _, err := s.Upload(ctx, key, reader, 8, ""); err == nil {
			t.Fatal("expected failed upload")
		}
		f, _, err := s.GetObject(ctx, key)
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(f)
		f.Close()
		if err != nil || string(data) != "original" {
			t.Fatalf("partial upload published: %q %v", data, err)
		}
	}
}
