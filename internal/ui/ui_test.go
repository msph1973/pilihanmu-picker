package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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
