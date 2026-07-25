package shared

import "time"

// User is an authenticated identity. A user may own any number of Devices
// (see Device.Owner) — a user is not tied to a single device.
type User struct {
	ID          string    `bson:"_id" json:"id"`
	Email       string    `bson:"email" json:"email"`
	DateCreated time.Time `bson:"date_created" json:"date_created"`
	DateUpdated time.Time `bson:"date_updated" json:"date_updated"`
}

// Device is a solar-irradiance sensor, owned by the user whose email is in
// Owner. Devices are created by their owner (see solar-sentinel's
// /sentinel/v1/device/ endpoint) — there is no separate admin role.
type Device struct {
	ID          string    `bson:"_id" json:"id"`
	Name        string    `bson:"name" json:"name"`
	Owner       string    `bson:"owner" json:"owner"`
	DateCreated time.Time `bson:"date_created" json:"date_created"`
	DateUpdated time.Time `bson:"date_updated" json:"date_updated"`
}

type SolarIrradiance struct {
	DeviceID    string                 `bson:"device_id" json:"device_id"`
	Data        map[string]interface{} `bson:"data" json:"data"`
	DateCreated time.Time              `bson:"date_created" json:"date_created"`
	DateUpdated time.Time              `bson:"date_updated" json:"date_updated"`
}

// RevokedKey records an access key that has been explicitly invalidated
// before its natural expiry (or that has no expiry at all).
type RevokedKey struct {
	Key         string    `bson:"key" json:"key"`
	Email       string    `bson:"email" json:"email"`
	DateCreated time.Time `bson:"date_created" json:"date_created"`
}