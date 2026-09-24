package list

import (
	"errors"
	"strings"
	"testing"
)

type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error simulation")
}

func TestParseRules(t *testing.T) {
	t.Run("P-1: CRLF, LF, CR", func(t *testing.T) {
		input := "DSC_0001\r\nDSC_0002\nDSC_0003\rDSC_0004\r\n"
		res, err := ParseText(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{"DSC_0001", "DSC_0002", "DSC_0003", "DSC_0004"}
		if len(res.Items) != len(expected) {
			t.Fatalf("expected %d items, got %d", len(expected), len(res.Items))
		}
		for i, item := range res.Items {
			if item != expected[i] {
				t.Errorf("item %d: expected %s, got %s", i, expected[i], item)
			}
		}
		if res.DuplicateCount != 0 {
			t.Errorf("expected 0 duplicates, got %d", res.DuplicateCount)
		}
	})

	t.Run("P-2: Trim space and tabs", func(t *testing.T) {
		input := "  \tDSC_0001 \t \n \t DSC_0002\t\n"
		res, err := ParseText(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{"DSC_0001", "DSC_0002"}
		if len(res.Items) != len(expected) {
			t.Fatalf("expected %d items, got %d", len(expected), len(res.Items))
		}
		if res.Items[0] != "DSC_0001" || res.Items[1] != "DSC_0002" {
			t.Errorf("unexpected items: %v", res.Items)
		}
	})

	t.Run("P-3: Discard empty lines", func(t *testing.T) {
		input := "\n\n  \n\t\nDSC_0001\n\n\nDSC_0002\n\n"
		res, err := ParseText(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Items) != 2 {
			t.Fatalf("expected 2 items, got %d", len(res.Items))
		}
		if res.NonEmptyLinesCount != 2 {
			t.Errorf("expected 2 non-empty lines, got %d", res.NonEmptyLinesCount)
		}
	})

	t.Run("P-4: Strip path prefixes", func(t *testing.T) {
		input := `D:\Photos\2026\DSC_0001
/Volumes/SDCard/DCIM/100CANON/DSC_0002
sub\nested/dir\DSC_0003`
		res, err := ParseText(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{"DSC_0001", "DSC_0002", "DSC_0003"}
		for i, item := range res.Items {
			if item != expected[i] {
				t.Errorf("item %d: expected %s, got %s", i, expected[i], item)
			}
		}
	})

	t.Run("P-5: Strip extension", func(t *testing.T) {
		input := `DSC_0001.CR3
_MG_0002.CR2
IMG_0003.ARW
photo.with.multiple.dots.dng
no_extension_0005`
		res, err := ParseText(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{
			"DSC_0001",
			"_MG_0002",
			"IMG_0003",
			"photo.with.multiple.dots",
			"no_extension_0005",
		}
		if len(res.Items) != len(expected) {
			t.Fatalf("expected %d items, got %d", len(expected), len(res.Items))
		}
		for i, item := range res.Items {
			if item != expected[i] {
				t.Errorf("item %d: expected %s, got %s", i, expected[i], item)
			}
		}
	})

	t.Run("P-6: Dedupe case-insensitive preserving first occurrence", func(t *testing.T) {
		input := `DSC_0001
dsc_0001
Dsc_0001
DSC_0002
DSC_0002.cr3
dsc_0002.CR3`
		res, err := ParseText(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{"DSC_0001", "DSC_0002"}
		if len(res.Items) != len(expected) {
			t.Fatalf("expected %d items, got %d", len(expected), len(res.Items))
		}
		if res.Items[0] != "DSC_0001" || res.Items[1] != "DSC_0002" {
			t.Errorf("unexpected items: %v", res.Items)
		}
		if res.DuplicateCount != 4 {
			t.Errorf("expected 4 duplicates, got %d", res.DuplicateCount)
		}
		if res.NonEmptyLinesCount != 6 {
			t.Errorf("expected 6 non-empty lines, got %d", res.NonEmptyLinesCount)
		}
	})

	t.Run("P-7: Preserve order", func(t *testing.T) {
		input := `IMG_9999
DSC_0001
_MG_4000
DSC_0002
IMG_9999`
		res, err := ParseText(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{"IMG_9999", "DSC_0001", "_MG_4000", "DSC_0002"}
		if len(res.Items) != len(expected) {
			t.Fatalf("expected %d items, got %d", len(expected), len(res.Items))
		}
		for i, item := range res.Items {
			if item != expected[i] {
				t.Errorf("item %d: expected %s, got %s", i, expected[i], item)
			}
		}
	})
}

func TestBOMAndUTF8(t *testing.T) {
	t.Run("UTF-8 with BOM", func(t *testing.T) {
		bom := []byte("\xef\xbb\xbfDSC_0001\r\nDSC_0002\r\n")
		res, err := ParseBytes(bom)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Items) != 2 || res.Items[0] != "DSC_0001" || res.Items[1] != "DSC_0002" {
			t.Errorf("unexpected parsed result: %v", res.Items)
		}
	})

	t.Run("Non-ASCII UTF-8 characters", func(t *testing.T) {
		input := "Café_0001.CR3\n日本語_0002.NEF\nFoto_Pelaminan_Üñíçödé.ARW\n"
		res, err := ParseText(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := []string{"Café_0001", "日本語_0002", "Foto_Pelaminan_Üñíçödé"}
		if len(res.Items) != len(expected) {
			t.Fatalf("expected %d items, got %d", len(expected), len(res.Items))
		}
		for i, item := range res.Items {
			if item != expected[i] {
				t.Errorf("item %d: expected %s, got %s", i, expected[i], item)
			}
		}
	})

	t.Run("Invalid UTF-8 (R-D7)", func(t *testing.T) {
		invalidBytes := []byte{0xff, 0xfe, 0x00, 0x44, 0x53, 0x43} // non UTF-8
		_, err := ParseBytes(invalidBytes)
		if !errors.Is(err, ErrInvalidUTF8) {
			t.Errorf("expected ErrInvalidUTF8, got %v", err)
		}
	})
}

func TestParseReaderAndEmpty(t *testing.T) {
	t.Run("Empty string", func(t *testing.T) {
		res, err := ParseText("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Items) != 0 {
			t.Errorf("expected 0 items, got %d", len(res.Items))
		}
	})

	t.Run("ParseReader success", func(t *testing.T) {
		r := strings.NewReader("DSC_0001\nDSC_0002\n")
		res, err := ParseReader(r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(res.Items))
		}
	})

	t.Run("ParseReader error", func(t *testing.T) {
		_, err := ParseReader(&errReader{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("Summary format", func(t *testing.T) {
		input := "DSC_0001\nDSC_0001\nDSC_0002\n"
		res, err := ParseText(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectedSummary := "3 baris -> 2 nama unik (1 duplikat dibuang)"
		if res.Summary() != expectedSummary {
			t.Errorf("expected summary %q, got %q", expectedSummary, res.Summary())
		}
	})
}
