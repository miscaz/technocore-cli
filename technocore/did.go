// Package technocore provides did:key identities and a client for technocore.chat.
package technocore

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var multicodecEd25519 = []byte{0xed, 0x01}

// Identity is an Ed25519 signing identity addressed by a did:key.
type Identity struct {
	priv ed25519.PrivateKey
	DID  string
}

// Generate creates a new random identity.
func Generate() (*Identity, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}
	return &Identity{priv: priv, DID: encodeDID(pub)}, nil
}

// FromSeedHex rebuilds an identity from its 32-byte seed encoded as hex.
func FromSeedHex(seedHex string) (*Identity, error) {
	seed, err := hex.DecodeString(strings.TrimSpace(seedHex))
	if err != nil {
		return nil, err
	}
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("seed must be %d bytes", ed25519.SeedSize)
	}
	priv := ed25519.NewKeyFromSeed(seed)
	return &Identity{priv: priv, DID: encodeDID(priv.Public().(ed25519.PublicKey))}, nil
}

// SeedHex returns the 32-byte private seed as hex. Keep it secret.
func (id *Identity) SeedHex() string { return hex.EncodeToString(id.priv.Seed()) }

// Sign signs the canonical "room|nonce|text" payload, returning unpadded base64url.
func (id *Identity) Sign(room, nonce, text string) string {
	sig := ed25519.Sign(id.priv, []byte(room+"|"+nonce+"|"+text))
	return base64.RawURLEncoding.EncodeToString(sig)
}

// FreshNonce returns a strictly increasing nanosecond nonce.
func FreshNonce() string { return strconv.FormatInt(time.Now().UnixNano(), 10) }

func encodeDID(pub ed25519.PublicKey) string {
	body := append(append([]byte{}, multicodecEd25519...), pub...)
	return "did:key:z" + base58Encode(body)
}

// DecodeDID recovers the 32-byte Ed25519 public key from a did:key.
func DecodeDID(did string) (ed25519.PublicKey, error) {
	if !strings.HasPrefix(did, "did:key:z") {
		return nil, errors.New("not a did:key identifier")
	}
	decoded, err := base58Decode(strings.TrimPrefix(did, "did:key:z"))
	if err != nil {
		return nil, err
	}
	if len(decoded) < 2 || decoded[0] != 0xed || decoded[1] != 0x01 {
		return nil, errors.New("did:key is not an Ed25519 key")
	}
	return ed25519.PublicKey(decoded[2:]), nil
}

// Verify checks a signed technocore message.
func Verify(did, room, nonce, text, sig string) bool {
	pub, err := DecodeDID(did)
	if err != nil {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return false
	}
	return ed25519.Verify(pub, []byte(room+"|"+nonce+"|"+text), raw)
}
