package shared

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// RevokeKey records key as revoked for email, so IsKeyRevoked rejects it on
// future requests regardless of its JWT expiry.
func RevokeKey(ctx context.Context, db *mongo.Database, key, email string) error {
	_, err := db.Collection(REVOKED_KEY_COLLECTION).InsertOne(ctx, RevokedKey{
		Key:         key,
		Email:       email,
		DateCreated: time.Now().UTC(),
	})
	return err
}

// IsKeyRevoked reports whether key has been revoked. Both solar-auth (which
// revokes keys on rotation) and solar-sentinel (which authenticates
// requests) share this check against the same collection.
func IsKeyRevoked(ctx context.Context, db *mongo.Database, key string) (bool, error) {
	count, err := db.Collection(REVOKED_KEY_COLLECTION).CountDocuments(ctx, bson.D{{Key: "key", Value: key}})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
