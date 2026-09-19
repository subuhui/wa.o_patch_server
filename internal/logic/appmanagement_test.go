package logic

import (
	"context"
	"strconv"
	"testing"

	"shorebird-server/internal/db"
	"shorebird-server/internal/types"
)

func TestRenameAndDeleteChannel(t *testing.T) {
	s, release, stable, _ := testService(t)
	if err := s.DB.AutoMigrate(&db.App{}); err != nil {
		t.Fatal(err)
	}
	app := db.App{ID: "app", DisplayName: "old"}
	if err := s.DB.Create(&app).Error; err != nil {
		t.Fatal(err)
	}
	l := NewAppLogic(context.Background(), s)
	if err := l.RenameApp("app", "新名称"); err != nil {
		t.Fatal(err)
	}
	if err := s.DB.First(&app, "id = ?", "app").Error; err != nil {
		t.Fatal(err)
	}
	if app.DisplayName != "新名称" {
		t.Fatal(app.DisplayName)
	}
	for _, name := range []string{"", "  "} {
		if err := l.RenameApp("app", name); err == nil {
			t.Fatal("blank name accepted")
		}
	}
	if err := l.RenameApp("missing", "name"); err == nil {
		t.Fatal("missing app accepted")
	}
	if err := l.DeleteChannel("app", strconv.Itoa(int(stable.ID))); err == nil {
		t.Fatal("deleted stable")
	}
	channel, err := l.CreateChannel("app", &types.CreateChannelReq{Channel: "qa"})
	if err != nil {
		t.Fatal(err)
	}
	p := addPatch(t, s, release, *channel, 1, "android", "arm64")
	id := strconv.Itoa(int(channel.ID))
	if err := l.DeleteChannel("other", id); err == nil {
		t.Fatal("cross-app deletion")
	}
	if err := l.DeleteChannel("app", id); err != nil {
		t.Fatal(err)
	}
	if err := s.DB.First(&p, p.ID).Error; err != nil {
		t.Fatal(err)
	}
	if p.ChannelID != nil {
		t.Fatal("dangling channel")
	}
	if err := l.DeleteChannel("app", id); err == nil {
		t.Fatal("missing channel accepted")
	}
}

func TestRollforwardPatch(t *testing.T) {
	s, release, stable, _ := testService(t)
	p := addPatch(t, s, release, stable, 1, "android", "arm64")
	l := NewPatchLogic(context.Background(), s)
	pid, rid := strconv.Itoa(int(p.ID)), strconv.Itoa(int(release.ID))
	if _, err := l.RollbackPatch(pid); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][3]string{{"other", rid, pid}, {"app", "999", pid}, {"app", rid, "999"}} {
		if _, err := l.RollforwardPatch(args[0], args[1], args[2]); err == nil {
			t.Fatal("invalid scope accepted")
		}
	}
	for _, want := range []bool{true, false} {
		changed, err := l.RollforwardPatch("app", rid, pid)
		if err != nil || changed != want {
			t.Fatalf("changed=%v err=%v", changed, err)
		}
	}
	resp, err := NewCheckLogic(context.Background(), s).CheckPatch(&types.PatchCheckReq{AppID: "app", ReleaseVersion: release.Version, Platform: "android", Arch: "arm64"})
	if err != nil || !resp.PatchAvailable || len(resp.RolledBackPatchNumbers) != 0 {
		t.Fatalf("%+v %v", resp, err)
	}
	for _, status := range []string{"draft", "rolled_back"} {
		if err := s.DB.Model(&p).Updates(map[string]any{"status": status, "channel_id": nil}).Error; err != nil {
			t.Fatal(err)
		}
		if _, err := l.RollforwardPatch("app", rid, pid); err == nil {
			t.Fatal("unpublished or unassigned patch restored")
		}
	}
}
