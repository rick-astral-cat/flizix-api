package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	db "github.com/rick-astral-cat/flizix-api/db/sqlc"
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
	return token.SignedString([]byte(api.JWTSecret))
}

func (api *Config) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
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

// HandleTelegramLogin manage login with Telegram
// @Summary      Login con Telegram
// @Description  Validate Telegram hash, search or create user emitting a JWT o a cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        data  body      TelegramAuthRequest  true  "telegram data"
// @Success      200   {object}  UserResponse
// @Router       /auth/telegram [post]
func (api *Config) HandleTelegramLogin(w http.ResponseWriter, r *http.Request) {
	var req TelegramAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := api.VerifyTelegramHash(req); err != nil {
		http.Error(w, "Invalid hash, not authorized", http.StatusUnauthorized)
		return
	}

	tgID := strconv.FormatInt(req.ID, 10)
	user, err := api.Queries.GetUserByTelegramId(r.Context(), sql.NullString{string(tgID), true})
	//Create user if not exists
	if err == sql.ErrNoRows {
		user, err = api.Queries.CreateUserWithTelegram(r.Context(), db.CreateUserWithTelegramParams{
			Name:       req.FirstName,
			Email:      sql.NullString{String: "", Valid: false},
			TelegramID: sql.NullString{String: tgID, Valid: true},
		})
		if err != nil {
			http.Error(w, "Error creating user with Telegram: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Error on database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//Generate JWT Token
	token, err := api.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Error generating token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Secure:   api.AppTLS,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MapUserToResponse(user))
}

// HandleLogout logout user session
// @Summary      Close session
// @Description  Delete cookie on browser session sending an expired cookie
// @Tags         auth
// @Success      200  {object}  map[string]string
// @Router       /auth/logout [post]
func (api *Config) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out"})
}
