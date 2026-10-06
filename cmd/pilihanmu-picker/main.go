package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/msph1973/pilihanmu-picker/internal/ui"
)

// version diisi saat rilis lewat penyuntik ldflags:
//
//	go build -ldflags "-X main.version=v0.1.0" ./cmd/pilihanmu-picker
//
// Lewat `fyne package` penyuntikannya berbeda: `GOFLAGS="-ldflags=-X=main.version=v0.1.0"`
// (tanda `=` wajib, tanpa spasi — lihat AGENTS.md §7 jebakan 12).
//
// Pada build pengembangan nilainya "dev".
var version = "dev"

func main() {
	// Inisialisasi aplikasi Fyne desktop
	a := app.NewWithID("id.web.pilihanmu.picker")
	a.Settings().SetTheme(ui.NewCustomTheme())

	// Buat jendela utama aplikasi. Versi ditampilkan di judul agar pengguna
	// bisa menyebutkannya saat melapor.
	w := a.NewWindow(ui.AppTitle + " — " + ui.AppSubtitle + " (" + version + ")")
	w.Resize(fyne.NewSize(980, 720))
	w.SetMaster()

	// Inisialisasi antarmuka pengguna
	ui.NewApp(w)

	// Jalankan event loop aplikasi
	w.ShowAndRun()
}
