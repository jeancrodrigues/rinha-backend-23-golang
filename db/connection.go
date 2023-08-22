package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
)

type PgxIface interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	Close()
}

var (
	db *pgxpool.Pool
)

func GetConnection() PgxIface {

	if db != nil {
		return db
	}

	connectionString := os.Getenv("DATABASE_URL")

	if connectionString == "" {
		connectionString = "postgres://pessoa:pessoa@localhost:5432/pessoa"
	}

	var err error

	db, err = pgxpool.New(context.Background(), connectionString)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	return db
}
