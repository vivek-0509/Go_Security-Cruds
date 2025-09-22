package security

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func CreateJWT(email string) (string, error) {
	claims := jwt.MapClaims{}
	claims["sub"] = email
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
