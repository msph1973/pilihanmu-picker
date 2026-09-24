package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultReportFileName adalah nama standar berkas laporan audit di folder tujuan (PRD §4.7, R-P7).
const DefaultReportFileName = "_laporan.txt"

// CollisionItem mencatat baris tabrakan nama untuk laporan.
type CollisionItem struct {
	SourceRelPath    string
	ResolvedFileName string
}

// FailureItem mencatat baris kegagalan berkas untuk laporan.
type FailureItem struct {
	FileName string
	Reason   string
}

// Data menyimpan informasi lengkap yang dibutuhkan untuk menghasilkan _laporan.txt.
type Data struct {
	// Timestamp adalah waktu eksekusi dimulai atau diselesaikan.
	Timestamp time.Time

	// Mode adalah "Salin" atau "Pindahkan".
	Mode string

	// SourceDir adalah path direktori sumber.
	SourceDir string

	// DestDir adalah path direktori tujuan.
	DestDir string

	// ListTitle adalah nama berkas daftar .txt atau "Clipboard".
	ListTitle string

	// UniqueTargetsCount adalah jumlah nama target unik dari daftar.
	UniqueTargetsCount int

	// MatchedFilesCount adalah jumlah berkas yang berhasil disalin/dipindahkan.
	MatchedFilesCount int

	// NotFoundNames adalah daftar nama yang tidak ditemukan di folder sumber.
	NotFoundNames []string

	// Collisions adalah daftar berkas yang mengalami tabrakan nama.
	Collisions []CollisionItem

	// Failures adalah daftar berkas yang gagal diproses.
	Failures []FailureItem

	// Cancelled menandai jika proses dihentikan sebelum selesai.
	Cancelled bool
}

// Generate menghasilkan isi teks laporan dalam format baku PRD §4.7.
func Generate(d Data) string {
	var b strings.Builder

	timestampStr := d.Timestamp.Format(time.RFC3339)
	if d.Timestamp.IsZero() {
		timestampStr = time.Now().Format(time.RFC3339)
	}

	listInfo := d.ListTitle
	if listInfo == "" {
		listInfo = "Daftar Seleksi"
	}
	if d.UniqueTargetsCount > 0 {
		listInfo = fmt.Sprintf("%s (%d nama unik)", listInfo, d.UniqueTargetsCount)
	}

	b.WriteString("Pilihanmu Picker — Laporan\n")
	b.WriteString(fmt.Sprintf("Waktu     : %s\n", timestampStr))
	b.WriteString(fmt.Sprintf("Mode      : %s\n", d.Mode))
	b.WriteString(fmt.Sprintf("Sumber    : %s\n", d.SourceDir))
	b.WriteString(fmt.Sprintf("Tujuan    : %s\n", d.DestDir))
	b.WriteString(fmt.Sprintf("Daftar    : %s\n", listInfo))
	if d.Cancelled {
		b.WriteString("Status    : Dibatalkan oleh pengguna\n")
	}
	b.WriteString("\n")

	// Ringkasan
	b.WriteString("Ringkasan\n")
	b.WriteString(fmt.Sprintf("  Cocok         : %d file\n", d.MatchedFilesCount))
	b.WriteString(fmt.Sprintf("  Tidak ketemu  : %d nama\n", len(d.NotFoundNames)))
	b.WriteString(fmt.Sprintf("  Tabrakan      : %d\n", len(d.Collisions)))
	b.WriteString(fmt.Sprintf("  Gagal         : %d\n", len(d.Failures)))
	b.WriteString("\n")

	// Tidak ketemu
	b.WriteString("Tidak ketemu\n")
	if len(d.NotFoundNames) == 0 {
		b.WriteString("  (tidak ada)\n")
	} else {
		for _, name := range d.NotFoundNames {
			b.WriteString(fmt.Sprintf("  %s\n", name))
		}
	}
	b.WriteString("\n")

	// Tabrakan
	b.WriteString("Tabrakan\n")
	if len(d.Collisions) == 0 {
		b.WriteString("  (tidak ada)\n")
	} else {
		for _, col := range d.Collisions {
			b.WriteString(fmt.Sprintf("  %s  ->  %s\n", col.SourceRelPath, col.ResolvedFileName))
		}
	}
	b.WriteString("\n")

	// Gagal
	b.WriteString("Gagal\n")
	if len(d.Failures) == 0 {
		b.WriteString("  (tidak ada)\n")
	} else {
		for _, fail := range d.Failures {
			if fail.Reason != "" {
				b.WriteString(fmt.Sprintf("  %s (%s)\n", fail.FileName, fail.Reason))
			} else {
				b.WriteString(fmt.Sprintf("  %s\n", fail.FileName))
			}
		}
	}

	return b.String()
}

// WriteFile menulis teks laporan ke berkas _laporan.txt di folder tujuan (R-P7).
func WriteFile(destDir string, d Data) (string, error) {
	if destDir == "" {
		return "", fmt.Errorf("destination directory cannot be empty")
	}

	content := Generate(d)
	reportPath := filepath.Join(destDir, DefaultReportFileName)

	if err := os.WriteFile(reportPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write report file %q: %w", reportPath, err)
	}

	return reportPath, nil
}
