package api

import (
	"net/http"

	_ "github.com/rick-astral-cat/flizix-api/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func RegisterRoutes(
	mux *http.ServeMux,
	env string,
	userH *UserHandler,
	authH *AuthHandler,
	midH *MiddlewareHandler,
	cardH *CardHandler,
	accH *AccountHandler,
) {
	mux.HandleFunc("POST /users", userH.HandleCreateUser)
	mux.HandleFunc("GET /health", HandleHealth)
	mux.HandleFunc("POST /auth/telegram", authH.HandleTelegramLogin)
	mux.HandleFunc("POST /auth/logout", authH.HandleLogout)

	if env == "development" {
		mux.Handle("GET /swagger/", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))
		mux.HandleFunc("POST /auth/dev-login", authH.HandleDevLogin)
	}

	// Private Routes
	profileHandler := http.HandlerFunc(userH.HandleGetProfile)
	cardsHandler := http.HandlerFunc(cardH.HandleCreateCard)

	mux.Handle("GET /me", midH.JWTMiddleware(profileHandler))
	mux.Handle("POST /cards", midH.JWTMiddleware(cardsHandler))
	mux.Handle("GET /cards", midH.JWTMiddleware(http.HandlerFunc(cardH.HandleListCards)))
	mux.Handle("POST /accounts", midH.JWTMiddleware(http.HandlerFunc(accH.HandleCreateAccount)))
	mux.Handle("GET /accounts", midH.JWTMiddleware(http.HandlerFunc(accH.HandleListAccounts)))
	mux.Handle("DELETE /accounts/{id}", midH.JWTMiddleware(http.HandlerFunc(accH.HandleDeleteAccount)))
}
