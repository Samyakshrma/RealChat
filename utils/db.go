package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	Db  *pgxpool.Pool
	Ctx = context.Background()
)

func InitDB() {
	dsn := os.Getenv("DATABASE_URL") // e.g. "postgres://user:pass@host:port/dbname"
	if dsn == "" {
		panic("DATABASE_URL environment variable not set")
	}

	var err error
	Db, err = pgxpool.New(Ctx, dsn)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Postgres: %v", err))
	}

	err = Db.Ping(Ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to ping Postgres: %v", err))
	}

	fmt.Println("Connected to Postgres successfully")
}
