package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync/atomic"
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
	listText   string
	listResult list.Result
	sourceDir  string
	destDir    string
	// listDestFiles adalah pintu I/O pembacaan nama berkas di folder tujuan (R-A6). Disimpan
	// sebagai field yang bisa disuntik supaya jalur peringatan R-C2 dapat diuji secara
	// deterministik di semua OS — memicunya lewat error filesystem tidak portabel (folder tujuan
	// berupa berkas biasa menghasilkan error di Linux tetapi tidak di Windows).
	listDestFiles     func(string) ([]string, error)
	includeSubfolders bool
	rawFilterEnabled  bool
	// isScanning dan isExecuting dibaca dari updateActionButtonsState (bisa dipanggil dari
	// berbagai pemicu UI) dan ditulis dari dalam goroutine latar belakang lewat fyne.Do; pada
	// driver tes, fyne.Do berjalan langsung di goroutine pemanggilnya (bukan goroutine UI
	// khusus), sehingga kedua flag ini dibuat atomic agar aman dibaca/ditulis lintas goroutine
	// tanpa pengecualian, bukan hanya mengandalkan konvensi "selalu lewat goroutine UI".
	isScanning  atomic.Bool
	isExecuting atomic.Bool
	// asyncGeneration dinaikkan (Add 1) sebagai pernyataan PALING TERAKHIR pada setiap callback
	// fyne.Do yang menandai tuntasnya satu siklus pembaruan UI asinkron (pemindaian atau
	// eksekusi). isScanning/isExecuting sengaja di-set false LEBIH AWAL pada callback yang sama
	// (sebelum updateActionButtonsState dipanggil, agar tombol dievaluasi dengan status
	// terbaru) sehingga keduanya TIDAK bisa dipakai sebagai sinyal "seluruh pembaruan widget
	// pada goroutine ini sudah rampung". asyncGeneration mengisi celah itu lewat happens-before
	// bawaan operasi atomic: begitu nilainya teramati naik, seluruh pernyataan sebelumnya pada
	// goroutine yang sama dijamin sudah selesai. Dipakai terutama oleh tes untuk menunggu
	// goroutine latar belakang benar-benar tuntas sebelum memeriksa state/dialog.
	asyncGeneration    atomic.Uint64
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
		listDestFiles:     fsops.ListDestFileNames,
	}

	a.buildUI()
	return a
}

// buildUI menyusun tata letak (layout) jendela aplikasi.
func (a *App) buildUI() {
	// 1. Header Aplikasi
	titleLabel := widget.NewLabelWithStyle(AppTitle, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	subtitleLabel := widget.NewLabelWithStyle(AppSubtitle, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	badgeLabel := widget.NewLabelWithStyle(AppBadgeStatus, fyne.TextAlignTrailing, fyne.TextStyle{Bold: true})

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

// createInputSection membuat komponen untuk Input Daftar Foto.
func (a *App) createInputSection() fyne.CanvasObject {
	sectionTitle := widget.NewLabelWithStyle(SectionInputTitle, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	a.inputEntry = widget.NewMultiLineEntry()
	a.inputEntry.SetPlaceHolder(InputListPlaceholder)
	a.inputEntry.Wrapping = fyne.TextWrapWord
	a.inputEntry.SetMinRowsVisible(7)
	a.inputEntry.OnChanged = func(s string) {
		a.listText = s
		a.parseInputList()
	}

	btnPaste := widget.NewButtonWithIcon(BtnPasteList, theme.ContentPasteIcon(), func() {
		clip := a.window.Clipboard().Content()
		if clip != "" {
			a.inputEntry.SetText(clip)
		}
	})

	btnFile := widget.NewButtonWithIcon(BtnOpenTxtFile, theme.FileTextIcon(), func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			data, readErr := io.ReadAll(reader)
			if readErr != nil {
				dialog.ShowError(fmt.Errorf(FmtErrReadFile, readErr), a.window)
				return
			}
			a.lastExecutionTitle = reader.URI().Name()
			a.inputEntry.SetText(string(data))
		}, a.window)
	})

	btnClear := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		a.inputEntry.SetText("")
	})

	a.listStatusLabel = widget.NewLabel(LabelListEmpty)
	a.listStatusLabel.Wrapping = fyne.TextWrapWord

	btnRow := container.NewHBox(btnPaste, btnFile, btnClear)
	headerRow := container.NewBorder(nil, nil, sectionTitle, btnRow)

	return container.NewVBox(
		headerRow,
		a.inputEntry,
		a.listStatusLabel,
	)
}

// createFolderSection membuat komponen untuk Folder & Saringan.
func (a *App) createFolderSection() fyne.CanvasObject {
	sectionTitle := widget.NewLabelWithStyle(SectionFoldersTitle, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Folder Sumber
	a.sourceEntry = widget.NewEntry()
	a.sourceEntry.SetPlaceHolder(PlaceholderSource)
	a.sourceEntry.OnChanged = func(s string) {
		a.sourceDir = strings.TrimSpace(s)
		a.triggerAsyncScan()
	}

	btnBrowseSource := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
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

	btnBrowseDest := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
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

	rawInfoLabel := widget.NewLabelWithStyle(fmt.Sprintf(FmtRawFormatsInfo, len(scan.RawExtensions)), fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
	rawInfoLabel.Wrapping = fyne.TextWrapWord

	// Jaring Pengaman JPEG (R-U4)
	a.jpegWarningLabel = widget.NewLabel(WarnJPEGOnly)
	a.jpegWarningLabel.Wrapping = fyne.TextWrapWord
	a.jpegWarningCard = widget.NewCard("", CardTitleJPEGSafetyNet, a.jpegWarningLabel)
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

// createPreviewSection membuat komponen untuk Pratinjau Pencocokan 4 Keranjang.
func (a *App) createPreviewSection() fyne.CanvasObject {
	sectionTitle := widget.NewLabelWithStyle(SectionPreviewTitle, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	a.statusScanLabel = widget.NewLabel(StatusSelectSourceFolder)
	a.statusScanLabel.Wrapping = fyne.TextWrapWord

	a.refreshBtn = widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
		a.triggerAsyncScan()
	})

	topRow := container.NewBorder(nil, nil, sectionTitle, a.refreshBtn, a.statusScanLabel)

	// Tombol Aksi Keranjang
	btnViewNotFound := widget.NewButtonWithIcon(BtnViewNotFound, theme.VisibilityIcon(), func() {
		ShowNotFoundNamesDialog(a.window, a.matchResult.NotFoundNames)
	})

	btnViewCollisions := widget.NewButtonWithIcon(BtnViewCollisions, theme.VisibilityIcon(), func() {
		ShowCollisionsDialog(a.window, a.matchResult.Collisions)
	})

	// 4 Keranjang, masing-masing dengan garis aksen warna semantik sesuai kategorinya
	a.matchedCard = NewBasketCard(LabelBasketMatched, LabelZeroFiles, SubMatchedInitial, nil, theme.Color(theme.ColorNameSuccess))
	a.notFoundCard = NewBasketCard(LabelBasketNotFound, LabelZeroNames, SubNotFoundInitial, btnViewNotFound, theme.Color(theme.ColorNameError))
	a.collisionCard = NewBasketCard(LabelBasketCollision, LabelZeroFiles, SubCollisionInitial, btnViewCollisions, theme.Color(theme.ColorNameWarning))
	a.skippedCard = NewBasketCard(LabelBasketSkipped, LabelZeroFiles, SubSkippedInitial, nil, theme.Color(theme.ColorNamePlaceHolder))

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

// createActionSection membuat komponen untuk Tombol Tindakan Eksekusi.
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

	notes := widget.NewLabel(NoteActionSection)
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
		a.statusScanLabel.SetText(StatusSelectSourceStart)
		a.resetBasketCards()
		a.updateActionButtonsState()
		return
	}

	// Jangan tumpuk scan jika sedang berjalan
	if a.isScanning.Load() {
		return
	}

	a.isScanning.Store(true)
	a.statusScanLabel.SetText(StatusScanning)
	a.refreshBtn.Disable()
	a.updateActionButtonsState()

	sourceDir := a.sourceDir
	destDir := a.destDir
	recursive := a.includeSubfolders
	rawFilter := a.rawFilterEnabled
	targets := a.listResult.Items
	listDestFiles := a.listDestFiles

	go func() {
		// 1. Pindai berkas sumber
		scanRes, err := scan.Scan(scan.Options{
			RootDir:   sourceDir,
			Recursive: recursive,
		})

		// 2. Baca nama berkas di tujuan jika direktori tujuan sudah ada.
		// R-A6: satu-satunya pintu I/O berkas adalah internal/fsops, jadi pembacaan dilakukan
		// lewat listDestFiles (default fsops.ListDestFileNames), bukan os.ReadDir langsung di
		// lapisan UI.
		destFiles, destErr := listDestFiles(destDir)

		// 3. Lakukan pencocokan 4 keranjang
		matchRes := match.Match(targets, scanRes.Files, match.Options{
			RawFilter:         rawFilter,
			ExistingDestFiles: destFiles,
		})

		// 4. Perbarui antarmuka di UI thread
		fyne.Do(func() {
			a.isScanning.Store(false)
			a.refreshBtn.Enable()

			// R-C2: peringatan pemindaian maupun pembacaan folder tujuan tidak pernah ditelan diam-diam.
			switch {
			case err != nil && destErr != nil:
				a.statusScanLabel.SetText(fmt.Sprintf(FmtScanAndDestWarning, err, destErr))
			case err != nil:
				a.statusScanLabel.SetText(fmt.Sprintf(FmtScanWarning, err))
			case destErr != nil:
				a.statusScanLabel.SetText(fmt.Sprintf(FmtDestReadWarning, destErr))
			default:
				a.statusScanLabel.SetText(fmt.Sprintf(FmtScanComplete, len(scanRes.Files)))
			}

			a.scanResult = scanRes
			a.matchResult = matchRes

			// Perbarui 4 Keranjang
			a.matchedCard.SetValues(
				fmt.Sprintf(FmtCountFiles, matchRes.TotalMatchedFiles),
				fmt.Sprintf(FmtTotalSize, FormatBytes(matchRes.TotalMatchedSize)),
			)

			a.notFoundCard.SetValues(
				fmt.Sprintf(FmtCountNames, len(matchRes.NotFoundNames)),
				SubNotFoundAfterScan,
			)

			a.collisionCard.SetValues(
				fmt.Sprintf(FmtCountFiles, len(matchRes.Collisions)),
				SubCollisionAfterScan,
			)

			a.skippedCard.SetValues(
				fmt.Sprintf(FmtCountFiles, matchRes.SkippedCount),
				SubSkippedAfterScan,
			)

			// Jaring pengaman JPEG (R-U4)
			if matchRes.JPEGOnlyWarning {
				a.jpegWarningContainer.Show()
			} else {
				a.jpegWarningContainer.Hide()
			}

			a.updateActionButtonsState()
			a.asyncGeneration.Add(1)
		})
	}()
}

func (a *App) resetBasketCards() {
	a.matchedCard.SetValues(LabelZeroFiles, SubMatchedInitial)
	a.notFoundCard.SetValues(LabelZeroNames, SubNotFoundInitial)
	a.collisionCard.SetValues(LabelZeroFiles, SubCollisionInitial)
	a.skippedCard.SetValues(LabelZeroFiles, SubSkippedInitial)
	a.jpegWarningContainer.Hide()
}

// updateActionButtonsState memperbarui status aktif/nonaktif tombol eksekusi (R-P4).
func (a *App) updateActionButtonsState() {
	canExecute := !a.isScanning.Load() &&
		!a.isExecuting.Load() &&
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
		dialog.ShowError(errors.New(WarnSameFolder), a.window)
		return
	}

	if a.matchResult.TotalMatchedFiles == 0 {
		dialog.ShowInformation(DialogNoFilesTitle, LabelNoMatches, a.window)
		return
	}

	var title string
	var message string

	totalFiles := a.matchResult.TotalMatchedFiles
	totalSize := FormatBytes(a.matchResult.TotalMatchedSize)

	if mode == fsops.ModeCopy {
		title = DialogConfirmCopyTitle
		message = fmt.Sprintf(FmtConfirmCopyMessage, totalFiles, totalSize, a.sourceDir, a.destDir)
	} else {
		title = DialogConfirmMoveTitle
		message = fmt.Sprintf(FmtConfirmMoveMessage, totalFiles, totalSize, a.sourceDir, a.destDir, DialogMoveWarningText)
	}

	content := widget.NewLabel(message)
	content.Wrapping = fyne.TextWrapWord

	confirmDialog := dialog.NewCustomConfirm(title, DialogBtnProceed, DialogBtnCancel, content, func(confirmed bool) {
		if confirmed {
			a.startExecution(mode)
		}
	}, a.window)

	if mode == fsops.ModeMove {
		confirmDialog.SetConfirmImportance(widget.DangerImportance)
	}

	confirmDialog.Resize(fyne.NewSize(480, 260))
	confirmDialog.Show()
}

// startExecution mengeksekusi operasi salin/pindah dengan modal progres dan pembatalan (R-U3, R-A4).
func (a *App) startExecution(mode fsops.OpMode) {
	a.isExecuting.Store(true)
	a.updateActionButtonsState()

	ctx, cancel := context.WithCancel(context.Background())
	a.execContextCancel = cancel

	// Buat Modal Progres
	progressBar := widget.NewProgressBar()
	progressBar.SetValue(0)

	fileLabel := widget.NewLabel(StatusPreparingExecution)
	fileLabel.Wrapping = fyne.TextWrapWord

	byteLabel := widget.NewLabel(fmt.Sprintf(FmtByteProgressInitial, FormatBytes(a.matchResult.TotalMatchedSize)))

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
		listTitle = LabelDefaultListTitle
	}
	uniqueTargets := len(a.listResult.Items)
	notFoundList := a.matchResult.NotFoundNames

	go func() {
		startTime := time.Now()

		// Jalankan operasi fsops
		execRes, execErr := fsops.Execute(ctx, fsops.Options{
			Mode:         mode,
			SourceDir:    sourceDir,
			DestDir:      destDir,
			MatchedItems: matchedItems,
			ProgressCallback: func(p fsops.Progress) {
				fyne.Do(func() {
					if p.TotalFiles > 0 {
						progressBar.SetValue(float64(p.CurrentFileIndex) / float64(p.TotalFiles))
					}
					fileLabel.SetText(fmt.Sprintf(FmtProcessingFile, p.CurrentFileName, p.CurrentFileIndex, p.TotalFiles))
					byteLabel.SetText(fmt.Sprintf(FmtByteProgress, FormatBytes(p.BytesProcessed), FormatBytes(p.TotalBytes)))
				})
			},
		})

		// Bug kritis hasil audit (R-C2/R-U6): kegagalan total (mis. folder tujuan tidak bisa dibuat)
		// WAJIB dilaporkan apa adanya ke pengguna, bukan ditelan diam-diam lalu ditampilkan sebagai sukses.
		if execErr != nil {
			fyne.Do(func() {
				a.isExecuting.Store(false)
				progressDialog.Hide()
				a.updateActionButtonsState()
				dialog.ShowError(fmt.Errorf(FmtExecutionFailedMessage, mode, execErr), a.window)
				a.triggerAsyncScan()
				a.asyncGeneration.Add(1)
			})
			return
		}

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

		// Tulis berkas _laporan.txt di folder tujuan (R-P7). Kegagalan tulis tidak membatalkan hasil
		// yang sudah diproses, tapi WAJIB dilaporkan eksplisit ke pengguna, bukan ditelan (R-C2).
		reportPath, reportErr := report.WriteFile(destDir, reportData)
		reportContent := report.Generate(reportData)

		fyne.Do(func() {
			a.isExecuting.Store(false)
			progressDialog.Hide()
			a.updateActionButtonsState()

			if reportErr != nil {
				dialog.ShowError(fmt.Errorf(FmtErrReportWriteFailed, reportErr), a.window)
			}

			// Tampilkan Ringkasan Hasil
			a.showSummaryDialog(execRes, reportPath, reportContent)

			// Pindai ulang folder untuk memperbarui status
			a.triggerAsyncScan()
			a.asyncGeneration.Add(1)
		})
	}()
}

// showSummaryDialog menampilkan modal ringkasan setelah eksekusi selesai.
// reportPath boleh kosong bila penulisan _laporan.txt gagal (lihat pesan error terpisah yang
// sudah ditampilkan oleh pemanggil) — ringkasan tetap jujur, tidak mengklaim path yang tidak ada.
func (a *App) showSummaryDialog(res fsops.Result, reportPath, reportContent string) {
	var statusMsg string
	if res.Cancelled {
		statusMsg = SummaryCancelledLabel
	} else if len(res.FailedFiles) > 0 {
		statusMsg = SummaryPartialLabel
	} else {
		statusMsg = SummarySuccessLabel
	}

	reportLine := MsgReportSaveFailed
	if reportPath != "" {
		reportLine = fmt.Sprintf(FmtReportSavedAt, reportPath)
	}

	detailText := fmt.Sprintf(
		FmtSummaryDetail,
		statusMsg,
		len(res.CopiedFiles),
		FormatBytes(res.TotalBytesCopied),
		len(res.FailedFiles),
		len(res.CollisionRecords),
		reportLine,
	)

	label := widget.NewLabel(detailText)
	label.Wrapping = fyne.TextWrapWord

	btnOpenFolder := widget.NewButtonWithIcon(BtnOpenDestFolder, theme.FolderOpenIcon(), func() {
		if err := OpenFolderInFileManager(res.DestDir); err != nil {
			dialog.ShowError(fmt.Errorf(FmtErrOpenFolderFailed, err), a.window)
		}
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
