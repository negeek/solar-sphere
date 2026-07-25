// Package accesskey mints and verifies the EdDSA-signed JWTs solar-sphere
// uses as bearer credentials ("access keys").
package accesskey

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/negeek/solar-sphere/solar-spectrum/idgen"
)

// Claims is the JWT payload carried by a solar-sphere access key.
type Claims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

// GenerateOptions controls how a new access key is minted.
type GenerateOptions struct {
	// ExpiresIn is how long the key stays valid. Zero means the key never
	// expires on its own — it stays valid until it is revoked.
	ExpiresIn time.Duration
}

// Generate mints a new access key for email, signed with signingKeyHex (a
// hex-encoded ed25519 private key).
func Generate(email string, signingKeyHex string, opts GenerateOptions) (string, error) {
	if opts.ExpiresIn < 0 {
		return "", errors.New("accesskey: ExpiresIn must not be negative")
	}

	signingKey, err := hex.DecodeString(signingKeyHex)
	if err != nil {
		return "", errors.New("accesskey: invalid signing key")
	}

	now := time.Now().UTC()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// ID makes every token unique even if two are issued for the
			// same email within the same second: EdDSA signing is
			// deterministic, so without it, two otherwise-identical claims
			// (same email, same truncated-to-the-second IssuedAt, no
			// expiry) would sign to the exact same token string.
			ID:       idgen.New(""),
			IssuedAt: jwt.NewNumericDate(now),
		},
		Email: email,
	}
	if opts.ExpiresIn > 0 {
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(opts.ExpiresIn))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(ed25519.PrivateKey(signingKey))
}

// Verify parses key and validates its signature and expiry against
// verificationKeyHex (a hex-encoded ed25519 public key).
//
// Verify does NOT check revocation — a key can have a valid signature and
// still have been revoked. Callers must additionally check the presented
// key string against the revocation store (shared.IsKeyRevoked) before
// trusting it.
func Verify(key string, verificationKeyHex string) (*Claims, error) {
	verificationKey, err := hex.DecodeString(verificationKeyHex)
	if err != nil {
		return nil, errors.New("accesskey: invalid verification key")
	}

	claims := &Claims{}
	_, err = jwt.ParseWithClaims(key, claims, func(token *jwt.Token) (interface{}, error) {
		return ed25519.PublicKey(verificationKey), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// GenerateKeyPair creates a new ed25519 signing/verification keypair, hex
// encoded, ready to drop into SIGNING_KEY/VERIFICATION_KEY.
func GenerateKeyPair() (verificationKeyHex, signingKeyHex string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", err
	}
	return hex.EncodeToString(pub), hex.EncodeToString(priv), nil
}
