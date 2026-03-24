// JWT utilities for user authentication
package user

import (
	"github.com/golang-jwt/jwt/v5"
	
	"time"
)

type Claims struct {
	UserID string `json:user_id`
	jwt.RegisteredClaims
}

func GenerateJWT(userID string , secretKey []byte) (string, error) {
	claims := &Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(45 * time.Minute)), // Token expires in 45 minutes
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secretKey)

	if err != nil {
		return "", err
	}
	return tokenString, nil

}

func VerifyJWT(tokenString string, secretKey []byte) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}