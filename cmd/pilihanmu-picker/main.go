package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/msph1973/pilihanmu-picker/internal/ui"
)

func main() {
	// Inisialisasi aplikasi Fyne desktop
	a := app.NewWithID("id.web.pilihanmu.picker")
	a.Settings().SetTheme(ui.NewCustomTheme())

	// Buat jendela utama aplikasi
	w := a.NewWindow(ui.AppTitle + " — " + ui.AppSubtitle)
	w.Resize(fyne.NewSize(980, 720))
	w.SetMaster()

	// Inisialisasi antarmuka pengguna
	ui.NewApp(w)

	// Jalankan event loop aplikasi
	w.ShowAndRun()
}
