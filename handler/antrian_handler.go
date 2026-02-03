package handler

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"sikupas/backend/config"
	"sikupas/backend/model"
)

// ─── GET /antrian ────────────────────────────────────────────────────────────

func GetAllAntrian(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	search := strings.TrimSpace(c.Query("search", ""))
	tanggal := strings.TrimSpace(c.Query("tanggal", ""))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}
	if tanggal == "" {
		tanggal = time.Now().Format("2006-01-02")
	}
	offset := (page - 1) * perPage

	var totalData int
	var rows []model.AntrianResponse

	if search != "" {
		config.DB.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM antrian a JOIN pasien p ON a.nik = p.nik
			 WHERE a.tanggal_kunjungan = $1 AND (a.nik LIKE '%'||$2||'%' OR p.nama_pasien ILIKE '%'||$2||'%')`,
			tanggal, search,
		).Scan(&totalData)

		queryRows, err := config.DB.Query(context.Background(),
			`SELECT a.id_antrian, a.nik, p.nama_pasien, a.nomor_antrian, a.tanggal_kunjungan, a.status
			 FROM antrian a JOIN pasien p ON a.nik = p.nik
			 WHERE a.tanggal_kunjungan = $1 AND (a.nik LIKE '%'||$2||'%' OR p.nama_pasien ILIKE '%'||$2||'%')
			 ORDER BY a.nomor_antrian ASC LIMIT $3 OFFSET $4`,
			tanggal, search, perPage, offset)
		if err != nil {
			return model.ErrorResponse(c, 500, "Gagal mengambil data antrian")
		}
		defer queryRows.Close()

		for queryRows.Next() {
			var a model.AntrianResponse
			var tg interface{}
			queryRows.Scan(&a.IDantrian, &a.NIK, &a.NamaPasien, &a.NomorAntrian, &tg, &a.Status)
			a.TanggalKunjungan = formatDate(tg)
			rows = append(rows, a)
		}
	} else {
		config.DB.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM antrian WHERE tanggal_kunjungan = $1`, tanggal,
		).Scan(&totalData)

		queryRows, err := config.DB.Query(context.Background(),
			`SELECT a.id_antrian, a.nik, p.nama_pasien, a.nomor_antrian, a.tanggal_kunjungan, a.status
			 FROM antrian a JOIN pasien p ON a.nik = p.nik
			 WHERE a.tanggal_kunjungan = $1
			 ORDER BY a.nomor_antrian ASC LIMIT $2 OFFSET $3`,
			tanggal, perPage, offset)
		if err != nil {
			return model.ErrorResponse(c, 500, "Gagal mengambil data antrian")
		}
		defer queryRows.Close()

		for queryRows.Next() {
			var a model.AntrianResponse
			var tg interface{}
			queryRows.Scan(&a.IDantrian, &a.NIK, &a.NamaPasien, &a.NomorAntrian, &tg, &a.Status)
			a.TanggalKunjungan = formatDate(tg)
			rows = append(rows, a)
		}
	}

	if rows == nil {
		rows = []model.AntrianResponse{}
	}

	return model.PaginatedSuccessResponse(c, rows, totalData, page, perPage)
}

// ─── GET /antrian/:id ────────────────────────────────────────────────────────

func GetAntrianByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return model.ErrorResponse(c, 400, "ID tidak valid")
	}

	var a model.AntrianResponse
	var tg interface{}
	err = config.DB.QueryRow(context.Background(),
		`SELECT a.id_antrian, a.nik, p.nama_pasien, a.nomor_antrian, a.tanggal_kunjungan, a.status
		 FROM antrian a JOIN pasien p ON a.nik = p.nik WHERE a.id_antrian = $1`, id,
	).Scan(&a.IDantrian, &a.NIK, &a.NamaPasien, &a.NomorAntrian, &tg, &a.Status)

	if err != nil {
		return model.ErrorResponse(c, 404, "Antrian tidak ditemukan")
	}

	a.TanggalKunjungan = formatDate(tg)
	return model.SuccessResponse(c, 200, "Berhasil", a)
}

// ─── POST /antrian ───────────────────────────────────────────────────────────

func CreateAntrian(c *fiber.Ctx) error {
	var req model.CreateAntrianRequest
	if err := c.BodyParser(&req); err != nil {
		return model.ErrorResponse(c, 400, "Format request body tidak valid")
	}

	req.NIK = strings.TrimSpace(req.NIK)
	req.TanggalKunjungan = strings.TrimSpace(req.TanggalKunjungan)

	if errs := req.Validate(); len(errs) > 0 {
		return model.ErrorResponse(c, 400, "Validasi gagal", errs)
	}

	// Cek pasien ada
	var pasienExists bool
	config.DB.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM pasien WHERE nik = $1)`, req.NIK,
	).Scan(&pasienExists)

	if !pasienExists {
		return model.ErrorResponse(c, 404, "Pasien dengan NIK tersebut tidak ditemukan")
	}

	// Cek sudah ada antrian hari yang sama
	var dupExists bool
	config.DB.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM antrian WHERE nik = $1 AND tanggal_kunjungan = $2)`,
		req.NIK, req.TanggalKunjungan,
	).Scan(&dupExists)

	if dupExists {
		return model.ErrorResponse(c, 409, "Pasien sudah memiliki antrian pada hari ini")
	}

	// Cek max antrian (50 per hari)
	var totalHari int
	config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM antrian WHERE tanggal_kunjungan = $1`, req.TanggalKunjungan,
	).Scan(&totalHari)

	if totalHari >= 50 {
		return model.ErrorResponse(c, 409, "Antrian hari ini sudah penuh (max 50)")
	}

	// Nomor antrian berikutnya
	nomorAntrian := totalHari + 1

	var idAntrian int
	err := config.DB.QueryRow(context.Background(),
		`INSERT INTO antrian (nik, nomor_antrian, tanggal_kunjungan, status)
		 VALUES ($1, $2, $3, 'belum_dikelola')
		 RETURNING id_antrian`,
		req.NIK, nomorAntrian, req.TanggalKunjungan,
	).Scan(&idAntrian)

	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal membuat antrian: "+err.Error())
	}

	// Ambil nama pasien
	var namaPasien string
	config.DB.QueryRow(context.Background(),
		`SELECT nama_pasien FROM pasien WHERE nik = $1`, req.NIK,
	).Scan(&namaPasien)

	return model.SuccessResponse(c, 201, "Antrian berhasil dibuat", model.AntrianResponse{
		IDantrian:        idAntrian,
		NIK:              req.NIK,
		NamaPasien:       namaPasien,
		NomorAntrian:     nomorAntrian,
		TanggalKunjungan: req.TanggalKunjungan,
		Status:           "belum_dikelola",
	})
}

// ─── PUT /antrian/:id ────────────────────────────────────────────────────────

func UpdateAntrianStatus(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return model.ErrorResponse(c, 400, "ID tidak valid")
	}

	var req model.UpdateAntrianStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return model.ErrorResponse(c, 400, "Format request body tidak valid")
	}

	req.Status = strings.TrimSpace(req.Status)
	if errs := req.Validate(); len(errs) > 0 {
		return model.ErrorResponse(c, 400, "Validasi gagal", errs)
	}

	result, err := config.DB.Exec(context.Background(),
		`UPDATE antrian SET status = $1, updated_at = NOW() WHERE id_antrian = $2`,
		req.Status, id)

	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal update status antrian")
	}
	if result.RowsAffected() == 0 {
		return model.ErrorResponse(c, 404, "Antrian tidak ditemukan")
	}

	return model.SuccessResponse(c, 200, "Status antrian berhasil diupdate", nil)
}

// ─── DELETE /antrian/:id ─────────────────────────────────────────────────────

func DeleteAntrian(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return model.ErrorResponse(c, 400, "ID tidak valid")
	}

	result, err := config.DB.Exec(context.Background(),
		`DELETE FROM antrian WHERE id_antrian = $1`, id)

	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal menghapus antrian")
	}
	if result.RowsAffected() == 0 {
		return model.ErrorResponse(c, 404, "Antrian tidak ditemukan")
	}

	return model.SuccessResponse(c, 200, "Antrian berhasil dihapus", nil)
}

// ─── GET /antrian/dashboard ─────────────────────────────────────────────────

func GetDashboardSummary(c *fiber.Ctx) error {
	tanggal := strings.TrimSpace(c.Query("tanggal", ""))
	if tanggal == "" {
		tanggal = time.Now().Format("2006-01-02")
	}

	var total, sudah, belum int

	config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM antrian WHERE tanggal_kunjungan = $1`, tanggal,
	).Scan(&total)

	config.DB.QueryRow(context.Background(),
		`SELECT COUNT(*) FROM antrian WHERE tanggal_kunjungan = $1 AND status = 'sudah_dikelola'`, tanggal,
	).Scan(&sudah)

	belum = total - sudah

	// Cari nomor antrian pertama yang belum dikelola
	nomorSekarang := 1
	config.DB.QueryRow(context.Background(),
		`SELECT COALESCE(MIN(nomor_antrian), 1) FROM antrian
		 WHERE tanggal_kunjungan = $1 AND status = 'belum_dikelola'`, tanggal,
	).Scan(&nomorSekarang)

	return model.SuccessResponse(c, 200, "Berhasil", model.DashboardSummary{
		TotalAntrian:         total,
		TotalSudahDikelola:   sudah,
		TotalBelumDikelola:   belum,
		NomorAntrianSekarang: nomorSekarang,
	})
}

// ─── GET /antrian/boxes ─────────────────────────────────────────────────────

func GetAntrianBoxes(c *fiber.Ctx) error {
	tanggal := strings.TrimSpace(c.Query("tanggal", ""))
	if tanggal == "" {
		tanggal = time.Now().Format("2006-01-02")
	}

	boxes := make([]model.AntrianBoxItem, 50)
	for i := 0; i < 50; i++ {
		boxes[i] = model.AntrianBoxItem{
			NomorAntrian: i + 1,
			Status:       "kosong",
		}
	}

	// Ambil semua antrian hari ini
	queryRows, err := config.DB.Query(context.Background(),
		`SELECT nomor_antrian, status FROM antrian WHERE tanggal_kunjungan = $1`, tanggal)
	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal mengambil data antrian")
	}
	defer queryRows.Close()

	for queryRows.Next() {
		var nomor int
		var status string
		queryRows.Scan(&nomor, &status)
		if nomor >= 1 && nomor <= 50 {
			boxes[nomor-1].Status = status
		}
	}

	return model.SuccessResponse(c, 200, "Berhasil", boxes)
}
