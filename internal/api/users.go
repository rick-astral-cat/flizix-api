package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	db "github.com/rick-astral-cat/flizix-api/db/sqlc"
)

type Config struct {
	Queries          *db.Queries
	JWTSecret        string
	TelegramBotToken string
	EnableCORS       bool
	AllowedOrigins   []string
}

type CreateUserRequest struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	PasskeyID string `json:"passkey_id"`
}

type UserResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
}

func MapUserToResponse(user db.User) UserResponse {
	return UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email.String,
	}
}

// HandleCreateUser godoc
// @Summary      Create new user
// @Description  Register new user on DB with name, email and passkey ID.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        user  body      CreateUserRequest  true  "User data"
// @Success      201   {object}	 UserResponse
// @Failure      400   {string}  string  "Invalid JSON"
// @Failure      500   {string}  string  "Internal Server Error"
// @Router       /users [post]
func (api *Config) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	passkey := sql.NullString{
		String: req.PasskeyID,
		Valid:  req.PasskeyID != "",
	}

	email := sql.NullString{
		String: req.Email,
		Valid:  true,
	}

	user, err := api.Queries.CreateUserWithPasskey(r.Context(), db.CreateUserWithPasskeyParams{
		Name:      req.Name,
		Email:     email,
		PasskeyID: passkey,
	})

	if err != nil {
		http.Error(w, "Error at creating user : "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := MapUserToResponse(user)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// HandleGetProfile return authenticated user data
// @Summary      Get profile
// @Description  Get userdata using JWT on cookie.
// @Tags         users
// @Security     BearerAuth
// @Success      200  {object}  UserResponse
// @Failure      401  {string}  string "Not authorized"
// @Router       /me [get]
func (api *Config) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		http.Error(w, "No user ID found in context", http.StatusInternalServerError)
		return
	}
	user, err := api.Queries.GetUserById(r.Context(), userID)
	if err != nil {
		http.Error(w, "Error at getting user, not found : "+err.Error(), http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MapUserToResponse(user))
}
