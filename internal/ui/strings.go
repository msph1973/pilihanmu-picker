package ui

// Konstanta teks antarmuka Bahasa Indonesia terpusat (R-C5, R-U5).
const (
	// Info Aplikasi & Header
	AppTitle       = "Pilihanmu Picker"
	AppSubtitle    = "Alat Pemindah Berkas RAW Hasil Seleksi Klien"
	AppBadgeStatus = "100% Offline • Bebas Jaringan • Aman"

	// Langkah 1: Daftar Foto
	SectionInputTitle    = "1. Daftar Foto Pilihan Klien"
	InputListPlaceholder = "Tempel daftar foto dari Dashboard Pilihanmu di sini...\nContoh:\nDSC_0001\nDSC_0002\n_MG_0003"
	BtnPasteClipboard    = "Tempel dari Clipboard"
	BtnClearInput        = "Kosongkan"
	BtnOpenFile          = "Pilih Berkas .txt..."
	LabelListEmpty       = "Daftar foto belum diisi (masukkan nama berkas atau pilih berkas .txt)"
	LabelListInvalidUTF8 = "Format teks tidak valid (harus UTF-8)"

	// Langkah 2: Folder & Saringan
	SectionFoldersTitle = "2. Lokasi Folder & Pengaturan"
	LabelSourceFolder   = "Folder Sumber (tempat RAW tersimpan):"
	LabelDestFolder     = "Folder Tujuan (hasil seleksi):"
	PlaceholderSource   = "Pilih folder yang berisi berkas RAW fotografer..."
	PlaceholderDest     = "Pilih folder baru untuk menampung foto terpilih..."
	BtnBrowseSource     = "Pilih Folder Sumber..."
	BtnBrowseDest       = "Pilih Folder Tujuan..."
	CheckRecursive      = "Sertakan sub-folder (pencarian rekursif)"
	CheckRawFilter      = "Hanya salin format foto RAW"
	LabelRawFormatsInfo = "Format RAW didukung: CR2, CR3, ARW, NEF, DNG, RAF, ORF, RW2, PEF, 3FR, dll (27 format)"
	WarnSameFolder      = "Folder sumber dan folder tujuan tidak boleh sama!"
	WarnJPEGOnly        = "Perhatian: Sepertinya Anda memotret format JPEG saja. Matikan centang 'Hanya salin format foto RAW' untuk memproses berkas JPEG."

	// Langkah 3: Pratinjau
	SectionPreviewTitle    = "3. Pratinjau Pencocokan"
	BtnRefreshPreview      = "Pindai Ulang Folder"
	StatusScanning         = "Sedang memindai folder sumber..."
	LabelBasketMatched     = "Cocok"
	LabelBasketNotFound    = "Tidak Ketemu"
	LabelBasketCollision   = "Tabrakan Nama"
	LabelBasketSkipped     = "Dilewati"
	BtnCopyNotFound        = "Salin Daftar Tidak Ketemu"
	CopiedToClipboardToast = "Daftar nama tidak ketemu berhasil disalin ke clipboard!"
	LabelNoMatches         = "Belum ada berkas yang cocok ditemukan."

	// Langkah 4: Tombol Tindakan
	SectionActionTitle = "4. Eksekusi Tindakan"
	BtnActionCopy      = "Salin ke Folder Tujuan"
	BtnActionMove      = "Pindahkan ke Folder Tujuan"
	TooltipCopy        = "Menyalin berkas terpilih ke folder tujuan. Folder sumber tetap utuh tanpa perubahan."
	TooltipMove        = "Menyalin ke tujuan, memverifikasi ukuran, lalu menghapus berkas sumber satu per satu."

	// Konfirmasi Dialog (R-U1)
	DialogConfirmCopyTitle = "Konfirmasi Salin Berkas"
	DialogConfirmMoveTitle = "Konfirmasi Pindahkan Berkas (Hapus Sumber)"
	DialogBtnCancel        = "Batal"
	DialogBtnProceed       = "Ya, Lanjutkan"
	DialogMoveWarningText  = "PERINGATAN KESELAMATAN DATA:\nBerkas RAW di folder sumber akan DIHAPUS satu per satu setelah diverifikasi ukurannya sama di folder tujuan.\nPastikan Anda memiliki cadangan (backup) jika diperlukan."

	// Modal Progres (R-U3)
	ProgressTitleExecuting = "Memproses Berkas..."
	BtnCancelProgress      = "Batalkan Proses"
	StatusCancelling       = "Membatalkan... Menunggu berkas yang sedang disalin selesai."

	// Hasil Akhir
	DialogSummaryTitle    = "Hasil Pemrosesan Selesai"
	SummarySuccessLabel   = "Proses selesai dengan sukses!"
	SummaryPartialLabel   = "Proses selesai dengan beberapa catatan/kegagalan."
	SummaryCancelledLabel = "Proses dibatalkan oleh pengguna."
	BtnOpenDestFolder     = "Buka Folder Tujuan"
	BtnViewReport         = "Lihat Berkas Laporan"
	BtnClose              = "Tutup"
)
