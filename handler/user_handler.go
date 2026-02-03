package handler

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"sikupas/backend/config"
	"sikupas/backend/middleware"
	"sikupas/backend/model"
)

// ─── Register ────────────────────────────────────────────────────────────────

func Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return model.ErrorResponse(c, 400, "Format request body tidak valid")
	}

	// Validasi
	if errs := req.Validate(); len(errs) > 0 {
		return model.ErrorResponse(c, 400, "Validasi gagal", errs)
	}

	req.Nama = strings.TrimSpace(req.Nama)
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	// Default role = admin jika kosong
	if req.Role == "" {
		req.Role = "admin"
	}

	// Hash password
	hashedPwd, err := model.HashPassword(req.Password)
	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal hash password")
	}

	// Insert ke DB
	var user model.User
	err = config.DB.QueryRow(context.Background(),
		`INSERT INTO users (nama, username, password, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, nama, username, role, created_at, updated_at`,
		req.Nama, req.Username, hashedPwd, req.Role,
	).Scan(&user.ID, &user.Nama, &user.Username, &user.Role, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if config.IsUniqueViolation(err) {
			return model.ErrorResponse(c, 409, "Username sudah terdaftar")
		}
		return model.ErrorResponse(c, 500, "Gagal mendaftar user: "+err.Error())
	}

	return model.SuccessResponse(c, 201, "Registrasi berhasil", model.UserResponse{
		ID:       user.ID,
		Nama:     user.Nama,
		Username: user.Username,
		Role:     user.Role,
	})
}

// ─── Login ───────────────────────────────────────────────────────────────────

func Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return model.ErrorResponse(c, 400, "Format request body tidak valid")
	}

	if errs := req.Validate(); len(errs) > 0 {
		return model.ErrorResponse(c, 400, "Validasi gagal", errs)
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	// Cari user di DB
	var user model.User
	var hashedPwd string
	err := config.DB.QueryRow(context.Background(),
		`SELECT id, nama, username, password, role FROM users WHERE username = $1`,
		req.Username,
	).Scan(&user.ID, &user.Nama, &user.Username, &hashedPwd, &user.Role)

	if err != nil {
		return model.ErrorResponse(c, 401, "Username atau password salah")
	}

	// Cek password
	if err := model.ComparePassword(hashedPwd, req.Password); err != nil {
		return model.ErrorResponse(c, 401, "Username atau password salah")
	}

	// Generate token
	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal generate token")
	}

	return model.SuccessResponse(c, 200, "Login berhasil", model.LoginResponse{
		Token: token,
		User: model.UserResponse{
			ID:       user.ID,
			Nama:     user.Nama,
			Username: user.Username,
			Role:     user.Role,
		},
	})
}

// ─── Get Current User (dari token) ──────────────────────────────────────────

func GetCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)

	var user model.UserResponse
	err := config.DB.QueryRow(context.Background(),
		`SELECT id, nama, username, role FROM users WHERE id = $1`, userID,
	).Scan(&user.ID, &user.Nama, &user.Username, &user.Role)

	if err != nil {
		return model.ErrorResponse(c, 404, "User tidak ditemukan")
	}

	return model.SuccessResponse(c, 200, "Berhasil", user)
}
