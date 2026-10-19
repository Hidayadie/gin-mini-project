package database

import (
	"context"
	//"database/sql"
	"errors"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm/logger"
)


func Connect() (*gorm.DB, error) {
  if err := godotenv.Load(); err != nil {
    log.Fatal("Error loading .env file")
  }

  dsn := os.Getenv("DATABASE_URL")
  if dsn == "" {
    return nil, errors.New("environment variable DATABASE_URL belum diatur")
  }

  db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
Logger:               logger.Default.LogMode(logger.Info), 
DisableAutomaticPing: true, // ping dilakukan manual di bawah, dengan batas waktu     
	}) 
  if err != nil {
    return nil, err
  }

// Ping untuk memastikan database benar-benar bisa dihubungi
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}     
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()         
		return nil, err     
	}     
	return db, nil 

 }
