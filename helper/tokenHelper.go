package helper

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// You should store this securely

// SignedDetails Claims for JWT tokens
type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	UserID    primitive.ObjectID
	jwt.RegisteredClaims
}

func init() {

}

func GenerateToken(email, firstName, lastName string, userID primitive.ObjectID, duration time.Duration) (accessToken string, err error) {
	tokenExpiresAt := time.Now().Add(duration) // Access token expires in 15 minutes

	// Create claims for the access token
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UserID:    userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(tokenExpiresAt),
		},
	}

	return createJWTToken(claims)
}

func createJWTToken(claims jwt.Claims) (string, error) {
	// Create a new JWT token with the given claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(Secretary)

}

func ValidateToken(signedToken string) (*SignedDetails, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return Secretary, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*SignedDetails)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}
