package database


import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
	"time"
	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect() (*sql.DB, error) {
  if err := godotenv.Load(); err != nil {
    log.Fatal("Error loading .env file")
  }

  dsn := os.Getenv("DATABASE_URL")
  if dsn == "" {
    return nil, errors.New("environment variable DATABASE_URL belum diatur")
  }

  var err error
  db, err := sql.Open("pgx", dsn)
  if err != nil {
    return nil, err
  }

  ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
  defer cancel()

  if err := db.PingContext(ctx); err != nil {
    db.Close()
    return nil, err
  }

  return db, nil
}
