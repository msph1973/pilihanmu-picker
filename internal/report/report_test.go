package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateReport(t *testing.T) {
	parsedTime, _ := time.Parse(time.RFC3339, "2026-09-24T14:03:11+08:00")

	data := Data{
		Timestamp:          parsedTime,
		Mode:               "Salin",
		SourceDir:          `D:\Photos\2026-09-20 Wedding`,
		DestDir:            `D:\Photos\2026-09-20 Wedding - SELECTED`,
		ListTitle:          "Wedding Rian - Daftar Foto.txt",
		UniqueTargetsCount: 120,
		MatchedFilesCount:  117,
		NotFoundNames:      []string{"DSC_0042", "IMG_8891", "_MG_2210"},
		Collisions: []CollisionItem{
			{
				SourceRelPath:    `subA\DSC_0001.CR3`,
				ResolvedFileName: `DSC_0001.CR3`,
			},
			{
				SourceRelPath:    `subB\DSC_0001.CR3`,
				ResolvedFileName: `DSC_0001 (2).CR3`,
			},
		},
		Failures: []FailureItem{},
	}

	reportText := Generate(data)

	// Verifikasi konten
	expectedSections := []string{
		"Pilihanmu Picker — Laporan",
		"Waktu     : 2026-09-24T14:03:11+08:00",
		"Mode      : Salin",
		"Sumber    : D:\\Photos\\2026-09-20 Wedding",
		"Tujuan    : D:\\Photos\\2026-09-20 Wedding - SELECTED",
		"Daftar    : Wedding Rian - Daftar Foto.txt (120 nama unik)",
		"Ringkasan",
		"  Cocok         : 117 file",
		"  Tidak ketemu  : 3 nama",
		"  Tabrakan      : 2",
		"  Gagal         : 0",
		"Tidak ketemu",
		"  DSC_0042",
		"  IMG_8891",
		"  _MG_2210",
		"Tabrakan",
		"  subA\\DSC_0001.CR3  ->  DSC_0001.CR3",
		"  subB\\DSC_0001.CR3  ->  DSC_0001 (2).CR3",
		"Gagal",
		"  (tidak ada)",
	}

	for _, sec := range expectedSections {
		if !strings.Contains(reportText, sec) {
			t.Errorf("expected report to contain %q, but missing", sec)
		}
	}
}

func TestWriteFile(t *testing.T) {
	tempDir := t.TempDir()

	data := Data{
		Mode:              "Pindahkan",
		SourceDir:         tempDir,
		DestDir:           tempDir,
		ListTitle:         "Seleksi.txt",
		MatchedFilesCount: 5,
		Cancelled:         true,
		Failures: []FailureItem{
			{
				FileName: "DSC_9999.CR3",
				Reason:   "berkas sedang dibuka aplikasi lain",
			},
		},
	}

	reportPath, err := WriteFile(tempDir, data)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if filepath.Base(reportPath) != DefaultReportFileName {
		t.Errorf("expected report filename %q, got %q", DefaultReportFileName, filepath.Base(reportPath))
	}

	content, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed to read written report file: %v", err)
	}

	strContent := string(content)
	if !strings.Contains(strContent, "Status    : Dibatalkan oleh pengguna") {
		t.Errorf("expected cancellation status in report")
	}
	if !strings.Contains(strContent, "DSC_9999.CR3 (berkas sedang dibuka aplikasi lain)") {
		t.Errorf("expected failure detail in report")
	}
}

func TestWriteFileError(t *testing.T) {
	_, err := WriteFile("", Data{})
	if err == nil {
		t.Error("expected error for empty destination directory")
	}
}
