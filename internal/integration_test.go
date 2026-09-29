package integration_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	db "github.com/rick-astral-cat/flizix-api/db/sqlc"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *db.Queries {
	t.Helper()

	schema, err := os.ReadFile("../db/schema.sql")
	if err != nil {
		t.Fatalf("could not read schema: %v", err)
	}

	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("could not open in-memory db: %v", err)
	}
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Errorf("could not close test db connection: %v", err)
		}
	})

	if _, err := conn.Exec(string(schema)); err != nil {
		t.Fatalf("could not apply schema: %v", err)
	}

	return db.New(conn)
}

func createTestUser(t *testing.T, q *db.Queries) db.User {
	t.Helper()
	user, err := q.CreateUserWithPasskey(context.Background(), db.CreateUserWithPasskeyParams{
		Name:      "Test User",
		Email:     sql.NullString{String: "test@example.com", Valid: true},
		PasskeyID: sql.NullString{String: "pk-test-123", Valid: true},
	})
	if err != nil {
		t.Fatalf("could not create test user: %v", err)
	}
	return user
}

func TestIntegration_AccountLifecycle(t *testing.T) {
	q := setupTestDB(t)
	ctx := context.Background()
	user := createTestUser(t, q)

	// Create account type
	accType, err := q.CreateAccountType(ctx, db.CreateAccountTypeParams{
		Name:     "checking",
		UserID:   sql.NullInt64{Int64: user.ID, Valid: true},
		IsSystem: 0,
	})
	if err != nil {
		t.Fatalf("CreateAccountType error: %v", err)
	}

	// Create account
	acc, err := q.CreateAccount(ctx, db.CreateAccountParams{
		Name:   "BBVA",
		Type:   sql.NullInt64{Int64: accType.ID, Valid: true},
		UserID: sql.NullInt64{Int64: user.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateAccount error: %v", err)
	}
	if acc.Name != "BBVA" {
		t.Errorf("expected name BBVA, got %s", acc.Name)
	}

	// List accounts
	accounts, err := q.ListAccountsByUser(ctx, sql.NullInt64{Int64: user.ID, Valid: true})
	if err != nil {
		t.Fatalf("ListAccountsByUser error: %v", err)
	}
	if len(accounts) != 1 {
		t.Errorf("expected 1 account, got %d", len(accounts))
	}

	// Get by ID
	found, err := q.GetAccountByID(ctx, db.GetAccountByIDParams{
		ID:     acc.ID,
		UserID: sql.NullInt64{Int64: user.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("GetAccountByID error: %v", err)
	}
	if found.ID != acc.ID {
		t.Errorf("expected ID %d, got %d", acc.ID, found.ID)
	}

	// Soft delete
	err = q.SoftDeleteAccount(ctx, db.SoftDeleteAccountParams{
		ID:     acc.ID,
		UserID: sql.NullInt64{Int64: user.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("SoftDeleteAccount error: %v", err)
	}

	// Should not appear in list after delete
	accounts, _ = q.ListAccountsByUser(ctx, sql.NullInt64{Int64: user.ID, Valid: true})
	if len(accounts) != 0 {
		t.Errorf("expected 0 accounts after soft delete, got %d", len(accounts))
	}
}

func TestIntegration_CardLifecycle(t *testing.T) {
	q := setupTestDB(t)
	ctx := context.Background()
	user := createTestUser(t, q)

	// Create credit card
	card, err := q.CreateCard(ctx, db.CreateCardParams{
		Name:        "Visa",
		Type:        "credit",
		CreditLimit: sql.NullInt64{Int64: 50000, Valid: true},
		CutoffDate:  sql.NullInt64{Int64: 15, Valid: true},
		AccountID:   sql.NullInt64{Valid: false},
		UserID:      sql.NullInt64{Int64: user.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateCard error: %v", err)
	}
	if card.Type != "credit" {
		t.Errorf("expected type credit, got %s", card.Type)
	}

	// Update card
	updated, err := q.UpdateCard(ctx, db.UpdateCardParams{
		Name:        "Visa Platinum",
		Type:        "credit",
		CreditLimit: sql.NullInt64{Int64: 100000, Valid: true},
		CutoffDate:  sql.NullInt64{Int64: 20, Valid: true},
		ID:          card.ID,
		UserID:      sql.NullInt64{Int64: user.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdateCard error: %v", err)
	}
	if updated.Name != "Visa Platinum" {
		t.Errorf("expected Visa Platinum, got %s", updated.Name)
	}

	// Soft delete
	err = q.SoftDeleteCard(ctx, db.SoftDeleteCardParams{
		ID:     card.ID,
		UserID: sql.NullInt64{Int64: user.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("SoftDeleteCard error: %v", err)
	}

	// Should not appear in list
	cards, _ := q.ListCardsByUser(ctx, sql.NullInt64{Int64: user.ID, Valid: true})
	if len(cards) != 0 {
		t.Errorf("expected 0 cards after soft delete, got %d", len(cards))
	}
}

func TestIntegration_AccountType_SystemSeed(t *testing.T) {
	q := setupTestDB(t)
	ctx := context.Background()

	// Seed system types
	for _, name := range []string{"account_type.checking", "account_type.cash"} {
		_, err := q.CreateAccountType(ctx, db.CreateAccountTypeParams{
			Name:     name,
			UserID:   sql.NullInt64{Valid: false},
			IsSystem: 1,
		})
		if err != nil {
			t.Fatalf("CreateAccountType %s error: %v", name, err)
		}
	}

	types, err := q.ListSystemAccountTypes(ctx)
	if err != nil {
		t.Fatalf("ListSystemAccountTypes error: %v", err)
	}
	if len(types) != 2 {
		t.Errorf("expected 2 system types, got %d", len(types))
	}
}

func TestIntegration_PurchaseLifecycle(t *testing.T) {
	q := setupTestDB(t)
	ctx := context.Background()
	user := createTestUser(t, q)

	// Create a credit card to use as payment source
	card, err := q.CreateCard(ctx, db.CreateCardParams{
		Name:        "Nu",
		Type:        "credit",
		CreditLimit: sql.NullInt64{Int64: 20000, Valid: true},
		CutoffDate:  sql.NullInt64{Int64: 10, Valid: true},
		AccountID:   sql.NullInt64{Valid: false},
		UserID:      sql.NullInt64{Int64: user.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateCard error: %v", err)
	}

	// Create a purchase paid with the card
	purchase, err := q.CreatePurchase(ctx, db.CreatePurchaseParams{
		Purchase:           "Groceries",
		Date:               "2026-09-23",
		Amount:             500,
		CardID:             sql.NullInt64{Int64: card.ID, Valid: true},
		AccountID:          sql.NullInt64{Valid: false},
		CarryOverNextMonth: sql.NullInt64{Int64: 0, Valid: true},
		Comment:            sql.NullString{Valid: false},
		ProjectID:          sql.NullInt64{Valid: false},
		UserID:             sql.NullInt64{Int64: user.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreatePurchase error: %v", err)
	}
	if purchase.Amount != 500 {
		t.Errorf("expected amount 500, got %d", purchase.Amount)
	}
	if purchase.Remaining.Int64 != 500 {
		t.Errorf("expected remaining 500, got %d", purchase.Remaining.Int64)
	}

	// List purchases — should return the one we just created
	purchases, err := q.ListPurchasesByUserId(ctx, sql.NullInt64{Int64: user.ID, Valid: true})
	if err != nil {
		t.Fatalf("ListPurchasesByUserId error: %v", err)
	}
	if len(purchases) != 1 {
		t.Fatalf("expected 1 purchase, got %d", len(purchases))
	}

	// Get by ID
	found, err := q.GetPurchaseById(ctx, db.GetPurchaseByIdParams{
		ID:     purchase.ID,
		UserID: sql.NullInt64{Int64: user.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("GetPurchaseById error: %v", err)
	}
	if found.Purchase != "Groceries" {
		t.Errorf("expected 'Groceries', got '%s'", found.Purchase)
	}

	// Update — register a partial payment
	updated, err := q.UpdatePurchase(ctx, db.UpdatePurchaseParams{
		ID:                 purchase.ID,
		Purchase:           purchase.Purchase,
		Date:               purchase.Date,
		Amount:             purchase.Amount,
		AmountPaid:         300,
		CarryOverNextMonth: sql.NullInt64{Int64: 0, Valid: true},
		Comment:            sql.NullString{Valid: false},
		CardID:             sql.NullInt64{Int64: card.ID, Valid: true},
		AccountID:          sql.NullInt64{Valid: false},
		ProjectID:          sql.NullInt64{Valid: false},
		UserID:             sql.NullInt64{Int64: user.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdatePurchase error: %v", err)
	}
	if updated.AmountPaid != 300 {
		t.Errorf("expected amount_paid 300, got %d", updated.AmountPaid)
	}
	if updated.Remaining.Int64 != 200 {
		t.Errorf("expected remaining 200, got %d", updated.Remaining.Int64)
	}
	if updated.Paid.Int64 != 0 {
		t.Error("expected paid=false when amount_paid < amount")
	}

	// Mark as fully paid
	fullyPaid, err := q.UpdatePurchase(ctx, db.UpdatePurchaseParams{
		ID:                 purchase.ID,
		Purchase:           purchase.Purchase,
		Date:               purchase.Date,
		Amount:             purchase.Amount,
		AmountPaid:         500,
		CarryOverNextMonth: sql.NullInt64{Int64: 0, Valid: true},
		Comment:            sql.NullString{Valid: false},
		CardID:             sql.NullInt64{Int64: card.ID, Valid: true},
		AccountID:          sql.NullInt64{Valid: false},
		ProjectID:          sql.NullInt64{Valid: false},
		UserID:             sql.NullInt64{Int64: user.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("UpdatePurchase (full payment) error: %v", err)
	}
	if fullyPaid.Paid.Int64 != 1 {
		t.Error("expected paid=true when amount_paid == amount")
	}

	// Soft delete
	err = q.SoftDeletePurchase(ctx, db.SoftDeletePurchaseParams{
		ID:     purchase.ID,
		UserID: sql.NullInt64{Int64: user.ID, Valid: true},
	})
	if err != nil {
		t.Fatalf("SoftDeletePurchase error: %v", err)
	}

	// Should not appear in list after delete
	purchases, _ = q.ListPurchasesByUserId(ctx, sql.NullInt64{Int64: user.ID, Valid: true})
	if len(purchases) != 0 {
		t.Errorf("expected 0 purchases after soft delete, got %d", len(purchases))
	}

	// Should not appear in get by ID after delete
	_, err = q.GetPurchaseById(ctx, db.GetPurchaseByIdParams{
		ID:     purchase.ID,
		UserID: sql.NullInt64{Int64: user.ID, Valid: true},
	})
	if err == nil {
		t.Error("expected error when getting soft-deleted purchase, got nil")
	}
}
