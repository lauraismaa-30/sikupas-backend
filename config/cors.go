package config

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSConfig mengembalikan middleware CORS yang dikonfigurasi
func CORSConfig() fiber.Handler {
	origins := os.Getenv("CORS_ORIGINS")
	if origins == "" {
		origins = "http://localhost:3000,http://localhost:8080,http://127.0.0.1:8080,http://127.0.0.1:5500"
	}

	allowOrigins := strings.Split(origins, ",")
	allowOriginsStr := strings.Join(allowOrigins, ",")

	return cors.New(cors.Config{
		AllowOrigins:     allowOriginsStr,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type,Authorization,X-Requested-With",
		ExposeHeaders:    "Content-Length,Content-Type",
		AllowCredentials: true,
		MaxAge:           86400, // 24 jam
	})
}
