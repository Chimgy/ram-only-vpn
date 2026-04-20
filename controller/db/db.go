package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func Connect() (*pgx.Conn, error) {
	url := os.Getenv("DB_URL") // for now have a .env var that i grab for the db
	if url == "" {
		return nil, fmt.Errorf("DB_URL not set")
	}

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		return nil, fmt.Errorf("connecting to db %w", err)

	}

	return conn, nil
}

func ValidateUser(conn *pgx.Conn, userID string) (bool, error) {

	var validUntil time.Time

	err := conn.QueryRow(
		context.Background(),
		"SELECT valid_until FROM users WHERE user_id = $1",
		userID,
	).Scan(&validUntil)

	if err == pgx.ErrNoRows {
		// user dont exist
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("Querying user: %w", err)
	}

	return validUntil.After(time.Now()), nil
}
