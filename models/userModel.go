package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"id"`
	FirstName *string            `bson:"first_name" binding:"required"`
	LastName  *string            `bson:"last_name" binding:"required"`
	Email     *string            `bson:"email"  binding:"required"`
	Password  *string            `bson:"password"  binding:"required" validate:"min=8"`
	CreateAt  time.Time          `json:"create_at"`
	UpdateAt  time.Time          `json:"update_at"`
	PhoneNo   *string            `json:"phone" validate:"required,min=10,max=10"`
	//TokenHash string             `json:"token_hash"`
}

type RefreshToken struct {
	ID        primitive.ObjectID `bson:"id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id""`
	TokenHash string             `json:"token_hash"`
}
