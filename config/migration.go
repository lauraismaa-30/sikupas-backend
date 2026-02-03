package config

import (
	"context"
	"log"
)

// RunMigrations menjalankan SQL migration untuk membuat tabel
func RunMigrations() {
	migrations := []string{
		// ===================== Tabel Users =====================
		`CREATE TABLE IF NOT EXISTS users (
			id          SERIAL PRIMARY KEY,
			nama        VARCHAR(100) NOT NULL,
			username    VARCHAR(50)  NOT NULL UNIQUE,
			password    VARCHAR(255) NOT NULL,
			role        VARCHAR(20)  NOT NULL DEFAULT 'admin' CHECK (role IN ('admin', 'kepala_puskesmas')),
			created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMP    NOT NULL DEFAULT NOW()
		);`,

		// ===================== Tabel Pasien =====================
		`CREATE TABLE IF NOT EXISTS pasien (
			nik             VARCHAR(20) PRIMARY KEY,
			nama_pasien     VARCHAR(100) NOT NULL,
			tanggal_lahir   DATE         NOT NULL,
			umur            INTEGER      NOT NULL,
			jenis_kelamin   VARCHAR(10)  NOT NULL CHECK (jenis_kelamin IN ('Laki-Laki', 'Perempuan')),
			alamat          TEXT         NOT NULL,
			created_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
			updated_at      TIMESTAMP    NOT NULL DEFAULT NOW()
		);`,

		// ===================== Tabel Antrian =====================
		`CREATE TABLE IF NOT EXISTS antrian (
			id_antrian        SERIAL      PRIMARY KEY,
			nik               VARCHAR(20) NOT NULL REFERENCES pasien(nik) ON DELETE CASCADE,
			nomor_antrian     INTEGER     NOT NULL,
			tanggal_kunjungan DATE        NOT NULL DEFAULT CURRENT_DATE,
			status            VARCHAR(20) NOT NULL DEFAULT 'belum_dikelola' CHECK (status IN ('belum_dikelola', 'sudah_dikelola')),
			created_at        TIMESTAMP   NOT NULL DEFAULT NOW(),
			updated_at        TIMESTAMP   NOT NULL DEFAULT NOW(),
			UNIQUE(nomor_antrian, tanggal_kunjungan)
		);`,

		// ===================== Tabel Poli (Master) =====================
		`CREATE TABLE IF NOT EXISTS poli (
			id_poli     SERIAL      PRIMARY KEY,
			nama_poli   VARCHAR(50) NOT NULL,
			nama_dokter VARCHAR(100) NOT NULL
		);`,

		// ===================== Seed Data Poli =====================
		`INSERT INTO poli (nama_poli, nama_dokter) VALUES
			('Poli Umum', 'Dr. Ahmad Suryadi'),
			('Poli Anak', 'Dr. Siti Nurhaliza'),
			('Poli Kandungan', 'Dr. Dewi Lestari'),
			('Poli Gigi', 'Drg. Budi Santoso'),
			('Poli Mata', 'Dr. Rini Wahyudi')
		ON CONFLICT DO NOTHING;`,

		// ===================== Tabel Pemeriksaan =====================
		`CREATE TABLE IF NOT EXISTS pemeriksaan (
			id_pemeriksaan       SERIAL      PRIMARY KEY,
			nik_pasien           VARCHAR(20) NOT NULL REFERENCES pasien(nik) ON DELETE CASCADE,
			tanggal_pemeriksaan  DATE        NOT NULL DEFAULT CURRENT_DATE,
			keluhan              TEXT        NOT NULL,
			id_poli              INTEGER     NOT NULL REFERENCES poli(id_poli),
			metode_pembayaran    VARCHAR(10) NOT NULL CHECK (metode_pembayaran IN ('Umum', 'BPJS')),
			nominal_pembayaran   NUMERIC(12,2) NOT NULL DEFAULT 0,
			created_at           TIMESTAMP   NOT NULL DEFAULT NOW(),
			updated_at           TIMESTAMP   NOT NULL DEFAULT NOW()
		);`,

		// ===================== Index untuk performa =====================
		`CREATE INDEX IF NOT EXISTS idx_antrian_tanggal ON antrian(tanggal_kunjungan);`,
		`CREATE INDEX IF NOT EXISTS idx_antrian_nik ON antrian(nik);`,
		`CREATE INDEX IF NOT EXISTS idx_pemeriksaan_tanggal ON pemeriksaan(tanggal_pemeriksaan);`,
		`CREATE INDEX IF NOT EXISTS idx_pemeriksaan_nik ON pemeriksaan(nik_pasien);`,
	}

	for i, sql := range migrations {
		_, err := DB.Exec(context.Background(), sql)
		if err != nil {
			log.Fatalf("❌ Migration %d gagal: %v\nSQL: %s", i+1, err, sql)
		}
	}

	log.Println("✅ Semua migrasi berhasil dijalankan")
}
