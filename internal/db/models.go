package db

import (
	"time"
)

// User represents the Shorebird user model.
type User struct {
	ID                    uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email                 string    `gorm:"size:128;uniqueIndex;not null" json:"email"`
	Name                  string    `gorm:"size:128" json:"name"`
	HasActiveSubscription bool      `gorm:"default:true" json:"has_active_subscription"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// App represents an application.
type App struct {
	ID          string    `gorm:"primaryKey;size:64" json:"id"` // App UUID
	DisplayName string    `gorm:"size:128;not null" json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Channel represents an update channel (e.g. stable, beta).
type Channel struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AppID     string    `gorm:"size:64;index;not null" json:"app_id"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// Release represents a base app release.
type Release struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AppID           string    `gorm:"size:64;index;not null" json:"app_id"`
	Version         string    `gorm:"size:64;not null" json:"version"` // e.g. "1.0.0+1"
	FlutterRevision string    `gorm:"size:64;not null" json:"flutter_revision"`
	FlutterVersion  string    `gorm:"size:64" json:"flutter_version"`
	Status          string    `gorm:"size:32;default:'active'" json:"status"` // "draft", "active"
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// ReleaseArtifact represents an artifact file associated with a Release.
type ReleaseArtifact struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ReleaseID       uint      `gorm:"index;not null" json:"release_id"`
	Arch            string    `gorm:"size:32;not null" json:"arch"`
	Platform        string    `gorm:"size:32;not null" json:"platform"`
	Hash            string    `gorm:"size:128;not null" json:"hash"`
	Size            int64     `gorm:"not null" json:"size"`
	URL             string    `gorm:"size:512;not null" json:"url"`
	StoragePath     string    `gorm:"size:255;not null" json:"-"`
	CanSideload     bool      `gorm:"default:false" json:"can_sideload"`
	Filename        string    `gorm:"size:128" json:"filename"`
	PodfileLockHash string    `gorm:"size:128" json:"podfile_lock_hash"`
	CreatedAt       time.Time `json:"created_at"`
}

// Patch represents a patch created on top of a Release.
type Patch struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ReleaseID uint      `gorm:"index;not null" json:"release_id"`
	ChannelID *uint     `gorm:"index" json:"channel_id"`
	Number    int       `gorm:"not null" json:"number"`                // 1, 2, 3...
	Status    string    `gorm:"size:32;default:'draft'" json:"status"` // "draft", "active", "rolled_back"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PatchArtifact represents the binary diff file for a patch on a specific arch/platform.
type PatchArtifact struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	PatchID         uint      `gorm:"index;not null" json:"patch_id"`
	Arch            string    `gorm:"size:32;not null" json:"arch"`
	Platform        string    `gorm:"size:32;not null" json:"platform"`
	Hash            string    `gorm:"size:128;not null" json:"hash"`
	HashSignature   string    `gorm:"type:text" json:"hash_signature"` // RSA-PKCS1v15-SHA256 signature
	Size            int64     `gorm:"not null" json:"size"`
	URL             string    `gorm:"size:512;not null" json:"url"`
	StoragePath     string    `gorm:"size:255;not null" json:"-"`
	PodfileLockHash string    `gorm:"size:128" json:"podfile_lock_hash"`
	CreatedAt       time.Time `json:"created_at"`
}

// PatchEvent tracks client-reported patch download/install events.
type PatchEvent struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AppID       string    `gorm:"size:64;index;not null" json:"app_id"`
	ClientID    string    `gorm:"size:128" json:"client_id"`
	PatchNumber int       `json:"patch_number"`
	Type        string    `gorm:"size:64" json:"type"`
	Status      string    `gorm:"size:64" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
