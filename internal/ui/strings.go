package ui

// Konstanta teks antarmuka Bahasa Indonesia terpusat (R-C5, R-U5).
//
// Seluruh teks yang terlihat oleh pengguna — label, judul dialog, tombol, dan templat pesan
// (dipakai lewat fmt.Sprintf/fmt.Errorf) — WAJIB didefinisikan di sini, bukan sebagai literal
// langsung di app.go/widgets.go, agar bisa ditinjau sekali jalan (R-C5).
const (
	// Info Aplikasi & Header
	AppTitle       = "Pilihanmu Picker"
	AppSubtitle    = "Pemilah Berkas RAW Seleksi Klien"
	AppBadgeStatus = "100% Offline | Aman"

	// Panel 1: Daftar Foto
	SectionInputTitle    = "Daftar Seleksi Klien"
	InputListPlaceholder = "Tempel daftar foto dari Dashboard Pilihanmu...\nContoh:\nDSC_0001\nDSC_0002\n_MG_0003"
	LabelListEmpty       = "Tempel daftar foto atau buka berkas .txt hasil seleksi"
	LabelListInvalidUTF8 = "Format teks tidak valid (harus UTF-8)"
	BtnPasteList         = "Tempel"
	BtnOpenTxtFile       = "Buka .txt"
	FmtErrReadFile       = "Gagal membaca berkas: %w"

	// Panel 2: Folder & Saringan
	SectionFoldersTitle    = "Direktori & Pengaturan"
	LabelSourceFolder      = "Folder Sumber RAW:"
	LabelDestFolder        = "Folder Tujuan:"
	PlaceholderSource      = "Pilih folder sumber RAW..."
	PlaceholderDest        = "Pilih folder tujuan..."
	CheckRecursive         = "Cari ke dalam sub-folder (rekursif)"
	CheckRawFilter         = "Hanya format foto RAW"
	FmtRawFormatsInfo      = "Mendukung %d format RAW populer (CR2, CR3, ARW, NEF, DNG, RAF, ORF, dll)"
	WarnSameFolder         = "Folder sumber dan folder tujuan tidak boleh sama!"
	WarnJPEGOnly           = "Perhatian: Folder sumber hanya berisi JPEG. Nonaktifkan filter RAW jika ingin memproses berkas JPEG."
	CardTitleJPEGSafetyNet = "Jaring Pengaman Format"

	// Panel 3: Pratinjau
	SectionPreviewTitle      = "Pratinjau Hasil"
	StatusScanning           = "Memindai folder sumber..."
	StatusSelectSourceFolder = "Pilih folder sumber untuk memindai berkas..."
	StatusSelectSourceStart  = "Pilih folder sumber untuk memulai pencarian."
	LabelBasketMatched       = "Cocok"
	LabelBasketNotFound      = "Tidak Ditemukan"
	LabelBasketCollision     = "Nama Ganda"
	LabelBasketSkipped       = "Dilewati"
	BtnViewNotFound          = "Lihat"
	BtnViewCollisions        = "Lihat"
	LabelNoMatches           = "Tidak ada berkas cocok yang ditemukan."

	// Nilai & keterangan awal/reset 4 Keranjang (dipakai saat inisialisasi dan saat direset)
	LabelZeroFiles      = "0 berkas"
	LabelZeroNames      = "0 nama"
	SubMatchedInitial   = "Belum ada berkas cocok"
	SubNotFoundInitial  = "Semua berkas ditemukan"
	SubCollisionInitial = "Tidak ada nama ganda"
	SubSkippedInitial   = "Format non-RAW / di luar pilihan"

	// Keterangan 4 Keranjang setelah pemindaian selesai
	FmtCountFiles         = "%d berkas"
	FmtCountNames         = "%d nama"
	FmtTotalSize          = "Total ukuran: %s"
	SubNotFoundAfterScan  = "Tidak ada di folder sumber"
	SubCollisionAfterScan = "Otomatis diberi akhiran (2)"
	SubSkippedAfterScan   = "Bukan pilihan / non-RAW"

	// Peringatan hasil pemindaian (R-C2: dilaporkan, tidak ditelan diam-diam)
	FmtScanWarning        = "Peringatan pemindaian: %v"
	FmtScanComplete       = "Pindai selesai: %d total berkas ditemukan di sumber."
	FmtDestReadWarning    = "Peringatan membaca folder tujuan: %v"
	FmtScanAndDestWarning = "Peringatan pemindaian: %v\nPeringatan folder tujuan: %v"

	// Panel 4: Tombol Tindakan
	SectionActionTitle = "Eksekusi"
	BtnActionCopy      = "Salin ke Tujuan"
	BtnActionMove      = "Pindahkan ke Tujuan"
	NoteActionSection  = "- Salin: Folder sumber tetap utuh aman.\n- Pindahkan: Sumber dihapus hanya setelah ukuran terverifikasi 100% identik."

	// Konfirmasi Dialog (R-U1)
	DialogConfirmCopyTitle = "Konfirmasi Salin"
	DialogConfirmMoveTitle = "Konfirmasi Pindah (Hapus Sumber)"
	DialogNoFilesTitle     = "Tidak Ada Berkas"
	DialogBtnCancel        = "Batal"
	DialogBtnProceed       = "Lanjutkan"
	DialogMoveWarningText  = "PERINGATAN KESELAMATAN DATA:\nBerkas di folder sumber akan DIHAPUS setelah diverifikasi 100% identik di tujuan.\nPastikan Anda memiliki cadangan (backup) bila diperlukan."
	FmtConfirmCopyMessage  = "Anda akan MENYALIN %d berkas foto (%s) ke folder tujuan.\n\nSumber: %s\nTujuan: %s\n\nFolder sumber tidak akan diubah atau dihapus."
	FmtConfirmMoveMessage  = "Anda akan MEMINDAHKAN %d berkas foto (%s) ke folder tujuan.\n\nSumber: %s\nTujuan: %s\n\n%s"

	// Modal Progres (R-U3)
	ProgressTitleExecuting   = "Memproses Berkas..."
	BtnCancelProgress        = "Batalkan"
	StatusCancelling         = "Membatalkan... Menunggu operasi saat ini selesai."
	StatusPreparingExecution = "Menyiapkan pemrosesan..."
	LabelDefaultListTitle    = "Daftar Teks"
	FmtByteProgressInitial   = "0 B / %s"
	FmtProcessingFile        = "Memproses: %s (%d dari %d berkas)"
	FmtByteProgress          = "%s / %s"

	// Kesalahan eksekusi (R-U6: actionable, bukan generik; R-C2: tidak ditelan)
	FmtExecutionFailedMessage = "Gagal menjalankan proses %s: %v\n\nTidak ada berkas yang berhasil diproses. Periksa apakah folder tujuan bisa dibuat (izin akses, ruang penyimpanan, atau path tidak valid)."
	FmtErrReportWriteFailed   = "Berkas foto berhasil diproses, tetapi berkas _laporan.txt GAGAL ditulis ke folder tujuan: %v"
	FmtErrOpenFolderFailed    = "Gagal membuka folder tujuan: %w"

	// Hasil Akhir
	DialogSummaryTitle    = "Hasil Pemrosesan"
	SummarySuccessLabel   = "Pemrosesan selesai dengan sukses!"
	SummaryPartialLabel   = "Pemrosesan selesai dengan beberapa catatan."
	SummaryCancelledLabel = "Pemrosesan dibatalkan pengguna."
	BtnOpenDestFolder     = "Buka Folder Tujuan"
	BtnViewReport         = "Lihat Laporan"
	BtnClose              = "Tutup"
	FmtSummaryDetail      = "%s\n\n- Berhasil: %d berkas (%s)\n- Gagal: %d berkas\n- Nama ganda diselesaikan: %d berkas\n\n%s"
	FmtReportSavedAt      = "Laporan disimpan di:\n%s"
	MsgReportSaveFailed   = "Laporan GAGAL disimpan ke folder tujuan (lihat pesan kesalahan terpisah)."

	// Dialog Daftar "Tidak Ditemukan" (widgets.go)
	DialogNotFoundEmptyTitle  = "Tidak Ditemukan"
	MsgAllNamesFound          = "Semua nama dalam daftar berhasil ditemukan di folder sumber!"
	DialogNotFoundListTitle   = "Daftar Tidak Ditemukan"
	FmtNotFoundHeader         = "Ditemukan %d nama dari daftar yang tidak ada di folder sumber:"
	BtnCopyToClipboard        = "Salin ke Clipboard"
	DialogCopiedTitle         = "Berhasil Disalin"
	FmtCopiedNamesToClipboard = "%d nama berhasil disalin ke clipboard."

	// Dialog Daftar "Tabrakan Nama" (widgets.go, R-U2/PRD §4.5)
	DialogCollisionListTitle       = "Daftar Tabrakan Nama"
	DialogNoCollisionsTitle        = "Tidak Ada Tabrakan"
	MsgNoCollisions                = "Tidak ada berkas yang mengalami tabrakan nama."
	FmtCollisionHeader             = "Ditemukan %d tabrakan nama. Berkas akan diberi akhiran otomatis agar tidak menimpa:"
	FmtCollisionLine               = "%s  ->  %s"
	FmtCopiedCollisionsToClipboard = "%d baris tabrakan berhasil disalin ke clipboard."

	// Dialog Pratinjau Laporan (widgets.go)
	DialogReportViewerTitle = "Laporan Hasil Seleksi (_laporan.txt)"
	BtnCopyReportContent    = "Salin Isi Laporan"
	MsgReportCopied         = "Isi laporan berhasil disalin ke clipboard."
	FmtReportFileLabel      = "Berkas laporan: %s"
	MsgReportNotSavedToDisk = "Berkas laporan: (gagal disimpan ke disk — isi berikut hanya tersedia di memori untuk sesi ini)"
)
