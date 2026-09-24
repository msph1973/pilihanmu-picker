package fsops

import (
	"context"
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
