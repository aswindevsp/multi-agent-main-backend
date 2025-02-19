package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

func NewConnection() (*pgx.Conn, error) {
	connectionUrl := "postgres://postgres:example@localhost:5432/postgres"
	conn, err := pgx.Connect(context.Background(), connectionUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	return conn, nil
}
