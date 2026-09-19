package crypto

import (
	"testing"
)

func TestRSAKeyGenAndSign(t *testing.T) {
	key, err := GenerateRSAKeyPair()
	if err != nil {
		t.Fatalf("GenerateRSAKeyPair failed: %v", err)
	}

	pemBytes := ExportPrivateKeyPEM(key)
	if len(pemBytes) == 0 {
		t.Fatal("ExportPrivateKeyPEM returned empty")
	}

	loadedKey, err := LoadPrivateKeyPEM(pemBytes)
	if err != nil {
		t.Fatalf("LoadPrivateKeyPEM failed: %v", err)
	}

	derBase64, err := ExportPublicKeyDERBase64(&loadedKey.PublicKey)
	if err != nil {
		t.Fatalf("ExportPublicKeyDERBase64 failed: %v", err)
	}
	if len(derBase64) == 0 {
		t.Fatal("ExportPublicKeyDERBase64 returned empty")
	}

	sampleHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	sig, err := SignPatchHash(sampleHash, loadedKey)
	if err != nil {
		t.Fatalf("SignPatchHash failed: %v", err)
	}
	if len(sig) == 0 {
		t.Fatal("SignPatchHash returned empty signature")
	}

	err = VerifyPatchHash(sampleHash, sig, &loadedKey.PublicKey)
	if err != nil {
		t.Fatalf("VerifyPatchHash failed: %v", err)
	}

	// Verify with corrupted hash
	err = VerifyPatchHash("bad_hash", sig, &loadedKey.PublicKey)
	if err == nil {
		t.Fatal("VerifyPatchHash should have failed with bad hash")
	}
}
