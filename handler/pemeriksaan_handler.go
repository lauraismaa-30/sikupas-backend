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

// ─── GET /poli ───────────────────────────────────────────────────────────────

func GetAllPoli(c *fiber.Ctx) error {
	rows, err := config.DB.Query(context.Background(),
		`SELECT id_poli, nama_poli, nama_dokter FROM poli ORDER BY id_poli`)
	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal mengambil data poli")
	}
	defer rows.Close()

	var polis []model.Poli
	for rows.Next() {
		var p model.Poli
		rows.Scan(&p.IDPoli, &p.NamaPoli, &p.NamaDokter)
		polis = append(polis, p)
	}

	if polis == nil {
		polis = []model.Poli{}
	}

	return model.SuccessResponse(c, 200, "Berhasil", polis)
}

// ─── GET /pemeriksaan ────────────────────────────────────────────────────────

func GetAllPemeriksaan(c *fiber.Ctx) error {
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
	offset := (page - 1) * perPage

	var totalData int
	var rows []model.PemeriksaanResponse

	baseWhere := `WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if tanggal != "" {
		baseWhere += ` AND pe.tanggal_pemeriksaan = $` + strconv.Itoa(argIdx)
		args = append(args, tanggal)
		argIdx++
	}

	if search != "" {
		baseWhere += ` AND (p.nik LIKE '%'||$` + strconv.Itoa(argIdx) + `||'%' OR p.nama_pasien ILIKE '%'||$` + strconv.Itoa(argIdx) + `||'%')`
		args = append(args, search)
		argIdx++
	}

	// Count
	countSQL := `SELECT COUNT(*) FROM pemeriksaan pe JOIN pasien p ON pe.nik_pasien = p.nik JOIN poli po ON pe.id_poli = po.id_poli ` + baseWhere
	config.DB.QueryRow(context.Background(), countSQL, args...).Scan(&totalData)

	// Fetch
	fetchSQL := `SELECT pe.id_pemeriksaan, pe.nik_pasien, p.nama_pasien,
		pe.tanggal_pemeriksaan, pe.keluhan, po.nama_poli, po.nama_dokter,
		pe.metode_pembayaran, pe.nominal_pembayaran
		FROM pemeriksaan pe
		JOIN pasien p ON pe.nik_pasien = p.nik
		JOIN poli po ON pe.id_poli = po.id_poli
		` + baseWhere + `
		ORDER BY pe.tanggal_pemeriksaan DESC, pe.id_pemeriksaan DESC
		LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)

	args = append(args, perPage, offset)

	queryRows, err := config.DB.Query(context.Background(), fetchSQL, args...)
	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal mengambil data pemeriksaan: "+err.Error())
	}
	defer queryRows.Close()

	for queryRows.Next() {
		var pm model.PemeriksaanResponse
		var tp interface{}
		queryRows.Scan(&pm.IDPemeriksaan, &pm.NIKPasien, &pm.NamaPasien,
			&tp, &pm.Keluhan, &pm.NamaPoli, &pm.NamaDokter,
			&pm.MetodePembayaran, &pm.NominalPembayaran)
		pm.TanggalPemeriksaan = formatDate(tp)
		rows = append(rows, pm)
	}

	if rows == nil {
		rows = []model.PemeriksaanResponse{}
	}

	return model.PaginatedSuccessResponse(c, rows, totalData, page, perPage)
}

// ─── GET /pemeriksaan/:id ────────────────────────────────────────────────────

func GetPemeriksaanByID(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return model.ErrorResponse(c, 400, "ID tidak valid")
	}

	var pm model.PemeriksaanResponse
	var tp interface{}
	err = config.DB.QueryRow(context.Background(),
		`SELECT pe.id_pemeriksaan, pe.nik_pasien, p.nama_pasien,
			pe.tanggal_pemeriksaan, pe.keluhan, po.nama_poli, po.nama_dokter,
			pe.metode_pembayaran, pe.nominal_pembayaran
			FROM pemeriksaan pe
			JOIN pasien p ON pe.nik_pasien = p.nik
			JOIN poli po ON pe.id_poli = po.id_poli
			WHERE pe.id_pemeriksaan = $1`, id,
	).Scan(&pm.IDPemeriksaan, &pm.NIKPasien, &pm.NamaPasien,
		&tp, &pm.Keluhan, &pm.NamaPoli, &pm.NamaDokter,
		&pm.MetodePembayaran, &pm.NominalPembayaran)

	if err != nil {
		return model.ErrorResponse(c, 404, "Pemeriksaan tidak ditemukan")
	}

	pm.TanggalPemeriksaan = formatDate(tp)
	return model.SuccessResponse(c, 200, "Berhasil", pm)
}

// ─── POST /pemeriksaan ───────────────────────────────────────────────────────

func CreatePemeriksaan(c *fiber.Ctx) error {
	var req model.CreatePemeriksaanRequest
	if err := c.BodyParser(&req); err != nil {
		return model.ErrorResponse(c, 400, "Format request body tidak valid")
	}

	req.NIKPasien = strings.TrimSpace(req.NIKPasien)
	req.Keluhan = strings.TrimSpace(req.Keluhan)
	req.MetodePembayaran = strings.TrimSpace(req.MetodePembayaran)

	if errs := req.Validate(); len(errs) > 0 {
		return model.ErrorResponse(c, 400, "Validasi gagal", errs)
	}

	// Cek pasien ada
	var pasienExists bool
	config.DB.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM pasien WHERE nik = $1)`, req.NIKPasien,
	).Scan(&pasienExists)
	if !pasienExists {
		return model.ErrorResponse(c, 404, "Pasien tidak ditemukan")
	}

	// Cek poli ada
	var poliExists bool
	config.DB.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM poli WHERE id_poli = $1)`, req.IDPoli,
	).Scan(&poliExists)
	if !poliExists {
		return model.ErrorResponse(c, 404, "Poli tidak ditemukan")
	}

	// Jika BPJS, set nominal = 0
	if req.MetodePembayaran == "BPJS" {
		req.NominalPembayaran = 0
	}

	tanggalHari := time.Now().Format("2006-01-02")

	var idPem int
	err := config.DB.QueryRow(context.Background(),
		`INSERT INTO pemeriksaan (nik_pasien, tanggal_pemeriksaan, keluhan, id_poli, metode_pembayaran, nominal_pembayaran)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id_pemeriksaan`,
		req.NIKPasien, tanggalHari, req.Keluhan, req.IDPoli, req.MetodePembayaran, req.NominalPembayaran,
	).Scan(&idPem)

	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal membuat pemeriksaan: "+err.Error())
	}

	// Auto update status antrian -> sudah_dikelola
	config.DB.Exec(context.Background(),
		`UPDATE antrian SET status = 'sudah_dikelola', updated_at = NOW()
		 WHERE nik = $1 AND tanggal_kunjungan = $2 AND status = 'belum_dikelola'`,
		req.NIKPasien, tanggalHari)

	// Ambil data lengkap untuk response
	var pm model.PemeriksaanResponse
	var tp interface{}
	config.DB.QueryRow(context.Background(),
		`SELECT pe.id_pemeriksaan, pe.nik_pasien, p.nama_pasien,
			pe.tanggal_pemeriksaan, pe.keluhan, po.nama_poli, po.nama_dokter,
			pe.metode_pembayaran, pe.nominal_pembayaran
			FROM pemeriksaan pe
			JOIN pasien p ON pe.nik_pasien = p.nik
			JOIN poli po ON pe.id_poli = po.id_poli
			WHERE pe.id_pemeriksaan = $1`, idPem,
	).Scan(&pm.IDPemeriksaan, &pm.NIKPasien, &pm.NamaPasien,
		&tp, &pm.Keluhan, &pm.NamaPoli, &pm.NamaDokter,
		&pm.MetodePembayaran, &pm.NominalPembayaran)
	pm.TanggalPemeriksaan = formatDate(tp)

	return model.SuccessResponse(c, 201, "Pemeriksaan berhasil ditambahkan", pm)
}

// ─── PUT /pemeriksaan/:id ────────────────────────────────────────────────────

func UpdatePemeriksaan(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return model.ErrorResponse(c, 400, "ID tidak valid")
	}

	// Cek pemeriksaan ada
	var exists bool
	config.DB.QueryRow(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM pemeriksaan WHERE id_pemeriksaan = $1)`, id,
	).Scan(&exists)
	if !exists {
		return model.ErrorResponse(c, 404, "Pemeriksaan tidak ditemukan")
	}

	var req model.UpdatePemeriksaanRequest
	if err := c.BodyParser(&req); err != nil {
		return model.ErrorResponse(c, 400, "Format request body tidak valid")
	}

	req.NIKPasien = strings.TrimSpace(req.NIKPasien)
	req.Keluhan = strings.TrimSpace(req.Keluhan)
	req.MetodePembayaran = strings.TrimSpace(req.MetodePembayaran)

	if errs := req.Validate(); len(errs) > 0 {
		return model.ErrorResponse(c, 400, "Validasi gagal", errs)
	}

	if req.MetodePembayaran == "BPJS" {
		req.NominalPembayaran = 0
	}

	_, err = config.DB.Exec(context.Background(),
		`UPDATE pemeriksaan SET nik_pasien=$1, keluhan=$2, id_poli=$3,
		 metode_pembayaran=$4, nominal_pembayaran=$5, updated_at=NOW()
		 WHERE id_pemeriksaan=$6`,
		req.NIKPasien, req.Keluhan, req.IDPoli, req.MetodePembayaran, req.NominalPembayaran, id)

	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal update pemeriksaan: "+err.Error())
	}

	// Ambil data lengkap
	var pm model.PemeriksaanResponse
	var tp interface{}
	config.DB.QueryRow(context.Background(),
		`SELECT pe.id_pemeriksaan, pe.nik_pasien, p.nama_pasien,
			pe.tanggal_pemeriksaan, pe.keluhan, po.nama_poli, po.nama_dokter,
			pe.metode_pembayaran, pe.nominal_pembayaran
			FROM pemeriksaan pe
			JOIN pasien p ON pe.nik_pasien = p.nik
			JOIN poli po ON pe.id_poli = po.id_poli
			WHERE pe.id_pemeriksaan = $1`, id,
	).Scan(&pm.IDPemeriksaan, &pm.NIKPasien, &pm.NamaPasien,
		&tp, &pm.Keluhan, &pm.NamaPoli, &pm.NamaDokter,
		&pm.MetodePembayaran, &pm.NominalPembayaran)
	pm.TanggalPemeriksaan = formatDate(tp)

	return model.SuccessResponse(c, 200, "Pemeriksaan berhasil diupdate", pm)
}

// ─── DELETE /pemeriksaan/:id ──────────────────────────────────────────────────

func DeletePemeriksaan(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return model.ErrorResponse(c, 400, "ID tidak valid")
	}

	result, err := config.DB.Exec(context.Background(),
		`DELETE FROM pemeriksaan WHERE id_pemeriksaan = $1`, id)

	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal menghapus pemeriksaan")
	}
	if result.RowsAffected() == 0 {
		return model.ErrorResponse(c, 404, "Pemeriksaan tidak ditemukan")
	}

	return model.SuccessResponse(c, 200, "Pemeriksaan berhasil dihapus", nil)
}

// ─── GET /laporan/pasien ─────────────────────────────────────────────────────

func GetReportPasien(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	search := strings.TrimSpace(c.Query("search", ""))
	tanggalDari := strings.TrimSpace(c.Query("tanggal_dari", ""))
	tanggalSampai := strings.TrimSpace(c.Query("tanggal_sampai", ""))

	if page < 1 { page = 1 }
	if perPage < 1 || perPage > 100 { perPage = 10 }
	offset := (page - 1) * perPage

	baseWhere := `WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if tanggalDari != "" {
		baseWhere += ` AND p.created_at::date >= $` + strconv.Itoa(argIdx)
		args = append(args, tanggalDari)
		argIdx++
	}
	if tanggalSampai != "" {
		baseWhere += ` AND p.created_at::date <= $` + strconv.Itoa(argIdx)
		args = append(args, tanggalSampai)
		argIdx++
	}
	if search != "" {
		baseWhere += ` AND (p.nik LIKE '%'||$` + strconv.Itoa(argIdx) + `||'%' OR p.nama_pasien ILIKE '%'||$` + strconv.Itoa(argIdx) + `||'%')`
		args = append(args, search)
		argIdx++
	}

	var totalData int
	countSQL := `SELECT COUNT(*) FROM pasien p ` + baseWhere
	config.DB.QueryRow(context.Background(), countSQL, args...).Scan(&totalData)

	fetchSQL := `SELECT p.nik, p.nama_pasien, p.tanggal_lahir, p.umur, p.jenis_kelamin, p.alamat
		FROM pasien p ` + baseWhere + `
		ORDER BY p.nama_pasien ASC
		LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, perPage, offset)

	queryRows, err := config.DB.Query(context.Background(), fetchSQL, args...)
	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal mengambil laporan pasien")
	}
	defer queryRows.Close()

	var rows []model.ReportPasien
	for queryRows.Next() {
		var r model.ReportPasien
		var tl interface{}
		queryRows.Scan(&r.NIK, &r.NamaPasien, &tl, &r.Umur, &r.JenisKelamin, &r.Alamat)
		r.TanggalLahir = formatDate(tl)
		rows = append(rows, r)
	}

	if rows == nil { rows = []model.ReportPasien{} }
	return model.PaginatedSuccessResponse(c, rows, totalData, page, perPage)
}

// ─── GET /laporan/pemeriksaan ────────────────────────────────────────────────

func GetReportPemeriksaan(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "10"))
	search := strings.TrimSpace(c.Query("search", ""))
	tanggalDari := strings.TrimSpace(c.Query("tanggal_dari", ""))
	tanggalSampai := strings.TrimSpace(c.Query("tanggal_sampai", ""))

	if page < 1 { page = 1 }
	if perPage < 1 || perPage > 100 { perPage = 10 }
	offset := (page - 1) * perPage

	baseWhere := `WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if tanggalDari != "" {
		baseWhere += ` AND pe.tanggal_pemeriksaan >= $` + strconv.Itoa(argIdx)
		args = append(args, tanggalDari)
		argIdx++
	}
	if tanggalSampai != "" {
		baseWhere += ` AND pe.tanggal_pemeriksaan <= $` + strconv.Itoa(argIdx)
		args = append(args, tanggalSampai)
		argIdx++
	}
	if search != "" {
		baseWhere += ` AND (p.nik LIKE '%'||$` + strconv.Itoa(argIdx) + `||'%' OR p.nama_pasien ILIKE '%'||$` + strconv.Itoa(argIdx) + `||'%')`
		args = append(args, search)
		argIdx++
	}

	var totalData int
	countSQL := `SELECT COUNT(*) FROM pemeriksaan pe JOIN pasien p ON pe.nik_pasien = p.nik JOIN poli po ON pe.id_poli = po.id_poli ` + baseWhere
	config.DB.QueryRow(context.Background(), countSQL, args...).Scan(&totalData)

	fetchSQL := `SELECT pe.id_pemeriksaan, pe.nik_pasien, p.nama_pasien,
		pe.tanggal_pemeriksaan, pe.keluhan, po.nama_poli, po.nama_dokter,
		pe.metode_pembayaran, pe.nominal_pembayaran
		FROM pemeriksaan pe
		JOIN pasien p ON pe.nik_pasien = p.nik
		JOIN poli po ON pe.id_poli = po.id_poli
		` + baseWhere + `
		ORDER BY pe.tanggal_pemeriksaan DESC
		LIMIT $` + strconv.Itoa(argIdx) + ` OFFSET $` + strconv.Itoa(argIdx+1)
	args = append(args, perPage, offset)

	queryRows, err := config.DB.Query(context.Background(), fetchSQL, args...)
	if err != nil {
		return model.ErrorResponse(c, 500, "Gagal mengambil laporan pemeriksaan")
	}
	defer queryRows.Close()

	var rows []model.ReportPemeriksaan
	for queryRows.Next() {
		var r model.ReportPemeriksaan
		var tp interface{}
		queryRows.Scan(&r.IDPemeriksaan, &r.NIKPasien, &r.NamaPasien,
			&tp, &r.Keluhan, &r.NamaPoli, &r.NamaDokter,
			&r.MetodePembayaran, &r.NominalPembayaran)
		r.TanggalPemeriksaan = formatDate(tp)
		rows = append(rows, r)
	}

	if rows == nil { rows = []model.ReportPemeriksaan{} }
	return model.PaginatedSuccessResponse(c, rows, totalData, page, perPage)
}
