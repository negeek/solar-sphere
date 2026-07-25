// Package migrate runs an ordered list of forward-only schema migrations
// against MongoDB, tracking which have already been applied in a
// "migrations" collection. Unlike the previous approach (compiling each
// migration file into a Go plugin at runtime via `go build
// -buildmode=plugin`), this requires no toolchain inside the running
// process, works on every platform Go plugins don't (e.g. Windows), and
// runs fine in a minimal container image.
package migrate

import (
	"context"
	"fmt"
	"time"

	"github.com/negeek/solar-sphere/solar-spectrum/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Migration is one forward-only schema change.
type Migration struct {
	Name string
	Up   func(ctx context.Context, db *mongo.Database) error
}

// Run applies each migration, in order, that hasn't been recorded as
// applied yet. Safe to call on every startup — already-applied migrations
// are skipped.
func Run(ctx context.Context, db *mongo.Database, migrations []Migration) error {
	applied := db.Collection(shared.MIGRATION_COLLECTION)

	for _, m := range migrations {
		count, err := applied.CountDocuments(ctx, bson.D{{Key: "name", Value: m.Name}})
		if err != nil {
			return fmt.Errorf("migrate: check %s: %w", m.Name, err)
		}
		if count > 0 {
			continue
		}

		if err := m.Up(ctx, db); err != nil {
			return fmt.Errorf("migrate: run %s: %w", m.Name, err)
		}

		if _, err := applied.InsertOne(ctx, bson.D{
			{Key: "name", Value: m.Name},
			{Key: "applied_at", Value: time.Now().UTC()},
		}); err != nil {
			return fmt.Errorf("migrate: record %s applied: %w", m.Name, err)
		}
	}
	return nil
}
