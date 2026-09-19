package types



// User types
type UserResp struct {
	ID                    uint    `json:"id"`
	Email                 string  `json:"email"`
	DisplayName           *string `json:"display_name,omitempty"`
	HasActiveSubscription bool    `json:"has_active_subscription"`
	JWTIssuer             string  `json:"jwt_issuer"`
	StripeCustomerID      *string `json:"stripe_customer_id,omitempty"`
	PatchOverageLimit     *int    `json:"patch_overage_limit,omitempty"`
}

// App types
type CreateAppReq struct {
	DisplayName    string `json:"display_name"`
	OrganizationID int    `json:"organization_id,optional"`
}

type AppResp struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type AppMetadataResp struct {
	AppID                string   `json:"app_id"`
	DisplayName          string   `json:"display_name"`
	LatestReleaseVersion *string  `json:"latest_release_version"`
	LatestPatchNumber    *int     `json:"latest_patch_number"`
	CreatedAt            string   `json:"created_at"`
	UpdatedAt            string   `json:"updated_at"`
	Platforms            []string `json:"platforms"`
}

type GetAppsResp struct {
	Apps []AppMetadataResp `json:"apps"`
}

type AppIDPathReq struct {
	AppID string `path:"app_id"`
}

type CreateChannelReq struct {
	AppID   string `path:"app_id"`
	Channel string `json:"channel"`
}

// Release types
type ReleasePathReq struct {
	AppID     string `path:"app_id"`
	ReleaseID string `path:"release_id"`
}

type GetReleaseArtifactsReq struct {
	AppID     string `path:"app_id"`
	ReleaseID string `path:"release_id"`
	Arch      string `form:"arch,optional"`
	Platform  string `form:"platform,optional"`
}

type CreateReleaseReq struct {
	AppID           string  `path:"app_id"`
	Version         string  `json:"version"`
	FlutterRevision string  `json:"flutter_revision"`
	FlutterVersion  *string `json:"flutter_version,optional"`
	DisplayName     *string `json:"display_name,optional"`
}

type ReleaseResp struct {
	ID               uint              `json:"id"`
	AppID            string            `json:"app_id"`
	Version          string            `json:"version"`
	FlutterRevision  string            `json:"flutter_revision"`
	FlutterVersion   *string           `json:"flutter_version,omitempty"`
	DisplayName      *string           `json:"display_name,omitempty"`
	PlatformStatuses map[string]string `json:"platform_statuses"`
	CreatedAt        string            `json:"created_at"`
	UpdatedAt        string            `json:"updated_at"`
}

type GetReleasesResp struct {
	Releases []ReleaseResp `json:"releases"`
}

type CreateReleaseResp struct {
	Release ReleaseResp `json:"release"`
}

type UpdateReleaseStatusReq struct {
	AppID     string `path:"app_id"`
	ReleaseID string `path:"release_id"`
	Status    string `json:"status,optional"`
}

type ReleaseArtifactResp struct {
	ID              uint   `json:"id"`
	ReleaseID       uint   `json:"release_id"`
	Arch            string `json:"arch"`
	Platform        string `json:"platform"`
	Hash            string `json:"hash"`
	Size            int64  `json:"size"`
	URL             string `json:"url"`
	CanSideload     bool   `json:"can_sideload"`
	PodfileLockHash string `json:"podfile_lock_hash,omitempty"`
}

type CreateReleaseArtifactResp struct {
	ID           uint   `json:"id"`
	ReleaseID    uint   `json:"release_id"`
	Arch         string `json:"arch"`
	Platform     string `json:"platform"`
	Hash         string `json:"hash"`
	Size         int64  `json:"size"`
	URL          string `json:"url"`
	UploadMethod string `json:"upload_method"`
}

type GetReleaseArtifactsResp struct {
	Artifacts []ReleaseArtifactResp `json:"artifacts"`
}

// Patch types
type PatchPathReq struct {
	AppID   string `path:"app_id"`
	PatchID string `path:"patch_id"`
}

type PatchArtifactResp struct {
	ID        uint   `json:"id"`
	PatchID   uint   `json:"patch_id"`
	Arch      string `json:"arch"`
	Platform  string `json:"platform"`
	Hash      string `json:"hash"`
	Size      int64  `json:"size"`
	CreatedAt string `json:"created_at"`
}

type ReleasePatchResp struct {
	ID           uint                `json:"id"`
	Number       int                 `json:"number"`
	Channel      *string             `json:"channel,omitempty"`
	Artifacts    []PatchArtifactResp `json:"artifacts"`
	IsRolledBack bool                `json:"is_rolled_back"`
}

type GetReleasePatchesResp struct {
	Patches []ReleasePatchResp `json:"patches"`
}

type CreatePatchReq struct {
	AppID     string                 `path:"app_id"`
	ReleaseID int                    `json:"release_id"`
	Metadata  map[string]interface{} `json:"metadata,optional"`
}

type CreatePatchResp struct {
	ID     uint `json:"id"`
	Number int  `json:"number"`
}

type UpdatePatchReq struct {
	AppID   string `path:"app_id"`
	PatchID string `path:"patch_id"`
	Status  string `json:"status,optional"`
}

type PromotePatchReq struct {
	AppID     string `path:"app_id"`
	PatchID   int    `json:"patch_id,optional"`
	ChannelID int    `json:"channel_id,optional"`
}

type RollbackPatchReq struct {
	AppID     string `path:"app_id"`
	ReleaseID string `path:"release_id,optional"`
	PatchID   string `path:"patch_id,optional"`
}

type CreatePatchArtifactResp struct {
	ID           uint   `json:"id"`
	PatchID      uint   `json:"patch_id"`
	Arch         string `json:"arch"`
	Platform     string `json:"platform"`
	Hash         string `json:"hash"`
	Size         int64  `json:"size"`
	URL          string `json:"url"`
	UploadMethod string `json:"upload_method"`
}

// Patch Check types
type PatchCheckReq struct {
	ReleaseVersion     string `json:"release_version"`
	Platform           string `json:"platform"`
	Arch               string `json:"arch"`
	AppID              string `json:"app_id"`
	Channel            string `json:"channel,optional"`
	PatchNumber        *int   `json:"patch_number,optional"`
	CurrentPatchNumber *int   `json:"current_patch_number,optional"`
	ClientID           string `json:"client_id,optional"`
}

type PatchCheckMetadata struct {
	Number        int    `json:"number"`
	DownloadURL   string `json:"download_url"`
	Hash          string `json:"hash"`
	HashSignature string `json:"hash_signature,omitempty"`
}

type PatchCheckResp struct {
	PatchAvailable         bool                `json:"patch_available"`
	Patch                  *PatchCheckMetadata `json:"patch,omitempty"`
	RolledBackPatchNumbers []int               `json:"rolled_back_patch_numbers,omitempty"`
}

type PatchEventPayload struct {
	AppID          string  `json:"app_id,optional"`
	Arch           string  `json:"arch,optional"`
	ClientID       string  `json:"client_id,optional"`
	Type           string  `json:"type,optional"`
	PatchNumber    int     `json:"patch_number,optional"`
	Platform       string  `json:"platform,optional"`
	ReleaseVersion string  `json:"release_version,optional"`
	Timestamp      int64   `json:"timestamp,optional"`
	Message        *string `json:"message,optional"`
}

type PatchEventReq struct {
	Event       *PatchEventPayload `json:"event,optional"`
	AppID       string             `json:"app_id,optional"`
	ClientID    string             `json:"client_id,optional"`
	PatchNumber int                `json:"patch_number,optional"`
	Type        string             `json:"type,optional"`
	Status      string             `json:"status,optional"`
}

// Storage types
type UploadResp struct {
	Message     string `json:"message"`
	Key         string `json:"key"`
	DownloadURL string `json:"download_url"`
}
