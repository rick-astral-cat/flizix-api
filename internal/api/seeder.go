package api

import (
	"context"
	"database/sql"
	"log"

	db "github.com/rick-astral-cat/flizix-api/db/sqlc"
)

func SeedDefaultAccountTypes(ctx context.Context, queries *db.Queries) error {
	defaults := []string{"account_type.checking", "account_type.cash"}

	accTypes, err := queries.ListSystemAccountTypes(ctx)
	if err != nil {
		return err
	}

	existingMap := make(map[string]struct{})
	for _, t := range accTypes {
		existingMap[t.Name] = struct{}{}
	}

	for _, name := range defaults {
		_, exists := existingMap[name]
		if !exists {
			_, err := queries.CreateAccountType(ctx, db.CreateAccountTypeParams{
				Name:     name,
				UserID:   sql.NullInt64{Valid: false},
				IsSystem: 1,
			})
			if err != nil {
				log.Printf("error at creating account type (%s) on seeding process: %v\n", name, err)
				return err
			}
		}
	}

	return nil
}
