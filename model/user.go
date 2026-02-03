package model

import (
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ─── User Model ──────────────────────────────────────────────────────────────

type User struct {
	ID        int       `json:"id"`
	Nama      string    `json:"nama"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // tidak dikembalikan ke response
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ─── Request DTO ─────────────────────────────────────────────────────────────

type RegisterRequest struct {
	Nama     string `json:"nama"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ─── Response DTO ───────────────────────────────────────────────────────────

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID       int    `json:"id"`
	Nama     string `json:"nama"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

// ─── Validation ──────────────────────────────────────────────────────────────

func (r *RegisterRequest) Validate() []string {
	var errs []string

	nama := strings.TrimSpace(r.Nama)
	username := strings.TrimSpace(r.Username)
	password := strings.TrimSpace(r.Password)

	if nama == "" {
		errs = append(errs, "Nama tidak boleh kosong")
	} else if len(nama) < 2 || len(nama) > 100 {
		errs = append(errs, "Nama harus antara 2-100 karakter")
	}

	if username == "" {
		errs = append(errs, "Username tidak boleh kosong")
	} else if len(username) < 3 || len(username) > 50 {
		errs = append(errs, "Username harus antara 3-50 karakter")
	} else if !isAlphanumeric(username) {
		errs = append(errs, "Username hanya boleh berisi huruf dan angka")
	}

	if password == "" {
		errs = append(errs, "Password tidak boleh kosong")
	} else if len(password) < 6 {
		errs = append(errs, "Password minimal 6 karakter")
	}

	role := strings.TrimSpace(r.Role)
	if role != "" && role != "admin" && role != "kepala_puskesmas" {
		errs = append(errs, "Role harus 'admin' atau 'kepala_puskesmas'")
	}

	return errs
}

func (r *LoginRequest) Validate() []string {
	var errs []string
	if strings.TrimSpace(r.Username) == "" {
		errs = append(errs, "Username tidak boleh kosong")
	}
	if strings.TrimSpace(r.Password) == "" {
		errs = append(errs, "Password tidak boleh kosong")
	}
	return errs
}

// ─── Password Helpers ────────────────────────────────────────────────────────

func HashPassword(plain string) (string, error) {
	cost := bcrypt.DefaultCost // 10
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	return string(hashed), err
}

func ComparePassword(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}

// ─── Simple Token (Base64-encoded JSON) ──────────────────────────────────────
// Untuk production gunakan JWT library yang proper. Ini simplified token.

func GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "sikupas_default_secret_key_2024"
	}
	return secret
}

// isAlphanumeric cek string hanya huruf dan angka
func isAlphanumeric(s string) bool {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
