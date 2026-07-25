package v1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/negeek/solar-sphere/solar-spectrum/accesskey"
	"github.com/negeek/solar-sphere/solar-spectrum/shared"
)

// fakeRepo is a hand-written in-memory stand-in for the repository, so
// these tests exercise AuthService's logic without needing a real database.
type fakeRepo struct {
	users         []shared.User
	revokedKeys   map[string]string // key -> email
	createUserErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{revokedKeys: map[string]string{}}
}

func (f *fakeRepo) CreateUser(_ context.Context, u *shared.User) error {
	if f.createUserErr != nil {
		return f.createUserErr
	}
	f.users = append(f.users, *u)
	return nil
}

func (f *fakeRepo) RevokeKey(_ context.Context, key, email string) error {
	f.revokedKeys[key] = email
	return nil
}

func (f *fakeRepo) IsKeyRevoked(_ context.Context, key string) (bool, error) {
	_, ok := f.revokedKeys[key]
	return ok, nil
}

func testKeyPair(t *testing.T) (verificationKey, signingKey string) {
	t.Helper()
	v, s, err := accesskey.GenerateKeyPair()
	if err != nil {
		t.Fatalf("generate key pair: %v", err)
	}
	return v, s
}

func TestSignUp(t *testing.T) {
	verifyKey, signKey := testKeyPair(t)
	repo := newFakeRepo()
	svc := NewAuthService(repo, signKey, verifyKey)

	result, err := svc.SignUp(context.Background(), SignUpInput{Email: "alice@example.com"})
	if err != nil {
		t.Fatalf("SignUp: %v", err)
	}
	if result.Email != "alice@example.com" {
		t.Errorf("Email = %q, want alice@example.com", result.Email)
	}
	if len(repo.users) != 1 {
		t.Fatalf("expected 1 user created, got %d", len(repo.users))
	}

	claims, err := accesskey.Verify(result.AccessKey, verifyKey)
	if err != nil {
		t.Fatalf("verify issued key: %v", err)
	}
	if claims.Email != "alice@example.com" {
		t.Errorf("claims.Email = %q, want alice@example.com", claims.Email)
	}
	if claims.ExpiresAt != nil {
		t.Errorf("expected no expiry by default, got %v", claims.ExpiresAt)
	}
}

func TestSignUpInvalidEmail(t *testing.T) {
	verifyKey, signKey := testKeyPair(t)
	repo := newFakeRepo()
	svc := NewAuthService(repo, signKey, verifyKey)

	if _, err := svc.SignUp(context.Background(), SignUpInput{Email: "not-an-email"}); !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("SignUp error = %v, want ErrInvalidEmail", err)
	}
	if len(repo.users) != 0 {
		t.Errorf("expected no user created for an invalid email")
	}
}

func TestSignUpRepositoryError(t *testing.T) {
	verifyKey, signKey := testKeyPair(t)
	repo := newFakeRepo()
	repo.createUserErr = errors.New("boom")
	svc := NewAuthService(repo, signKey, verifyKey)

	if _, err := svc.SignUp(context.Background(), SignUpInput{Email: "alice@example.com"}); err == nil {
		t.Fatal("expected an error when the repository fails")
	}
}

func TestSignUpWithExpiry(t *testing.T) {
	verifyKey, signKey := testKeyPair(t)
	repo := newFakeRepo()
	svc := NewAuthService(repo, signKey, verifyKey)

	result, err := svc.SignUp(context.Background(), SignUpInput{Email: "bob@example.com", ExpiresIn: time.Hour})
	if err != nil {
		t.Fatalf("SignUp: %v", err)
	}

	claims, err := accesskey.Verify(result.AccessKey, verifyKey)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if claims.ExpiresAt == nil {
		t.Fatal("expected expiry to be set when ExpiresIn is non-zero")
	}
}

func TestRotateKey(t *testing.T) {
	verifyKey, signKey := testKeyPair(t)
	repo := newFakeRepo()
	svc := NewAuthService(repo, signKey, verifyKey)

	signedUp, err := svc.SignUp(context.Background(), SignUpInput{Email: "carol@example.com"})
	if err != nil {
		t.Fatalf("SignUp: %v", err)
	}

	rotated, err := svc.RotateKey(context.Background(), RotateKeyInput{
		OldKey: signedUp.AccessKey,
		Email:  "carol@example.com",
	})
	if err != nil {
		t.Fatalf("RotateKey: %v", err)
	}
	if rotated.AccessKey == signedUp.AccessKey {
		t.Error("expected a new access key, got the same one back")
	}

	revoked, err := repo.IsKeyRevoked(context.Background(), signedUp.AccessKey)
	if err != nil {
		t.Fatalf("IsKeyRevoked: %v", err)
	}
	if !revoked {
		t.Error("expected the old key to be revoked after rotation")
	}

	// The old key must now be rejected outright, not just accepted because
	// its signature and expiry still check out.
	if _, err := svc.RotateKey(context.Background(), RotateKeyInput{OldKey: signedUp.AccessKey, Email: "carol@example.com"}); !errors.Is(err, ErrAlreadyRevoked) {
		t.Fatalf("reusing a revoked key: error = %v, want ErrAlreadyRevoked", err)
	}
}

func TestRotateKeyEmailMismatch(t *testing.T) {
	verifyKey, signKey := testKeyPair(t)
	repo := newFakeRepo()
	svc := NewAuthService(repo, signKey, verifyKey)

	signedUp, err := svc.SignUp(context.Background(), SignUpInput{Email: "dave@example.com"})
	if err != nil {
		t.Fatalf("SignUp: %v", err)
	}

	_, err = svc.RotateKey(context.Background(), RotateKeyInput{OldKey: signedUp.AccessKey, Email: "someone-else@example.com"})
	if !errors.Is(err, ErrEmailMismatch) {
		t.Fatalf("RotateKey error = %v, want ErrEmailMismatch", err)
	}
}

func TestRotateKeyInvalidKey(t *testing.T) {
	verifyKey, signKey := testKeyPair(t)
	repo := newFakeRepo()
	svc := NewAuthService(repo, signKey, verifyKey)

	_, err := svc.RotateKey(context.Background(), RotateKeyInput{OldKey: "not-a-jwt", Email: "eve@example.com"})
	if !errors.Is(err, ErrInvalidAccessKey) {
		t.Fatalf("RotateKey error = %v, want ErrInvalidAccessKey", err)
	}
}
