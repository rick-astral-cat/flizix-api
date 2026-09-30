package api

import (
	"context"
	"database/sql"
	"errors"
	"log"

	db "github.com/rick-astral-cat/flizix-api/db/sqlc"
)

// devUser holds the fixed credentials for the development user.
// These are intentionally fake values used only in the development environment.
const (
	devUserEmail   = "dev@flizix.local"
	devUserPasskey = "dev-passkey-flizix-001"
	devUserName    = "Dev User"
)

// SeedDevData inserts a predictable set of data into the development database
// so that all endpoints can be exercised via Swagger or manual testing without
// any prior manual setup. All insertions are idempotent: running the server
// multiple times will not create duplicates.
func SeedDevData(ctx context.Context, queries db.Querier, telegramID string) error {
	log.Println("[dev-seeder] Starting development data seeding...")

	user, err := ensureDevUser(ctx, queries, telegramID)
	if err != nil {
		return err
	}

	accountTypes, err := ensureDevAccountTypes(ctx, queries, user.ID)
	if err != nil {
		return err
	}

	accounts, err := ensureDevAccounts(ctx, queries, user.ID, accountTypes)
	if err != nil {
		return err
	}

	cards, err := ensureDevCards(ctx, queries, user.ID, accounts)
	if err != nil {
		return err
	}

	if err := ensureDevPurchases(ctx, queries, user.ID, accounts, cards); err != nil {
		return err
	}

	log.Println("[dev-seeder] Development data seeding complete.")
	return nil
}

// ensureDevUser returns the dev user, creating it if it does not exist yet.
// If a telegramID is provided the user is created with it linked directly,
// so the normal Telegram login flow works without any manual setup.
// If no telegramID is configured, the user is created with a passkey only
// and can still be used via the dev-login endpoint in Swagger.
func ensureDevUser(ctx context.Context, queries db.Querier, telegramID string) (db.User, error) {
	user, err := queries.GetUserByEmail(ctx, sql.NullString{String: devUserEmail, Valid: true})
	if err == nil {
		log.Printf("[dev-seeder] Dev user already exists (id=%d), skipping.\n", user.ID)
		return user, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return db.User{}, err
	}

	if telegramID != "" {
		user, err = queries.CreateUserWithTelegram(ctx, db.CreateUserWithTelegramParams{
			Name:       devUserName,
			Email:      sql.NullString{String: devUserEmail, Valid: true},
			TelegramID: sql.NullString{String: telegramID, Valid: true},
		})
	} else {
		log.Println("[dev-seeder] DEV_TELEGRAM_ID not set, creating dev user with passkey only.")
		user, err = queries.CreateUserWithPasskey(ctx, db.CreateUserWithPasskeyParams{
			Name:      devUserName,
			Email:     sql.NullString{String: devUserEmail, Valid: true},
			PasskeyID: sql.NullString{String: devUserPasskey, Valid: true},
		})
	}
	if err != nil {
		return db.User{}, err
	}

	log.Printf("[dev-seeder] Created dev user (id=%d).\n", user.ID)
	return user, nil
}

// devAccountTypeNames maps a short key to the display name stored in the DB.
// System types (account_type.checking, account_type.cash) are seeded by
// SeedDefaultAccountTypes, so here we only add user-level custom types.
var devAccountTypeNames = map[string]string{
	"bank":    "Banco",
	"payroll": "Nómina",
}

// ensureDevAccountTypes creates the custom account types for the dev user if
// they do not already exist. Returns a map[key]AccountType for use downstream.
func ensureDevAccountTypes(ctx context.Context, queries db.Querier, userID int64) (map[string]db.AccountType, error) {
	result := make(map[string]db.AccountType, len(devAccountTypeNames))

	existing, err := queries.ListAccountTypesByUser(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		return nil, err
	}

	existingMap := make(map[string]db.AccountType, len(existing))
	for _, t := range existing {
		existingMap[t.Name] = t
	}

	for key, name := range devAccountTypeNames {
		if t, found := existingMap[name]; found {
			result[key] = t
			continue
		}
		t, err := queries.CreateAccountType(ctx, db.CreateAccountTypeParams{
			Name:     name,
			UserID:   sql.NullInt64{Int64: userID, Valid: true},
			IsSystem: 0,
		})
		if err != nil {
			return nil, err
		}
		log.Printf("[dev-seeder] Created account type '%s'.\n", name)
		result[key] = t
	}

	return result, nil
}

// ensureDevAccounts creates the dev bank/cash accounts if they do not exist.
// Returns a map[name]Account for use downstream.
func ensureDevAccounts(ctx context.Context, queries db.Querier, userID int64, accountTypes map[string]db.AccountType) (map[string]db.Account, error) {
	result := make(map[string]db.Account)

	existing, err := queries.ListAccountsByUser(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		return nil, err
	}

	existingMap := make(map[string]db.Account, len(existing))
	for _, a := range existing {
		existingMap[a.Name] = a
	}

	wanted := []struct {
		name    string
		typeKey string
	}{
		{"Santander", "bank"},
		{"BBVA Nómina", "payroll"},
		{"Efectivo", "bank"},
	}

	for _, w := range wanted {
		if a, found := existingMap[w.name]; found {
			result[w.name] = a
			continue
		}
		a, err := queries.CreateAccount(ctx, db.CreateAccountParams{
			Name:   w.name,
			Type:   sql.NullInt64{Int64: accountTypes[w.typeKey].ID, Valid: true},
			UserID: sql.NullInt64{Int64: userID, Valid: true},
		})
		if err != nil {
			return nil, err
		}
		log.Printf("[dev-seeder] Created account '%s'.\n", w.name)
		result[w.name] = a
	}

	return result, nil
}

// ensureDevCards creates the dev credit/debit cards if they do not exist.
// Returns a map[name]Card for use downstream.
func ensureDevCards(ctx context.Context, queries db.Querier, userID int64, accounts map[string]db.Account) (map[string]db.Card, error) {
	result := make(map[string]db.Card)

	existing, err := queries.ListCardsByUser(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		return nil, err
	}

	existingMap := make(map[string]db.Card, len(existing))
	for _, c := range existing {
		existingMap[c.Name] = c
	}

	wanted := []struct {
		name        string
		cardType    string
		creditLimit *int64
		cutoffDate  *int64
		accountName string
	}{
		{
			name:        "Nu",
			cardType:    "credit",
			creditLimit: int64Ptr(20000),
			cutoffDate:  int64Ptr(10),
		},
		{
			name:        "Rappicard",
			cardType:    "credit",
			creditLimit: int64Ptr(8000),
			cutoffDate:  int64Ptr(25),
		},
		{
			name:        "Santander Débito",
			cardType:    "debit",
			accountName: "Santander",
		},
	}

	for _, w := range wanted {
		if c, found := existingMap[w.name]; found {
			result[w.name] = c
			continue
		}

		params := db.CreateCardParams{
			Name:   w.name,
			Type:   w.cardType,
			UserID: sql.NullInt64{Int64: userID, Valid: true},
		}

		if w.creditLimit != nil {
			params.CreditLimit = sql.NullInt64{Int64: *w.creditLimit, Valid: true}
			params.CutoffDate = sql.NullInt64{Int64: *w.cutoffDate, Valid: true}
		}
		if w.accountName != "" {
			params.AccountID = sql.NullInt64{Int64: accounts[w.accountName].ID, Valid: true}
		}

		c, err := queries.CreateCard(ctx, params)
		if err != nil {
			return nil, err
		}
		log.Printf("[dev-seeder] Created card '%s'.\n", w.name)
		result[w.name] = c
	}

	return result, nil
}

// ensureDevPurchases creates a varied set of purchases if none exist yet for
// the dev user. We check once at the list level to keep the logic simple —
// if any purchase exists we assume the set was already seeded.
func ensureDevPurchases(ctx context.Context, queries db.Querier, userID int64, accounts map[string]db.Account, cards map[string]db.Card) error {
	existing, err := queries.ListPurchasesByUserId(ctx, sql.NullInt64{Int64: userID, Valid: true})
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		log.Printf("[dev-seeder] Purchases already exist (%d found), skipping.\n", len(existing))
		return nil
	}

	type purchaseSeed struct {
		name        string
		date        string
		amount      int64
		amountPaid  int64
		cardName    string
		accountName string
		comment     string
		carryOver   bool
	}

	seeds := []purchaseSeed{
		// Fully paid with credit card
		{name: "Netflix", date: "2026-09-01", amount: 199, amountPaid: 199, cardName: "Nu", comment: "Suscripción mensual"},
		// Partially paid with credit card
		{name: "Laptop Stand", date: "2026-09-05", amount: 850, amountPaid: 400, cardName: "Rappicard", comment: "Accesorio de trabajo"},
		// Unpaid, carry over to next month
		{name: "Dentista", date: "2026-09-10", amount: 1200, cardName: "Nu", carryOver: true},
		// Paid with bank account (transfer)
		{name: "Renta", date: "2026-09-01", amount: 7500, amountPaid: 7500, accountName: "Santander", comment: "Pago de renta mensual"},
		// Partial cash payment
		{name: "Supermercado", date: "2026-09-15", amount: 650, amountPaid: 650, accountName: "Efectivo"},
		// Recent unpaid purchase
		{name: "Curso de Go", date: "2026-09-20", amount: 499, cardName: "Nu", comment: "Udemy"},
		// Small debit card purchase
		{name: "Gasolina", date: "2026-09-22", amount: 300, amountPaid: 300, cardName: "Santander Débito"},
	}

	for _, s := range seeds {
		params := db.CreatePurchaseParams{
			Purchase: s.name,
			Date:     s.date,
			Amount:   s.amount,
			UserID:   sql.NullInt64{Int64: userID, Valid: true},
		}

		if s.comment != "" {
			params.Comment = sql.NullString{String: s.comment, Valid: true}
		}
		if s.carryOver {
			params.CarryOverNextMonth = sql.NullInt64{Int64: 1, Valid: true}
		}
		if s.cardName != "" {
			params.CardID = sql.NullInt64{Int64: cards[s.cardName].ID, Valid: true}
		} else {
			params.AccountID = sql.NullInt64{Int64: accounts[s.accountName].ID, Valid: true}
		}

		purchase, err := queries.CreatePurchase(ctx, params)
		if err != nil {
			return err
		}

		// Register amount_paid if the purchase has one
		if s.amountPaid > 0 {
			_, err = queries.UpdatePurchase(ctx, db.UpdatePurchaseParams{
				ID:                 purchase.ID,
				Purchase:           purchase.Purchase,
				Date:               purchase.Date,
				Amount:             purchase.Amount,
				AmountPaid:         s.amountPaid,
				CarryOverNextMonth: purchase.CarryOverNextMonth,
				Comment:            purchase.Comment,
				CardID:             purchase.CardID,
				AccountID:          purchase.AccountID,
				ProjectID:          purchase.ProjectID,
				UserID:             sql.NullInt64{Int64: userID, Valid: true},
			})
			if err != nil {
				return err
			}
		}

		log.Printf("[dev-seeder] Created purchase '%s'.\n", s.name)
	}

	return nil
}

// int64Ptr is a helper to get a pointer to an int64 literal.
func int64Ptr(v int64) *int64 {
	return &v
}
