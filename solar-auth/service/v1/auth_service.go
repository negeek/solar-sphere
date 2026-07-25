// Package v1 is solar-auth's service layer: it owns validation and business
// rules for sign-up and access-key rotation. Handlers only decode requests
// and call into this package; this package is the only caller of the
// repository.
package v1

import (
	"context"
	"net/mail"
	"time"

	repo "github.com/negeek/solar-sphere/solar-auth/repository/v1"
	"github.com/negeek/solar-sphere/solar-spectrum/accesskey"
	"github.com/negeek/solar-sphere/solar-spectrum/idgen"
	"github.com/negeek/solar-sphere/solar-spectrum/shared"
)

type AuthService struct {
	repo            *repo.Repository
	signingKey      string
	verificationKey string
}

func NewAuthService(r *repo.Repository, signingKey, verificationKey string) *AuthService {
	return &AuthService{repo: r, signingKey: signingKey, verificationKey: verificationKey}
}

// AccessKeyResult is returned by both SignUp and RotateKey.
type AccessKeyResult struct {
	Email     string
	AccessKey string
}

type SignUpInput struct {
	Email string
	// ExpiresIn is how long the issued key stays valid. Zero means it never
	// expires on its own (revocation is the only way to invalidate it).
	ExpiresIn time.Duration
}

// SignUp creates a new user identity and issues their first access key.
// Device creation is handled separately by solar-sentinel, once the user
// has a key to authenticate with — a user may go on to register any number
// of devices.
func (s *AuthService) SignUp(ctx context.Context, in SignUpInput) (*AccessKeyResult, error) {
	if _, err := mail.ParseAddress(in.Email); err != nil {
		return nil, ErrInvalidEmail
	}

	user := &shared.User{
		ID:    idgen.New(in.Email),
		Email: in.Email,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	key, err := accesskey.Generate(in.Email, s.signingKey, accesskey.GenerateOptions{ExpiresIn: in.ExpiresIn})
	if err != nil {
		return nil, err
	}

	return &AccessKeyResult{Email: in.Email, AccessKey: key}, nil
}

type RotateKeyInput struct {
	OldKey    string
	Email     string
	ExpiresIn time.Duration
}

// RotateKey verifies the caller's current key, revokes it, and issues a
// replacement. The old key stops working immediately — revocation is
// checked (not just signature/expiry) everywhere access keys are verified.
func (s *AuthService) RotateKey(ctx context.Context, in RotateKeyInput) (*AccessKeyResult, error) {
	claims, err := accesskey.Verify(in.OldKey, s.verificationKey)
	if err != nil {
		return nil, ErrInvalidAccessKey
	}
	if claims.Email != in.Email {
		return nil, ErrEmailMismatch
	}

	revoked, err := s.repo.IsKeyRevoked(ctx, in.OldKey)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, ErrAlreadyRevoked
	}

	if err := s.repo.RevokeKey(ctx, in.OldKey, in.Email); err != nil {
		return nil, err
	}

	newKey, err := accesskey.Generate(in.Email, s.signingKey, accesskey.GenerateOptions{ExpiresIn: in.ExpiresIn})
	if err != nil {
		return nil, err
	}

	return &AccessKeyResult{Email: in.Email, AccessKey: newKey}, nil
}

// RecoverAccessKey is intentionally not implemented yet: it needs an email
// delivery mechanism to send the caller a fresh key/link, which nothing in
// this codebase provides today.
