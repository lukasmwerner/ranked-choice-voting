package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/osu-acm/acm-votes/database"
)

func GetSession(db *sql.DB, ctx context.Context, token string) (string, error) {
	queries := database.New(db)
	// SQL query handles ignoring out-of-date sessions
	return queries.GetSession(ctx, token)
}

func CreateSession(db *sql.DB, ctx context.Context, email string) (string, error) {
	queries := database.New(db)

	uid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	err = queries.CreateSession(ctx, database.CreateSessionParams{
		ID:        uid.String(),
		User:      email,
		ExpiresAt: time.Now().Add(SessionExpiry).Unix(),
	})

	return uid.String(), err
}
