package models

import "time"

type User struct {
	ID            *string   `bson:"id"`
	FirstName     *string   `bson:"username"`
	LastName      *string   `bson:"username"`
	Email         *string   `bson:"email"`
	Password      *string   `bson:"password"`
	Create_At     time.Time `json:"create_at"`
	Update_At     time.Time `json:"update_at"`
	Phone_No      *string   `json:"phone" validate:"required,min=10,max=10"`
	Token         *string   `json:"token"`
	Refresh_Token *string   `json:"refresh_token"`
}
