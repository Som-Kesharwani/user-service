package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/Som-Kesharwani/shared-service/database"
	"github.com/go-playground/validator/v10"
	"net/http"
	"time"

	"github.com/Som-Kesharwani/shared-service/logger"
	"github.com/Som-Kesharwani/user-service/helper"
	"github.com/Som-Kesharwani/user-service/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection
var validate *validator.Validate
var refreshTokenCollection *mongo.Collection

type UserResponse struct {
	ID        primitive.ObjectID `bson:"id"`
	FirstName *string            `bson:"first_name" binding:"required"`
	LastName  *string            `bson:"last_name" binding:"required"`
	Email     *string            `bson:"email"  binding:"required"`
	PhoneNo   *string            `json:"phone" validate:"required,min=10,max=10"`
}

func init() {
	userCollection = database.OpenCollection(database.Client, "user")
	refreshTokenCollection = database.OpenCollection(database.Client, "refreshToken")
	validate = validator.New()
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var user UserResponse
		userID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = userCollection.FindOne(ctx, bson.M{"id": userID}).Decode(&user)
		if err != nil {
			logger.Error.Println("User not found")
			c.JSON(http.StatusBadRequest, gin.H{"error": "User not found!!"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user []UserResponse
		cursor, err := userCollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := cursor.All(ctx, &user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		defer func(cursor *mongo.Cursor, ctx context.Context) {
			err := cursor.Close(ctx)
			if err != nil {
				logger.Error.Printf("Failed to close cursor: %v", err)
			}
		}(cursor, ctx)

		c.JSON(http.StatusOK, user)
	}
}

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			//logger.Error.Printf("Json Parser Failed with error %s", err.Error())
			return
		}
		//logger.Info.Printf("User Data recieved : %s", user)

		validationErr := validate.Struct(user)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			logger.Error.Printf("Validation Error : %s", validationErr.Error())
			return
		}

		var existingUser models.User
		err := userCollection.FindOne(ctx, bson.M{"email": *user.Email}).Decode(&existingUser)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
			logger.Error.Printf("User with this email already exists")
			return
		}

		userID := primitive.NewObjectID()
		user.ID = userID
		user.CreateAt = time.Now()
		user.UpdateAt = time.Now()

		passwordHash, err := PasswordHash(*user.Password)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing the password"})
			return
		}

		user.Password = &passwordHash

		_, err = userCollection.InsertOne(ctx, user)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating the user"})
			return
		}

		userResponse := &UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
			PhoneNo:   user.PhoneNo,
		}

		c.JSON(http.StatusOK, gin.H{"message": "user successfully Signed up!", "user": userResponse})

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var loginDetails struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required" validate:"min=8"`
		}

		if err := c.BindJSON(&loginDetails); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		validatorErr := validate.Struct(loginDetails)

		if validatorErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validatorErr.Error()})
			return
		}

		// Check if the user exists by email
		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"email": loginDetails.Email}).Decode(&user)
		if err != nil {
			// If no user is found with the provided email
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		// Verify the password
		passwordIsValid, msg := VerifyHash(loginDetails.Password, *user.Password)
		if !passwordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}

		// Generate new tokens (access and refresh)
		var accessToken string
		accessToken, err = helper.GenerateToken(*user.Email, *user.FirstName, *user.LastName, user.ID, time.Hour*1)
		if err != nil {
			logger.Error.Printf("Failed to generate access token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		}

		var refreshToken string
		refreshToken, err = helper.GenerateToken(*user.Email, *user.FirstName, *user.LastName, user.ID, time.Hour*10)
		//	_, refreshToken, expiredAt, err := helper.GenerateTokens(*user.Email, *user.FirstName, *user.LastName, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating tokens"})
			return
		}

		err = StoreRefreshToken(ctx, user.ID, refreshToken, time.Now().Add(time.Hour*10))
		if err != nil {
			logger.Error.Printf("Failed to store refresh token: %v", err)
		}

		// Respond with user details and tokens
		c.JSON(http.StatusOK, gin.H{
			"message":       "Successfully logged in",
			"token":         accessToken,
			"refresh_token": refreshToken,
		})

	}
}

func GetUserToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		var refreshToken []models.RefreshToken

		cursor, err := refreshTokenCollection.Find(ctx, bson.M{})

		if err != nil {
			logger.Error.Printf("Error finding all tokens: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching tokens"})
			return
		}

		err = cursor.All(ctx, &refreshToken)
		if err != nil {
			logger.Error.Printf("Error getting all tokens: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching tokens"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"tokens": refreshToken})

	}
}

func RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var requestBody struct {
			RefreshToken string `json:"refresh_token"`
		}

		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Decode JWT to get userID
		claims, err := helper.ValidateToken(requestBody.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		// Cross-check with DB
		valid, err := ValidateRefreshToken(ctx, claims.UserID, requestBody.RefreshToken)
		if !valid || err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		// Generate new tokens
		accessToken, err := helper.GenerateToken(claims.Email, claims.FirstName, claims.LastName, claims.UserID, time.Hour*10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating tokens"})
			return
		}

		// Send the new tokens
		c.JSON(http.StatusOK, gin.H{
			"access_token": accessToken,
			"expires_in":   time.Now().Add(time.Hour * 10).Local(),
		})
	}
}

func ValidateRefreshToken(ctx context.Context, userID primitive.ObjectID, refreshToken string) (bool, error) {
	var existingUser models.User

	err := userCollection.FindOne(ctx, bson.M{"id": userID}).Decode(&existingUser)

	if err != nil {
		logger.Error.Printf("Error finding user: %v", err)
		return false, err
	}
	var refreshedToken models.RefreshToken
	err = refreshTokenCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&refreshedToken)

	token := HashToken(refreshToken)
	if ok, msg := VerifyTokenHash(refreshedToken.TokenHash, token); !ok {
		return false, errors.New(msg)
	}

	return true, nil
}
func StoreRefreshToken(ctx context.Context, userID primitive.ObjectID, refreshToken string, expiredAt time.Time) error {
	var refreshTokenObj models.RefreshToken
	refreshTokenObj.UserID = userID
	refreshTokenObj.TokenHash = HashToken(refreshToken)

	_, err := refreshTokenCollection.DeleteOne(ctx, bson.M{"user_id": userID})
	if err != nil {
		return err
	}

	_, err = refreshTokenCollection.InsertOne(ctx, refreshTokenObj)
	if err != nil {
		return err
	}

	return nil

}

/*
	func DeleteTokens() gin.HandlerFunc {
			return func(c *gin.Context) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				refreshTokenCollection.DeleteMany(ctx, bson.M{})
				c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted tokens"})
			}
		}

func RevokeRefreshToken(ctx context.Context, userID string) error {

		_, err := refreshTokenCollection.DeleteOne(ctx, bson.M{"user_id": userID})
		return err
	}

func StoreRefreshToken(ctx context.Context, userID primitive.ObjectID, refreshToken string) error {

		var refreshTokenDoc models.RefreshToken

		refreshTokenDoc.ID = userID
		refreshTokenDoc.TokenHash = HashToken(refreshToken)
		refreshTokenDoc.ExpiresAt = time.Now().Add(time.Hour * 24 * 7)

		_, err := refreshTokenCollection.InsertOne(ctx, refreshTokenDoc)
		return err
	}
*/
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
func PasswordHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func VerifyHash(providedPassword string, storedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(providedPassword))
	if err != nil {
		logger.Error.Printf(err.Error())
		return false, "Invalid email or password"
	}
	return true, ""
}

func VerifyTokenHash(userToken string, storedToken string) (bool, string) {
	if userToken != storedToken {
		return false, "Invalid token"
	}
	return true, ""
}
