package main

import (
	"fmt"
	"log"
	"os"

	"sikupas/backend/config"
	"sikupas/backend/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	// ─── Load .env ─────────────────────────────────────────────────────
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  File .env tidak ditemukan, menggunakan environment variabel sistem")
	}

	// ─── Koneksi Database ──────────────────────────────────────────────
	config.ConnectDB()
	defer config.CloseDB()

	// ─── Jalankan Migrasi ──────────────────────────────────────────────
	config.RunMigrations()

	// ─── Inisialisasi Fiber ────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:   "SIKUPAS - Sistem Informasi Kunjungan Pasien",
		BodyLimit: 10 * 1024 * 1024, // 10 MB
		Views:     nil,
	})

	// Logger middleware
	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} ${latency} ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
	}))

	// CORS middleware
	app.Use(config.CORSConfig())

	// ─── Setup Routes ──────────────────────────────────────────────────
	route.SetupRoutes(app)

	// ─── Health Check ──────────────────────────────────────────────────
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "SIKUPAS API",
			"version": "1.0.0",
		})
	})

	// ─── 404 Handler ───────────────────────────────────────────────────
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"status":  false,
			"message": "Endpoint tidak ditemukan",
		})
	})

	// ─── Start Server ──────────────────────────────────────────────────
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║   SIKUPAS - Sistem Informasi Kunjungan Pasien   ║")
	fmt.Println("║         Puskesmas Sukasari  v1.0.0               ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Printf("🚀 Server berjalan di: http://localhost:%s\n", port)
	fmt.Printf("📌 Environment: %s\n", env)
	fmt.Printf("📌 API Base: http://localhost:%s/api\n", port)
	fmt.Println("─────────────────────────────────────────────────────")

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ Gagal memulai server: %v", err)
	}
}
