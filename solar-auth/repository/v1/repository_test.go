package v1

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/negeek/solar-sphere/solar-spectrum/idgen"
	"github.com/negeek/solar-sphere/solar-spectrum/mongoutil"
	"github.com/negeek/solar-sphere/solar-spectrum/shared"
)

// testRepository connects to a real MongoDB for integration testing. Set
// TEST_DATABASE_URL (e.g. mongodb://localhost:27017) to run these — they're
// skipped otherwise. Each test gets its own throwaway database, dropped and
// disconnected in cleanup.
func testRepository(t *testing.T) *Repository {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping repository integration test")
	}

	ctx := context.Background()
	client, db, err := mongoutil.Connect(ctx, url, "solarsphere_auth_test_"+idgen.New(""))
	if err != nil {
		t.Fatalf("connect to test mongo: %v", err)
	}

	t.Cleanup(func() {
		dropCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := db.Drop(dropCtx); err != nil {
			t.Errorf("drop test database: %v", err)
		}
		if err := mongoutil.Disconnect(context.Background(), client, 10*time.Second); err != nil {
			t.Errorf("disconnect test mongo: %v", err)
		}
	})

	return NewRepository(db)
}

func TestCreateUser(t *testing.T) {
	repo := testRepository(t)

	user := &shared.User{ID: "user-1", Email: "alice@example.com"}
	if err := repo.CreateUser(context.Background(), user); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.DateCreated.IsZero() || user.DateUpdated.IsZero() {
		t.Error("expected DateCreated/DateUpdated to be set by CreateUser")
	}
}

func TestRevokeKeyAndIsKeyRevoked(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()

	revoked, err := repo.IsKeyRevoked(ctx, "some-key")
	if err != nil {
		t.Fatalf("IsKeyRevoked: %v", err)
	}
	if revoked {
		t.Fatal("expected key to not be revoked yet")
	}

	if err := repo.RevokeKey(ctx, "some-key", "alice@example.com"); err != nil {
		t.Fatalf("RevokeKey: %v", err)
	}

	revoked, err = repo.IsKeyRevoked(ctx, "some-key")
	if err != nil {
		t.Fatalf("IsKeyRevoked: %v", err)
	}
	if !revoked {
		t.Fatal("expected key to be revoked after RevokeKey")
	}
}
