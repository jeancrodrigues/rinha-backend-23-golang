package db

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"strconv"
	"time"
)

type PgxIface interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	Close()
}

var (
	Conn *pgxpool.Pool
)

func GetConnection() PgxIface {

	if Conn != nil {
		//log.Println("getting existent connection")
		return Conn
	} else {
		log.Println("opening connections")
	}

	connectionString := os.Getenv("DATABASE_URL")

	if connectionString == "" {
		connectionString = "postgres://pessoa:pessoa@localhost:5433/pessoa"
	}

	var err error

	maxConnectionsS := os.Getenv("MAX_CONNECTIONS")
	if maxConnectionsS == "" {
		maxConnectionsS = "50"
	}

	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		log.Fatal(err)
	}

	maxConnections, _ := strconv.Atoi(maxConnectionsS)

	config.MaxConns = int32(maxConnections)
	config.MinConns = int32(maxConnections)

	config.MaxConnIdleTime = time.Minute * 3

	Conn, err = pgxpool.NewWithConfig(context.Background(), config)

	if err != nil {
		log.Fatal(err)
	}

	err = Conn.Ping(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	return Conn
}
