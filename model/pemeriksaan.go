package model

import (
	"strings"
	"time"
)

// ─── Poli Model ──────────────────────────────────────────────────────────────

type Poli struct {
	IDPoli     int    `json:"id_poli"`
	NamaPoli   string `json:"nama_poli"`
	NamaDokter string `json:"nama_dokter"`
}

// ─── Pemeriksaan Model ───────────────────────────────────────────────────────

type Pemeriksaan struct {
	IDPemeriksaan      int       `json:"id_pemeriksaan"`
	NIKPasien          string    `json:"nik_pasien"`
	NamaPasien         string    `json:"nama_pasien"`           // dari JOIN pasien
	TanggalPemeriksaan string    `json:"tanggal_pemeriksaan"`   // YYYY-MM-DD
	Keluhan            string    `json:"keluhan"`
	IDPoli             int       `json:"id_poli"`
	NamaPoli           string    `json:"nama_poli"`             // dari JOIN poli
	NamaDokter         string    `json:"nama_dokter"`           // dari JOIN poli
	MetodePembayaran   string    `json:"metode_pembayaran"`     // Umum / BPJS
	NominalPembayaran  float64   `json:"nominal_pembayaran"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ─── Request DTO ─────────────────────────────────────────────────────────────

type CreatePemeriksaanRequest struct {
	NIKPasien         string  `json:"nik_pasien"`
	Keluhan           string  `json:"keluhan"`
	IDPoli            int     `json:"id_poli"`
	MetodePembayaran  string  `json:"metode_pembayaran"`
	NominalPembayaran float64 `json:"nominal_pembayaran"`
}

type UpdatePemeriksaanRequest struct {
	NIKPasien         string  `json:"nik_pasien"`
	Keluhan           string  `json:"keluhan"`
	IDPoli            int     `json:"id_poli"`
	MetodePembayaran  string  `json:"metode_pembayaran"`
	NominalPembayaran float64 `json:"nominal_pembayaran"`
}

// ─── Response DTO ───────────────────────────────────────────────────────────

type PemeriksaanResponse struct {
	IDPemeriksaan      int     `json:"id_pemeriksaan"`
	NIKPasien          string  `json:"nik_pasien"`
	NamaPasien         string  `json:"nama_pasien"`
	TanggalPemeriksaan string  `json:"tanggal_pemeriksaan"`
	Keluhan            string  `json:"keluhan"`
	NamaPoli           string  `json:"nama_poli"`
	NamaDokter         string  `json:"nama_dokter"`
	MetodePembayaran   string  `json:"metode_pembayaran"`
	NominalPembayaran  float64 `json:"nominal_pembayaran"`
}

// ─── Report Response ─────────────────────────────────────────────────────────

type ReportPasien struct {
	NIK          string `json:"nik"`
	NamaPasien   string `json:"nama_pasien"`
	TanggalLahir string `json:"tanggal_lahir"`
	Umur         int    `json:"umur"`
	JenisKelamin string `json:"jenis_kelamin"`
	Alamat       string `json:"alamat"`
}

type ReportPemeriksaan struct {
	IDPemeriksaan      int     `json:"id_pemeriksaan"`
	NIKPasien          string  `json:"nik_pasien"`
	NamaPasien         string  `json:"nama_pasien"`
	TanggalPemeriksaan string  `json:"tanggal_pemeriksaan"`
	Keluhan            string  `json:"keluhan"`
	NamaPoli           string  `json:"nama_poli"`
	NamaDokter         string  `json:"nama_dokter"`
	MetodePembayaran   string  `json:"metode_pembayaran"`
	NominalPembayaran  float64 `json:"nominal_pembayaran"`
}

// ─── Validation ──────────────────────────────────────────────────────────────

func (r *CreatePemeriksaanRequest) Validate() []string {
	var errs []string

	nik := strings.TrimSpace(r.NIKPasien)
	keluhan := strings.TrimSpace(r.Keluhan)

	if nik == "" {
		errs = append(errs, "NIK Pasien tidak boleh kosong")
	} else if len(nik) != 16 || !isNumericStr(nik) {
		errs = append(errs, "NIK Pasien harus 16 angka")
	}

	if keluhan == "" {
		errs = append(errs, "Keluhan tidak boleh kosong")
	} else if len(keluhan) < 3 {
		errs = append(errs, "Keluhan minimal 3 karakter")
	}

	if r.IDPoli <= 0 {
		errs = append(errs, "Poli harus dipilih")
	}

	mp := strings.TrimSpace(r.MetodePembayaran)
	if mp != "Umum" && mp != "BPJS" {
		errs = append(errs, "Metode Pembayaran harus 'Umum' atau 'BPJS'")
	}

	// Jika BPJS, nominal harus 0
	if mp == "BPJS" && r.NominalPembayaran != 0 {
		errs = append(errs, "Nominal Pembayaran harus Rp 0 untuk BPJS")
	}

	// Jika Umum, nominal harus >= 0
	if mp == "Umum" && r.NominalPembayaran < 0 {
		errs = append(errs, "Nominal Pembayaran tidak boleh negatif")
	}

	return errs
}

func (r *UpdatePemeriksaanRequest) Validate() []string {
	var errs []string

	nik := strings.TrimSpace(r.NIKPasien)
	keluhan := strings.TrimSpace(r.Keluhan)

	if nik == "" {
		errs = append(errs, "NIK Pasien tidak boleh kosong")
	} else if len(nik) != 16 || !isNumericStr(nik) {
		errs = append(errs, "NIK Pasien harus 16 angka")
	}

	if keluhan == "" {
		errs = append(errs, "Keluhan tidak boleh kosong")
	} else if len(keluhan) < 3 {
		errs = append(errs, "Keluhan minimal 3 karakter")
	}

	if r.IDPoli <= 0 {
		errs = append(errs, "Poli harus dipilih")
	}

	mp := strings.TrimSpace(r.MetodePembayaran)
	if mp != "Umum" && mp != "BPJS" {
		errs = append(errs, "Metode Pembayaran harus 'Umum' atau 'BPJS'")
	}

	if mp == "BPJS" && r.NominalPembayaran != 0 {
		errs = append(errs, "Nominal Pembayaran harus Rp 0 untuk BPJS")
	}

	if mp == "Umum" && r.NominalPembayaran < 0 {
		errs = append(errs, "Nominal Pembayaran tidak boleh negatif")
	}

	return errs
}
