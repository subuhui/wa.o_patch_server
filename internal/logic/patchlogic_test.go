package logic

import (
	"context"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"path/filepath"
	"shorebird-server/internal/db"
	"shorebird-server/internal/storage"
	"shorebird-server/internal/svc"
	"shorebird-server/internal/types"
	"strconv"
	"strings"
	"testing"
)

func testService(t *testing.T) (*svc.ServiceContext, db.Release, db.Channel, db.Channel) {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if err := database.AutoMigrate(&db.Release{}, &db.Patch{}, &db.Channel{}, &db.PatchArtifact{}); err != nil {
		t.Fatal(err)
	}
	st, err := storage.NewLocalStorage(t.TempDir(), "http://localhost")
	if err != nil {
		t.Fatal(err)
	}
	release := db.Release{AppID: "app", Version: "1.0.0", Status: "active"}
	stable := db.Channel{AppID: "app", Name: "stable"}
	beta := db.Channel{AppID: "app", Name: "beta"}
	for _, record := range []any{&release, &stable, &beta} {
		if err := database.Create(record).Error; err != nil {
			t.Fatal(err)
		}
	}
	return &svc.ServiceContext{DB: database, Storage: st}, release, stable, beta
}

func addPatch(t *testing.T, s *svc.ServiceContext, release db.Release, channel db.Channel, number int, platform, arch string) db.Patch {
	t.Helper()
	p := db.Patch{ReleaseID: release.ID, ChannelID: &channel.ID, Number: number, Status: "active"}
	if err := s.DB.Create(&p).Error; err != nil {
		t.Fatal(err)
	}
	key := fmt.Sprintf("patches/%d/%s_%s", p.ID, platform, arch)
	if _, err := s.Storage.Upload(context.Background(), key, strings.NewReader("data"), 4, ""); err != nil {
		t.Fatal(err)
	}
	a := db.PatchArtifact{PatchID: p.ID, Platform: platform, Arch: arch, Size: 4, StoragePath: key}
	if err := s.DB.Create(&a).Error; err != nil {
		t.Fatal(err)
	}
	return p
}

func TestCheckSelectsChannelPlatformAndArchitecture(t *testing.T) {
	s, release, stable, beta := testService(t)
	addPatch(t, s, release, stable, 1, "android", "arm64")
	addPatch(t, s, release, beta, 2, "android", "arm64")
	addPatch(t, s, release, stable, 3, "ios", "arm64")
	addPatch(t, s, release, stable, 4, "android", "x86_64")
	for _, tc := range []struct {
		channel, platform, arch string
		want                    int
	}{
		{"", "android", "arm64", 1}, {"stable", "android", "arm64", 1},
		{"beta", "android", "arm64", 2}, {"stable", "ios", "arm64", 3},
		{"stable", "android", "x86_64", 4}, {"unknown", "android", "arm64", 0},
	} {
		resp, err := NewCheckLogic(context.Background(), s).CheckPatch(&types.PatchCheckReq{AppID: "app", ReleaseVersion: release.Version, Channel: tc.channel, Platform: tc.platform, Arch: tc.arch})
		if err != nil {
			t.Fatal(err)
		}
		if tc.want == 0 {
			if resp.PatchAvailable {
				t.Fatal("unknown channel received patch")
			}
			continue
		}
		if !resp.PatchAvailable || resp.Patch.Number != tc.want {
			t.Fatalf("%+v: %+v", tc, resp)
		}
	}
}

func TestDraftRequiresCompleteUploadAndExplicitPromotion(t *testing.T) {
	s, release, stable, beta := testService(t)
	addPatch(t, s, release, stable, 1, "android", "arm64")
	l := NewPatchLogic(context.Background(), s)
	p, err := l.CreatePatch(&types.CreatePatchReq{AppID: "app", ReleaseID: int(release.ID)})
	if err != nil {
		t.Fatal(err)
	}
	id := strconv.Itoa(int(p.ID))
	if err := l.UpdatePatch(id, "active"); err == nil {
		t.Fatal("draft activation bypassed promotion")
	}
	if err := l.PromotePatch("app", id, int(beta.ID)); err == nil {
		t.Fatal("published without artifacts")
	}
	_, err = l.CreatePatchArtifact("app", id, "arm64", "android", "hash", "", "", 4)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.PromotePatch("app", id, int(beta.ID)); err == nil {
		t.Fatal("published before upload")
	}
	key := fmt.Sprintf("patches/app/%s/android_arm64.diff", id)
	if _, err := s.Storage.Upload(context.Background(), key, strings.NewReader("bad"), 3, ""); err != nil {
		t.Fatal(err)
	}
	if err := l.PromotePatch("app", id, int(beta.ID)); err == nil {
		t.Fatal("published wrong size")
	}
	if _, err := s.Storage.Upload(context.Background(), key, strings.NewReader("data"), 4, ""); err != nil {
		t.Fatal(err)
	}
	check := NewCheckLogic(context.Background(), s)
	req := &types.PatchCheckReq{AppID: "app", ReleaseVersion: release.Version, Platform: "android", Arch: "arm64"}
	resp, err := check.CheckPatch(req)
	if err != nil || !resp.PatchAvailable || resp.Patch.Number != 1 {
		t.Fatalf("draft hid old patch: %+v %v", resp, err)
	}
	if err := l.PromotePatch("other-app", id, int(beta.ID)); err == nil {
		t.Fatal("cross-app promotion allowed")
	}
	if err := l.PromotePatch("app", id, int(beta.ID)); err != nil {
		t.Fatal(err)
	}
	req.Channel = "beta"
	resp, err = check.CheckPatch(req)
	if err != nil || !resp.PatchAvailable || resp.Patch.Number != p.Number {
		t.Fatalf("promoted patch missing: %+v %v", resp, err)
	}
	if _, err := l.CreatePatchArtifact("app", id, "arm64", "android", "new-hash", "", "", 4); err == nil {
		t.Fatal("published metadata changed")
	}
	patches, err := l.GetPatches(strconv.Itoa(int(release.ID)))
	if err != nil {
		t.Fatal(err)
	}
	if patches.Patches[0].Channel == nil || *patches.Patches[0].Channel != "beta" {
		t.Fatal("wrong channel in patch list")
	}
	if err := l.RollbackPatch(id); err != nil {
		t.Fatal(err)
	}
	resp, err = check.CheckPatch(req)
	if err != nil || resp.PatchAvailable || len(resp.RolledBackPatchNumbers) != 1 {
		t.Fatalf("rollback failed: %+v %v", resp, err)
	}
}

func TestLegacyPatchNeedsExplicitChannel(t *testing.T) {
	s, release, stable, _ := testService(t)
	p := addPatch(t, s, release, stable, 1, "android", "arm64")
	if err := s.DB.Model(&p).Update("channel_id", nil).Error; err != nil {
		t.Fatal(err)
	}
	req := &types.PatchCheckReq{AppID: "app", ReleaseVersion: release.Version, Platform: "android", Arch: "arm64"}
	resp, err := NewCheckLogic(context.Background(), s).CheckPatch(req)
	if err != nil || resp.PatchAvailable {
		t.Fatalf("unassigned legacy patch served: %+v %v", resp, err)
	}
	if err := NewPatchLogic(context.Background(), s).PromotePatch("app", strconv.Itoa(int(p.ID)), int(stable.ID)); err != nil {
		t.Fatal(err)
	}
	resp, err = NewCheckLogic(context.Background(), s).CheckPatch(req)
	if err != nil || !resp.PatchAvailable {
		t.Fatalf("legacy promotion failed: %+v %v", resp, err)
	}
}
