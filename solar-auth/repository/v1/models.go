package v1

import "time"

type User struct {
	ID			string	  `bson:"_id" json:"id"`
	Email       string    `bson:"email" json:"email"`
	DateCreated time.Time `bson:"date_created" json:"date_created"`
	DateUpdated time.Time `bson:"date_updated" json:"date_updated"`
}


type AccessKey struct {
	Key			string	  `json:"key"`
	Email       string    `json:"email"`
	Revoked		bool	  `bson:"revoked"`	
}
