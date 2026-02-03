package route

import (
	"sikupas/backend/handler"
	"sikupas/backend/middleware"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes mendefinisikan semua route API
func SetupRoutes(app *fiber.App) {

	// ─── Auth Routes (public) ──────────────────────────────────────────
	auth := app.Group("/api/auth")
	{
		auth.Post("/register", handler.Register)
		auth.Post("/login", handler.Login)
	}

	// ─── Protected Routes ──────────────────────────────────────────────
	api := app.Group("/api", middleware.AuthRequired())

	// Endpoint untuk mendapatkan info user login
	api.Get("/me", handler.GetCurrentUser)

	// ─── Pasien CRUD (Admin only) ──────────────────────────────────────
	pasien := api.Group("/pasien", middleware.RoleRequired("admin"))
	{
		pasien.Get("/", handler.GetAllPasien)        // Get /api/pasien
		pasien.Get("/:nik", handler.GetPasienByNIK)  // Get /api/pasien/:nik
		pasien.Post("/", handler.CreatePasien)       // Post /api/pasien
		pasien.Put("/:nik", handler.UpdatePasien)    // Put /api/pasien/:nik
		pasien.Delete("/:nik", handler.DeletePasien) // Delete /api/pasien/:nik
	}

	// ─── Antrian CRUD (Admin only) ─────────────────────────────────────
	antrian := api.Group("/antrian", middleware.RoleRequired("admin"))
	{
		antrian.Get("/", handler.GetAllAntrian)                // Get /api/antrian
		antrian.Get("/dashboard", handler.GetDashboardSummary) // Get /api/antrian/dashboard
		antrian.Get("/boxes", handler.GetAntrianBoxes)         // Get /api/antrian/boxes
		antrian.Get("/:id", handler.GetAntrianByID)            // Get /api/antrian/:id
		antrian.Post("/", handler.CreateAntrian)               // Post /api/antrian
		antrian.Put("/:id", handler.UpdateAntrianStatus)       // Put /api/antrian/:id
		antrian.Delete("/:id", handler.DeleteAntrian)          // Delete /api/antrian/:id
	}

	// ─── Poli (Admin only) ─────────────────────────────────────────────
	poli := api.Group("/poli", middleware.RoleRequired("admin"))
	{
		poli.Get("/", handler.GetAllPoli) // Get /api/poli
	}

	// ─── Pemeriksaan CRUD (Admin only) ─────────────────────────────────
	pemeriksaan := api.Group("/pemeriksaan", middleware.RoleRequired("admin"))
	{
		pemeriksaan.Get("/", handler.GetAllPemeriksaan)       // Get /api/pemeriksaan
		pemeriksaan.Get("/:id", handler.GetPemeriksaanByID)   // Get /api/pemeriksaan/:id
		pemeriksaan.Post("/", handler.CreatePemeriksaan)      // Post /api/pemeriksaan
		pemeriksaan.Put("/:id", handler.UpdatePemeriksaan)    // Put /api/pemeriksaan/:id
		pemeriksaan.Delete("/:id", handler.DeletePemeriksaan) // Delete /api/pemeriksaan/:id
	}

	// ─── Laporan (Admin + Kepala Puskesmas) ────────────────────────────
	laporan := api.Group("/laporan", middleware.RoleRequired("admin", "kepala_puskesmas"))
	{
		laporan.Get("/pasien", handler.GetReportPasien)           // Get /api/laporan/pasien
		laporan.Get("/pemeriksaan", handler.GetReportPemeriksaan) // Get /api/laporan/pemeriksaan
	}
}
