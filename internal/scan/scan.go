package scan

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RawExtensions adalah daftar tertutup ekstensi RAW yang didukung (PRD §4.4, R-D5).
var RawExtensions = map[string]struct{}{
	"3FR": {},
	"ARW": {},
	"CR2": {},
	"CR3": {},
	"CRW": {},
	"DCR": {},
	"DNG": {},
	"ERF": {},
	"FFF": {},
	"IIQ": {},
	"KDC": {},
	"MEF": {},
	"MOS": {},
	"MRW": {},
	"NEF": {},
	"NRW": {},
	"ORF": {},
	"PEF": {},
	"PTX": {},
	"RAF": {},
	"RAW": {},
	"RWL": {},
	"RW2": {},
	"SR2": {},
	"SRF": {},
	"SRW": {},
	"X3F": {},
}

// IsRAW mengembalikan true jika ekstensi yang diberikan merupakan format RAW terdaftar.
func IsRAW(ext string) bool {
	ext = strings.TrimPrefix(ext, ".")
	ext = strings.ToUpper(strings.TrimSpace(ext))
	_, ok := RawExtensions[ext]
	return ok
}

// IsXMP mengembalikan true jika berkas berekstensi .xmp (case-insensitive).
func IsXMP(ext string) bool {
	ext = strings.TrimPrefix(ext, ".")
	return strings.EqualFold(ext, "xmp")
}

// FileEntry mewakili informasi satu berkas yang ditemukan selama pemindaian.
type FileEntry struct {
	// Path adalah path absolut/bersih berkas di disk.
	Path string

	// RelPath adalah path berkas relatif terhadap root folder sumber.
	RelPath string

	// FileName adalah nama berkas lengkap dengan ekstensi asli (misal: "DSC_0001.CR3").
	FileName string

	// BaseName adalah nama berkas tanpa ekstensi (misal: "DSC_0001").
	BaseName string

	// Ext adalah ekstensi berkas asli termasuk titik (misal: ".CR3").
	Ext string

	// Size adalah ukuran berkas dalam byte.
	Size int64

	// ModTime adalah waktu modifikasi terakhir berkas.
	ModTime time.Time

	// IsRAW menandai apakah berkas merupakan format RAW.
	IsRAW bool

	// HasXMP bernilai true jika berkas pendamping .xmp ditemukan di direktori yang sama.
	HasXMP bool

	// XMPPath adalah path lengkap berkas .xmp pendamping jika ditemukan.
	XMPPath string
}

// Options menyimpan konfigurasi parameter pemindaian direktori sumber.
type Options struct {
	// RootDir adalah path direktori sumber yang akan dipindai.
	RootDir string

	// Recursive menentukan apakah pemindaian dilakukan ke dalam sub-folder.
	Recursive bool
}

// Result menyimpan hasil keseluruhan dari proses pemindaian.
type Result struct {
	// Files berisi seluruh entri berkas yang valid ditemukan.
	Files []FileEntry

	// RAWCount adalah jumlah berkas foto RAW yang ditemukan.
	RAWCount int

	// NonRAWCount adalah jumlah berkas bukan RAW (termasuk JPEG, TIFF, PNG, dll).
	NonRAWCount int

	// Warnings mencatat peringatan direktori yang tidak dapat diakses atau dilewati.
	Warnings []string

	// ScannedDirsCount adalah jumlah direktori yang diperiksa.
	ScannedDirsCount int
}

// Scan memindai folder sumber sesuai opsi yang diberikan.
func Scan(opts Options) (Result, error) {
	if opts.RootDir == "" {
		return Result{}, fmt.Errorf("source root directory cannot be empty")
	}

	cleanRoot, err := filepath.Abs(filepath.Clean(opts.RootDir))
	if err != nil {
		return Result{}, fmt.Errorf("failed to resolve root directory: %w", err)
	}

	rootInfo, err := os.Stat(cleanRoot)
	if err != nil {
		return Result{}, fmt.Errorf("cannot access source directory %q: %w", cleanRoot, err)
	}
	if !rootInfo.IsDir() {
		return Result{}, fmt.Errorf("source path is not a directory: %q", cleanRoot)
	}

	res := Result{
		Files:    make([]FileEntry, 0),
		Warnings: make([]string, 0),
	}

	// Peta untuk menyimpan berkas XMP per direktori: map[dirPath]map[lowercaseBaseName]fullXMPPath
	xmpMap := make(map[string]map[string]string)

	// Koleksi sementara sebelum linking sidecar XMP
	var candidateFiles []FileEntry

	if !opts.Recursive {
		// Pemindaian non-rekursif (hanya isi folder root)
		res.ScannedDirsCount = 1
		entries, err := os.ReadDir(cleanRoot)
		if err != nil {
			return Result{}, fmt.Errorf("cannot read directory %q: %w", cleanRoot, err)
		}

		for _, entry := range entries {
			fullPath := filepath.Join(cleanRoot, entry.Name())
			if entry.Type()&fs.ModeSymlink != 0 {
				targetInfo, err := os.Stat(fullPath)
				if err != nil {
					res.Warnings = append(res.Warnings, fmt.Sprintf("cannot resolve symlink %q: %v", fullPath, err))
					continue
				}
				if targetInfo.IsDir() {
					// R-D6: Symlink direktori tidak diikuti
					res.Warnings = append(res.Warnings, fmt.Sprintf("directory symlink skipped: %q", fullPath))
					continue
				}
				processFileEntry(cleanRoot, fullPath, entry.Name(), targetInfo, &candidateFiles, xmpMap)
				continue
			}

			if entry.IsDir() {
				continue
			}
			entryInfo, err := entry.Info()
			if err != nil {
				res.Warnings = append(res.Warnings, fmt.Sprintf("cannot read file info for %q: %v", entry.Name(), err))
				continue
			}

			processFileEntry(cleanRoot, fullPath, entry.Name(), entryInfo, &candidateFiles, xmpMap)
		}
	} else {
		// Pemindaian rekursif (PRD §4.3)
		walkErr := filepath.WalkDir(cleanRoot, func(currentPath string, d fs.DirEntry, err error) error {
			if err != nil {
				// Lewati direktori tak terbaca tanpa menghentikan pemindaian (PRD §4.3)
				res.Warnings = append(res.Warnings, fmt.Sprintf("cannot access path %q: %v", currentPath, err))
				if d != nil && d.IsDir() && currentPath != cleanRoot {
					return filepath.SkipDir
				}
				return nil
			}

			if currentPath == cleanRoot {
				res.ScannedDirsCount++
				return nil
			}

			// Periksa apakah ini symlink
			if d.Type()&fs.ModeSymlink != 0 {
				targetInfo, err := os.Stat(currentPath)
				if err != nil {
					res.Warnings = append(res.Warnings, fmt.Sprintf("cannot resolve symlink %q: %v", currentPath, err))
					return nil
				}
				if targetInfo.IsDir() {
					// R-D6: Symlink direktori tidak diikuti
					res.Warnings = append(res.Warnings, fmt.Sprintf("directory symlink skipped: %q", currentPath))
					return nil
				}
				// Symlink berkas boleh disalin isinya (R-D6)
				processFileEntry(cleanRoot, currentPath, d.Name(), targetInfo, &candidateFiles, xmpMap)
				return nil
			}

			if d.IsDir() {
				res.ScannedDirsCount++
				return nil
			}

			fileInfo, err := d.Info()
			if err != nil {
				res.Warnings = append(res.Warnings, fmt.Sprintf("cannot read file info for %q: %v", currentPath, err))
				return nil
			}

			processFileEntry(cleanRoot, currentPath, d.Name(), fileInfo, &candidateFiles, xmpMap)
			return nil
		})

		if walkErr != nil {
			return Result{}, fmt.Errorf("error during directory scan: %w", walkErr)
		}
	}

	// Hubungkan berkas foto dengan berkas pendamping XMP yang sesuai
	for i := range candidateFiles {
		file := &candidateFiles[i]
		dir := filepath.Dir(file.Path)
		if dirXMPs, ok := xmpMap[dir]; ok {
			baseKey := strings.ToLower(file.BaseName)
			if xmpPath, exists := dirXMPs[baseKey]; exists {
				file.HasXMP = true
				file.XMPPath = xmpPath
			}
		}

		if file.IsRAW {
			res.RAWCount++
		} else {
			res.NonRAWCount++
		}
		res.Files = append(res.Files, *file)
	}

	return res, nil
}

func processFileEntry(
	root string,
	fullPath string,
	fileName string,
	info fs.FileInfo,
	candidates *[]FileEntry,
	xmpMap map[string]map[string]string,
) {
	relPath, err := filepath.Rel(root, fullPath)
	if err != nil {
		relPath = fileName
	}

	ext := filepath.Ext(fileName)
	baseName := fileName
	if ext != "" {
		baseName = strings.TrimSuffix(fileName, ext)
	}

	dir := filepath.Dir(fullPath)
	if IsXMP(ext) {
		// Catat berkas XMP untuk linking sidecar
		if _, ok := xmpMap[dir]; !ok {
			xmpMap[dir] = make(map[string]string)
		}
		// Dukung format DSC_0001.xmp dan DSC_0001.CR3.xmp
		baseKey := strings.ToLower(baseName)
		xmpMap[dir][baseKey] = fullPath
		// Jika baseName berformat DSC_0001.CR3, simpan juga kunci DSC_0001
		innerExt := filepath.Ext(baseName)
		if innerExt != "" {
			innerBase := strings.TrimSuffix(baseName, innerExt)
			xmpMap[dir][strings.ToLower(innerBase)] = fullPath
		}
	}

	isRaw := IsRAW(ext)

	*candidates = append(*candidates, FileEntry{
		Path:     fullPath,
		RelPath:  relPath,
		FileName: fileName,
		BaseName: baseName,
		Ext:      ext,
		Size:     info.Size(),
		ModTime:  info.ModTime(),
		IsRAW:    isRaw,
	})
}
