package model

import (
	"strings"
	"time"
)

// ─── Pasien Model ────────────────────────────────────────────────────────────

type Pasien struct {
	NIK          string    `json:"nik"`
	NamaPasien   string    `json:"nama_pasien"`
	TanggalLahir string    `json:"tanggal_lahir"` // format: YYYY-MM-DD
	Umur         int       `json:"umur"`
	JenisKelamin string    `json:"jenis_kelamin"`
	Alamat       string    `json:"alamat"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ─── Request DTO ─────────────────────────────────────────────────────────────

type CreatePasienRequest struct {
	NIK          string `json:"nik"`
	NamaPasien   string `json:"nama_pasien"`
	TanggalLahir string `json:"tanggal_lahir"`
	Umur         int    `json:"umur"`
	JenisKelamin string `json:"jenis_kelamin"`
	Alamat       string `json:"alamat"`
}

type UpdatePasienRequest struct {
	NamaPasien   string `json:"nama_pasien"`
	TanggalLahir string `json:"tanggal_lahir"`
	Umur         int    `json:"umur"`
	JenisKelamin string `json:"jenis_kelamin"`
	Alamat       string `json:"alamat"`
}

// ─── Response DTO ───────────────────────────────────────────────────────────

type PasienResponse struct {
	NIK          string `json:"nik"`
	NamaPasien   string `json:"nama_pasien"`
	TanggalLahir string `json:"tanggal_lahir"`
	Umur         int    `json:"umur"`
	JenisKelamin string `json:"jenis_kelamin"`
	Alamat       string `json:"alamat"`
}

// ─── Validation ──────────────────────────────────────────────────────────────

func (r *CreatePasienRequest) Validate() []string {
	var errs []string

	nik := strings.TrimSpace(r.NIK)
	nama := strings.TrimSpace(r.NamaPasien)
	tl := strings.TrimSpace(r.TanggalLahir)
	alamat := strings.TrimSpace(r.Alamat)

	// NIK
	if nik == "" {
		errs = append(errs, "NIK tidak boleh kosong")
	} else if len(nik) != 16 {
		errs = append(errs, "NIK harus tepat 16 angka")
	} else if !isNumericStr(nik) {
		errs = append(errs, "NIK harus berisi angka saja")
	}

	// Nama
	if nama == "" {
		errs = append(errs, "Nama Pasien tidak boleh kosong")
	} else if len(nama) < 2 || len(nama) > 100 {
		errs = append(errs, "Nama Pasien harus antara 2-100 karakter")
	}

	// Tanggal Lahir
	if tl == "" {
		errs = append(errs, "Tanggal Lahir tidak boleh kosong")
	} else {
		_, err := time.Parse("2006-01-02", tl)
		if err != nil {
			errs = append(errs, "Format Tanggal Lahir harus YYYY-MM-DD")
		}
	}

	// Umur
	if r.Umur <= 0 || r.Umur > 150 {
		errs = append(errs, "Umur harus antara 1-150")
	}

	// Jenis Kelamin
	if r.JenisKelamin != "Laki-Laki" && r.JenisKelamin != "Perempuan" {
		errs = append(errs, "Jenis Kelamin harus 'Laki-Laki' atau 'Perempuan'")
	}

	// Alamat
	if alamat == "" {
		errs = append(errs, "Alamat tidak boleh kosong")
	} else if len(alamat) < 5 {
		errs = append(errs, "Alamat minimal 5 karakter")
	}

	return errs
}

func (r *UpdatePasienRequest) Validate() []string {
	var errs []string

	nama := strings.TrimSpace(r.NamaPasien)
	tl := strings.TrimSpace(r.TanggalLahir)
	alamat := strings.TrimSpace(r.Alamat)

	if nama == "" {
		errs = append(errs, "Nama Pasien tidak boleh kosong")
	} else if len(nama) < 2 || len(nama) > 100 {
		errs = append(errs, "Nama Pasien harus antara 2-100 karakter")
	}

	if tl == "" {
		errs = append(errs, "Tanggal Lahir tidak boleh kosong")
	} else {
		_, err := time.Parse("2006-01-02", tl)
		if err != nil {
			errs = append(errs, "Format Tanggal Lahir harus YYYY-MM-DD")
		}
	}

	if r.Umur <= 0 || r.Umur > 150 {
		errs = append(errs, "Umur harus antara 1-150")
	}

	if r.JenisKelamin != "Laki-Laki" && r.JenisKelamin != "Perempuan" {
		errs = append(errs, "Jenis Kelamin harus 'Laki-Laki' atau 'Perempuan'")
	}

	if alamat == "" {
		errs = append(errs, "Alamat tidak boleh kosong")
	} else if len(alamat) < 5 {
		errs = append(errs, "Alamat minimal 5 karakter")
	}

	return errs
}

// ─── Helper ──────────────────────────────────────────────────────────────────

func isNumericStr(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
