package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"sikupas/backend/model"
)

// ─── Simple JWT Implementation ───────────────────────────────────────────────
// Menggunakan HS256 signing manual tanpa dependency tambahan

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtPayload struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}

// GenerateToken membuat JWT token
func GenerateToken(userID int, username, role string) (string, error) {
	header := jwtHeader{Alg: "HS256", Typ: "JWT"}
	now := time.Now()
	payload := jwtPayload{
		UserID:   userID,
		Username: username,
		Role:     role,
		Iat:      now.Unix(),
		Exp:      now.Add(24 * time.Hour).Unix(), // Token berlaku 24 jam
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	signingInput := headerB64 + "." + payloadB64
	signature := signHS256(signingInput, model.GetJWTSecret())

	token := signingInput + "." + signature
	return token, nil
}

// ParseToken memvalidasi dan mengekstrak payload dari JWT token
func ParseToken(tokenString string) (*jwtPayload, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("format token tidak valid")
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSig := signHS256(signingInput, model.GetJWTSecret())

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, fmt.Errorf("signature token tidak valid")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("gagal decode payload")
	}

	var payload jwtPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return nil, fmt.Errorf("gagal parse payload")
	}

	// Cek expiration
	if time.Now().Unix() > payload.Exp {
		return nil, fmt.Errorf("token telah kedaluwarsa")
	}

	return &payload, nil
}

func signHS256(input, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// ─── Auth Middleware ─────────────────────────────────────────────────────────

// AuthRequired middleware – cek token ada dan valid
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return model.ErrorResponse(c, 401, "Token autentikasi diperlukan")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return model.ErrorResponse(c, 401, "Format Authorization harus 'Bearer <token>'")
		}

		payload, err := ParseToken(tokenString)
		if err != nil {
			return model.ErrorResponse(c, 401, err.Error())
		}

		// Simpan user info di context
		c.Locals("user_id", payload.UserID)
		c.Locals("username", payload.Username)
		c.Locals("role", payload.Role)

		return c.Next()
	}
}

// RoleRequired middleware – cek role tertentu diizinkan
func RoleRequired(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok {
			return model.ErrorResponse(c, 403, "Role tidak ditemukan")
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				return c.Next()
			}
		}

		return model.ErrorResponse(c, 403, "Anda tidak memiliki akses untuk fitur ini")
	}
}
