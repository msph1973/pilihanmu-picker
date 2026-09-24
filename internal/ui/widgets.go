package ui

import (
	"fmt"
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// BasketCard adalah widget kustom untuk menampilkan satu dari empat keranjang hasil pratinjau.
type BasketCard struct {
	widget.BaseWidget
	Title       string
	CountText   string
	SubText     string
	ActionBtn   *widget.Button
	AccentColor color.Color
	titleLabel  *widget.Label
	countLabel  *widget.Label
	subLabel    *widget.Label
	bgRectangle *canvas.Rectangle
	container   *fyne.Container
}

// NewBasketCard membuat card pratinjau untuk keranjang pencocokan, dengan garis aksen
// accentColor di sisi kiri untuk menandai kategori keranjang (mis. hijau untuk Cocok).
func NewBasketCard(title, countText, subText string, actionBtn *widget.Button, accentColor color.Color) *BasketCard {
	bc := &BasketCard{
		Title:       title,
		CountText:   countText,
		SubText:     subText,
		ActionBtn:   actionBtn,
		AccentColor: accentColor,
	}
	bc.ExtendBaseWidget(bc)
	bc.initUI()
	return bc
}

func (bc *BasketCard) initUI() {
	bc.titleLabel = widget.NewLabelWithStyle(bc.Title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	bc.countLabel = widget.NewLabelWithStyle(bc.CountText, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	bc.subLabel = widget.NewLabel(bc.SubText)
	bc.subLabel.Wrapping = fyne.TextWrapWord

	var right fyne.CanvasObject
	if bc.ActionBtn != nil {
		right = bc.ActionBtn
	}
	headerRow := container.NewBorder(nil, nil, bc.titleLabel, right)

	content := container.NewVBox(
		headerRow,
		bc.countLabel,
		bc.subLabel,
	)

	card := widget.NewCard("", "", content)

	// Garis aksen warna semantik di sisi kiri (menandai kategori keranjang)
	bc.bgRectangle = canvas.NewRectangle(bc.AccentColor)
	bc.bgRectangle.SetMinSize(fyne.NewSize(4, 0))

	bc.container = container.NewBorder(nil, nil, bc.bgRectangle, nil, card)
}

// SetValues memperbarui angka dan teks keterangan card.
func (bc *BasketCard) SetValues(countText, subText string) {
	bc.CountText = countText
	bc.SubText = subText
	if bc.countLabel != nil {
		bc.countLabel.SetText(countText)
	}
	if bc.subLabel != nil {
		bc.subLabel.SetText(subText)
	}
}

// CreateRenderer mengimplementasikan antarmuka fyne.Widget.
func (bc *BasketCard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(bc.container)
}

// ShowNotFoundNamesDialog menampilkan dialog daftar nama berkas yang tidak ditemukan.
func ShowNotFoundNamesDialog(parent fyne.Window, notFound []string) {
	if len(notFound) == 0 {
		dialog.ShowInformation("Tidak Ditemukan", "Semua nama dalam daftar berhasil ditemukan di folder sumber!", parent)
		return
	}

	content := strings.Join(notFound, "\n")
	entry := widget.NewMultiLineEntry()
	entry.SetText(content)
	entry.Wrapping = fyne.TextWrapWord
	entry.Disable()

	entry.SetMinRowsVisible(8)

	copyBtn := widget.NewButtonWithIcon("Salin ke Clipboard", theme.ContentCopyIcon(), func() {
		parent.Clipboard().SetContent(content)
		dialog.ShowInformation("Berhasil Disalin", fmt.Sprintf("%d nama berhasil disalin ke clipboard.", len(notFound)), parent)
	})

	scroll := container.NewScroll(entry)

	body := container.NewBorder(
		widget.NewLabel(fmt.Sprintf("Ditemukan %d nama dari daftar yang tidak ada di folder sumber:", len(notFound))),
		copyBtn,
		nil,
		nil,
		scroll,
	)

	d := dialog.NewCustom("Daftar Tidak Ditemukan", "Tutup", body, parent)
	d.Resize(fyne.NewSize(480, 400))
	d.Show()
}

// ShowReportViewerDialog menampilkan dialog pratinjau isi _laporan.txt.
func ShowReportViewerDialog(parent fyne.Window, reportText string, reportPath string) {
	entry := widget.NewMultiLineEntry()
	entry.SetText(reportText)
	entry.Wrapping = fyne.TextWrapWord
	entry.Disable()
	entry.SetMinRowsVisible(12)

	copyBtn := widget.NewButtonWithIcon("Salin Isi Laporan", theme.ContentCopyIcon(), func() {
		parent.Clipboard().SetContent(reportText)
		dialog.ShowInformation("Berhasil Disalin", "Isi laporan berhasil disalin ke clipboard.", parent)
	})

	scroll := container.NewScroll(entry)

	body := container.NewBorder(
		widget.NewLabel(fmt.Sprintf("Berkas laporan: %s", reportPath)),
		copyBtn,
		nil,
		nil,
		scroll,
	)

	d := dialog.NewCustom("Laporan Hasil Seleksi (_laporan.txt)", "Tutup", body, parent)
	d.Resize(fyne.NewSize(580, 480))
	d.Show()
}

// FormatBytes mengubah ukuran byte menjadi teks ramah baca (MB / GB).
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
