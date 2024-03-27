package v1

import "time"

type User struct {
	ID			string	  `bson:"_id" json:"id"`
	Email       string    `bson:"email" json:"email"`
	DeviceID    string    `bson:"device_id" json:"device_id"`
	DateCreated time.Time `bson:"date_created" json:"date_created"`
	DateUpdated time.Time `bson:"date_updated" json:"date_updated"`
}


type RevokedKey struct {
	Key			string	  `bson:"key" json:"key"`
	Email       string    `bson:"email" json:"email"`	
}
