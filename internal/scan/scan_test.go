package scan

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRawExtensions(t *testing.T) {
	// 27 ekstensi terdaftar di PRD §4.4
	expectedRAWs := []string{
		"3FR", "ARW", "CR2", "CR3", "CRW", "DCR", "DNG", "ERF", "FFF", "IIQ",
		"KDC", "MEF", "MOS", "MRW", "NEF", "NRW", "ORF", "PEF", "PTX", "RAF",
		"RAW", "RWL", "RW2", "SR2", "SRF", "SRW", "X3F",
	}

	if len(RawExtensions) != 27 {
		t.Fatalf("expected exactly 27 RAW extensions, got %d", len(RawExtensions))
	}

	for _, ext := range expectedRAWs {
		if !IsRAW(ext) {
			t.Errorf("expected %s to be RAW", ext)
		}
		if !IsRAW("." + ext) {
			t.Errorf("expected .%s to be RAW", ext)
		}
		if !IsRAW(filepath.Ext("photo." + ext)) {
			t.Errorf("expected file with .%s to be RAW", ext)
		}
	}

	nonRAWs := []string{"jpg", "jpeg", "png", "tif", "tiff", "txt", "mp4", "xmp", "exe"}
	for _, ext := range nonRAWs {
		if IsRAW(ext) {
			t.Errorf("expected %s NOT to be RAW", ext)
		}
	}
}

func TestIsXMP(t *testing.T) {
	if !IsXMP(".xmp") || !IsXMP("xmp") || !IsXMP(".XMP") || !IsXMP("XMP") {
		t.Errorf("expected xmp variations to return true")
	}
	if IsXMP(".jpg") || IsXMP(".cr3") {
		t.Errorf("expected non-xmp to return false")
	}
}

func TestScanDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Buat struktur direktori untuk pengujian
	// root/
	//   DSC_0001.CR3
	//   DSC_0001.xmp (sidecar)
	//   DSC_0002.JPG
	//   sub1/
	//     IMG_0003.ARW
	//     IMG_0003.ARW.xmp (sidecar bentuk 2)
	//     ignore.txt
	//   sub2/nested/
	//     _MG_0004.NEF

	if err := os.WriteFile(filepath.Join(tempDir, "DSC_0001.CR3"), []byte("raw1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "DSC_0001.xmp"), []byte("<xmp/>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "DSC_0002.JPG"), []byte("jpeg"), 0644); err != nil {
		t.Fatal(err)
	}

	sub1 := filepath.Join(tempDir, "sub1")
	if err := os.MkdirAll(sub1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub1, "IMG_0003.ARW"), []byte("raw2"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub1, "IMG_0003.ARW.xmp"), []byte("<xmp2/>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub1, "ignore.txt"), []byte("text"), 0644); err != nil {
		t.Fatal(err)
	}

	sub2Nested := filepath.Join(tempDir, "sub2", "nested")
	if err := os.MkdirAll(sub2Nested, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub2Nested, "_MG_0004.NEF"), []byte("raw3"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("Non-recursive scan", func(t *testing.T) {
		res, err := Scan(Options{
			RootDir:   tempDir,
			Recursive: false,
		})
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}

		// Harus hanya mendeteksi berkas di root: DSC_0001.CR3, DSC_0001.xmp, DSC_0002.JPG
		if len(res.Files) != 3 {
			t.Fatalf("expected 3 files, got %d", len(res.Files))
		}
		if res.RAWCount != 1 {
			t.Errorf("expected 1 RAW file, got %d", res.RAWCount)
		}
		if res.NonRAWCount != 2 {
			t.Errorf("expected 2 non-RAW files, got %d", res.NonRAWCount)
		}

		// Cek sidecar xmp
		var foundRaw *FileEntry
		for _, f := range res.Files {
			if f.FileName == "DSC_0001.CR3" {
				fCopy := f
				foundRaw = &fCopy
				break
			}
		}
		if foundRaw == nil {
			t.Fatal("DSC_0001.CR3 not found")
		}
		if !foundRaw.HasXMP {
			t.Errorf("expected DSC_0001.CR3 to have HasXMP = true")
		}
		if foundRaw.XMPPath == "" {
			t.Errorf("expected XMPPath to be populated")
		}
	})

	t.Run("Recursive scan", func(t *testing.T) {
		res, err := Scan(Options{
			RootDir:   tempDir,
			Recursive: true,
		})
		if err != nil {
			t.Fatalf("scan failed: %v", err)
		}

		if res.RAWCount != 3 { // DSC_0001.CR3, IMG_0003.ARW, _MG_0004.NEF
			t.Errorf("expected 3 RAW files, got %d", res.RAWCount)
		}
		if len(res.Files) != 7 {
			t.Errorf("expected 7 total files, got %d", len(res.Files))
		}

		// Cek sidecar IMG_0003.ARW.xmp
		var foundARW *FileEntry
		for _, f := range res.Files {
			if f.FileName == "IMG_0003.ARW" {
				fCopy := f
				foundARW = &fCopy
				break
			}
		}
		if foundARW == nil {
			t.Fatal("IMG_0003.ARW not found")
		}
		if !foundARW.HasXMP {
			t.Errorf("expected IMG_0003.ARW to have HasXMP = true")
		}
	})
}

func TestSymlinkHandling(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping symlink tests on Windows runner without privilege")
	}

	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "real_dir")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	realFile := filepath.Join(sourceDir, "TARGET.CR3")
	if err := os.WriteFile(realFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Buat symlink file di root
	fileSymlink := filepath.Join(tempDir, "LINKED.CR3")
	if err := os.Symlink(realFile, fileSymlink); err != nil {
		t.Fatal(err)
	}

	// Buat symlink direktori di root
	dirSymlink := filepath.Join(tempDir, "LINKED_DIR")
	if err := os.Symlink(sourceDir, dirSymlink); err != nil {
		t.Fatal(err)
	}

	res, err := Scan(Options{
		RootDir:   tempDir,
		Recursive: true,
	})
	if err != nil {
		t.Fatalf("scan error: %v", err)
	}

	// Symlink direktori tidak boleh diikuti (R-D6)
	if len(res.Warnings) == 0 {
		t.Errorf("expected warning about skipped directory symlink")
	}

	// Symlink file tetap boleh disalin/ditemukan
	var foundLinkedFile bool
	for _, f := range res.Files {
		if f.FileName == "LINKED.CR3" {
			foundLinkedFile = true
			break
		}
	}
	if !foundLinkedFile {
		t.Errorf("expected file symlink LINKED.CR3 to be detected")
	}

	t.Run("Non-recursive symlink handling", func(t *testing.T) {
		res, err := Scan(Options{
			RootDir:   tempDir,
			Recursive: false,
		})
		if err != nil {
			t.Fatalf("scan error: %v", err)
		}
		if len(res.Warnings) == 0 {
			t.Errorf("expected warning about skipped directory symlink")
		}
	})
}

func TestScanErrors(t *testing.T) {
	t.Run("Empty root dir", func(t *testing.T) {
		_, err := Scan(Options{RootDir: ""})
		if err == nil {
			t.Error("expected error for empty root dir")
		}
	})

	t.Run("Non-existent root dir", func(t *testing.T) {
		_, err := Scan(Options{RootDir: filepath.Join(t.TempDir(), "not_exist")})
		if err == nil {
			t.Error("expected error for non-existent root dir")
		}
	})

	t.Run("File instead of directory", func(t *testing.T) {
		tmp := filepath.Join(t.TempDir(), "file.txt")
		if err := os.WriteFile(tmp, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := Scan(Options{RootDir: tmp})
		if err == nil {
			t.Error("expected error when root dir is a file")
		}
	})
}
