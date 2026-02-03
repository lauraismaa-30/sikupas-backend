package handler

import (
	"context"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"sikupas/backend/config"
	"sikupas/backend/model"
)

// ─── GET /pasien (all with pagination) ───────────────────────────────────────

func GetAllPasien(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	search := strings.TrimSpace(c.Query("search", ""))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}
	offset := (page - 1) * perPage

	var totalData int
	var rows []model.PasienResponse

	if search != "" {
		// Count
		config.DB.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM pasien WHERE nik LIKE '%'||$1||'%' OR nama_pasien ILIKE '%'||$1||'%'`, search,
		).Scan(&totalData)

		// Fetch
		queryRows, err := config.DB.Query(context.Background(),
			`SELECT nik, nama_pasien, tanggal_lahir, umur, jenis_kelamin, alamat
			 FROM pasien
			 WHERE nik LIKE '%'||$1||'%' OR nama_pasien ILIKE '%'||$1||'%'
			 ORDER BY nama_pasien ASC
			 LIMIT $2 OFFSET $3`,
			search, perPage, offset)
		if err != nil {
			return model.ErrorResponse(c, 500, "Gagal mengambil data pasien")
		}
		defer queryRows.Close()

		for queryRows.Next() {
			var p model.PasienResponse
			var tl interface{}
			queryRows.Scan(&p.NIK, &p.NamaPasien, &tl, &p.Umur, &p.JenisKelamin, &p.Alamat)
			p.TanggalLahir = formatDate(tl)
			rows = append(rows, p)
		}
	} else {
		config.DB.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM pasien`,
		).Scan(&totalData)

		queryRows, err := config.DB.Query(context.Background(),
			`SELECT nik, nama_pasien, tanggal_lahir, umur, jenis_kelamin, alamat
			 FROM pasien ORDER BY nama_pasien ASC LIMIT $1 OFFSET $2`,
			perPage, offset)
		if err != nil {
			return model.ErrorResponse(c, 500, "Gagal mengambil data pasien")
		}
		defer queryRows.Close()

		for queryRows.Next() {
			var p model.PasienResponse
			var tl interface{}
			queryRows.Scan(&p.NIK, &p.NamaPasien, &tl, &p.Umur, &p.JenisKelamin, &p.Alamat)
			p.TanggalLahir = formatDate(tl)
			rows = append(rows, p)
		}
	}

	if rows == nil {
		rows = []model.PasienResponse{}
	}

	return model.PaginatedSuccessResponse(c, rows, totalData, page, perPage)
}

// ─── GET /pasien/:nik ────────────────────────────────────────────────────────

func GetPasienByNIK(c *fiber.Ctx) error {
	nik := strings.TrimSpace(c.Params("nik"))

	var p model.PasienResponse
	var tl interface{}
	err := config.DB.QueryRow(context.Background(),
		`SELECT nik, nama_pasien, tanggal_lahir, umur, jenis_kelamin, alamat FROM pasien WHERE nik = $1`, nik,
	).Scan(&p.NIK, &p.NamaPasien, &tl, &p.Umur, &p.JenisKelamin, &p.Alamat)

	if err != nil {
		return model.ErrorResponse(c, 404, "Pasien tidak ditemukan")
	}

	p.TanggalLahir = formatDate(tl)
	return model.SuccessResponse(c, 200, "Berhasil", p)
}

// ─── POST /pasien ────────────────────────────────────────────────────────────

func CreatePasien(c *fiber.Ctx) error {
	var req model.CreatePasienRequest
	if err := c.BodyParser(&req); err != nil {
		return model.ErrorResponse(c, 400, "Format request body tidak valid")
	}

	req.NIK = strings.TrimSpace(req.NIK)
	req.NamaPasien = strings.TrimSpace(req.NamaPasien)
	req.TanggalLahir = strings.TrimSpace(req.TanggalLahir)
	req.Alamat = strings.TrimSpace(req.Alamat)

	if errs := req.Validate(); len(errs) > 0 {
		return model.ErrorResponse(c, 400, "Validasi gagal", errs)
	}

	_, err := config.DB.Exec(context.Background(),
		`INSERT INTO pasien (nik, nama_pasien, tanggal_lahir, umur, jenis_kelamin, alamat)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		req.NIK, req.NamaPasien, req.TanggalLahir, req.Umur, req.JenisKelamin, req.Alamat)

	if err != nil {
		if config.IsUniqueViolation(err) {
			return model.ErrorResponse(c, 409, "NIK sudah terdaftar")
		}
		return model.ErrorResponse(c, 500, "Gagal menambahkan pasien: "+err.Error())
	}

	return model.SuccessResponse(c, 201, "Pasien berhasil ditambahkan", model.PasienResponse{
		NIK:          req.NIK,
		NamaPasien:   req.NamaPasien,
		TanggalLahir: req.TanggalLahir,
		Umur:         req.Umur,
		JenisKelamin: req.JenisKelamin,
		Alamat:       req.Alamat,
	})
}

// ─── PUT /pasien/:nik ────────────────────────────────────────────────────────

func UpdatePasien(c *fiber.Ctx) error {
	nik := strings.TrimSpace(c.Params("nik"))

	// Cek pasien ada
	var exists bool
	config.DB.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM pasien WHERE nik = $1)`, nik,
	).Scan(&exists)

	if !exists {
		return model.ErrorResponse(c, 404, "Pasien tidak ditemukan")
	}

	var req model.UpdatePasienRequest
	if err := c.BodyParser(&req); err != nil {
		return model.ErrorResponse(c, 400, "Format request body tidak valid")
	}

	req.NamaPasien = strings.TrimSpace(req.NamaPasien)
	req.TanggalLahir = strings.TrimSpace(req.TanggalLahir)
	req.Alamat = strings.TrimSpace(req.Alamat)

	if errs := req.Validate(); len(errs) > 0 {
		return model.ErrorResponse(c, 400, "Validasi gagal", errs)
	}

	_, err := config.DB.Exec(context.Background(),
		`UPDATE pasien SET nama_pasien=$1, tanggal_lahir=$2, umur=$3, jenis_kelamin=$4, alamat=$5, updated_at=NOW()
		 WHERE nik=$6`,
		req.NamaPasien, req.TanggalLahir, req.Umur, req.JenisKelamin, req.Alamat, nik)

	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal update pasien: "+err.Error())
	}

	return model.SuccessResponse(c, 200, "Pasien berhasil diupdate", model.PasienResponse{
		NIK:          nik,
		NamaPasien:   req.NamaPasien,
		TanggalLahir: req.TanggalLahir,
		Umur:         req.Umur,
		JenisKelamin: req.JenisKelamin,
		Alamat:       req.Alamat,
	})
}

// ─── DELETE /pasien/:nik ─────────────────────────────────────────────────────

func DeletePasien(c *fiber.Ctx) error {
	nik := strings.TrimSpace(c.Params("nik"))

	result, err := config.DB.Exec(context.Background(),
		`DELETE FROM pasien WHERE nik = $1`, nik)

	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal menghapus pasien: "+err.Error())
	}

	if result.RowsAffected() == 0 {
		return model.ErrorResponse(c, 404, "Pasien tidak ditemukan")
	}

	return model.SuccessResponse(c, 200, "Pasien berhasil dihapus", nil)
}

// ─── Helper ──────────────────────────────────────────────────────────────────

func formatDate(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		if len(v) >= 10 {
			return v[:10]
		}
		return v
	case []byte:
		s := string(v)
		if len(s) >= 10 {
			return s[:10]
		}
		return s
	default:
		return ""
	}
}
