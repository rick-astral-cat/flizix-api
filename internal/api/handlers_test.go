package api

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"

	db "github.com/rick-astral-cat/flizix-api/db/sqlc"
)

// mockQuerier implements db.Querier for testing without a real database.
type mockQuerier struct {
	createAccount               func(ctx context.Context, arg db.CreateAccountParams) (db.Account, error)
	createAccountType           func(ctx context.Context, arg db.CreateAccountTypeParams) (db.AccountType, error)
	createCard                  func(ctx context.Context, arg db.CreateCardParams) (db.Card, error)
	createPurchase              func(ctx context.Context, arg db.CreatePurchaseParams) (db.Purchase, error)
	createUserWithPasskey       func(ctx context.Context, arg db.CreateUserWithPasskeyParams) (db.User, error)
	createUserWithTelegram      func(ctx context.Context, arg db.CreateUserWithTelegramParams) (db.User, error)
	getAccountByID              func(ctx context.Context, arg db.GetAccountByIDParams) (db.Account, error)
	getAccountTypeByUser        func(ctx context.Context, arg db.GetAccountTypeByUserParams) (db.AccountType, error)
	getCardByID                 func(ctx context.Context, arg db.GetCardByIDParams) (db.Card, error)
	getPurchaseById             func(ctx context.Context, arg db.GetPurchaseByIdParams) (db.Purchase, error)
	getUserByEmail              func(ctx context.Context, email sql.NullString) (db.User, error)
	getUserById                 func(ctx context.Context, id int64) (db.User, error)
	getUserByPassKey            func(ctx context.Context, passkeyID sql.NullString) (db.User, error)
	getUserByTelegramId         func(ctx context.Context, telegramID sql.NullString) (db.User, error)
	listAccountTypesByUser      func(ctx context.Context, userID sql.NullInt64) ([]db.AccountType, error)
	listAccountsByUser          func(ctx context.Context, userID sql.NullInt64) ([]db.Account, error)
	listCardsByUser             func(ctx context.Context, userID sql.NullInt64) ([]db.Card, error)
	listPurchasesByUserId       func(ctx context.Context, userID sql.NullInt64) ([]db.Purchase, error)
	listSystemAccountTypes      func(ctx context.Context) ([]db.AccountType, error)
	softDeleteAccount           func(ctx context.Context, arg db.SoftDeleteAccountParams) error
	softDeleteAccountTypeByUser func(ctx context.Context, arg db.SoftDeleteAccountTypeByUserParams) error
	softDeleteCard              func(ctx context.Context, arg db.SoftDeleteCardParams) error
	softDeletePurchase          func(ctx context.Context, arg db.SoftDeletePurchaseParams) error
	softDeleteUser              func(ctx context.Context, id int64) error
	updateAccount               func(ctx context.Context, arg db.UpdateAccountParams) (db.Account, error)
	updateCard                  func(ctx context.Context, arg db.UpdateCardParams) (db.Card, error)
	updatePurchase              func(ctx context.Context, arg db.UpdatePurchaseParams) (db.Purchase, error)
	updateUserPasskey           func(ctx context.Context, arg db.UpdateUserPasskeyParams) error
}

func (m *mockQuerier) CreateAccount(ctx context.Context, arg db.CreateAccountParams) (db.Account, error) {
	return m.createAccount(ctx, arg)
}
func (m *mockQuerier) CreateAccountType(ctx context.Context, arg db.CreateAccountTypeParams) (db.AccountType, error) {
	return m.createAccountType(ctx, arg)
}
func (m *mockQuerier) CreateCard(ctx context.Context, arg db.CreateCardParams) (db.Card, error) {
	return m.createCard(ctx, arg)
}
func (m *mockQuerier) CreateUserWithPasskey(ctx context.Context, arg db.CreateUserWithPasskeyParams) (db.User, error) {
	return m.createUserWithPasskey(ctx, arg)
}
func (m *mockQuerier) CreateUserWithTelegram(ctx context.Context, arg db.CreateUserWithTelegramParams) (db.User, error) {
	return m.createUserWithTelegram(ctx, arg)
}
func (m *mockQuerier) GetAccountByID(ctx context.Context, arg db.GetAccountByIDParams) (db.Account, error) {
	return m.getAccountByID(ctx, arg)
}
func (m *mockQuerier) GetAccountTypeByUser(ctx context.Context, arg db.GetAccountTypeByUserParams) (db.AccountType, error) {
	return m.getAccountTypeByUser(ctx, arg)
}
func (m *mockQuerier) GetCardByID(ctx context.Context, arg db.GetCardByIDParams) (db.Card, error) {
	return m.getCardByID(ctx, arg)
}
func (m *mockQuerier) GetUserByEmail(ctx context.Context, email sql.NullString) (db.User, error) {
	return m.getUserByEmail(ctx, email)
}
func (m *mockQuerier) GetUserById(ctx context.Context, id int64) (db.User, error) {
	return m.getUserById(ctx, id)
}
func (m *mockQuerier) GetUserByPassKey(ctx context.Context, passkeyID sql.NullString) (db.User, error) {
	return m.getUserByPassKey(ctx, passkeyID)
}
func (m *mockQuerier) GetUserByTelegramId(ctx context.Context, telegramID sql.NullString) (db.User, error) {
	return m.getUserByTelegramId(ctx, telegramID)
}
func (m *mockQuerier) ListAccountTypesByUser(ctx context.Context, userID sql.NullInt64) ([]db.AccountType, error) {
	return m.listAccountTypesByUser(ctx, userID)
}
func (m *mockQuerier) ListAccountsByUser(ctx context.Context, userID sql.NullInt64) ([]db.Account, error) {
	return m.listAccountsByUser(ctx, userID)
}
func (m *mockQuerier) ListCardsByUser(ctx context.Context, userID sql.NullInt64) ([]db.Card, error) {
	return m.listCardsByUser(ctx, userID)
}
func (m *mockQuerier) ListSystemAccountTypes(ctx context.Context) ([]db.AccountType, error) {
	return m.listSystemAccountTypes(ctx)
}
func (m *mockQuerier) SoftDeleteAccount(ctx context.Context, arg db.SoftDeleteAccountParams) error {
	return m.softDeleteAccount(ctx, arg)
}
func (m *mockQuerier) SoftDeleteAccountTypeByUser(ctx context.Context, arg db.SoftDeleteAccountTypeByUserParams) error {
	return m.softDeleteAccountTypeByUser(ctx, arg)
}
func (m *mockQuerier) SoftDeleteCard(ctx context.Context, arg db.SoftDeleteCardParams) error {
	return m.softDeleteCard(ctx, arg)
}
func (m *mockQuerier) SoftDeleteUser(ctx context.Context, id int64) error {
	return m.softDeleteUser(ctx, id)
}
func (m *mockQuerier) UpdateAccount(ctx context.Context, arg db.UpdateAccountParams) (db.Account, error) {
	return m.updateAccount(ctx, arg)
}
func (m *mockQuerier) UpdateCard(ctx context.Context, arg db.UpdateCardParams) (db.Card, error) {
	return m.updateCard(ctx, arg)
}
func (m *mockQuerier) UpdateUserPasskey(ctx context.Context, arg db.UpdateUserPasskeyParams) error {
	return m.updateUserPasskey(ctx, arg)
}
func (m *mockQuerier) CreatePurchase(ctx context.Context, arg db.CreatePurchaseParams) (db.Purchase, error) {
	return m.createPurchase(ctx, arg)
}
func (m *mockQuerier) GetPurchaseById(ctx context.Context, arg db.GetPurchaseByIdParams) (db.Purchase, error) {
	return m.getPurchaseById(ctx, arg)
}
func (m *mockQuerier) ListPurchasesByUserId(ctx context.Context, userID sql.NullInt64) ([]db.Purchase, error) {
	return m.listPurchasesByUserId(ctx, userID)
}
func (m *mockQuerier) SoftDeletePurchase(ctx context.Context, arg db.SoftDeletePurchaseParams) error {
	return m.softDeletePurchase(ctx, arg)
}
func (m *mockQuerier) UpdatePurchase(ctx context.Context, arg db.UpdatePurchaseParams) (db.Purchase, error) {
	return m.updatePurchase(ctx, arg)
}

const testJWTSecret = "test-secret-key"

// authenticatedRequest injects a valid JWT cookie into the request.
func authenticatedRequest(r *http.Request, userID int64) *http.Request {
	auth := &AuthHandler{JWTSecret: testJWTSecret}
	token, _ := auth.GenerateToken(userID)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	return r
}

func executeRequest(handler http.Handler, r *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, r)
	return rr
}
