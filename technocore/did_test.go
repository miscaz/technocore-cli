package technocore

import (
	"strings"
	"testing"
)

func TestDIDPrefix(t *testing.T) {
	id, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(id.DID, "did:key:z6Mk") {
		t.Fatalf("unexpected did prefix: %s", id.DID)
	}
}

func TestSeedRoundTrip(t *testing.T) {
	id, _ := Generate()
	clone, err := FromSeedHex(id.SeedHex())
	if err != nil {
		t.Fatal(err)
	}
	if clone.DID != id.DID {
		t.Fatalf("seed round-trip mismatch: %s != %s", clone.DID, id.DID)
	}
}

func TestSignVerify(t *testing.T) {
	id, _ := Generate()
	nonce := FreshNonce()
	sig := id.Sign("lobby", nonce, "gm")
	if len(sig) != 86 {
		t.Fatalf("expected 86-char signature, got %d", len(sig))
	}
	if !Verify(id.DID, "lobby", nonce, "gm", sig) {
		t.Fatal("valid signature failed to verify")
	}
	if Verify(id.DID, "lobby", nonce, "good morning", sig) {
		t.Fatal("tampered message verified")
	}
}

func TestKnownSeed(t *testing.T) {
	// Cross-language vector: this seed must yield this exact DID.
	id, err := FromSeedHex("06e0e75c3d37f7df0edf76c45547af575b61fe18d1dd8c807b2eabce93228b5b")
	if err != nil {
		t.Fatal(err)
	}
	want := "did:key:z6MkqaWnfiBjUSjxQFcMuVm8FQQgtQKgmLSYTnVgdccri8eV"
	if id.DID != want {
		t.Fatalf("did mismatch:\n got %s\nwant %s", id.DID, want)
	}
}
