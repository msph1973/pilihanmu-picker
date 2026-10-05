package fsops

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/msph1973/pilihanmu-picker/internal/match"
	"github.com/msph1973/pilihanmu-picker/internal/scan"
)

func TestCopyMode(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	file1Path := filepath.Join(srcDir, "DSC_0001.CR3")
	content1 := []byte("RAW_FILE_CONTENT_1234567890")
	if err := os.WriteFile(file1Path, content1, 0644); err != nil {
		t.Fatal(err)
	}

	srcInfoBefore, err := os.Stat(file1Path)
	if err != nil {
		t.Fatal(err)
	}

	matched := []match.MatchItem{
		{
			TargetName: "DSC_0001",
			Files: []scan.FileEntry{
				{
					Path:     file1Path,
					RelPath:  "DSC_0001.CR3",
					FileName: "DSC_0001.CR3",
					BaseName: "DSC_0001",
					Ext:      ".CR3",
					Size:     int64(len(content1)),
					IsRAW:    true,
				},
			},
		},
	}

	res, err := Execute(context.Background(), Options{
		Mode:         ModeCopy,
		SourceDir:    srcDir,
		DestDir:      destDir,
		MatchedItems: matched,
		RetryDelays:  []time.Duration{1 * time.Millisecond},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(res.CopiedFiles) != 1 {
		t.Fatalf("expected 1 copied file, got %d", len(res.CopiedFiles))
	}

	// Verifikasi file tujuan ada dengan isi sama
	destFilePath := filepath.Join(destDir, "DSC_0001.CR3")
	destData, err := os.ReadFile(destFilePath)
	if err != nil {
		t.Fatalf("failed to read dest file: %v", err)
	}
	if string(destData) != string(content1) {
		t.Errorf("destination content mismatch")
	}

	// Verifikasi R-P3: File sumber TIDAK disentuh sama sekali pada mode Salin
	srcInfoAfter, err := os.Stat(file1Path)
	if err != nil {
		t.Fatalf("source file was removed in Copy mode: %v", err)
	}
	if srcInfoAfter.Size() != srcInfoBefore.Size() || srcInfoAfter.ModTime() != srcInfoBefore.ModTime() {
		t.Errorf("source file was modified in Copy mode")
	}
}

func TestMoveMode(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	file1Path := filepath.Join(srcDir, "DSC_0001.CR3")
	content1 := []byte("RAW_FILE_CONTENT_TO_MOVE")
	if err := os.WriteFile(file1Path, content1, 0644); err != nil {
		t.Fatal(err)
	}

	matched := []match.MatchItem{
		{
			TargetName: "DSC_0001",
			Files: []scan.FileEntry{
				{
					Path:     file1Path,
					RelPath:  "DSC_0001.CR3",
					FileName: "DSC_0001.CR3",
					BaseName: "DSC_0001",
					Ext:      ".CR3",
					Size:     int64(len(content1)),
					IsRAW:    true,
				},
			},
		},
	}

	res, err := Execute(context.Background(), Options{
		Mode:         ModeMove,
		SourceDir:    srcDir,
		DestDir:      destDir,
		MatchedItems: matched,
		RetryDelays:  []time.Duration{1 * time.Millisecond},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(res.CopiedFiles) != 1 {
		t.Fatalf("expected 1 file moved, got %d", len(res.CopiedFiles))
	}
	if !res.CopiedFiles[0].SourceDeleted {
		t.Errorf("expected SourceDeleted = true in Move mode")
	}

	// Verifikasi tujuan ada
	destData, err := os.ReadFile(filepath.Join(destDir, "DSC_0001.CR3"))
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if string(destData) != string(content1) {
		t.Errorf("destination content mismatch")
	}

	// Verifikasi sumber terhapus (E-3)
	if _, err := os.Stat(file1Path); !os.IsNotExist(err) {
		t.Errorf("expected source file to be deleted in Move mode")
	}
}

func TestCollisionResolution(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	subA := filepath.Join(srcDir, "subA")
	subB := filepath.Join(srcDir, "subB")
	_ = os.MkdirAll(subA, 0755)
	_ = os.MkdirAll(subB, 0755)

	pathA := filepath.Join(subA, "DSC_0001.CR3")
	pathB := filepath.Join(subB, "DSC_0001.CR3")

	_ = os.WriteFile(pathA, []byte("CONTENT_A"), 0644)
	_ = os.WriteFile(pathB, []byte("CONTENT_B"), 0644)

	matched := []match.MatchItem{
		{
			TargetName: "DSC_0001",
			Files: []scan.FileEntry{
				{
					Path:     pathA,
					RelPath:  "subA/DSC_0001.CR3",
					FileName: "DSC_0001.CR3",
					Size:     int64(len("CONTENT_A")),
					IsRAW:    true,
				},
				{
					Path:     pathB,
					RelPath:  "subB/DSC_0001.CR3",
					FileName: "DSC_0001.CR3",
					Size:     int64(len("CONTENT_B")),
					IsRAW:    true,
				},
			},
		},
	}

	res, err := Execute(context.Background(), Options{
		Mode:         ModeCopy,
		SourceDir:    srcDir,
		DestDir:      destDir,
		MatchedItems: matched,
		RetryDelays:  []time.Duration{1 * time.Millisecond},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(res.CopiedFiles) != 2 {
		t.Fatalf("expected 2 files copied, got %d", len(res.CopiedFiles))
	}
	if len(res.CollisionRecords) != 1 {
		t.Fatalf("expected 1 collision record, got %d", len(res.CollisionRecords))
	}

	// Cek kedua file di tujuan: DSC_0001.CR3 dan DSC_0001 (2).CR3 (R-P5, E-4)
	file1 := filepath.Join(destDir, "DSC_0001.CR3")
	file2 := filepath.Join(destDir, "DSC_0001 (2).CR3")

	data1, err := os.ReadFile(file1)
	if err != nil {
		t.Fatalf("file1 missing: %v", err)
	}
	data2, err := os.ReadFile(file2)
	if err != nil {
		t.Fatalf("file2 missing: %v", err)
	}

	if string(data1) != "CONTENT_A" {
		t.Errorf("file1 content mismatch")
	}
	if string(data2) != "CONTENT_B" {
		t.Errorf("file2 content mismatch")
	}
}

func TestSidecarXMPCopying(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	rawPath := filepath.Join(srcDir, "DSC_0001.CR3")
	xmpPath := filepath.Join(srcDir, "DSC_0001.xmp")

	_ = os.WriteFile(rawPath, []byte("RAW_DATA"), 0644)
	_ = os.WriteFile(xmpPath, []byte("<xmp_data/>"), 0644)

	matched := []match.MatchItem{
		{
			TargetName: "DSC_0001",
			Files: []scan.FileEntry{
				{
					Path:     rawPath,
					RelPath:  "DSC_0001.CR3",
					FileName: "DSC_0001.CR3",
					Size:     int64(len("RAW_DATA")),
					IsRAW:    true,
					HasXMP:   true,
					XMPPath:  xmpPath,
				},
			},
		},
	}

	res, err := Execute(context.Background(), Options{
		Mode:         ModeCopy,
		SourceDir:    srcDir,
		DestDir:      destDir,
		MatchedItems: matched,
		RetryDelays:  []time.Duration{1 * time.Millisecond},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(res.CopiedFiles) != 2 {
		t.Fatalf("expected 2 copied files (RAW + XMP), got %d", len(res.CopiedFiles))
	}

	// Cek keberadaan XMP di folder tujuan
	destXMP := filepath.Join(destDir, "DSC_0001.xmp")
	xmpData, err := os.ReadFile(destXMP)
	if err != nil {
		t.Fatalf("destination xmp not found: %v", err)
	}
	if string(xmpData) != "<xmp_data/>" {
		t.Errorf("xmp content mismatch")
	}
}

func TestWindowsReservedNames(t *testing.T) {
	err := ValidateWindowsFileName("CON.CR3")
	if err == nil {
		t.Error("expected error for reserved name CON")
	}

	err = ValidateWindowsFileName("photo?.CR3")
	if err == nil {
		t.Error("expected error for forbidden character '?'")
	}

	err = ValidateWindowsFileName("normal_DSC_0001.CR3")
	if err != nil {
		t.Errorf("unexpected error for normal filename: %v", err)
	}
}

func TestCancellationAndProgress(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	rawPath := filepath.Join(srcDir, "DSC_0001.CR3")
	_ = os.WriteFile(rawPath, []byte("DATA"), 0644)

	matched := []match.MatchItem{
		{
			TargetName: "DSC_0001",
			Files: []scan.FileEntry{
				{
					Path:     rawPath,
					RelPath:  "DSC_0001.CR3",
					FileName: "DSC_0001.CR3",
					Size:     4,
					IsRAW:    true,
				},
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // langsung batalkan

	progressReported := false
	res, err := Execute(ctx, Options{
		Mode:         ModeCopy,
		SourceDir:    srcDir,
		DestDir:      destDir,
		MatchedItems: matched,
		ProgressCallback: func(p Progress) {
			progressReported = true
		},
		RetryDelays: []time.Duration{1 * time.Millisecond},
	})
	if err != nil {
		t.Fatalf("unexpected Execute error: %v", err)
	}

	if !res.Cancelled {
		t.Error("expected result.Cancelled = true")
	}
	_ = progressReported
}

func TestErrorsAndEdgeCases(t *testing.T) {
	t.Run("Empty directories", func(t *testing.T) {
		_, err := Execute(context.Background(), Options{})
		if err == nil {
			t.Error("expected error for empty options")
		}
	})

	t.Run("Source file does not exist (retry fail)", func(t *testing.T) {
		srcDir := t.TempDir()
		destDir := t.TempDir()

		matched := []match.MatchItem{
			{
				TargetName: "MISSING",
				Files: []scan.FileEntry{
					{
						Path:     filepath.Join(srcDir, "MISSING.CR3"),
						FileName: "MISSING.CR3",
						Size:     100,
						IsRAW:    true,
					},
				},
			},
		}

		res, err := Execute(context.Background(), Options{
			Mode:         ModeCopy,
			SourceDir:    srcDir,
			DestDir:      destDir,
			MatchedItems: matched,
			RetryDelays:  []time.Duration{1 * time.Millisecond, 2 * time.Millisecond},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.FailedFiles) != 1 {
			t.Errorf("expected 1 failed file record, got %d", len(res.FailedFiles))
		}
	})

	t.Run("Size verification mismatch does not delete source in Move mode", func(t *testing.T) {
		srcDir := t.TempDir()
		destDir := t.TempDir()

		srcFile := filepath.Join(srcDir, "CORRUPT.CR3")
		_ = os.WriteFile(srcFile, []byte("REAL_DATA"), 0644)

		// FileEntry mencatat ukuran yang salah (misal 9999 byte padahal 9 byte)
		matched := []match.MatchItem{
			{
				TargetName: "CORRUPT",
				Files: []scan.FileEntry{
					{
						Path:     srcFile,
						FileName: "CORRUPT.CR3",
						Size:     9999,
						IsRAW:    true,
					},
				},
			},
		}

		res, err := Execute(context.Background(), Options{
			Mode:         ModeMove,
			SourceDir:    srcDir,
			DestDir:      destDir,
			MatchedItems: matched,
			RetryDelays:  []time.Duration{1 * time.Millisecond},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Harus tercatat gagal
		if len(res.FailedFiles) != 1 {
			t.Fatalf("expected 1 failed file, got %d", len(res.FailedFiles))
		}

		// R-P3 & E-2: File sumber TIDAK boleh terhapus jika verifikasi gagal!
		if _, err := os.Stat(srcFile); os.IsNotExist(err) {
			t.Errorf("source file was deleted despite size verification failure!")
		}
	})

	t.Run("NormalizePath", func(t *testing.T) {
		shortPath := "/tmp/file.txt"
		normalized := NormalizePath(shortPath)
		if normalized == "" {
			t.Error("normalized path is empty")
		}

		longPath := "C:\\" + strings.Repeat("sub\\", 70) + "file.CR3"
		normLong := NormalizePath(longPath)
		if normLong == "" {
			t.Error("normLong is empty")
		}
	})
}

// TestSidecarXMPInheritsCollisionSuffix membuktikan bahwa sidecar .xmp harus mewarisi akhiran
// tabrakan dari RAW induknya (PRD §4.6 E-5, K-10), bukan diberi penomoran independen.
//
// Catatan penting: skenario "2 sub-folder simetris" (subA/subB masing-masing punya RAW+XMP
// dengan nama identik) TIDAK cukup untuk membuktikan bug ini, karena pada kasus simetris murni,
// penomoran independen RAW dan XMP kebetulan tetap sejajar (keduanya sama-sama tabrakan pada
// percobaan ke-2). Bug baru benar-benar terlihat ketika distribusi sidecar TIDAK simetris —
// di sini subB sengaja TIDAK punya sidecar .xmp, sehingga penomoran independen XMP (yang hanya
// menghitung kemunculan "DSC_0001.xmp") akan tertinggal satu langkah dari penomoran independen
// RAW (yang menghitung kemunculan "DSC_0001.CR3", termasuk RAW_B yang tidak punya sidecar).
// Diverifikasi secara empiris: kode lama menghasilkan XMP_C -> "DSC_0001 (2).xmp" padahal
// RAW_C -> "DSC_0001 (3).CR3" (salah pasang); kode baru menghasilkan "DSC_0001 (3).xmp" (benar).
func TestSidecarXMPInheritsCollisionSuffix(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	subA := filepath.Join(srcDir, "subA")
	subB := filepath.Join(srcDir, "subB")
	subC := filepath.Join(srcDir, "subC")
	_ = os.MkdirAll(subA, 0755)
	_ = os.MkdirAll(subB, 0755)
	_ = os.MkdirAll(subC, 0755)

	rawA := filepath.Join(subA, "DSC_0001.CR3")
	xmpA := filepath.Join(subA, "DSC_0001.xmp")
	rawB := filepath.Join(subB, "DSC_0001.CR3") // subB sengaja TANPA sidecar .xmp
	rawC := filepath.Join(subC, "DSC_0001.CR3")
	xmpC := filepath.Join(subC, "DSC_0001.xmp")

	_ = os.WriteFile(rawA, []byte("RAW_CONTENT_A"), 0644)
	_ = os.WriteFile(xmpA, []byte("<xmp_data_A/>"), 0644)
	_ = os.WriteFile(rawB, []byte("RAW_CONTENT_B"), 0644)
	_ = os.WriteFile(rawC, []byte("RAW_CONTENT_C"), 0644)
	_ = os.WriteFile(xmpC, []byte("<xmp_data_C/>"), 0644)

	matched := []match.MatchItem{
		{
			TargetName: "DSC_0001",
			Files: []scan.FileEntry{
				{
					Path:     rawA,
					RelPath:  "subA/DSC_0001.CR3",
					FileName: "DSC_0001.CR3",
					Size:     int64(len("RAW_CONTENT_A")),
					IsRAW:    true,
					HasXMP:   true,
					XMPPath:  xmpA,
				},
				{
					Path:     rawB,
					RelPath:  "subB/DSC_0001.CR3",
					FileName: "DSC_0001.CR3",
					Size:     int64(len("RAW_CONTENT_B")),
					IsRAW:    true,
				},
				{
					Path:     rawC,
					RelPath:  "subC/DSC_0001.CR3",
					FileName: "DSC_0001.CR3",
					Size:     int64(len("RAW_CONTENT_C")),
					IsRAW:    true,
					HasXMP:   true,
					XMPPath:  xmpC,
				},
			},
		},
	}

	res, err := Execute(context.Background(), Options{
		Mode:         ModeCopy,
		SourceDir:    srcDir,
		DestDir:      destDir,
		MatchedItems: matched,
		RetryDelays:  []time.Duration{1 * time.Millisecond},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(res.FailedFiles) != 0 {
		t.Fatalf("expected no failures, got %+v", res.FailedFiles)
	}
	if len(res.CopiedFiles) != 5 {
		t.Fatalf("expected 5 copied files (3 RAW + 2 XMP), got %d", len(res.CopiedFiles))
	}

	rawAPath := filepath.Join(destDir, "DSC_0001.CR3")
	rawBPath := filepath.Join(destDir, "DSC_0001 (2).CR3")
	rawCPath := filepath.Join(destDir, "DSC_0001 (3).CR3")
	xmpAPath := filepath.Join(destDir, "DSC_0001.xmp")
	xmpCPath := filepath.Join(destDir, "DSC_0001 (3).xmp") // HARUS (3) sejajar dengan RAW_C, bukan (2)

	for _, p := range []string{rawAPath, rawBPath, rawCPath, xmpAPath, xmpCPath} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected file missing: %s: %v", p, err)
		}
	}

	rawCData, err := os.ReadFile(rawCPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", rawCPath, err)
	}
	xmpCData, err := os.ReadFile(xmpCPath)
	if err != nil {
		t.Fatalf("failed to read %s (sidecar harus mengikuti akhiran RAW induknya): %v", xmpCPath, err)
	}

	if string(rawCData) != "RAW_CONTENT_C" {
		t.Errorf("rawC content mismatch: got %q", rawCData)
	}
	if string(xmpCData) != "<xmp_data_C/>" {
		t.Errorf("xmpC content mismatch: got %q, pasangan RAW-XMP tidak konsisten", xmpCData)
	}

	xmpAData, err := os.ReadFile(xmpAPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", xmpAPath, err)
	}
	if string(xmpAData) != "<xmp_data_A/>" {
		t.Errorf("xmpA content mismatch: got %q", xmpAData)
	}
}

// TestSidecarXMPDoubleExtensionInheritsCollisionSuffix memverifikasi pola penamaan sidecar dengan
// ekstensi ganda "DSC_0001.CR3.xmp" (bukan hanya "DSC_0001.xmp") tetap mewarisi akhiran tabrakan
// RAW induknya dengan benar, menghasilkan "DSC_0001 (3).CR3.xmp".
func TestSidecarXMPDoubleExtensionInheritsCollisionSuffix(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	subA := filepath.Join(srcDir, "subA")
	subB := filepath.Join(srcDir, "subB")
	subC := filepath.Join(srcDir, "subC")
	_ = os.MkdirAll(subA, 0755)
	_ = os.MkdirAll(subB, 0755)
	_ = os.MkdirAll(subC, 0755)

	rawA := filepath.Join(subA, "DSC_0001.CR3")
	xmpA := filepath.Join(subA, "DSC_0001.CR3.xmp")
	rawB := filepath.Join(subB, "DSC_0001.CR3") // tanpa sidecar
	rawC := filepath.Join(subC, "DSC_0001.CR3")
	xmpC := filepath.Join(subC, "DSC_0001.CR3.xmp")

	_ = os.WriteFile(rawA, []byte("RAW_CONTENT_A"), 0644)
	_ = os.WriteFile(xmpA, []byte("<xmp_double_A/>"), 0644)
	_ = os.WriteFile(rawB, []byte("RAW_CONTENT_B"), 0644)
	_ = os.WriteFile(rawC, []byte("RAW_CONTENT_C"), 0644)
	_ = os.WriteFile(xmpC, []byte("<xmp_double_C/>"), 0644)

	matched := []match.MatchItem{
		{
			TargetName: "DSC_0001",
			Files: []scan.FileEntry{
				{
					Path:     rawA,
					RelPath:  "subA/DSC_0001.CR3",
					FileName: "DSC_0001.CR3",
					Size:     int64(len("RAW_CONTENT_A")),
					IsRAW:    true,
					HasXMP:   true,
					XMPPath:  xmpA,
				},
				{
					Path:     rawB,
					RelPath:  "subB/DSC_0001.CR3",
					FileName: "DSC_0001.CR3",
					Size:     int64(len("RAW_CONTENT_B")),
					IsRAW:    true,
				},
				{
					Path:     rawC,
					RelPath:  "subC/DSC_0001.CR3",
					FileName: "DSC_0001.CR3",
					Size:     int64(len("RAW_CONTENT_C")),
					IsRAW:    true,
					HasXMP:   true,
					XMPPath:  xmpC,
				},
			},
		},
	}

	res, err := Execute(context.Background(), Options{
		Mode:         ModeCopy,
		SourceDir:    srcDir,
		DestDir:      destDir,
		MatchedItems: matched,
		RetryDelays:  []time.Duration{1 * time.Millisecond},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(res.FailedFiles) != 0 {
		t.Fatalf("expected no failures, got %+v", res.FailedFiles)
	}

	xmpCPath := filepath.Join(destDir, "DSC_0001 (3).CR3.xmp")
	data, err := os.ReadFile(xmpCPath)
	if err != nil {
		t.Fatalf("expected %s to exist: %v", xmpCPath, err)
	}
	if string(data) != "<xmp_double_C/>" {
		t.Errorf("xmp double-extension content mismatch: got %q", data)
	}
}

// TestListDestFileNames memverifikasi helper ListDestFileNames (R-A6: satu-satunya pintu I/O
// berkas untuk kebutuhan pratinjau pencocokan di lapisan UI).
func TestListDestFileNames(t *testing.T) {
	t.Run("destination directory does not exist", func(t *testing.T) {
		base := t.TempDir()
		missing := filepath.Join(base, "does-not-exist")
		names, err := ListDestFileNames(missing)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 0 {
			t.Errorf("expected empty slice, got %v", names)
		}
	})

	t.Run("mixed files and subfolders", func(t *testing.T) {
		destDir := t.TempDir()
		_ = os.WriteFile(filepath.Join(destDir, "DSC_0001.CR3"), []byte("a"), 0644)
		_ = os.WriteFile(filepath.Join(destDir, "DSC_0002.CR3"), []byte("b"), 0644)
		_ = os.MkdirAll(filepath.Join(destDir, "subfolder"), 0755)

		names, err := ListDestFileNames(destDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(names) != 2 {
			t.Fatalf("expected 2 file names, got %d: %v", len(names), names)
		}
		found := map[string]bool{}
		for _, n := range names {
			found[n] = true
		}
		if !found["DSC_0001.CR3"] || !found["DSC_0002.CR3"] {
			t.Errorf("expected DSC_0001.CR3 and DSC_0002.CR3 in result, got %v", names)
		}
		if found["subfolder"] {
			t.Errorf("subfolder should not be included in result")
		}
	})

	t.Run("empty destDir argument", func(t *testing.T) {
		names, err := ListDestFileNames("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if names != nil {
			t.Errorf("expected nil slice for empty destDir, got %v", names)
		}
	})
}

// TestNonASCIIFilenameEndToEnd memverifikasi alur PENUH pindai -> cocok -> salin untuk nama
// berkas yang mengandung karakter non-ASCII (PRD §9). Sebelumnya skenario ini hanya diuji di
// level parser daftar (lihat internal/list/list_test.go), belum lintas scan+match+fsops sekaligus.
func TestNonASCIIFilenameEndToEnd(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	fileName := "Café_0001.CR3"
	content := []byte("NON_ASCII_RAW_CONTENT")
	if err := os.WriteFile(filepath.Join(srcDir, fileName), content, 0644); err != nil {
		t.Fatal(err)
	}

	scanRes, err := scan.Scan(scan.Options{RootDir: srcDir, Recursive: true})
	if err != nil {
		t.Fatalf("scan.Scan failed: %v", err)
	}
	if len(scanRes.Files) != 1 {
		t.Fatalf("expected 1 scanned file, got %d", len(scanRes.Files))
	}

	matchRes := match.Match([]string{"Café_0001"}, scanRes.Files, match.Options{RawFilter: true})
	if len(matchRes.Matched) != 1 {
		t.Fatalf("expected 1 matched target, got %d (not found: %v)", len(matchRes.Matched), matchRes.NotFoundNames)
	}

	res, err := Execute(context.Background(), Options{
		Mode:         ModeCopy,
		SourceDir:    srcDir,
		DestDir:      destDir,
		MatchedItems: matchRes.Matched,
		RetryDelays:  []time.Duration{1 * time.Millisecond},
	})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if len(res.CopiedFiles) != 1 {
		t.Fatalf("expected 1 copied file, got %d", len(res.CopiedFiles))
	}

	destPath := filepath.Join(destDir, fileName)
	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("expected non-ASCII filename to exist at destination: %v", err)
	}
	if string(data) != string(content) {
		t.Errorf("content mismatch for non-ASCII filename")
	}
}

// TestGetUniqueDestinationNameAttemptLimit memverifikasi getUniqueDestinationName tidak berputar
// tanpa batas bila stat terus-menerus gagal karena alasan SELAIN "tidak ditemukan" (mis.
// permission denied), lalu jatuh ke kandidat best-effort setelah batas percobaan tercapai.
//
// Kegagalan stat DISUNTIK, bukan dipicu lewat izin berkas: os.Chmod(0o000) menolak akses baca
// di Unix tetapi TIDAK di Windows (hanya menyetel atribut read-only), sehingga pemicu lama gagal
// di CI windows-latest. Dengan suntikan ini perilakunya identik di semua OS.
func TestGetUniqueDestinationNameAttemptLimit(t *testing.T) {
	const limit = 5

	statSelaluGagal := func(string) (os.FileInfo, error) {
		return nil, errors.New("permission denied")
	}

	candidate, hadCollision := getUniqueDestinationNameLimited(
		"/folder/tujuan", "test.txt", map[string]struct{}{}, limit, statSelaluGagal)

	if !hadCollision {
		t.Errorf("expected hadCollision=true when attempt limit is reached (anti-hang fallback)")
	}

	expected := fmt.Sprintf("test (%d).txt", limit+1)
	if candidate != expected {
		t.Errorf("expected best-effort candidate %q, got %q", expected, candidate)
	}
}
