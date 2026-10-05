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

	"github.com/msph1973/pilihanmu-picker/internal/match"
)

// BasketCard adalah widget kustom untuk menampilkan satu dari empat keranjang hasil pratinjau.
//
// PENTING: field publik (Title, CountText, SubText, ActionBtn, AccentColor) hanya dibaca saat
// NewBasketCard menginisialisasi tampilan. Mengubah field ini secara langsung SETELAH widget
// dibuat TIDAK akan merender ulang label/warnanya — gunakan SetValues untuk memperbarui
// CountText dan SubText secara aman.
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
		dialog.ShowInformation(DialogNotFoundEmptyTitle, MsgAllNamesFound, parent)
		return
	}

	content := strings.Join(notFound, "\n")
	entry := widget.NewMultiLineEntry()
	entry.SetText(content)
	entry.Wrapping = fyne.TextWrapWord
	entry.Disable()

	entry.SetMinRowsVisible(8)

	copyBtn := widget.NewButtonWithIcon(BtnCopyToClipboard, theme.ContentCopyIcon(), func() {
		parent.Clipboard().SetContent(content)
		dialog.ShowInformation(DialogCopiedTitle, fmt.Sprintf(FmtCopiedNamesToClipboard, len(notFound)), parent)
	})

	scroll := container.NewScroll(entry)

	body := container.NewBorder(
		widget.NewLabel(fmt.Sprintf(FmtNotFoundHeader, len(notFound))),
		copyBtn,
		nil,
		nil,
		scroll,
	)

	d := dialog.NewCustom(DialogNotFoundListTitle, BtnClose, body, parent)
	d.Resize(fyne.NewSize(480, 400))
	d.Show()
}

// ShowCollisionsDialog menampilkan dialog daftar pasangan path sumber yang mengalami tabrakan
// nama, beserta nama berkas hasil resolusi otomatisnya (R-U2, PRD §4.5).
func ShowCollisionsDialog(parent fyne.Window, collisions []match.Collision) {
	if len(collisions) == 0 {
		dialog.ShowInformation(DialogNoCollisionsTitle, MsgNoCollisions, parent)
		return
	}

	lines := make([]string, 0, len(collisions))
	for _, c := range collisions {
		lines = append(lines, fmt.Sprintf(FmtCollisionLine, c.RelPath, c.ResolvedFileName))
	}
	content := strings.Join(lines, "\n")

	entry := widget.NewMultiLineEntry()
	entry.SetText(content)
	entry.Wrapping = fyne.TextWrapWord
	entry.Disable()
	entry.SetMinRowsVisible(8)

	copyBtn := widget.NewButtonWithIcon(BtnCopyToClipboard, theme.ContentCopyIcon(), func() {
		parent.Clipboard().SetContent(content)
		dialog.ShowInformation(DialogCopiedTitle, fmt.Sprintf(FmtCopiedCollisionsToClipboard, len(collisions)), parent)
	})

	scroll := container.NewScroll(entry)

	body := container.NewBorder(
		widget.NewLabel(fmt.Sprintf(FmtCollisionHeader, len(collisions))),
		copyBtn,
		nil,
		nil,
		scroll,
	)

	d := dialog.NewCustom(DialogCollisionListTitle, BtnClose, body, parent)
	d.Resize(fyne.NewSize(480, 400))
	d.Show()
}

// ShowReportViewerDialog menampilkan dialog pratinjau isi _laporan.txt.
// reportPath boleh kosong bila penulisan berkas laporan ke disk gagal — isi laporan tetap
// ditampilkan dari memori, dengan keterangan yang jelas bahwa berkas tidak tersimpan (R-C2, R-U6).
func ShowReportViewerDialog(parent fyne.Window, reportText string, reportPath string) {
	entry := widget.NewMultiLineEntry()
	entry.SetText(reportText)
	entry.Wrapping = fyne.TextWrapWord
	entry.Disable()
	entry.SetMinRowsVisible(12)

	copyBtn := widget.NewButtonWithIcon(BtnCopyReportContent, theme.ContentCopyIcon(), func() {
		parent.Clipboard().SetContent(reportText)
		dialog.ShowInformation(DialogCopiedTitle, MsgReportCopied, parent)
	})

	scroll := container.NewScroll(entry)

	pathLabel := fmt.Sprintf(FmtReportFileLabel, reportPath)
	if reportPath == "" {
		pathLabel = MsgReportNotSavedToDisk
	}

	body := container.NewBorder(
		widget.NewLabel(pathLabel),
		copyBtn,
		nil,
		nil,
		scroll,
	)

	d := dialog.NewCustom(DialogReportViewerTitle, BtnClose, body, parent)
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
