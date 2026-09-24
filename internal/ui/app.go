package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/msph1973/pilihanmu-picker/internal/fsops"
	"github.com/msph1973/pilihanmu-picker/internal/list"
	"github.com/msph1973/pilihanmu-picker/internal/match"
	"github.com/msph1973/pilihanmu-picker/internal/report"
	"github.com/msph1973/pilihanmu-picker/internal/scan"
)

// App adalah pengelola antarmuka utama aplikasi desktop Pilihanmu Picker.
type App struct {
	window fyne.Window

	// State internal
	listText           string
	listResult         list.Result
	sourceDir          string
	destDir            string
	includeSubfolders  bool
	rawFilterEnabled   bool
	isScanning         bool
	isExecuting        bool
	scanResult         scan.Result
	matchResult        match.Result
	execContextCancel  context.CancelFunc
	lastExecutionTitle string

	// Komponen UI Masukan
	inputEntry      *widget.Entry
	listStatusLabel *widget.Label

	// Komponen UI Folder & Pengaturan
	sourceEntry          *widget.Entry
	destEntry            *widget.Entry
	subfolderCheck       *widget.Check
	rawFilterCheck       *widget.Check
	jpegWarningCard      *widget.Card
	jpegWarningLabel     *widget.Label
	jpegWarningContainer *fyne.Container

	// Komponen UI Pratinjau
	statusScanLabel *widget.Label
	refreshBtn      *widget.Button
	matchedCard     *BasketCard
	notFoundCard    *BasketCard
	collisionCard   *BasketCard
	skippedCard     *BasketCard

	// Komponen UI Tindakan
	copyBtn *widget.Button
	moveBtn *widget.Button
}

// NewApp menginisialisasi dan membangun antarmuka aplikasi.
func NewApp(w fyne.Window) *App {
	a := &App{
		window:            w,
		includeSubfolders: true, // Default aktif (PRD §4.3)
		rawFilterEnabled:  true, // Default aktif (PRD §4.4)
	}

	a.buildUI()
	return a
}

// buildUI menyusun tata letak (layout) jendela aplikasi.
func (a *App) buildUI() {
	// 1. Header Aplikasi
	titleLabel := widget.NewLabelWithStyle(AppTitle, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	subtitleLabel := widget.NewLabel(AppSubtitle)
	badgeLabel := widget.NewLabelWithStyle(AppBadgeStatus, fyne.TextAlignTrailing, fyne.TextStyle{Italic: true})

	headerLeft := container.NewVBox(titleLabel, subtitleLabel)
	headerBar := container.NewBorder(nil, nil, headerLeft, badgeLabel)

	// 2. Panel Kiri: Input Daftar & Folder
	inputSection := a.createInputSection()
	folderSection := a.createFolderSection()
	leftPanel := container.NewVBox(inputSection, widget.NewSeparator(), folderSection)

	// 3. Panel Kanan: Pratinjau & Tombol Tindakan
	previewSection := a.createPreviewSection()
	actionSection := a.createActionSection()
	rightPanel := container.NewVBox(previewSection, widget.NewSeparator(), actionSection)

	// 4. Tata Letak Dua Kolom
	mainGrid := container.NewGridWithColumns(2, leftPanel, rightPanel)
	mainScroll := container.NewScroll(mainGrid)

	// Jendela Utama
	rootContainer := container.NewBorder(
		container.NewVBox(headerBar, widget.NewSeparator()),
		nil,
		nil,
		nil,
		mainScroll,
	)

	a.window.SetContent(container.NewPadded(rootContainer))
	a.updateListStatus()
	a.updateActionButtonsState()
}

// createInputSection membuat komponen untuk Langkah 1 (Input Daftar Foto).
func (a *App) createInputSection() fyne.CanvasObject {
	sectionTitle := widget.NewLabelWithStyle(SectionInputTitle, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	a.inputEntry = widget.NewMultiLineEntry()
	a.inputEntry.SetPlaceHolder(InputListPlaceholder)
	a.inputEntry.Wrapping = fyne.TextWrapWord
	a.inputEntry.SetMinRowsVisible(6)
	a.inputEntry.OnChanged = func(s string) {
		a.listText = s
		a.parseInputList()
	}

	scrollInput := container.NewScroll(a.inputEntry)

	btnPaste := widget.NewButtonWithIcon(BtnPasteClipboard, theme.ContentPasteIcon(), func() {
		clip := a.window.Clipboard().Content()
		if clip != "" {
			a.inputEntry.SetText(clip)
		}
	})

	btnClear := widget.NewButtonWithIcon(BtnClearInput, theme.DeleteIcon(), func() {
		a.inputEntry.SetText("")
	})

	btnFile := widget.NewButtonWithIcon(BtnOpenFile, theme.FolderOpenIcon(), func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			res, parseErr := list.ParseReader(reader)
			if parseErr != nil {
				dialog.ShowError(fmt.Errorf("Gagal membaca berkas: %w", parseErr), a.window)
				return
			}
			a.lastExecutionTitle = reader.URI().Name()
			a.inputEntry.SetText(strings.Join(res.Items, "\n"))
		}, a.window)
	})

	a.listStatusLabel = widget.NewLabel(LabelListEmpty)
	a.listStatusLabel.Wrapping = fyne.TextWrapWord

	btnRow := container.NewHBox(btnPaste, btnFile, btnClear)

	return container.NewVBox(
		sectionTitle,
		btnRow,
		scrollInput,
		a.listStatusLabel,
	)
}

// createFolderSection membuat komponen untuk Langkah 2 (Folder & Saringan).
func (a *App) createFolderSection() fyne.CanvasObject {
	sectionTitle := widget.NewLabelWithStyle(SectionFoldersTitle, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Folder Sumber
	a.sourceEntry = widget.NewEntry()
	a.sourceEntry.SetPlaceHolder(PlaceholderSource)
	a.sourceEntry.OnChanged = func(s string) {
		a.sourceDir = strings.TrimSpace(s)
		a.triggerAsyncScan()
	}

	btnBrowseSource := widget.NewButtonWithIcon(BtnBrowseSource, theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil || list == nil {
				return
			}
			path := list.Path()
			a.sourceEntry.SetText(path)
		}, a.window)
	})

	sourceBox := container.NewBorder(nil, nil, nil, btnBrowseSource, a.sourceEntry)

	// Folder Tujuan
	a.destEntry = widget.NewEntry()
	a.destEntry.SetPlaceHolder(PlaceholderDest)
	a.destEntry.OnChanged = func(s string) {
		a.destDir = strings.TrimSpace(s)
		a.updateActionButtonsState()
	}

	btnBrowseDest := widget.NewButtonWithIcon(BtnBrowseDest, theme.FolderNewIcon(), func() {
		dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil || list == nil {
				return
			}
			path := list.Path()
			a.destEntry.SetText(path)
		}, a.window)
	})

	destBox := container.NewBorder(nil, nil, nil, btnBrowseDest, a.destEntry)

	// Opsi-opsi
	a.subfolderCheck = widget.NewCheck(CheckRecursive, func(checked bool) {
		a.includeSubfolders = checked
		a.triggerAsyncScan()
	})
	a.subfolderCheck.Checked = a.includeSubfolders

	a.rawFilterCheck = widget.NewCheck(CheckRawFilter, func(checked bool) {
		a.rawFilterEnabled = checked
		a.triggerAsyncScan()
	})
	a.rawFilterCheck.Checked = a.rawFilterEnabled

	rawInfoLabel := widget.NewLabelWithStyle(LabelRawFormatsInfo, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	rawInfoLabel.Wrapping = fyne.TextWrapWord

	// Jaring Pengaman JPEG (R-U4)
	a.jpegWarningLabel = widget.NewLabel(WarnJPEGOnly)
	a.jpegWarningLabel.Wrapping = fyne.TextWrapWord
	a.jpegWarningCard = widget.NewCard("", "Jaring Pengaman Format", a.jpegWarningLabel)
	a.jpegWarningContainer = container.NewVBox(a.jpegWarningCard)
	a.jpegWarningContainer.Hide() // Sembunyi default sampai terdeteksi

	return container.NewVBox(
		sectionTitle,
		widget.NewLabel(LabelSourceFolder),
		sourceBox,
		widget.NewLabel(LabelDestFolder),
		destBox,
		container.NewHBox(a.subfolderCheck, a.rawFilterCheck),
		rawInfoLabel,
		a.jpegWarningContainer,
	)
}

// createPreviewSection membuat komponen untuk Langkah 3 (Pratinjau Pencocokan 4 Keranjang).
func (a *App) createPreviewSection() fyne.CanvasObject {
	sectionTitle := widget.NewLabelWithStyle(SectionPreviewTitle, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	a.statusScanLabel = widget.NewLabel("Pilih folder sumber untuk memindai berkas...")
	a.statusScanLabel.Wrapping = fyne.TextWrapWord

	a.refreshBtn = widget.NewButtonWithIcon(BtnRefreshPreview, theme.ViewRefreshIcon(), func() {
		a.triggerAsyncScan()
	})

	topRow := container.NewBorder(nil, nil, sectionTitle, a.refreshBtn, a.statusScanLabel)

	// Tombol Aksi Keranjang
	btnCopyNotFound := widget.NewButtonWithIcon("Lihat / Salin", theme.ContentCopyIcon(), func() {
		ShowNotFoundNamesDialog(a.window, a.matchResult.NotFoundNames)
	})

	// 4 Keranjang
	a.matchedCard = NewBasketCard(LabelBasketMatched, "0 berkas", "Belum ada berkas cocok", nil)
	a.notFoundCard = NewBasketCard(LabelBasketNotFound, "0 nama", "Tidak ada catatan", btnCopyNotFound)
	a.collisionCard = NewBasketCard(LabelBasketCollision, "0 berkas", "Tidak ada tabrakan nama", nil)
	a.skippedCard = NewBasketCard(LabelBasketSkipped, "0 berkas", "Berkas tidak dipilih klien", nil)

	basketGrid := container.NewGridWithColumns(2,
		a.matchedCard,
		a.notFoundCard,
		a.collisionCard,
		a.skippedCard,
	)

	return container.NewVBox(
		topRow,
		basketGrid,
	)
}

// createActionSection membuat komponen untuk Langkah 4 (Tombol Tindakan Eksekusi).
func (a *App) createActionSection() fyne.CanvasObject {
	sectionTitle := widget.NewLabelWithStyle(SectionActionTitle, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	a.copyBtn = widget.NewButtonWithIcon(BtnActionCopy, theme.DocumentSaveIcon(), func() {
		a.confirmAndExecute(fsops.ModeCopy)
	})
	a.copyBtn.Importance = widget.HighImportance

	a.moveBtn = widget.NewButtonWithIcon(BtnActionMove, theme.MailForwardIcon(), func() {
		a.confirmAndExecute(fsops.ModeMove)
	})
	a.moveBtn.Importance = widget.WarningImportance

	actionGrid := container.NewGridWithColumns(2, a.copyBtn, a.moveBtn)

	notes := widget.NewLabel("• Salin: Folder sumber tetap utuh.\n• Pindahkan: Berkas sumber dihapus setelah diverifikasi sama ukurannya.")
	notes.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		sectionTitle,
		actionGrid,
		notes,
	)
}

// parseInputList memproses teks input menggunakan parser internal/list.
func (a *App) parseInputList() {
	if strings.TrimSpace(a.listText) == "" {
		a.listResult = list.Result{}
		a.listStatusLabel.SetText(LabelListEmpty)
		a.triggerAsyncScan()
		return
	}

	res, err := list.ParseText(a.listText)
	if err != nil {
		a.listStatusLabel.SetText(LabelListInvalidUTF8)
		return
	}

	a.listResult = res
	a.listStatusLabel.SetText(res.Summary())
	a.triggerAsyncScan()
}

// updateListStatus memperbarui label status daftar foto.
func (a *App) updateListStatus() {
	if len(a.listResult.Items) == 0 {
		a.listStatusLabel.SetText(LabelListEmpty)
	} else {
		a.listStatusLabel.SetText(a.listResult.Summary())
	}
}

// triggerAsyncScan memicu pemindaian dan pencocokan di goroutine latar belakang (R-A4).
func (a *App) triggerAsyncScan() {
	if a.sourceDir == "" {
		a.statusScanLabel.SetText("Pilih folder sumber untuk memulai pencarian.")
		a.resetBasketCards()
		a.updateActionButtonsState()
		return
	}

	// Jangan tumpuk scan jika sedang berjalan
	if a.isScanning {
		return
	}

	a.isScanning = true
	a.statusScanLabel.SetText(StatusScanning)
	a.refreshBtn.Disable()
	a.updateActionButtonsState()

	sourceDir := a.sourceDir
	destDir := a.destDir
	recursive := a.includeSubfolders
	rawFilter := a.rawFilterEnabled
	targets := a.listResult.Items

	go func() {
		// 1. Pindai berkas sumber
		scanRes, err := scan.Scan(scan.Options{
			RootDir:   sourceDir,
			Recursive: recursive,
		})

		// 2. Baca nama berkas di tujuan jika direktori tujuan sudah ada
		var destFiles []string
		if destDir != "" {
			entries, _ := os.ReadDir(destDir)
			for _, entry := range entries {
				if !entry.IsDir() {
					destFiles = append(destFiles, entry.Name())
				}
			}
		}

		// 3. Lakukan pencocokan 4 keranjang
		matchRes := match.Match(targets, scanRes.Files, match.Options{
			RawFilter:         rawFilter,
			ExistingDestFiles: destFiles,
		})

		// 4. Perbarui antarmuka di UI thread
		fyne.Do(func() {
			a.isScanning = false
			a.refreshBtn.Enable()

			if err != nil {
				a.statusScanLabel.SetText(fmt.Sprintf("Peringatan pemindaian: %v", err))
			} else {
				a.statusScanLabel.SetText(fmt.Sprintf("Pindai selesai: %d total berkas ditemukan di sumber.", len(scanRes.Files)))
			}

			a.scanResult = scanRes
			a.matchResult = matchRes

			// Perbarui 4 Keranjang
			a.matchedCard.SetValues(
				fmt.Sprintf("%d berkas", matchRes.TotalMatchedFiles),
				fmt.Sprintf("Total ukuran: %s", FormatBytes(matchRes.TotalMatchedSize)),
			)

			a.notFoundCard.SetValues(
				fmt.Sprintf("%d nama", len(matchRes.NotFoundNames)),
				"Tidak ditemukan di sumber",
			)

			a.collisionCard.SetValues(
				fmt.Sprintf("%d berkas", len(matchRes.Collisions)),
				"Akan otomatis diberi sufiks (2)",
			)

			a.skippedCard.SetValues(
				fmt.Sprintf("%d berkas", matchRes.SkippedCount),
				"Berkas sumber tidak dipilih klien",
			)

			// Jaring pengaman JPEG (R-U4)
			if matchRes.JPEGOnlyWarning {
				a.jpegWarningContainer.Show()
			} else {
				a.jpegWarningContainer.Hide()
			}

			a.updateActionButtonsState()
		})
	}()
}

func (a *App) resetBasketCards() {
	a.matchedCard.SetValues("0 berkas", "Belum ada berkas cocok")
	a.notFoundCard.SetValues("0 nama", "Tidak ada catatan")
	a.collisionCard.SetValues("0 berkas", "Tidak ada tabrakan")
	a.skippedCard.SetValues("0 berkas", "0 berkas")
	a.jpegWarningContainer.Hide()
}

// updateActionButtonsState memperbarui status aktif/nonaktif tombol eksekusi (R-P4).
func (a *App) updateActionButtonsState() {
	canExecute := !a.isScanning &&
		!a.isExecuting &&
		a.sourceDir != "" &&
		a.destDir != "" &&
		a.sourceDir != a.destDir &&
		a.matchResult.TotalMatchedFiles > 0

	if canExecute {
		a.copyBtn.Enable()
		a.moveBtn.Enable()
	} else {
		a.copyBtn.Disable()
		a.moveBtn.Disable()
	}
}

// confirmAndExecute menampilkan dialog konfirmasi eksplisit sebelum menjalankan aksi (R-U1).
func (a *App) confirmAndExecute(mode fsops.OpMode) {
	if a.sourceDir == a.destDir {
		dialog.ShowError(fmt.Errorf(WarnSameFolder), a.window)
		return
	}

	if a.matchResult.TotalMatchedFiles == 0 {
		dialog.ShowInformation("Tidak Ada Berkas", LabelNoMatches, a.window)
		return
	}

	var title string
	var message string

	totalFiles := a.matchResult.TotalMatchedFiles
	totalSize := FormatBytes(a.matchResult.TotalMatchedSize)

	if mode == fsops.ModeCopy {
		title = DialogConfirmCopyTitle
		message = fmt.Sprintf("Anda akan MENYALIN %d berkas foto (%s) ke folder tujuan.\n\nSumber: %s\nTujuan: %s\n\nFolder sumber tidak akan diubah atau dihapus.",
			totalFiles, totalSize, a.sourceDir, a.destDir)
	} else {
		title = DialogConfirmMoveTitle
		message = fmt.Sprintf("Anda akan MEMINDAHKAN %d berkas foto (%s) ke folder tujuan.\n\nSumber: %s\nTujuan: %s\n\n%s",
			totalFiles, totalSize, a.sourceDir, a.destDir, DialogMoveWarningText)
	}

	dialog.ShowConfirm(title, message, func(confirmed bool) {
		if confirmed {
			a.startExecution(mode)
		}
	}, a.window)
}

// startExecution mengeksekusi operasi salin/pindah dengan modal progres dan pembatalan (R-U3, R-A4).
func (a *App) startExecution(mode fsops.OpMode) {
	a.isExecuting = true
	a.updateActionButtonsState()

	ctx, cancel := context.WithCancel(context.Background())
	a.execContextCancel = cancel

	// Buat Modal Progres
	progressBar := widget.NewProgressBar()
	progressBar.SetValue(0)

	fileLabel := widget.NewLabel("Menyiapkan pemrosesan...")
	fileLabel.Wrapping = fyne.TextWrapWord

	byteLabel := widget.NewLabel("0 B / " + FormatBytes(a.matchResult.TotalMatchedSize))

	var progressDialog dialog.Dialog

	cancelBtn := widget.NewButtonWithIcon(BtnCancelProgress, theme.CancelIcon(), func() {
		cancel()
		fileLabel.SetText(StatusCancelling)
	})
	cancelBtn.Importance = widget.DangerImportance

	progressContent := container.NewVBox(
		fileLabel,
		progressBar,
		byteLabel,
		cancelBtn,
	)

	progressDialog = dialog.NewCustomWithoutButtons(ProgressTitleExecuting, progressContent, a.window)
	progressDialog.Resize(fyne.NewSize(450, 180))
	progressDialog.Show()

	sourceDir := a.sourceDir
	destDir := a.destDir
	matchedItems := a.matchResult.Matched
	listTitle := a.lastExecutionTitle
	if listTitle == "" {
		listTitle = "Daftar Teks"
	}
	uniqueTargets := len(a.listResult.Items)
	notFoundList := a.matchResult.NotFoundNames

	go func() {
		startTime := time.Now()

		// Jalankan operasi fsops
		execRes, _ := fsops.Execute(ctx, fsops.Options{
			Mode:         mode,
			SourceDir:    sourceDir,
			DestDir:      destDir,
			MatchedItems: matchedItems,
			ProgressCallback: func(p fsops.Progress) {
				fyne.Do(func() {
					if p.TotalFiles > 0 {
						progressBar.SetValue(float64(p.CurrentFileIndex) / float64(p.TotalFiles))
					}
					fileLabel.SetText(fmt.Sprintf("Memproses: %s (%d dari %d berkas)", p.CurrentFileName, p.CurrentFileIndex, p.TotalFiles))
					byteLabel.SetText(fmt.Sprintf("%s / %s", FormatBytes(p.BytesProcessed), FormatBytes(p.TotalBytes)))
				})
			},
		})

		// Siapkan data laporan (R-P7)
		var reportCollisions []report.CollisionItem
		for _, col := range execRes.CollisionRecords {
			reportCollisions = append(reportCollisions, report.CollisionItem{
				SourceRelPath:    col.SourceRelPath,
				ResolvedFileName: col.ResolvedFileName,
			})
		}

		var reportFailures []report.FailureItem
		for _, f := range execRes.FailedFiles {
			reportFailures = append(reportFailures, report.FailureItem{
				FileName: f.FileName,
				Reason:   f.Reason,
			})
		}

		reportData := report.Data{
			Timestamp:          startTime,
			Mode:               string(mode),
			SourceDir:          sourceDir,
			DestDir:            destDir,
			ListTitle:          listTitle,
			UniqueTargetsCount: uniqueTargets,
			MatchedFilesCount:  len(execRes.CopiedFiles),
			NotFoundNames:      notFoundList,
			Collisions:         reportCollisions,
			Failures:           reportFailures,
			Cancelled:          execRes.Cancelled,
		}

		// Tulis berkas _laporan.txt di folder tujuan (R-P7)
		reportPath, _ := report.WriteFile(destDir, reportData)
		reportContent := report.Generate(reportData)

		fyne.Do(func() {
			a.isExecuting = false
			progressDialog.Hide()
			a.updateActionButtonsState()

			// Tampilkan Ringkasan Hasil
			a.showSummaryDialog(execRes, reportPath, reportContent)

			// Pindai ulang folder untuk memperbarui status
			a.triggerAsyncScan()
		})
	}()
}

// showSummaryDialog menampilkan modal ringkasan setelah eksekusi selesai.
func (a *App) showSummaryDialog(res fsops.Result, reportPath, reportContent string) {
	var statusMsg string
	if res.Cancelled {
		statusMsg = SummaryCancelledLabel
	} else if len(res.FailedFiles) > 0 {
		statusMsg = SummaryPartialLabel
	} else {
		statusMsg = SummarySuccessLabel
	}

	detailText := fmt.Sprintf(
		"%s\n\n• Berhasil: %d berkas (%s)\n• Gagal: %d berkas\n• Tabrakan nama diselesaikan: %d berkas\n\nBerkas laporan disimpan di:\n%s",
		statusMsg,
		len(res.CopiedFiles),
		FormatBytes(res.TotalBytesCopied),
		len(res.FailedFiles),
		len(res.CollisionRecords),
		filepath.Join(res.DestDir, report.DefaultReportFileName),
	)

	label := widget.NewLabel(detailText)
	label.Wrapping = fyne.TextWrapWord

	btnOpenFolder := widget.NewButtonWithIcon(BtnOpenDestFolder, theme.FolderOpenIcon(), func() {
		_ = OpenFolderInFileManager(res.DestDir)
	})

	btnViewReport := widget.NewButtonWithIcon(BtnViewReport, theme.DocumentIcon(), func() {
		ShowReportViewerDialog(a.window, reportContent, reportPath)
	})

	btnRow := container.NewHBox(btnOpenFolder, btnViewReport)

	content := container.NewVBox(
		label,
		widget.NewSeparator(),
		btnRow,
	)

	d := dialog.NewCustom(DialogSummaryTitle, BtnClose, content, a.window)
	d.Resize(fyne.NewSize(520, 320))
	d.Show()
}
