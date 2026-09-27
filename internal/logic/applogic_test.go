package logic

import (
	"context"
	"testing"

	"shorebird-server/internal/db"
)

func TestGetAppsCountsPublishedPatchesAcrossReleases(t *testing.T) {
	s, firstRelease, _, _ := testService(t)
	if err := s.DB.AutoMigrate(&db.App{}); err != nil {
		t.Fatal(err)
	}
	if err := s.DB.Create(&db.App{ID: "app", DisplayName: "App"}).Error; err != nil {
		t.Fatal(err)
	}
	secondRelease := db.Release{AppID: "app", Version: "2.0.0", Status: "active"}
	if err := s.DB.Create(&secondRelease).Error; err != nil {
		t.Fatal(err)
	}
	patches := []db.Patch{
		{ReleaseID: firstRelease.ID, Number: 1, Status: "active"},
		{ReleaseID: firstRelease.ID, Number: 2, Status: "rolled_back"},
		{ReleaseID: secondRelease.ID, Number: 1, Status: "active"},
		{ReleaseID: secondRelease.ID, Number: 2, Status: "draft"},
	}
	for i := range patches {
		if err := s.DB.Create(&patches[i]).Error; err != nil {
			t.Fatal(err)
		}
	}

	resp, err := NewAppLogic(context.Background(), s).GetApps()
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Apps) != 1 {
		t.Fatalf("apps=%d", len(resp.Apps))
	}
	app := resp.Apps[0]
	if app.PatchCount != 3 {
		t.Fatalf("patch_count=%d", app.PatchCount)
	}
	if app.LatestPatchNumber == nil || *app.LatestPatchNumber != 1 {
		t.Fatalf("latest_patch_number=%v", app.LatestPatchNumber)
	}
}
