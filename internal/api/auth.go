package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type TelegramAuthRequest struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

// GenerateToken Generate new signed JWT
func (api *Config) GenerateToken(userID int64) (string, error) {
	claims := CustomClaims{
		userID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "flizix-api",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(api.JWTSecret)
}

func (api *Config) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(api.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (api *Config) VerifyTelegramHash(req TelegramAuthRequest) error {
	data := []string{
		fmt.Sprintf("auth_date=%d", req.AuthDate),
		fmt.Sprintf("first_name=%s", req.FirstName),
		fmt.Sprintf("id=%d", req.ID),
	}
	if req.Username != "" {
		data = append(data, fmt.Sprintf("username=%s", req.Username))
	}
	if req.PhotoURL != "" {
		data = append(data, fmt.Sprintf("photo_url=%s", req.PhotoURL))
	}
	sort.Strings(data)
	dataCheckString := strings.Join(data, "\n")
	sha := sha256.New()
	sha.Write([]byte(api.TelegramBotToken))
	secretKey := sha.Sum(nil)
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(h.Sum(nil))
	if calculatedHash != req.Hash {
		return fmt.Errorf("invalid hash")
	}

	return nil
}
