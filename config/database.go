package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

// ConnectDB menghubungkan ke database Supabase PostgreSQL
func ConnectDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")

		if host == "" || port == "" || user == "" || password == "" || dbname == "" {
			log.Fatal("❌ Database environment variables belum lengkap. Cek .env anda.")
		}
		connStr = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=require", user, password, host, port, dbname)
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("❌ Gagal parse database config: %v", err)
	}

	// Pool settings
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("❌ Gagal membuat connection pool: %v", err)
	}

	// Ping test
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("❌ Gagal ping database: %v", err)
	}

	DB = pool
	log.Println("✅ Berhasil terhubung ke Supabase PostgreSQL")
}

// CloseDB menutup connection pool
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("📌 Database connection pool ditutup")
	}
}

// IsUniqueViolation cek apakah error adalah duplicate key
func IsUniqueViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23505"
	}
	return false
}
