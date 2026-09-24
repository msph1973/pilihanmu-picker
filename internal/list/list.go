package list

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

var (
	// ErrInvalidUTF8 dikembalikan jika masukan bukan merupakan teks UTF-8 yang valid.
	ErrInvalidUTF8 = errors.New("input contains invalid UTF-8 characters")
)

// Result menyimpan hasil parsing dari daftar teks nama berkas.
type Result struct {
	// Items berisi daftar nama dasar (basename tanpa path & ekstensi) yang unik,
	// dengan urutan dan huruf kapital sesuai kemunculan pertama.
	Items []string

	// RawLinesCount adalah jumlah seluruh baris pada masukan (sebelum filter/dedup).
	RawLinesCount int

	// NonEmptyLinesCount adalah jumlah baris tidak kosong yang valid sebelum deduplikasi.
	NonEmptyLinesCount int

	// DuplicateCount adalah jumlah baris duplikat yang dibuang saat deduplikasi.
	DuplicateCount int
}

// Summary menghasilkan ringkasan pembacaan daftar dalam format ramah pengguna.
// Contoh: "120 baris → 117 nama unik (3 duplikat dibuang)"
func (r Result) Summary() string {
	return fmt.Sprintf("%d baris → %d nama unik (%d duplikat dibuang)",
		r.NonEmptyLinesCount, len(r.Items), r.DuplicateCount)
}

// ParseText mem-parsing string masukan daftar nama berkas berdasarkan aturan P-1 sampai P-7.
func ParseText(raw string) (Result, error) {
	return ParseBytes([]byte(raw))
}

// ParseReader membaca seluruh data dari io.Reader dan mem-parsing hasilnya.
func ParseReader(r io.Reader) (Result, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Result{}, fmt.Errorf("failed to read input list: %w", err)
	}
	return ParseBytes(data)
}

// ParseBytes mem-parsing slice byte masukan daftar nama berkas.
func ParseBytes(data []byte) (Result, error) {
	// Toleransi UTF-8 BOM (Byte Order Mark: \xef\xbb\xbf)
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	// R-D7: Validasi UTF-8
	if !utf8.Valid(data) {
		return Result{}, ErrInvalidUTF8
	}

	content := string(data)
	if len(content) == 0 {
		return Result{}, nil
	}

	// P-1: Pisah pada \r\n, \n, dan \r
	// Normalisasikan semua variasi baris baru ke \n
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	rawLines := strings.Split(normalized, "\n")
	rawLinesCount := len(rawLines)

	seen := make(map[string]struct{})
	var items []string
	nonEmptyCount := 0
	duplicateCount := 0

	for _, line := range rawLines {
		// P-2: Trim spasi dan tab di awal/akhir baris
		trimmed := strings.TrimSpace(line)

		// P-3: Buang baris kosong
		if trimmed == "" {
			continue
		}

		// P-4: Buang prefiks path bila ada (Windows \ maupun Unix /)
		cleanName := extractBaseName(trimmed)
		if cleanName == "" {
			continue
		}

		// P-5: Buang ekstensi bila ada
		cleanBase := stripExtension(cleanName)
		if cleanBase == "" {
			continue
		}

		nonEmptyCount++

		// P-6: Dedupe case-insensitive, pertahankan kemunculan pertama (P-7: urutan dipertahankan)
		key := strings.ToLower(cleanBase)
		if _, exists := seen[key]; exists {
			duplicateCount++
			continue
		}

		seen[key] = struct{}{}
		items = append(items, cleanBase)
	}

	return Result{
		Items:              items,
		RawLinesCount:      rawLinesCount,
		NonEmptyLinesCount: nonEmptyCount,
		DuplicateCount:     duplicateCount,
	}, nil
}

// extractBaseName mengekstrak nama berkas dari path (mendukung pemisah Windows '\' dan Unix '/').
func extractBaseName(p string) string {
	// Cari indeks terakhir dari '/' atau '\'
	lastSlash := strings.LastIndexAny(p, `/\`)
	if lastSlash >= 0 {
		p = p[lastSlash+1:]
	}
	return strings.TrimSpace(p)
}

// stripExtension menghapus ekstensi berkas terakhir jika ada.
func stripExtension(name string) string {
	lastDot := strings.LastIndex(name, ".")
	if lastDot > 0 { // dot bukan di karakter pertama (untuk menghindari file tersembunyi seperti .gitignore menjadi kosong)
		return name[:lastDot]
	}
	return name
}
