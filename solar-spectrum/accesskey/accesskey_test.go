package accesskey

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateVerifyRoundTrip(t *testing.T) {
	verifyKey, signKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	key, err := Generate("alice@example.com", signKey, GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	claims, err := Verify(key, verifyKey)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.Email != "alice@example.com" {
		t.Errorf("Email = %q, want alice@example.com", claims.Email)
	}
	if claims.ExpiresAt != nil {
		t.Errorf("expected no expiry when ExpiresIn is zero, got %v", claims.ExpiresAt)
	}
	if claims.ID == "" {
		t.Error("expected a non-empty jti")
	}
}

func TestGenerateTwiceProducesDifferentKeys(t *testing.T) {
	verifyKey, signKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	first, err := Generate("alice@example.com", signKey, GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	second, err := Generate("alice@example.com", signKey, GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if first == second {
		t.Fatal("expected two keys issued for the same email to differ, got identical tokens")
	}

	if _, err := Verify(second, verifyKey); err != nil {
		t.Fatalf("Verify(second): %v", err)
	}
}

func TestGenerateRejectsNegativeExpiresIn(t *testing.T) {
	_, signKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	if _, err := Generate("alice@example.com", signKey, GenerateOptions{ExpiresIn: -time.Hour}); err == nil {
		t.Fatal("expected Generate to reject a negative ExpiresIn")
	}
}

func TestVerifyRejectsExpiredKey(t *testing.T) {
	verifyKey, signKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	// Built directly (bypassing Generate, which now refuses to mint an
	// already-expired token) so this test exercises Verify's own expiry
	// check in isolation.
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
		Email: "alice@example.com",
	}
	signingKey, err := hex.DecodeString(signKey)
	if err != nil {
		t.Fatalf("decode signing key: %v", err)
	}
	key, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(ed25519.PrivateKey(signingKey))
	if err != nil {
		t.Fatalf("sign expired token: %v", err)
	}

	if _, err := Verify(key, verifyKey); err == nil {
		t.Fatal("expected an error verifying an already-expired key")
	}
}

func TestVerifyRejectsWrongKeyPair(t *testing.T) {
	_, signKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	otherVerifyKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	key, err := Generate("alice@example.com", signKey, GenerateOptions{})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if _, err := Verify(key, otherVerifyKey); err == nil {
		t.Fatal("expected verification to fail against an unrelated verification key")
	}
}

func TestVerifyRejectsGarbage(t *testing.T) {
	verifyKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	if _, err := Verify("not-a-jwt", verifyKey); err == nil {
		t.Fatal("expected an error verifying a non-JWT string")
	}
}
