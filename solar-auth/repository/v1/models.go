package v1

import "time"

type User struct {
	ID			string	  `bson:"_id" json:"id"`
	Email       string    `bson:"email" json:"email"`
	Key         string    `bson:"key" json:"key"`
	DateCreated time.Time `bson:"date_created" json:"date_created"`
	DateUpdated time.Time `bson:"date_updated" json:"date_updated"`
}
