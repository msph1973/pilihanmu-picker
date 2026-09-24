package fsops

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/msph1973/pilihanmu-picker/internal/match"
	"github.com/msph1973/pilihanmu-picker/internal/scan"
)

// OpMode mendefinisikan mode eksekusi: Salin atau Pindahkan.
type OpMode string

const (
	// ModeCopy adalah mode Salin: folder sumber tidak pernah disentuh sama sekali (R-P3).
	ModeCopy OpMode = "Salin"

	// ModeMove adalah mode Pindahkan: sumber dihapus per file hanya setelah verifikasi ukuran berhasil (E-2, E-3).
	ModeMove OpMode = "Pindahkan"
)

// Windows reserved device names (R-D3)
var reservedWindowsNames = map[string]struct{}{
	"CON": {}, "PRN": {}, "AUX": {}, "NUL": {},
	"COM1": {}, "COM2": {}, "COM3": {}, "COM4": {}, "COM5": {},
	"COM6": {}, "COM7": {}, "COM8": {}, "COM9": {},
	"LPT1": {}, "LPT2": {}, "LPT3": {}, "LPT4": {}, "LPT5": {},
	"LPT6": {}, "LPT7": {}, "LPT8": {}, "LPT9": {},
}

// Progress mencatat status progres pengerjaan terkini.
type Progress struct {
	CurrentFileIndex int
	TotalFiles       int
	CurrentFileName  string
	BytesProcessed   int64
	TotalBytes       int64
	IsComplete       bool
}

// FileRecord mencatat berkas yang berhasil diproses.
type FileRecord struct {
	SourcePath    string
	SourceRelPath string
	DestPath      string
	DestFileName  string
	Size          int64
	HadCollision  bool
	OriginalName  string
	IsSidecarXMP  bool
	SourceDeleted bool
}

// FailureRecord mencatat berkas yang gagal diproses beserta alasannya (R-U6).
type FailureRecord struct {
	SourcePath string
	FileName   string
	Error      error
	Reason     string
}

// CollisionRecord mencatat detail tabrakan nama dan resolusi penamaannya (PRD §4.7).
type CollisionRecord struct {
	SourceRelPath    string
	OriginalFileName string
	ResolvedFileName string
}

// Result menyimpan hasil eksekusi keseluruhan operasi fsops.
type Result struct {
	Mode             OpMode
	SourceDir        string
	DestDir          string
	CopiedFiles      []FileRecord
	FailedFiles      []FailureRecord
	CollisionRecords []CollisionRecord
	TotalBytesCopied int64
	Cancelled        bool
	StartTime        time.Time
	EndTime          time.Time
}

// Options menyimpan konfigurasi eksekusi I/O.
type Options struct {
	Mode             OpMode
	SourceDir        string
	DestDir          string
	MatchedItems     []match.MatchItem
	ProgressCallback func(p Progress)
	// RetryDelays memungkinkan override jeda retry untuk pengujian cepat (default: 250ms, 500ms, 1000ms)
	RetryDelays []time.Duration
}

// NormalizePath menyiapkan path untuk OS, termasuk prefiks path panjang Windows jika diperlukan (R-D1).
func NormalizePath(p string) string {
	clean := filepath.Clean(p)
	if runtime.GOOS == "windows" {
		abs, err := filepath.Abs(clean)
		if err == nil {
			clean = abs
		}
		if len(clean) >= 260 && !strings.HasPrefix(clean, `\\?\`) {
			// Format \\?\ hanya untuk path absolut dengan pemisah \
			clean = `\\?\` + strings.ReplaceAll(clean, `/`, `\`)
		}
	}
	return clean
}

// ValidateWindowsFileName memvalidasi nama berkas terhadap karakter dan nama reserved Windows (R-D3).
func ValidateWindowsFileName(name string) error {
	// Karakter terlarang di Windows: < > : " / \ | ? *
	forbiddenChars := `<>:"/\|?*`
	for _, c := range forbiddenChars {
		if strings.ContainsRune(name, c) {
			return fmt.Errorf("filename contains forbidden character %q", string(c))
		}
	}

	base := strings.TrimSuffix(name, filepath.Ext(name))
	upperBase := strings.ToUpper(strings.TrimSpace(base))
	if _, ok := reservedWindowsNames[upperBase]; ok {
		return fmt.Errorf("filename uses reserved Windows device name %q", upperBase)
	}

	return nil
}

// Execute menjalankan proses penyalinan atau pemindahan berkas ke folder tujuan (PRD §4.6).
func Execute(ctx context.Context, opts Options) (Result, error) {
	if opts.SourceDir == "" || opts.DestDir == "" {
		return Result{}, errors.New("source and destination directories cannot be empty")
	}

	res := Result{
		Mode:             opts.Mode,
		SourceDir:        opts.SourceDir,
		DestDir:          opts.DestDir,
		CopiedFiles:      make([]FileRecord, 0),
		FailedFiles:      make([]FailureRecord, 0),
		CollisionRecords: make([]CollisionRecord, 0),
		StartTime:        time.Now(),
	}

	cleanDest := NormalizePath(opts.DestDir)
	if err := os.MkdirAll(cleanDest, 0755); err != nil {
		return res, fmt.Errorf("failed to create destination directory %q: %w", opts.DestDir, err)
	}

	retryDelays := opts.RetryDelays
	if len(retryDelays) == 0 {
		retryDelays = []time.Duration{250 * time.Millisecond, 500 * time.Millisecond, 1000 * time.Millisecond}
	}

	// Kumpulkan seluruh item yang akan diproses (foto RAW + sidecar XMP)
	type workUnit struct {
		fileEntry    scan.FileEntry
		isSidecar    bool
		parentRAWRel string
	}

	var queue []workUnit
	var totalBytes int64

	for _, item := range opts.MatchedItems {
		for _, f := range item.Files {
			queue = append(queue, workUnit{
				fileEntry: f,
				isSidecar: false,
			})
			totalBytes += f.Size

			if f.HasXMP && f.XMPPath != "" {
				xmpInfo, err := os.Stat(NormalizePath(f.XMPPath))
				var xmpSize int64
				if err == nil {
					xmpSize = xmpInfo.Size()
				}
				xmpEntry := scan.FileEntry{
					Path:     f.XMPPath,
					RelPath:  filepath.Join(filepath.Dir(f.RelPath), filepath.Base(f.XMPPath)),
					FileName: filepath.Base(f.XMPPath),
					BaseName: strings.TrimSuffix(filepath.Base(f.XMPPath), filepath.Ext(f.XMPPath)),
					Ext:      filepath.Ext(f.XMPPath),
					Size:     xmpSize,
					IsRAW:    false,
				}
				queue = append(queue, workUnit{
					fileEntry:    xmpEntry,
					isSidecar:    true,
					parentRAWRel: f.RelPath,
				})
				totalBytes += xmpSize
			}
		}
	}

	totalFiles := len(queue)
	var processedBytes int64

	// Peta untuk melacak nama berkas yang sudah ada di folder tujuan
	destNamesMap := make(map[string]struct{})

	// Isi awal dari folder tujuan untuk mencegah menimpa berkas yang sudah ada (R-P5)
	if existingEntries, err := os.ReadDir(cleanDest); err == nil {
		for _, e := range existingEntries {
			destNamesMap[strings.ToLower(e.Name())] = struct{}{}
		}
	}

	// Buffer streaming (64 KB)
	buf := make([]byte, 64*1024)

	for idx, unit := range queue {
		// Cek pembatalan (E-7, R-U3)
		if ctx.Err() != nil {
			res.Cancelled = true
			break
		}

		f := unit.fileEntry

		if opts.ProgressCallback != nil {
			opts.ProgressCallback(Progress{
				CurrentFileIndex: idx + 1,
				TotalFiles:       totalFiles,
				CurrentFileName:  f.FileName,
				BytesProcessed:   processedBytes,
				TotalBytes:       totalBytes,
				IsComplete:       false,
			})
		}

		// Validasi nama Windows (R-D3)
		if err := ValidateWindowsFileName(f.FileName); err != nil {
			res.FailedFiles = append(res.FailedFiles, FailureRecord{
				SourcePath: f.Path,
				FileName:   f.FileName,
				Error:      err,
				Reason:     fmt.Sprintf("Nama berkas terlarang di Windows: %v", err),
			})
			continue
		}

		// Tentukan nama tujuan bebas tabrakan (R-P5, E-4)
		destFileName, hadCollision := getUniqueDestinationName(cleanDest, f.FileName, destNamesMap)
		destFilePath := filepath.Join(cleanDest, destFileName)

		// Lakukan operasi salin streaming dengan mekanisme retry jika terkunci (E-1, E-8, R-D4)
		copyErr := copyWithRetry(ctx, f.Path, destFilePath, f.Size, retryDelays, buf)
		if copyErr != nil {
			res.FailedFiles = append(res.FailedFiles, FailureRecord{
				SourcePath: f.Path,
				FileName:   f.FileName,
				Error:      copyErr,
				Reason:     fmt.Sprintf("Gagal menyalin: %v", copyErr),
			})
			continue
		}

		// Verifikasi ukuran berkas tujuan (E-2, R-P3, R-P5)
		destInfo, err := os.Stat(destFilePath)
		if err != nil || destInfo.Size() != f.Size {
			_ = os.Remove(destFilePath) // Bersihkan tujuan yang tidak valid
			res.FailedFiles = append(res.FailedFiles, FailureRecord{
				SourcePath: f.Path,
				FileName:   f.FileName,
				Error:      fmt.Errorf("size verification failed: expected %d bytes, got %d", f.Size, destInfo.Size()),
				Reason:     "Verifikasi ukuran gagal: ukuran berkas tujuan tidak cocok dengan sumber",
			})
			continue
		}

		sourceDeleted := false
		// Mode Pindahkan: hapus sumber per berkas hanya setelah verifikasi berhasil (E-3, R-P3)
		if opts.Mode == ModeMove {
			rmErr := os.Remove(NormalizePath(f.Path))
			if rmErr != nil {
				res.FailedFiles = append(res.FailedFiles, FailureRecord{
					SourcePath: f.Path,
					FileName:   f.FileName,
					Error:      fmt.Errorf("failed to remove source file after move: %w", rmErr),
					Reason:     fmt.Sprintf("Berkas tersalin tetapi gagal menghapus sumber: %v", rmErr),
				})
			} else {
				sourceDeleted = true
			}
		}

		destNamesMap[strings.ToLower(destFileName)] = struct{}{}
		processedBytes += f.Size
		res.TotalBytesCopied += f.Size

		fileRecord := FileRecord{
			SourcePath:    f.Path,
			SourceRelPath: f.RelPath,
			DestPath:      destFilePath,
			DestFileName:  destFileName,
			Size:          f.Size,
			HadCollision:  hadCollision,
			OriginalName:  f.FileName,
			IsSidecarXMP:  unit.isSidecar,
			SourceDeleted: sourceDeleted,
		}
		res.CopiedFiles = append(res.CopiedFiles, fileRecord)

		if hadCollision {
			res.CollisionRecords = append(res.CollisionRecords, CollisionRecord{
				SourceRelPath:    f.RelPath,
				OriginalFileName: f.FileName,
				ResolvedFileName: destFileName,
			})
		}
	}

	res.EndTime = time.Now()

	if opts.ProgressCallback != nil {
		opts.ProgressCallback(Progress{
			CurrentFileIndex: len(res.CopiedFiles),
			TotalFiles:       totalFiles,
			BytesProcessed:   processedBytes,
			TotalBytes:       totalBytes,
			IsComplete:       true,
		})
	}

	return res, nil
}

// getUniqueDestinationName memeriksa apakah nama berkas sudah ada di folder tujuan,
// dan jika ada menghasilkan "nama (2).EXT", "nama (3).EXT", dst (E-4, R-P5).
func getUniqueDestinationName(destDir string, fileName string, trackedNames map[string]struct{}) (string, bool) {
	lowerName := strings.ToLower(fileName)
	if _, inMap := trackedNames[lowerName]; !inMap {
		// Cek juga langsung di disk
		targetPath := filepath.Join(destDir, fileName)
		if _, err := os.Stat(targetPath); errors.Is(err, os.ErrNotExist) {
			return fileName, false
		}
	}

	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)

	index := 2
	for {
		candidate := fmt.Sprintf("%s (%d)%s", base, index, ext)
		lowerCandidate := strings.ToLower(candidate)
		if _, inMap := trackedNames[lowerCandidate]; !inMap {
			targetPath := filepath.Join(destDir, candidate)
			if _, err := os.Stat(targetPath); errors.Is(err, os.ErrNotExist) {
				return candidate, true
			}
		}
		index++
	}
}

// copyWithRetry melakukan penyalinan streaming dengan perulangan coba ulang jika berkas terkunci (E-8, R-D4, R-A5).
func copyWithRetry(
	ctx context.Context,
	srcPath string,
	destPath string,
	expectedSize int64,
	delays []time.Duration,
	buf []byte,
) error {
	var lastErr error

	for attempt := 0; attempt <= len(delays); attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := copyFileStreaming(ctx, srcPath, destPath, buf)
		if err == nil {
			return nil
		}

		lastErr = err

		// Bersihkan file parsial di tujuan jika gagal (E-7)
		_ = os.Remove(destPath)

		if attempt < len(delays) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delays[attempt]):
			}
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", len(delays)+1, lastErr)
}

// copyFileStreaming mengalirkan data dari berkas sumber ke berkas tujuan menggunakan buffer (R-A5).
func copyFileStreaming(ctx context.Context, srcPath string, destPath string, buf []byte) error {
	src, err := os.Open(NormalizePath(srcPath))
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer src.Close()

	dest, err := os.OpenFile(NormalizePath(destPath), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dest.Close()

	// Gunakan CopyBuffer untuk menghindari pemuatan seluruh isi berkas ke RAM (R-A5)
	_, copyErr := io.CopyBuffer(dest, src, buf)
	if copyErr != nil {
		return fmt.Errorf("error streaming data: %w", copyErr)
	}

	if err := dest.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}
