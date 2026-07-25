// Package v1 is solar-auth's repository layer: it owns the *mongo.Database
// handle passed to it and is the only place that talks to MongoDB directly.
package v1

import (
	"context"
	"time"

	"github.com/negeek/solar-sphere/solar-spectrum/shared"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, u *shared.User) error {
	now := time.Now().UTC()
	u.DateCreated, u.DateUpdated = now, now
	_, err := r.db.Collection(shared.USER_COLLECTION).InsertOne(ctx, u)
	return err
}

func (r *Repository) RevokeKey(ctx context.Context, key, email string) error {
	return shared.RevokeKey(ctx, r.db, key, email)
}

func (r *Repository) IsKeyRevoked(ctx context.Context, key string) (bool, error) {
	return shared.IsKeyRevoked(ctx, r.db, key)
}
