package v1

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

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
	client, db, err := mongoutil.Connect(ctx, url, "solarsphere_sentinel_test_"+idgen.New(""))
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

func TestRepositoryCreateAndFindDevice(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()

	device := &shared.Device{ID: "device-1", Name: "backyard sensor", Owner: "alice@example.com"}
	if err := repo.CreateDevice(ctx, device); err != nil {
		t.Fatalf("CreateDevice: %v", err)
	}

	found, err := repo.FindDeviceByID(ctx, "device-1")
	if err != nil {
		t.Fatalf("FindDeviceByID: %v", err)
	}
	if found.Owner != "alice@example.com" {
		t.Errorf("Owner = %q, want alice@example.com", found.Owner)
	}
}

func TestRepositoryFindDeviceByIDNotFound(t *testing.T) {
	repo := testRepository(t)

	_, err := repo.FindDeviceByID(context.Background(), "does-not-exist")
	if !errors.Is(err, mongo.ErrNoDocuments) {
		t.Fatalf("error = %v, want mongo.ErrNoDocuments", err)
	}
}

func TestRepositoryIrradianceRoundTrip(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()

	reading := &shared.SolarIrradiance{DeviceID: "device-1", Data: map[string]interface{}{"lux": 100}}
	if err := repo.CreateIrradianceReading(ctx, reading); err != nil {
		t.Fatalf("CreateIrradianceReading: %v", err)
	}

	readings, err := repo.FindIrradianceByDevice(ctx, "device-1")
	if err != nil {
		t.Fatalf("FindIrradianceByDevice: %v", err)
	}
	if len(readings) != 1 {
		t.Fatalf("expected 1 reading, got %d", len(readings))
	}
}

func TestRepositoryUserExists(t *testing.T) {
	repo := testRepository(t)
	ctx := context.Background()

	exists, err := repo.UserExists(ctx, "nobody@example.com")
	if err != nil {
		t.Fatalf("UserExists: %v", err)
	}
	if exists {
		t.Fatal("expected user to not exist")
	}

	if _, err := repo.db.Collection(shared.USER_COLLECTION).InsertOne(ctx, shared.User{ID: "user-1", Email: "alice@example.com"}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	exists, err = repo.UserExists(ctx, "alice@example.com")
	if err != nil {
		t.Fatalf("UserExists: %v", err)
	}
	if !exists {
		t.Fatal("expected user to exist after seeding")
	}
}
