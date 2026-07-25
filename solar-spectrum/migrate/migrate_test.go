package migrate

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/negeek/solar-sphere/solar-spectrum/idgen"
	"github.com/negeek/solar-sphere/solar-spectrum/mongoutil"
)

// testDB connects to a real MongoDB for integration testing. Set
// TEST_DATABASE_URL (e.g. mongodb://localhost:27017) to run these — they're
// skipped otherwise. Each test gets its own throwaway database, dropped and
// disconnected in cleanup.
func testDB(t *testing.T) *mongo.Database {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping migrate integration test")
	}

	ctx := context.Background()
	client, db, err := mongoutil.Connect(ctx, url, "solarsphere_migrate_test_"+idgen.New(""))
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

	return db
}

func TestRunAppliesEachMigrationOnce(t *testing.T) {
	db := testDB(t)

	runs := 0
	migrations := []Migration{
		{Name: "001_first", Up: func(context.Context, *mongo.Database) error { runs++; return nil }},
	}

	if err := Run(context.Background(), db, migrations); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if err := Run(context.Background(), db, migrations); err != nil {
		t.Fatalf("second Run: %v", err)
	}

	if runs != 1 {
		t.Fatalf("Up ran %d times, want 1 (already-applied migrations must be skipped)", runs)
	}
}

func TestRunAppliesInOrder(t *testing.T) {
	db := testDB(t)

	var order []string
	migrations := []Migration{
		{Name: "001_first", Up: func(context.Context, *mongo.Database) error { order = append(order, "001"); return nil }},
		{Name: "002_second", Up: func(context.Context, *mongo.Database) error { order = append(order, "002"); return nil }},
	}

	if err := Run(context.Background(), db, migrations); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(order) != 2 || order[0] != "001" || order[1] != "002" {
		t.Fatalf("order = %v, want [001 002]", order)
	}
}

func TestRunStopsOnFirstError(t *testing.T) {
	db := testDB(t)

	secondRan := false
	migrations := []Migration{
		{Name: "001_fails", Up: func(context.Context, *mongo.Database) error { return errors.New("boom") }},
		{Name: "002_never_runs", Up: func(context.Context, *mongo.Database) error { secondRan = true; return nil }},
	}

	if err := Run(context.Background(), db, migrations); err == nil {
		t.Fatal("expected Run to return an error when a migration fails")
	}
	if secondRan {
		t.Error("expected the second migration not to run after the first failed")
	}
}
