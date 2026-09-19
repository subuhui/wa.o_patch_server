package logic

import (
	"context"
	"encoding/json"
	"testing"

	"shorebird-server/internal/db"
	"shorebird-server/internal/types"
)

func TestCheckPatchSignatureJSON(t *testing.T) {
	for _, signature := range []string{"", "signed-hash"} {
		t.Run(signature, func(t *testing.T) {
			s, release, stable, _ := testService(t)
			p := addPatch(t, s, release, stable, 1, "android", "arm64")
			if err := s.DB.Model(&db.PatchArtifact{}).Where("patch_id = ?", p.ID).Update("hash_signature", signature).Error; err != nil {
				t.Fatal(err)
			}
			resp, err := NewCheckLogic(context.Background(), s).CheckPatch(&types.PatchCheckReq{
				AppID: "app", ReleaseVersion: release.Version, Platform: "android", Arch: "arm64",
			})
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(resp)
			if err != nil {
				t.Fatal(err)
			}
			var wire struct {
				Patch map[string]json.RawMessage
			}
			if err := json.Unmarshal(data, &wire); err != nil {
				t.Fatal(err)
			}
			value, present := wire.Patch["hash_signature"]
			if signature == "" {
				if present {
					t.Fatalf("unsigned patch includes signature: %s", data)
				}
			} else if string(value) != `"signed-hash"` {
				t.Fatalf("signature lost: %s", data)
			}
		})
	}
}

func TestReleaseStatusesFollowArtifacts(t *testing.T) {
	s, _, _, _ := testService(t)
	if err := s.DB.AutoMigrate(&db.ReleaseArtifact{}); err != nil {
		t.Fatal(err)
	}
	l := NewReleaseLogic(context.Background(), s)
	created, err := l.CreateRelease(&types.CreateReleaseReq{AppID: "desktop", Version: "1.0"})
	if err != nil {
		t.Fatal(err)
	}
	if created.Release.PlatformStatuses == nil || len(created.Release.PlatformStatuses) != 0 {
		t.Fatalf("new release advertises platforms: %+v", created.Release.PlatformStatuses)
	}
	for _, platform := range []string{"android", "windows", "macos", "linux", "ios"} {
		for _, arch := range []string{"arm64", "x86_64"} {
			artifact := db.ReleaseArtifact{ReleaseID: created.Release.ID, Platform: platform, Arch: arch}
			if err := s.DB.Create(&artifact).Error; err != nil {
				t.Fatal(err)
			}
		}
		listed, err := l.GetReleases("desktop")
		if err != nil {
			t.Fatal(err)
		}
		statuses := listed.Releases[0].PlatformStatuses
		if statuses[platform] != "active" {
			t.Fatalf("missing platform %s: %+v", platform, statuses)
		}
		if platform == "android" && len(statuses) != 1 {
			t.Fatalf("unreleased platforms advertised: %+v", statuses)
		}
	}
	if err := s.DB.Model(&db.Release{}).Where("id = ?", created.Release.ID).Update("status", "draft").Error; err != nil {
		t.Fatal(err)
	}
	listed, err := l.GetReleases("desktop")
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.Releases[0].PlatformStatuses) != 5 {
		t.Fatal(listed)
	}
	for platform, status := range listed.Releases[0].PlatformStatuses {
		if status != "draft" {
			t.Fatalf("%s: %s", platform, status)
		}
	}
	empty, err := l.GetReleases("missing")
	if err != nil || len(empty.Releases) != 0 {
		t.Fatalf("%+v %v", empty, err)
	}
}
