package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

type AuthHandler struct {
	Queries          *db.Queries
	JWTSecret        string
	TelegramBotToken string
	AppTLS           bool
}

func NewAuthHandler(Queries *db.Queries, JWTSecret string, TelegramBotToken string, AppTLS bool) *AuthHandler {
	return &AuthHandler{
		Queries:          Queries,
		JWTSecret:        JWTSecret,
		TelegramBotToken: TelegramBotToken,
		AppTLS:           AppTLS,
	}
}

// GenerateToken Generate new signed JWT
func (h *AuthHandler) GenerateToken(userID int64) (string, error) {
	claims := CustomClaims{
		userID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "flizix-api",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.JWTSecret))
}

func (h *AuthHandler) ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (h *AuthHandler) VerifyTelegramHash(req TelegramAuthRequest) error {
	data := []string{
		fmt.Sprintf("auth_date=%d", req.AuthDate),
		fmt.Sprintf("first_name=%s", req.FirstName),
		fmt.Sprintf("id=%d", req.ID),
	}
	if req.Username != "" {
		data = append(data, fmt.Sprintf("username=%s", req.Username))
	}
	if req.LastName != "" {
		data = append(data, fmt.Sprintf("last_name=%s", req.LastName))
	}
	if req.PhotoURL != "" {
		data = append(data, fmt.Sprintf("photo_url=%s", req.PhotoURL))
	}
	sort.Strings(data)
	dataCheckString := strings.Join(data, "\n")
	sha := sha256.New()
	sha.Write([]byte(h.TelegramBotToken))
	secretKey := sha.Sum(nil)
	hm := hmac.New(sha256.New, secretKey)
	hm.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(hm.Sum(nil))
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
func (h *AuthHandler) HandleTelegramLogin(w http.ResponseWriter, r *http.Request) {
	var req TelegramAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if err := h.VerifyTelegramHash(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid hash, not authorized")
		return
	}

	tgID := strconv.FormatInt(req.ID, 10)
	user, err := h.Queries.GetUserByTelegramId(r.Context(), sql.NullString{String: string(tgID), Valid: true})
	//Create user if not exists
	if err == sql.ErrNoRows {
		user, err = h.Queries.CreateUserWithTelegram(r.Context(), db.CreateUserWithTelegramParams{
			Name:       req.FirstName,
			Email:      sql.NullString{String: "", Valid: false},
			TelegramID: sql.NullString{String: tgID, Valid: true},
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error creating user with Telegram: "+err.Error())
			return
		}
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error on database: "+err.Error())
		return
	}

	//Generate JWT Token
	token, err := h.GenerateToken(user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating token: "+err.Error())
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Secure:   h.AppTLS,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	respondWithJSON(w, http.StatusOK, MapUserToResponse(user))
}

// HandleDevLogin Manage simulated login for development environment.
// @Summary      Development login
// @Description  Omits Telegram validation and generate a JWT token for a testing user.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /auth/dev-login [post]
func (h *AuthHandler) HandleDevLogin(w http.ResponseWriter, r *http.Request) {
	tgID := "dev_test_user"
	user, err := h.Queries.GetUserByTelegramId(r.Context(), sql.NullString{String: string(tgID), Valid: true})
	if errors.Is(err, sql.ErrNoRows) {
		user, err = h.Queries.CreateUserWithTelegram(r.Context(), db.CreateUserWithTelegramParams{
			Name:       "Test User",
			Email:      sql.NullString{String: "dev@example.com", Valid: false},
			TelegramID: sql.NullString{String: tgID, Valid: true},
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Error creating user with Telegram: "+err.Error())
			return
		}
	} else if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error on database: "+err.Error())
		return
	}

	token, err := h.GenerateToken(user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating token: "+err.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(1 * time.Hour),
		HttpOnly: true,
		Secure:   h.AppTLS,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"user":  MapUserToResponse(user),
		"token": token,
	})

}

// HandleLogout logout user session
// @Summary      Close session
// @Description  Delete cookie on browser session sending an expired cookie
// @Tags         auth
// @Success      200  {object}  map[string]string
// @Router       /auth/logout [post]
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
		Secure:   false,
	})
	respondWithJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}
