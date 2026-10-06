package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/msph1973/pilihanmu-picker/internal/ui"
)

// version ditampilkan di judul jendela. Pada build pengembangan nilainya "dev".
//
// Saat rilis, `.github/workflows/release.yml` menulis berkas `version_gen.go` di
// direktori ini (berisi `func init() { version = "<tag>" }`) SEBELUM `fyne package`
// dijalankan, jadi nilainya tidak bergantung pada ldflags.
//
// Jangan kembali ke penyuntikan lewat `-X main.version`/`GOFLAGS`: `fyne package
// -os windows` membangun dua kali, dan build kedua — yang menimpa `.exe` — sudah
// kehilangan GOFLAGS karena Fyne menghapusnya setelah build pertama. Versi jadi
// hilang tanpa error (pernah kejadian di rilis v0.1.0). Lihat AGENTS.md §7 jebakan 12.
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
