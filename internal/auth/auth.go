package auth




import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)




type contextKey string

const UserContextKey contextKey = "user_id"

func GetUserID(r *http.Request) (string, error) {
	val := r.Context().Value(UserContextKey)
	if val == nil {
		return "", errors.New("no user found in context")
	}

	userID, ok := val.(string)
	if !ok {
		return "", errors.New("invalid user ID type in context")
	}

	return userID, nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}