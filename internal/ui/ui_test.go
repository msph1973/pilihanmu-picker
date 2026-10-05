package ui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/msph1973/pilihanmu-picker/internal/fsops"
	"github.com/msph1973/pilihanmu-picker/internal/list"
	"github.com/msph1973/pilihanmu-picker/internal/match"
	"github.com/msph1973/pilihanmu-picker/internal/report"
	"github.com/msph1973/pilihanmu-picker/internal/scan"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{52428800, "50.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		got := FormatBytes(tt.bytes)
		if got != tt.expected {
			t.Errorf("FormatBytes(%d) = %q, expected %q", tt.bytes, got, tt.expected)
		}
	}
}

func TestCustomTheme(t *testing.T) {
	th := NewCustomTheme()

	// Test colors
	bgDark := th.Color(theme.ColorNameBackground, theme.VariantDark)
	if bgDark == nil {
		t.Error("expected dark background color not to be nil")
	}

	bgLight := th.Color(theme.ColorNameBackground, theme.VariantLight)
	if bgLight == nil {
		t.Error("expected light background color not to be nil")
	}

	primary := th.Color(theme.ColorNamePrimary, theme.VariantDark)
	if primary == nil {
		t.Error("expected primary color not to be nil")
	}

	// Test sizes
	padding := th.Size(theme.SizeNamePadding)
	if padding <= 0 {
		t.Errorf("expected positive padding, got %f", padding)
	}

	textSize := th.Size(theme.SizeNameText)
	if textSize <= 0 {
		t.Errorf("expected positive text size, got %f", textSize)
	}
}

func TestBasketCard(t *testing.T) {
	test.NewApp()
	btn := widget.NewButton("Aksi", func() {})
	card := NewBasketCard("Keranjang Uji", "10 berkas", "Keterangan", btn, theme.Color(theme.ColorNamePrimary))

	if card.Title != "Keranjang Uji" {
		t.Errorf("expected title %q, got %q", "Keranjang Uji", card.Title)
	}
	if card.CountText != "10 berkas" {
		t.Errorf("expected count %q, got %q", "10 berkas", card.CountText)
	}

	card.SetValues("25 berkas", "Keterangan Baru")
	if card.CountText != "25 berkas" || card.SubText != "Keterangan Baru" {
		t.Errorf("SetValues failed to update fields")
	}

	renderer := card.CreateRenderer()
	if renderer == nil {
		t.Error("expected renderer not to be nil")
	}
}

func TestAppInitializationAndState(t *testing.T) {
	testApp := test.NewApp()
	testApp.Settings().SetTheme(NewCustomTheme())
	win := test.NewWindow(nil)

	appUI := NewApp(win)
	if appUI == nil {
		t.Fatal("expected NewApp not to return nil")
	}

	// Verifikasi nilai default
	if !appUI.includeSubfolders {
		t.Error("expected includeSubfolders default to be true")
	}
	if !appUI.rawFilterEnabled {
		t.Error("expected rawFilterEnabled default to be true")
	}

	// Verifikasi tombol eksekusi nonaktif di awal (R-P4)
	if !appUI.copyBtn.Disabled() {
		t.Error("expected copyBtn to be disabled initially")
	}
	if !appUI.moveBtn.Disabled() {
		t.Error("expected moveBtn to be disabled initially")
	}

	// Test input text parsing
	appUI.inputEntry.SetText("DSC_0001\nDSC_0002\nDSC_0001\n")
	if len(appUI.listResult.Items) != 2 {
		t.Errorf("expected 2 unique items parsed, got %d", len(appUI.listResult.Items))
	}
	if appUI.listResult.DuplicateCount != 1 {
		t.Errorf("expected 1 duplicate, got %d", appUI.listResult.DuplicateCount)
	}

	// Test clearing input
	appUI.inputEntry.SetText("")
	if len(appUI.listResult.Items) != 0 {
		t.Errorf("expected 0 items after clearing, got %d", len(appUI.listResult.Items))
	}
}

func TestDialogHelpers(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(nil)

	// Test ShowNotFoundNamesDialog dengan daftar kosong dan isi
	ShowNotFoundNamesDialog(win, []string{})
	ShowNotFoundNamesDialog(win, []string{"DSC_0001", "DSC_0002"})

	// Test ShowReportViewerDialog
	ShowReportViewerDialog(win, "Isi laporan pengujian...", "/tmp/_laporan.txt")
}

func TestOSHelper(t *testing.T) {
	err := OpenFolderInFileManager("")
	if err == nil {
		t.Error("expected error for empty path in OpenFolderInFileManager")
	}
}

// collectOverlayText menelusuri pohon fyne.CanvasObject secara rekursif (lewat Objects milik
// container maupun renderer widget) dan mengumpulkan seluruh teks widget.Label/widget.Entry yang
// ditemukan. Dipakai untuk memverifikasi isi dialog yang sedang tampil tanpa bergantung pada tipe
// internal paket dialog yang tidak diekspor (dialog.ShowError, dialog.NewCustom, dst. membungkus
// pesannya dengan cara yang berbeda-beda).
func collectOverlayText(o fyne.CanvasObject, depth int) string {
	if o == nil || depth > 15 {
		return ""
	}

	var sb strings.Builder
	switch w := o.(type) {
	case *widget.Label:
		sb.WriteString(w.Text)
		sb.WriteString("\n")
	case *widget.Entry:
		sb.WriteString(w.Text)
		sb.WriteString("\n")
	}

	if c, ok := o.(*fyne.Container); ok {
		for _, child := range c.Objects {
			sb.WriteString(collectOverlayText(child, depth+1))
		}
		return sb.String()
	}

	if wid, ok := o.(fyne.Widget); ok {
		if r := wid.CreateRenderer(); r != nil {
			for _, child := range r.Objects() {
				sb.WriteString(collectOverlayText(child, depth+1))
			}
		}
	}
	return sb.String()
}

// topOverlayText mengembalikan seluruh teks label yang tampil pada overlay (dialog) teratas
// window, atau string kosong bila tidak ada overlay yang sedang tampil.
func topOverlayText(win fyne.Window) string {
	top := win.Canvas().Overlays().Top()
	return collectOverlayText(top, 0)
}

// waitForAsyncGeneration menunggu hingga a.asyncGeneration naik setidaknya sebesar minIncrements
// dibanding nilai from, atau hingga timeout tercapai.
//
// Dipakai alih-alih memeriksa isExecuting/isScanning secara langsung: kedua flag itu sengaja
// di-set selesai (false) LEBIH AWAL di dalam callback fyne.Do yang sama (sebelum
// updateActionButtonsState dipanggil, supaya tombol dievaluasi dengan status terbaru), sehingga
// tidak bisa dipakai sebagai sinyal "seluruh pembaruan widget pada goroutine ini sudah rampung".
// Bila tes memakai flag itu saja, goroutine latar belakang (termasuk rantai pemindaian ulang
// otomatis yang dipicu di akhir startExecution) bisa saja masih menyentuh widget Fyne saat tes
// berikutnya mulai membangun App/Window baru — dua goroutine tak tersinkron memakai mesin
// perender teks Fyne secara bersamaan, teramati sebagai panic acak di pustaka shaping font saat
// dijalankan dengan -race. asyncGeneration dinaikkan sebagai pernyataan PALING TERAKHIR di tiap
// siklus, sehingga begitu nilainya teramati naik, goroutine tsb dijamin benar-benar selesai.
func waitForAsyncGeneration(a *App, from uint64, minIncrements uint64, timeout time.Duration) bool {
	target := from + minIncrements
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if a.asyncGeneration.Load() >= target {
			// Jeda tambahan kecil: driver tes Fyne (berbeda dari driver glfw produksi sejak Fyne
			// v2.6) TIDAK memarshalkan fyne.Do ke satu goroutine render khusus — ia sekadar
			// memanggil callback langsung di goroutine pemanggilnya. Ini dikenal menyebabkan race
			// pada cache internal Fyne (mis. cache metrik font) bila goroutine latar belakang
			// aplikasi tumpang tindih dengan goroutine tes lain yang sedang membangun widget
			// (lih. fyne-io/fyne#6346, #5619). asyncGeneration sudah menjamin SELURUH pembaruan
			// widget milik siklus ini rampung; jeda kecil ini murni jaring pengaman tambahan
			// untuk pekerjaan internal Fyne yang mungkin masih menyusul (timer/refresh tertunda)
			// di luar kendali kode aplikasi, supaya tes berikutnya benar-benar tidak tumpang
			// tindih saat dijalankan dengan -race.
			time.Sleep(50 * time.Millisecond)
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return a.asyncGeneration.Load() >= target
}

// TestTriggerAsyncScanFullFlow menguji jalur pemindaian+pencocokan asinkron penuh (sebelumnya
// nol cakupan) dengan folder sumber sungguhan berisi berkas RAW dan satu nama yang tidak ketemu.
//
// Catatan: pemicu memakai pengisian field + pemanggilan triggerAsyncScan() langsung, BUKAN lewat
// sourceEntry.SetText(...). widget.Entry.SetText memanggil OnChanged secara sinkron DI TENGAH
// proses internalnya sendiri (sebelum tuntas menata ulang kursor/scroll), sehingga goroutine
// latar belakang yang dipicu OnChanged bisa benar-benar berjalan bersamaan dengan sisa pekerjaan
// internal SetText itu sendiri pada driver tes Fyne (yang tidak memarshalkan fyne.Do ke satu
// goroutine render seperti driver produksi sejak Fyne v2.6) — ini keterbatasan driver tes Fyne,
// bukan bug di kode aplikasi, dan terbukti hilang total begitu dipicu lewat pemanggilan langsung.
func TestTriggerAsyncScanFullFlow(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(nil)
	a := NewApp(win)

	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "DSC_0001.CR3"), []byte("RAW_DATA"), 0644); err != nil {
		t.Fatal(err)
	}

	a.listResult = list.Result{Items: []string{"DSC_0001", "DSC_9999"}}
	a.sourceDir = srcDir

	before := a.asyncGeneration.Load()
	a.triggerAsyncScan()

	if !waitUntilScanDone(a, before, 5*time.Second) {
		t.Fatal("expected scan to finish within timeout")
	}

	if a.matchedCard.CountText != "1 berkas" {
		t.Errorf("expected matchedCard count '1 berkas', got %q", a.matchedCard.CountText)
	}
	if a.notFoundCard.CountText != "1 nama" {
		t.Errorf("expected notFoundCard count '1 nama', got %q", a.notFoundCard.CountText)
	}
	if !strings.Contains(a.statusScanLabel.Text, "Pindai selesai") {
		t.Errorf("expected scan-complete status, got %q", a.statusScanLabel.Text)
	}
}

// TestTriggerAsyncScanDestReadWarning membuktikan perbaikan R-C2: kegagalan membaca folder tujuan
// WAJIB dilaporkan di label status, bukan ditelan diam-diam.
//
// Kegagalan DISUNTIK lewat field listDestFiles, bukan dipicu dengan menjadikan folder tujuan
// sebuah berkas biasa: pemicu lama menghasilkan error di Linux (ENOTDIR) tetapi TIDAK di Windows,
// sehingga tes ini gagal di CI windows-latest. Dengan suntikan ini jalurnya identik di semua OS.
//
// Catatan: pemicu memakai pengisian field + triggerAsyncScan() langsung (lihat komentar di
// TestTriggerAsyncScanFullFlow soal kenapa Entry.SetText() dihindari di sini).
func TestTriggerAsyncScanDestReadWarning(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(nil)
	a := NewApp(win)

	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "DSC_0001.CR3"), []byte("RAW_DATA"), 0644); err != nil {
		t.Fatal(err)
	}

	a.listDestFiles = func(string) ([]string, error) {
		return nil, errors.New("destination directory is not readable")
	}
	a.destDir = t.TempDir()
	a.listResult = list.Result{Items: []string{"DSC_0001"}}
	a.sourceDir = srcDir

	before := a.asyncGeneration.Load()
	a.triggerAsyncScan()

	if !waitUntilScanDone(a, before, 5*time.Second) {
		t.Fatal("expected scan to finish within timeout")
	}

	if !strings.Contains(a.statusScanLabel.Text, "Peringatan membaca folder tujuan") {
		t.Errorf("expected destination read warning in status label, got %q", a.statusScanLabel.Text)
	}
}

// waitUntilScanDone adalah alias waitForAsyncGeneration khusus untuk satu siklus pemindaian saja
// (tanpa eksekusi salin/pindah berantai di belakangnya).
func waitUntilScanDone(a *App, before uint64, timeout time.Duration) bool {
	return waitForAsyncGeneration(a, before, 1, timeout)
}

// TestConfirmAndExecuteSameFolderShowsError menguji pagar (guard) sumber == tujuan pada
// confirmAndExecute: harus menampilkan dialog error eksplisit, bukan diam-diam tidak melakukan apa pun.
func TestConfirmAndExecuteSameFolderShowsError(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(nil)
	a := NewApp(win)

	same := t.TempDir()
	a.sourceDir = same
	a.destDir = same

	a.confirmAndExecute(fsops.ModeCopy)

	got := topOverlayText(win)
	if !strings.Contains(got, WarnSameFolder) {
		t.Errorf("expected overlay to contain %q, got %q", WarnSameFolder, got)
	}
}

// TestConfirmAndExecuteNoMatchesShowsInfo menguji pagar "tidak ada berkas cocok" pada
// confirmAndExecute.
func TestConfirmAndExecuteNoMatchesShowsInfo(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(nil)
	a := NewApp(win)

	a.sourceDir = t.TempDir()
	a.destDir = t.TempDir()
	a.matchResult = match.Result{}

	a.confirmAndExecute(fsops.ModeCopy)

	got := topOverlayText(win)
	if !strings.Contains(got, LabelNoMatches) {
		t.Errorf("expected overlay to contain %q, got %q", LabelNoMatches, got)
	}
}

// TestStartExecutionSuccessShowsSummary menjalankan startExecution end-to-end dengan berkas
// sungguhan dan memverifikasi dialog ringkasan sukses benar-benar tampil serta _laporan.txt
// tertulis di folder tujuan (R-P7).
func TestStartExecutionSuccessShowsSummary(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(nil)
	a := NewApp(win)

	srcDir := t.TempDir()
	rawPath := filepath.Join(srcDir, "DSC_0001.CR3")
	content := []byte("RAW_DATA_SUCCESS")
	if err := os.WriteFile(rawPath, content, 0644); err != nil {
		t.Fatal(err)
	}
	destDir := filepath.Join(t.TempDir(), "output")

	a.sourceDir = srcDir
	a.destDir = destDir
	a.listResult = list.Result{Items: []string{"DSC_0001"}}
	a.matchResult = match.Result{
		Matched: []match.MatchItem{{
			TargetName: "DSC_0001",
			Files: []scan.FileEntry{{
				Path:     rawPath,
				RelPath:  "DSC_0001.CR3",
				FileName: "DSC_0001.CR3",
				BaseName: "DSC_0001",
				Ext:      ".CR3",
				Size:     int64(len(content)),
				IsRAW:    true,
			}},
		}},
		TotalMatchedFiles: 1,
		TotalMatchedSize:  int64(len(content)),
	}

	before := a.asyncGeneration.Load()
	a.startExecution(fsops.ModeCopy)

	// 2 siklus: penyelesaian eksekusi itu sendiri + pemindaian ulang otomatis yang dipicunya.
	if !waitForAsyncGeneration(a, before, 2, 10*time.Second) {
		t.Fatal("expected startExecution to finish within timeout")
	}

	got := topOverlayText(win)
	if !strings.Contains(got, SummarySuccessLabel) {
		t.Errorf("expected success summary overlay containing %q, got %q", SummarySuccessLabel, got)
	}

	reportPath := filepath.Join(destDir, report.DefaultReportFileName)
	if _, err := os.Stat(reportPath); err != nil {
		t.Errorf("expected report file at %s: %v", reportPath, err)
	}
}

// TestStartExecutionFailureDoesNotShowFalseSuccess adalah tes regresi untuk bug kritis hasil
// audit: ketika fsops.Execute gagal total (mis. folder tujuan ternyata berkas biasa sehingga
// MkdirAll gagal), aplikasi TIDAK BOLEH menampilkan dialog ringkasan "sukses" palsu. Pada kode
// lama, error dari fsops.Execute dibuang (execRes, _ := fsops.Execute(...)) sehingga alur tetap
// lanjut ke showSummaryDialog dengan hasil kosong yang dibaca sebagai sukses.
func TestStartExecutionFailureDoesNotShowFalseSuccess(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(nil)
	a := NewApp(win)

	srcDir := t.TempDir()
	rawPath := filepath.Join(srcDir, "DSC_0001.CR3")
	content := []byte("RAW_DATA_FAILURE")
	if err := os.WriteFile(rawPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	// destDir menunjuk ke berkas biasa (bukan direktori): os.MkdirAll di fsops.Execute wajib gagal.
	blockingFile := filepath.Join(t.TempDir(), "blocker-dest")
	if err := os.WriteFile(blockingFile, []byte("bukan folder"), 0644); err != nil {
		t.Fatal(err)
	}

	a.sourceDir = srcDir
	a.destDir = blockingFile
	a.listResult = list.Result{Items: []string{"DSC_0001"}}
	a.matchResult = match.Result{
		Matched: []match.MatchItem{{
			TargetName: "DSC_0001",
			Files: []scan.FileEntry{{
				Path:     rawPath,
				RelPath:  "DSC_0001.CR3",
				FileName: "DSC_0001.CR3",
				BaseName: "DSC_0001",
				Ext:      ".CR3",
				Size:     int64(len(content)),
				IsRAW:    true,
			}},
		}},
		TotalMatchedFiles: 1,
		TotalMatchedSize:  int64(len(content)),
	}

	before := a.asyncGeneration.Load()
	a.startExecution(fsops.ModeCopy)

	// 2 siklus: penyelesaian eksekusi itu sendiri + pemindaian ulang otomatis yang dipicunya.
	if !waitForAsyncGeneration(a, before, 2, 10*time.Second) {
		t.Fatal("expected startExecution to finish within timeout")
	}

	got := topOverlayText(win)
	if strings.Contains(got, SummarySuccessLabel) {
		t.Fatalf("BUG: false success summary shown after total execution failure, overlay text: %q", got)
	}
	if !strings.Contains(got, "Gagal menjalankan proses") {
		t.Errorf("expected actionable execution-failed error overlay, got %q", got)
	}

	// _laporan.txt tidak boleh dianggap tertulis di lokasi yang bukan direktori.
	reportPath := filepath.Join(blockingFile, report.DefaultReportFileName)
	if _, err := os.Stat(reportPath); err == nil {
		t.Errorf("did not expect a report file to exist at invalid destination %s", reportPath)
	}
}

// TestResetBasketCardsUsesConsistentLabels adalah tes regresi untuk bug kecil hasil audit:
// resetBasketCards sebelumnya mengisi skippedCard dengan subteks "0 berkas" (duplikat dari
// CountText), bukan deskripsi kategori seperti tiga keranjang lainnya.
func TestResetBasketCardsUsesConsistentLabels(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(nil)
	a := NewApp(win)

	// Pastikan nilai sempat berubah dari default sebelum direset.
	a.skippedCard.SetValues("3 berkas", "Bukan pilihan / non-RAW")

	a.resetBasketCards()

	if a.skippedCard.SubText != SubSkippedInitial {
		t.Errorf("expected skippedCard subtext %q after reset, got %q", SubSkippedInitial, a.skippedCard.SubText)
	}
	if a.skippedCard.SubText == LabelZeroFiles {
		t.Error("skippedCard subtext must not duplicate its own count text after reset")
	}
}

// TestShowCollisionsDialogHelper menguji dialog detail "Tabrakan" baru (R-U2, PRD §4.5): kasus
// kosong menampilkan info netral, kasus berisi menampilkan pasangan RelPath -> ResolvedFileName.
func TestShowCollisionsDialogHelper(t *testing.T) {
	test.NewApp()
	win := test.NewWindow(nil)

	ShowCollisionsDialog(win, nil)
	gotEmpty := topOverlayText(win)
	if !strings.Contains(gotEmpty, MsgNoCollisions) {
		t.Errorf("expected empty-collisions overlay to contain %q, got %q", MsgNoCollisions, gotEmpty)
	}

	collisions := []match.Collision{
		{SourcePath: "/src/subA/DSC_0001.CR3", RelPath: "subA/DSC_0001.CR3", TargetFileName: "DSC_0001.CR3", ResolvedFileName: "DSC_0001 (2).CR3"},
	}
	ShowCollisionsDialog(win, collisions)
	got := topOverlayText(win)
	if !strings.Contains(got, "subA/DSC_0001.CR3") || !strings.Contains(got, "DSC_0001 (2).CR3") {
		t.Errorf("expected collision pair in overlay text, got %q", got)
	}
}
