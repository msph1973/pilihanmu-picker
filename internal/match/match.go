package match

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/msph1973/pilihanmu-picker/internal/scan"
)

const (
	// JPEGOnlyWarningMessage adalah pesan peringatan standar jaring pengaman JPEG (PRD §4.4, R-U4).
	JPEGOnlyWarningMessage = "Sepertinya Anda memotret JPEG saja. Matikan saringan RAW untuk memasukkan file JPEG."
)

// MatchItem merepresentasikan satu nama target dan berkas sumber yang cocok.
type MatchItem struct {
	// TargetName adalah nama dasar target dari daftar seleksi klien.
	TargetName string

	// Files adalah daftar berkas sumber yang cocok dengan target ini.
	Files []scan.FileEntry
}

// Collision merepresentasikan berkas yang memiliki nama dasar dan ekstensi sama dari lokasi sumber berbeda,
// atau bertabrakan dengan berkas yang sudah ada di folder tujuan.
type Collision struct {
	// SourcePath adalah path lengkap berkas sumber.
	SourcePath string

	// RelPath adalah path berkas relatif terhadap root sumber.
	RelPath string

	// TargetFileName adalah nama berkas tujuan yang diharapkan sebelum/sesudah penomoran.
	TargetFileName string

	// ResolvedFileName adalah nama akhir setelah resolusi penomoran (misal: "DSC_0001 (2).CR3").
	ResolvedFileName string
}

// Options menyimpan konfigurasi parameter pencocokan.
type Options struct {
	// RawFilter menentukan apakah hanya ekstensi RAW yang diizinkan (default: true).
	RawFilter bool

	// ExistingDestFiles adalah daftar nama berkas (lowercase) yang sudah ada di direktori tujuan (opsional).
	ExistingDestFiles []string
}

// Result menyimpan hasil pengelompokan ke dalam empat keranjang (PRD §4.3).
type Result struct {
	// Matched berisi kelompok berkas yang cocok dengan nama target.
	Matched []MatchItem

	// NotFoundNames berisi daftar nama dari seleksi klien yang tidak ditemukan di sumber.
	NotFoundNames []string

	// Collisions mencatat tabrakan nama berkas antar sub-folder atau dengan folder tujuan.
	Collisions []Collision

	// SkippedCount adalah jumlah berkas di folder sumber yang tidak dipilih klien.
	SkippedCount int

	// TotalMatchedFiles adalah total jumlah berkas foto + berkas pendamping .xmp yang akan disalin/dipindah.
	TotalMatchedFiles int

	// TotalMatchedSize adalah total ukuran byte dari seluruh berkas yang cocok.
	TotalMatchedSize int64

	// JPEGOnlyWarning bernilai true jika 0 berkas RAW cocok tetapi ditemukan berkas non-RAW yang cocok (R-U4).
	JPEGOnlyWarning bool

	// JPEGOnlyMessage berisi pesan teks peringatan jika JPEGOnlyWarning aktif.
	JPEGOnlyMessage string
}

// Match mencocokkan daftar target dengan hasil pemindaian berkas sumber.
func Match(targetNames []string, scannedFiles []scan.FileEntry, opts Options) Result {
	res := Result{
		Matched:       make([]MatchItem, 0),
		NotFoundNames: make([]string, 0),
		Collisions:    make([]Collision, 0),
	}

	// Buat index pencarian berkas dari hasil scan: map[lowercaseBaseName][]scan.FileEntry
	scanMap := make(map[string][]scan.FileEntry)
	for _, file := range scannedFiles {
		// Abaikan berkas .xmp murni dari pencocokan utama (xmp ditangani sebagai pendamping atau via HasXMP)
		if scan.IsXMP(file.Ext) {
			continue
		}
		key := strings.ToLower(file.BaseName)
		scanMap[key] = append(scanMap[key], file)
	}

	// Lacak berkas sumber yang cocok untuk menghitung keranjang "Dilewati"
	matchedPathSet := make(map[string]struct{})

	// Lacak tabrakan penamaan di folder tujuan (case-insensitive)
	destNameCounts := make(map[string]int)
	for _, name := range opts.ExistingDestFiles {
		key := strings.ToLower(name)
		destNameCounts[key]++
	}

	potentialNonRAWMatches := 0

	// Proses setiap target dalam urutan aslinya (P-7, R-P6)
	for _, targetName := range targetNames {
		key := strings.ToLower(targetName)
		files, exists := scanMap[key]

		if !exists || len(files) == 0 {
			res.NotFoundNames = append(res.NotFoundNames, targetName)
			continue
		}

		var matchingFiles []scan.FileEntry
		for _, f := range files {
			if opts.RawFilter {
				if f.IsRAW {
					matchingFiles = append(matchingFiles, f)
					matchedPathSet[f.Path] = struct{}{}
					if f.HasXMP && f.XMPPath != "" {
						matchedPathSet[f.XMPPath] = struct{}{}
					}
				} else {
					potentialNonRAWMatches++
				}
			} else {
				// Tanpa filter RAW: semua jenis berkas diterima
				matchingFiles = append(matchingFiles, f)
				matchedPathSet[f.Path] = struct{}{}
				if f.HasXMP && f.XMPPath != "" {
					matchedPathSet[f.XMPPath] = struct{}{}
				}
			}
		}

		if len(matchingFiles) == 0 {
			res.NotFoundNames = append(res.NotFoundNames, targetName)
			continue
		}

		res.Matched = append(res.Matched, MatchItem{
			TargetName: targetName,
			Files:      matchingFiles,
		})

		for _, f := range matchingFiles {
			res.TotalMatchedFiles++
			res.TotalMatchedSize += f.Size

			// Periksa tabrakan nama berkas di tujuan
			destKey := strings.ToLower(f.FileName)
			count := destNameCounts[destKey]
			destNameCounts[destKey]++

			if count > 0 {
				// Terjadi tabrakan: berkas sebelumnya sudah memakai nama ini
				resolved := resolveCollisionName(f.FileName, count+1)
				res.Collisions = append(res.Collisions, Collision{
					SourcePath:       f.Path,
					RelPath:          f.RelPath,
					TargetFileName:   f.FileName,
					ResolvedFileName: resolved,
				})
			}

			// Jika ada berkas pendamping XMP, hitung juga ukurannya
			if f.HasXMP && f.XMPPath != "" {
				res.TotalMatchedFiles++
				// Cari ukuran XMP dari scannedFiles jika ada
				for _, sf := range scannedFiles {
					if sf.Path == f.XMPPath {
						res.TotalMatchedSize += sf.Size
						break
					}
				}
			}
		}
	}

	// Hitung keranjang "Dilewati": semua file yang tidak masuk dalam matchedPathSet
	for _, f := range scannedFiles {
		if _, matched := matchedPathSet[f.Path]; !matched {
			res.SkippedCount++
		}
	}

	// R-U4: Jaring pengaman JPEG
	if len(res.Matched) == 0 && opts.RawFilter && potentialNonRAWMatches > 0 {
		res.JPEGOnlyWarning = true
		res.JPEGOnlyMessage = JPEGOnlyWarningMessage
	}

	return res
}

// resolveCollisionName menghasilkan nama berkas dengan sufiks penomoran: "nama (2).EXT".
func resolveCollisionName(fileName string, index int) string {
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	return fmt.Sprintf("%s (%d)%s", base, index, ext)
}
