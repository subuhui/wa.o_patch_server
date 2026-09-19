package logic

import (
	"context"
	"io"
	"strconv"
	"testing"

	"shorebird-server/internal/db"
	"shorebird-server/internal/storage"
	"shorebird-server/internal/types"
)

type beforeReadStorage struct {
	storage.Storage
	beforeRead func()
}

func (s beforeReadStorage) GetObject(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	s.beforeRead()
	return s.Storage.GetObject(ctx, key)
}

func TestPromotionCannotRestoreDeletedChannel(t *testing.T) {
	s, release, stable, _ := testService(t)
	app := NewAppLogic(context.Background(), s)
	channel, err := app.CreateChannel("app", &types.CreateChannelReq{Channel: "qa"})
	if err != nil {
		t.Fatal(err)
	}
	patch := addPatch(t, s, release, stable, 1, "android", "arm64")
	// Delete after promotion starts, while it is checking storage, before publication.
	s.Storage = beforeReadStorage{Storage: s.Storage, beforeRead: func() {
		if err := app.DeleteChannel("app", strconv.Itoa(int(channel.ID))); err != nil {
			t.Fatal(err)
		}
	}}
	err = NewPatchLogic(context.Background(), s).PromotePatch("app", strconv.Itoa(int(patch.ID)), int(channel.ID))
	if err == nil {
		t.Fatal("promotion succeeded after channel was deleted")
	}
	var got db.Patch
	if err := s.DB.First(&got, patch.ID).Error; err != nil {
		t.Fatal(err)
	}
	if got.ChannelID == nil || *got.ChannelID != stable.ID {
		t.Fatalf("original channel changed: %+v", got)
	}
}
