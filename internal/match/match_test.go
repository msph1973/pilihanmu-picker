package match

import (
	"testing"
	"time"

	"github.com/msph1973/pilihanmu-picker/internal/scan"
)

func TestMatchFourBaskets(t *testing.T) {
	targets := []string{"DSC_0001", "DSC_0002", "IMG_9999"}

	scanned := []scan.FileEntry{
		// Target 1: Cocok dengan 1 file + sidecar XMP
		{
			Path:     "/source/DSC_0001.CR3",
			RelPath:  "DSC_0001.CR3",
			FileName: "DSC_0001.CR3",
			BaseName: "DSC_0001",
			Ext:      ".CR3",
			Size:     25000000,
			IsRAW:    true,
			HasXMP:   true,
			XMPPath:  "/source/DSC_0001.xmp",
			ModTime:  time.Now(),
		},
		{
			Path:     "/source/DSC_0001.xmp",
			RelPath:  "DSC_0001.xmp",
			FileName: "DSC_0001.xmp",
			BaseName: "DSC_0001",
			Ext:      ".xmp",
			Size:     1024,
			IsRAW:    false,
			ModTime:  time.Now(),
		},
		// Target 2: Tabrakan (muncul di 2 folder sumber berbeda)
		{
			Path:     "/source/subA/DSC_0002.CR3",
			RelPath:  "subA/DSC_0002.CR3",
			FileName: "DSC_0002.CR3",
			BaseName: "DSC_0002",
			Ext:      ".CR3",
			Size:     26000000,
			IsRAW:    true,
			ModTime:  time.Now(),
		},
		{
			Path:     "/source/subB/DSC_0002.CR3",
			RelPath:  "subB/DSC_0002.CR3",
			FileName: "DSC_0002.CR3",
			BaseName: "DSC_0002",
			Ext:      ".CR3",
			Size:     26000000,
			IsRAW:    true,
			ModTime:  time.Now(),
		},
		// File yang tidak dipilih (Dilewati)
		{
			Path:     "/source/DSC_8888.CR3",
			RelPath:  "DSC_8888.CR3",
			FileName: "DSC_8888.CR3",
			BaseName: "DSC_8888",
			Ext:      ".CR3",
			Size:     30000000,
			IsRAW:    true,
			ModTime:  time.Now(),
		},
	}

	res := Match(targets, scanned, Options{RawFilter: true})

	// 1. Keranjang Cocok
	if len(res.Matched) != 2 { // DSC_0001 dan DSC_0002
		t.Fatalf("expected 2 matched targets, got %d", len(res.Matched))
	}

	// Total file cocok = 1 (DSC_0001.CR3) + 1 (DSC_0001.xmp) + 2 (DSC_0002.CR3) = 4
	if res.TotalMatchedFiles != 4 {
		t.Errorf("expected 4 total matched files, got %d", res.TotalMatchedFiles)
	}

	expectedSize := int64(25000000 + 1024 + 26000000 + 26000000)
	if res.TotalMatchedSize != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, res.TotalMatchedSize)
	}

	// 2. Keranjang Tidak ketemu
	if len(res.NotFoundNames) != 1 || res.NotFoundNames[0] != "IMG_9999" {
		t.Errorf("expected IMG_9999 in not found, got %v", res.NotFoundNames)
	}

	// 3. Keranjang Tabrakan
	if len(res.Collisions) != 1 {
		t.Fatalf("expected 1 collision, got %d", len(res.Collisions))
	}
	if res.Collisions[0].ResolvedFileName != "DSC_0002 (2).CR3" {
		t.Errorf("expected resolved collision filename 'DSC_0002 (2).CR3', got %s", res.Collisions[0].ResolvedFileName)
	}

	// 4. Keranjang Dilewati
	if res.SkippedCount != 1 {
		t.Errorf("expected 1 skipped file, got %d", res.SkippedCount)
	}
}

func TestCollisionWithExistingDestinationFiles(t *testing.T) {
	targets := []string{"DSC_0001"}
	scanned := []scan.FileEntry{
		{
			Path:     "/source/DSC_0001.CR3",
			FileName: "DSC_0001.CR3",
			BaseName: "DSC_0001",
			Ext:      ".CR3",
			Size:     100,
			IsRAW:    true,
		},
	}

	// Sudah ada DSC_0001.CR3 di folder tujuan
	res := Match(targets, scanned, Options{
		RawFilter:         true,
		ExistingDestFiles: []string{"dsc_0001.cr3"},
	})

	if len(res.Collisions) != 1 {
		t.Fatalf("expected 1 collision with existing dest file, got %d", len(res.Collisions))
	}
	if res.Collisions[0].ResolvedFileName != "DSC_0001 (2).CR3" {
		t.Errorf("expected resolved filename 'DSC_0001 (2).CR3', got %s", res.Collisions[0].ResolvedFileName)
	}
}

func TestMultipleCollisions(t *testing.T) {
	resolved := resolveCollisionName("PHOTO.CR3", 10)
	if resolved != "PHOTO (10).CR3" {
		t.Errorf("expected 'PHOTO (10).CR3', got %s", resolved)
	}
}

func TestJPEGOnlyWarning(t *testing.T) {
	targets := []string{"DSC_0001"}
	scanned := []scan.FileEntry{
		{
			Path:     "/source/DSC_0001.JPG",
			FileName: "DSC_0001.JPG",
			BaseName: "DSC_0001",
			Ext:      ".JPG",
			Size:     5000000,
			IsRAW:    false,
		},
	}

	t.Run("With RawFilter = true", func(t *testing.T) {
		res := Match(targets, scanned, Options{RawFilter: true})
		if !res.JPEGOnlyWarning {
			t.Error("expected JPEGOnlyWarning = true when only JPEG matches")
		}
		if res.JPEGOnlyMessage != JPEGOnlyWarningMessage {
			t.Errorf("expected standard warning message, got %q", res.JPEGOnlyMessage)
		}
		if len(res.Matched) != 0 {
			t.Errorf("expected 0 matched RAW files")
		}
	})

	t.Run("With RawFilter = false", func(t *testing.T) {
		res := Match(targets, scanned, Options{RawFilter: false})
		if res.JPEGOnlyWarning {
			t.Error("expected JPEGOnlyWarning = false when RawFilter is disabled")
		}
		if len(res.Matched) != 1 {
			t.Fatalf("expected 1 matched file when RawFilter is off, got %d", len(res.Matched))
		}
	})
}
