// Package v1 is solar-sentinel's repository layer: it owns the
// *mongo.Database handle passed to it and is the only place that talks to
// MongoDB directly.
package v1

import (
	"context"
	"time"

	"github.com/negeek/solar-sphere/solar-spectrum/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateDevice(ctx context.Context, d *shared.Device) error {
	now := time.Now().UTC()
	d.DateCreated, d.DateUpdated = now, now
	_, err := r.db.Collection(shared.DEVICE_COLLECTION).InsertOne(ctx, d)
	return err
}

func (r *Repository) FindDeviceByID(ctx context.Context, id string) (*shared.Device, error) {
	var d shared.Device
	if err := r.db.Collection(shared.DEVICE_COLLECTION).FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repository) CreateIrradianceReading(ctx context.Context, reading *shared.SolarIrradiance) error {
	now := time.Now().UTC()
	reading.DateCreated, reading.DateUpdated = now, now
	_, err := r.db.Collection(shared.IRR_COLLECTION).InsertOne(ctx, reading)
	return err
}

func (r *Repository) FindIrradianceByDevice(ctx context.Context, deviceID string) ([]shared.SolarIrradiance, error) {
	cursor, err := r.db.Collection(shared.IRR_COLLECTION).Find(ctx, bson.D{{Key: "device_id", Value: deviceID}})
	if err != nil {
		return nil, err
	}
	var data []shared.SolarIrradiance
	if err := cursor.All(ctx, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (r *Repository) UserExists(ctx context.Context, email string) (bool, error) {
	count, err := r.db.Collection(shared.USER_COLLECTION).CountDocuments(ctx, bson.D{{Key: "email", Value: email}})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) IsKeyRevoked(ctx context.Context, key string) (bool, error) {
	return shared.IsKeyRevoked(ctx, r.db, key)
}
