package model

import (
	"strings"
	"time"
)

// ─── Antrian Model ───────────────────────────────────────────────────────────

type Antrian struct {
	IDantrian        int       `json:"id_antrian"`
	NIK              string    `json:"nik"`
	NamaPasien       string    `json:"nama_pasien"`       // dari JOIN pasien
	NomorAntrian     int       `json:"nomor_antrian"`
	TanggalKunjungan string    `json:"tanggal_kunjungan"` // YYYY-MM-DD
	Status           string    `json:"status"`            // belum_dikelola / sudah_dikelola
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ─── Request DTO ─────────────────────────────────────────────────────────────

type CreateAntrianRequest struct {
	NIK              string `json:"nik"`
	TanggalKunjungan string `json:"tanggal_kunjungan"`
}

type UpdateAntrianStatusRequest struct {
	Status string `json:"status"`
}

// ─── Response DTO ───────────────────────────────────────────────────────────

type AntrianResponse struct {
	IDantrian        int    `json:"id_antrian"`
	NIK              string `json:"nik"`
	NamaPasien       string `json:"nama_pasien"`
	NomorAntrian     int    `json:"nomor_antrian"`
	TanggalKunjungan string `json:"tanggal_kunjungan"`
	Status           string `json:"status"`
}

// ─── Dashboard Summary ──────────────────────────────────────────────────────

type DashboardSummary struct {
	TotalAntrian       int `json:"total_antrian"`
	TotalSudahDikelola int `json:"total_sudah_dikelola"`
	TotalBelumDikelola int `json:"total_belum_dikelola"`
	NomorAntrianSekarang int `json:"nomor_antrian_sekarang"`
}

type AntrianBoxItem struct {
	NomorAntrian int    `json:"nomor_antrian"`
	Status       string `json:"status"` // belum_dikelola / sudah_dikelola / kosong
}

// ─── Validation ──────────────────────────────────────────────────────────────

func (r *CreateAntrianRequest) Validate() []string {
	var errs []string

	nik := strings.TrimSpace(r.NIK)
	tg := strings.TrimSpace(r.TanggalKunjungan)

	if nik == "" {
		errs = append(errs, "NIK tidak boleh kosong")
	} else if len(nik) != 16 {
		errs = append(errs, "NIK harus tepat 16 angka")
	} else if !isNumericStr(nik) {
		errs = append(errs, "NIK harus berisi angka saja")
	}

	if tg == "" {
		errs = append(errs, "Tanggal Kunjungan tidak boleh kosong")
	} else {
		_, err := time.Parse("2006-01-02", tg)
		if err != nil {
			errs = append(errs, "Format Tanggal Kunjungan harus YYYY-MM-DD")
		}
	}

	return errs
}

func (r *UpdateAntrianStatusRequest) Validate() []string {
	var errs []string
	s := strings.TrimSpace(r.Status)
	if s != "belum_dikelola" && s != "sudah_dikelola" {
		errs = append(errs, "Status harus 'belum_dikelola' atau 'sudah_dikelola'")
	}
	return errs
}
