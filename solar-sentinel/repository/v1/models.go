package v1

import "time"

type Device struct {
	ID			string	  `bson:"_id" json:"id"`
	Name       string    `bson:"name" json:"name"`
	DateCreated time.Time `bson:"date_created" json:"date_created"`
	DateUpdated time.Time `bson:"date_updated" json:"date_updated"`
}

type SolarIrradiance struct {
	DeviceID	string	  `bson:"device_id" json:"device_id"`
	Data	 map[string]interface{}	`bson:"data" json:"data"`
	DateCreated time.Time `bson:"date_created" json:"date_created"`
	DateUpdated time.Time `bson:"date_updated" json:"date_updated"`	
}

