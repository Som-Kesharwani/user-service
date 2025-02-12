package helper

import (
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"github.com/Som-Kesharwani/shared-service/logger"
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

func GenerateTokens(email, firstName, lastName string, userID primitive.ObjectID) (accessToken, refreshToken string, expiredAt time.Time, err error) {
	// Generate a new JWT token

	accessTokenExpiresAt := time.Now().Add(time.Hour * 2)       // Access token expires in 15 minutes
	refreshTokenExpiresAt := time.Now().Add(time.Hour * 24 * 7) // Refresh token expires in 7 days

	// Create claims for the access token
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UserID:    userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExpiresAt),
		},
	}

	// Create the access token
	accessToken, err = createJWTToken(claims)

	if err != nil {
		logger.Error.Printf("Error generating access token: ", err)
		return "", "", time.Time{}, err
	}
	refreshClaim := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		UserID:    userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExpiresAt),
		},
	}
	refreshToken, err = createJWTToken(refreshClaim)

	if err != nil {
		logger.Error.Printf("Error generating refresh token: ", err)
		return "", "", time.Time{}, err
	}

	return accessToken, refreshToken, refreshTokenExpiresAt, nil

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
