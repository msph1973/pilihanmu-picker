package ui

// Konstanta teks antarmuka Bahasa Indonesia terpusat (R-C5, R-U5).
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

	// Panel 2: Folder & Saringan
	SectionFoldersTitle = "Direktori & Pengaturan"
	LabelSourceFolder   = "Folder Sumber RAW:"
	LabelDestFolder     = "Folder Tujuan:"
	PlaceholderSource   = "Pilih folder sumber RAW..."
	PlaceholderDest     = "Pilih folder tujuan..."
	CheckRecursive      = "Cari ke dalam sub-folder (rekursif)"
	CheckRawFilter      = "Hanya format foto RAW"
	LabelRawFormatsInfo = "Mendukung 27 format RAW populer (CR2, CR3, ARW, NEF, DNG, RAF, ORF, dll)"
	WarnSameFolder      = "Folder sumber dan folder tujuan tidak boleh sama!"
	WarnJPEGOnly        = "Perhatian: Folder sumber hanya berisi JPEG. Nonaktifkan filter RAW jika ingin memproses berkas JPEG."

	// Panel 3: Pratinjau
	SectionPreviewTitle    = "Pratinjau Hasil"
	StatusScanning         = "Memindai folder sumber..."
	LabelBasketMatched     = "Cocok"
	LabelBasketNotFound    = "Tidak Ditemukan"
	LabelBasketCollision   = "Nama Ganda"
	LabelBasketSkipped     = "Dilewati"
	BtnViewNotFound        = "Lihat"
	CopiedToClipboardToast = "Daftar nama berhasil disalin!"
	LabelNoMatches         = "Tidak ada berkas cocok yang ditemukan."

	// Panel 4: Tombol Tindakan
	SectionActionTitle = "Eksekusi"
	BtnActionCopy      = "Salin ke Tujuan"
	BtnActionMove      = "Pindahkan ke Tujuan"
	TooltipCopy        = "Menyalin berkas terpilih ke folder tujuan tanpa mengubah folder sumber."
	TooltipMove        = "Menyalin ke tujuan, memverifikasi ukuran berkas, lalu menghapus sumber."

	// Konfirmasi Dialog (R-U1)
	DialogConfirmCopyTitle = "Konfirmasi Salin"
	DialogConfirmMoveTitle = "Konfirmasi Pindah (Hapus Sumber)"
	DialogBtnCancel        = "Batal"
	DialogBtnProceed       = "Lanjutkan"
	DialogMoveWarningText  = "PERINGATAN KESELAMATAN DATA:\nBerkas di folder sumber akan DIHAPUS setelah diverifikasi 100% identik di tujuan.\nPastikan Anda memiliki cadangan (backup) bila diperlukan."

	// Modal Progres (R-U3)
	ProgressTitleExecuting = "Memproses Berkas..."
	BtnCancelProgress      = "Batalkan"
	StatusCancelling       = "Membatalkan... Menunggu operasi saat ini selesai."

	// Hasil Akhir
	DialogSummaryTitle    = "Hasil Pemrosesan"
	SummarySuccessLabel   = "Pemrosesan selesai dengan sukses!"
	SummaryPartialLabel   = "Pemrosesan selesai dengan beberapa catatan."
	SummaryCancelledLabel = "Pemrosesan dibatalkan pengguna."
	BtnOpenDestFolder     = "Buka Folder Tujuan"
	BtnViewReport         = "Lihat Laporan"
	BtnClose              = "Tutup"
)
