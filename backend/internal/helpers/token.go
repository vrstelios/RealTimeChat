package helpers

import (
	"RealTimeChat/backend/internal/database"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/v2/bson"

	"golang.org/x/crypto/bcrypt"
	"time"
)

const (
	CtxUserId = "userId"
	CtxEmail  = "email"
)

var jwtKey []byte

func InitJWT(key []byte) {
	jwtKey = key
}

type Claims struct {
	UserId bson.ObjectID `bson:"userId"`
	Email  string        `bson:"email"`

	jwt.StandardClaims
}

func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func HashPassword(password string) *string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	hasHedPwd := string(bytes)
	return &hasHedPwd
}

func VerifyPassword(userPwd, pwd string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(userPwd), []byte(pwd))
	return err == nil, err
}

func GenerateToken(email string, userId bson.ObjectID) (string, string) {
	tokenExpiry := time.Now().Add(24 * time.Hour).Unix()
	refreshTokenExpiry := time.Now().Add(7 * 24 * time.Hour).Unix()

	claims := &Claims{
		Email:  email,
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: tokenExpiry,
		},
	}
	refreshClaims := &Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: refreshTokenExpiry,
		},
	}

	// Generating the tokens
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedAccessRefreshToken, err := accessToken.SignedString(jwtKey)
	if err != nil {
		panic(err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err := refreshToken.SignedString(jwtKey)
	if err != nil {
		panic(err)
	}

	return signedAccessRefreshToken, signedRefreshToken
}

func UpdatedAllToken(signedToken string, signedRefreshToken string, userId bson.ObjectID) error {
	user, err := database.GetUser(userId, "")
	if err != nil {
		return err
	}
	user[0].Token = &signedToken
	user[0].RefreshToken = &signedRefreshToken

	err = database.PutUser(user[0].Id, user[0])
	if err != nil {
		return err
	}

	return nil
}
