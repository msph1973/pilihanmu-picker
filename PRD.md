# PRD — Pilihanmu Picker (standalone)

> **Status**: DRAF SPESIFIKASI — belum ada satu baris kode pun.
> **Tanggal**: 24 September 2026
> **Pemilik**: msph1973
> **Repo induk (tidak boleh diimpor)**: `msph1973/pilih`

Aplikasi desktop **offline** yang membantu fotografer memindahkan file RAW hasil
seleksi klien ke folder baru, berdasarkan daftar `.txt` yang sudah bisa diunduh
dari Dashboard Pilihanmu.

Repo ini **berdiri sendiri**: tidak mengimpor paket, skema, atau kode apa pun
dari repo induk. Satu-satunya keterkaitan adalah **format `.txt`** yang dibekukan
di §6.

---

## 1. Masalah

Alur hari ini di Dashboard Pilihanmu:

1. Fotografer membuka `https://pilihanmu.web.id/dashboard`
2. Menemukan galeri yang sudah disubmit klien
3. Menekan **Unduh (.txt)** → dapat `<judul> - Daftar Foto.txt`, atau menekan
   **Salin Daftar Foto** → daftar masuk clipboard
4. File `.txt` itu berisi nama file yang dipilih klien, **satu basename per baris**

Masalahnya muncul di langkah 5, yang sepenuhnya manual: fotografer harus mencari
satu per satu nama di folder RAW-nya, menyeleksinya, lalu memindahkannya ke folder
kerja. Untuk 120 pilihan dari 800 file, ini melelahkan, rawan salah, dan tidak ada
laporan apa pun kalau ada nama yang tidak ketemu.

**Yang hilang**: alat yang mengubah "daftar nama" menjadi "folder berisi file
RAW yang benar", dengan jejak audit.

## 2. Pengguna

Fotografer (atau asistennya) yang:
- Menyimpan RAW di Windows dalam satu folder atau pohon folder per-tanggal
- Sudah menerima submission dari klien di Pilihanmu
- Tidak ingin membuka terminal
- Mengharapkan alat yang bisa dijalankan tanpa instalasi rumit

Bukan pengguna: klien (mereka tidak pernah menyentuh alat ini), dan bukan
pengembang.

## 3. Alur utama (dikunci)

```text
1. Fotografer unduh / salin  <judul> - Daftar Foto.txt  dari Dashboard
2. Buka app
3. Tempel daftar ke kotak teks  ATAU  pilih file .txt
4. Pilih folder sumber (tempat RAW berada)
5. Pilih folder tujuan (folder baru)
6. App memindai + mencocokkan  ->  PRATINJAU (wajib)
7. Fotografer menekan "Salin ke folder tujuan" atau "Pindahkan ke folder tujuan"
8. App mengeksekusi, menampilkan progres, menulis _laporan.txt
9. Selesai — folder tujuan siap diimpor ke Lightroom
```

Langkah 6 **tidak boleh dilewati**. Lihat §4.5.

## 4. Spesifikasi fungsional

### 4.1 Masukan

Dua jalur, keduanya harus didukung:

| Jalur | Perilaku |
|---|---|
| **Tempel teks** | Kotak multi-baris. Isi clipboard dari tombol "Salin Daftar Foto" boleh langsung ditempel. |
| **Pilih file** | Dialog file, filter `.txt`. Isi file dibaca UTF-8 (toleran BOM). |

Keduanya menghasilkan struktur data yang sama; tidak ada jalur yang diistimewakan.

### 4.2 Parsing daftar

Aturan parsing (semua wajib, masing-masing punya tes):

| # | Aturan | Alasan |
|---|---|---|
| P-1 | Pisah pada `\r\n`, `\n`, dan `\r` | File `.txt` yang diunduh di Windows berakhir CRLF. Kalau `\r` tidak dibuang, **semua** nama akan dianggap "tidak ketemu" — ini kegagalan paling membingungkan yang bisa terjadi. |
| P-2 | Trim spasi dan tab di awal/akhir baris | Hasil tempel dari clipboard sering membawa spasi tak terlihat. |
| P-3 | Buang baris kosong | — |
| P-4 | Buang prefiks path bila ada (`D:\x\DSC_0001.CR3` → `DSC_0001`) | Toleransi untuk daftar yang dibuat manual. |
| P-5 | Buang ekstensi bila ada (`DSC_0001.CR3` → `DSC_0001`) | Toleransi untuk daftar yang dibuat manual. |
| P-6 | Dedupe case-insensitive, pertahankan kemunculan pertama | Klien bisa memilih file yang sama dua kali; jangan salin dua kali. |
| P-7 | Urutan tidak diubah | Urutan = `pick_order`; berguna untuk menamai ulang atau menelusuri. |

Hasil parsing dilaporkan ke pengguna: `N baris → M nama unik (K duplikat dibuang)`.

### 4.3 Pemindaian & pencocokan

- Pindai folder sumber. **Rekursif** bila checkbox "Sertakan sub-folder" aktif
  (**default: aktif**) — banyak fotografer menyimpan per-tanggal.
- Untuk setiap file: ambil basename tanpa ekstensi, lowercase, bandingkan dengan
  himpunan nama hasil §4.2.
- Pencocokan **case-insensitive**, tetapi nama file asli di disk dipertahankan
  (R-P6).
- Jangan ikuti symlink direktori (cegah loop tak terbatas).
- Lewati direktori sistem/tak terbaca tanpa menghentikan pemindaian; kumpulkan
  sebagai peringatan.

Hasil pemindaian dikelompokkan menjadi empat keranjang:

| Keranjang | Isi |
|---|---|
| **Cocok** | nama → satu atau lebih path sumber |
| **Tidak ketemu** | nama dari daftar yang tidak punya file pasangan |
| **Tabrakan** | basename sama **dan** ekstensi sama, berasal dari path sumber berbeda |
| **Dilewati** | file di folder sumber yang tidak dipilih klien (dihitung saja, tidak ditampilkan satu per satu) |

### 4.4 Saringan ekstensi RAW

Hanya file dengan ekstensi berikut yang ikut disalin (case-insensitive):

```text
3FR ARW CR2 CR3 CRW DCR DNG ERF FFF IIQ KDC MEF MOS MRW NEF NRW
ORF PEF PTX RAF RAW RWL RW2 SR2 SRF SRW X3F
```

Saringan ini **terlihat oleh pengguna** (ditampilkan sebagai daftar, bukan
disembunyikan) dan bisa dimatikan lewat satu kontrol.

**Jaring pengaman wajib (R-U4)**: bila hasil cocok = 0 tetapi ada nama dari daftar
yang punya file dengan ekstensi non-RAW (mis. `.JPG`), tampilkan peringatan
khusus — *"Sepertinya Anda memotret JPEG saja. Matikan saringan RAW untuk
memasukkan file JPEG."* Tanpa ini, fotografer JPEG-only hanya melihat "0 cocok"
tanpa tahu sebabnya.

### 4.5 Pratinjau (wajib)

Sebelum tombol eksekusi bisa ditekan, tampilkan:

- Jumlah nama unik di daftar
- **Cocok**: jumlah file (bukan jumlah nama — satu nama bisa jadi beberapa file)
- **Tidak ketemu**: jumlah + **daftar namanya**, bisa disalin
- **Tabrakan**: jumlah + pasangan path sumber
- **Sudah ada di tujuan**: file yang akan diberi sufiks ` (2)`
- Peringatan §4.4 bila berlaku

Tombol eksekusi **nonaktif** selama pemindaian berjalan.

Pratinjau tidak boleh diringkas menjadi satu angka. Fotografer harus bisa melihat
nama mana yang tidak ketemu — itu satu-satunya cara ia tahu ada yang salah.

### 4.6 Eksekusi

Dua tombol, **tanpa default** (keputusan pemilik 2026-09-24):

| Tombol | Perilaku |
|---|---|
| **Salin ke folder tujuan** | Menyalin. Folder sumber tidak pernah disentuh. |
| **Pindahkan ke folder tujuan** | Menyalin, memverifikasi, lalu menghapus sumber. |

Keduanya membuka dialog konfirmasi yang menyebutkan jumlah file dan folder tujuan.

Aturan eksekusi:

| # | Aturan |
|---|---|
| E-1 | Penyalinan **streaming** (`io.Copy`) — file RAW bisa 60 MB+; jangan pernah `os.ReadFile`. |
| E-2 | Verifikasi = bandingkan ukuran file tujuan dengan sumber. Bila tidak sama → **jangan** hapus sumber, tandai gagal. |
| E-3 | Mode Pindahkan: hapus sumber **per file**, segera setelah file itu terverifikasi. Bukan borongan di akhir. |
| E-4 | Tabrakan nama di tujuan → tulis `nama (2).EXT`. **Tidak pernah menimpa.** Bila `(2)` juga ada, coba `(3)`, dst. |
| E-5 | Bila `.xmp` pendamping ada di folder sumber dengan basename sama, ikutkan (diperlakukan sebagai file biasa, termasuk aturan tabrakan). |
| E-6 | Satu file gagal **tidak** menghentikan sisanya. Kumpulkan kegagalan, lanjutkan. |
| E-7 | Progres terlihat dan bisa dibatalkan. Pembatalan menghentikan setelah file yang sedang berjalan selesai — tidak ada file setengah tertulis. |
| E-8 | File terkunci (Lightroom sedang memegang) → retry 3× dengan jeda 250 ms / 500 ms / 1000 ms, lalu tandai gagal dan lanjut. |

### 4.7 Keluaran

- Folder tujuan **rata (flatten)** — tanpa subfolder, keputusan pemilik 2026-09-24.
- Berkas `_laporan.txt` di folder tujuan, ditulis **selalu** — termasuk saat
  sebagian gagal atau dibatalkan.

Format laporan:

```text
Pilihanmu Picker — Laporan
Waktu     : 2026-09-24T14:03:11+08:00
Mode      : Salin
Sumber    : D:\Photos\2026-09-20 Wedding
Tujuan    : D:\Photos\2026-09-20 Wedding - SELECTED
Daftar    : Wedding Rian - Daftar Foto.txt (120 nama unik)

Ringkasan
  Cocok         : 117 file
  Tidak ketemu  : 3 nama
  Tabrakan      : 1
  Gagal         : 0

Tidak ketemu
  DSC_0042
  IMG_8891
  _MG_2210

Tabrakan
  subA\DSC_0001.CR3  ->  DSC_0001.CR3
  subB\DSC_0001.CR3  ->  DSC_0001 (2).CR3

Gagal
  (tidak ada)
```

## 5. Non-fungsional

| Aspek | Ketentuan |
|---|---|
| **Platform** | Windows 10+ (utama). Linux x86_64 (sekunder). macOS **di luar cakupan**. |
| **Distribusi** | Windows: `.exe` portabel + `.zip`, **tanpa installer, tanpa hak admin**. Linux: AppImage. |
| **Jaringan** | **Nol.** Aplikasi tidak boleh melakukan satu pun panggilan jaringan. Tidak ada telemetri, tidak ada pemeriksaan pembaruan otomatis. Ini properti keamanan, bukan kebetulan. |
| **Ukuran** | Satu binary, target < 40 MB. |
| **Memori** | Tidak pernah memuat isi file RAW ke memori. |
| **Skala** | Folder sumber 5.000 file harus selesai dipindai tanpa UI membeku. |
| **Lokalisasi** | Teks antarmuka Bahasa Indonesia. Nama file non-ASCII (UTF-8) wajib ditangani dengan benar. |
| **Jejak** | Tidak menulis apa pun ke luar folder tujuan, kecuali berkas konfigurasi kecil di lokasi standar OS (lihat §10). |

## 6. Kontrak `.txt` (dibekukan)

**Sumber otoritatif di repo induk**: `apps/web/src/lib/export.ts` → fungsi
`listText()`.

```ts
export function listText(picks: PickItem[]): string {
  return picks.map((p) => baseName(p.file_name)).join("\n") + "\n";
}
```

Karakteristik yang **dibekukan** — app ini bergantung padanya:

| Sifat | Nilai |
|---|---|
| Isi per baris | basename **tanpa ekstensi** |
| Pemisah | `\n` (LF); berkas yang diunduh di Windows dapat berisi `\r\n` |
| Urutan | mengikuti `pick_order` (urutan klien memilih) |
| Akhir berkas | diakhiri newline |
| Encoding | UTF-8 |
| Nama berkas | `<judul-galeri> - Daftar Foto.txt`, dengan judul disanitasi `safeName()` |

> **Konsekuensi yang harus disadari.** App ini menggandakan `listText()`. Bila
> `export.ts` di repo induk berubah bentuknya, app ini rusak **tanpa peringatan**.
> Karena itu: (a) parser wajib toleran terhadap variasi (§4.2), dan (b) perubahan
> pada `listText()` di repo induk harus disertai perubahan di sini.

## 7. Keputusan yang dikunci

Semua keputusan berikut diambil 24 September 2026 dan **tidak boleh dibuka ulang
tanpa persetujuan pemilik**.

| # | Keputusan | Alasan |
|---|---|---|
| K-1 | **Offline total** — tanpa API, tanpa token, tanpa login | Infrastruktur token (`api_tokens` + gerbang `isUnlocked`) ada di milestone M3 repo induk dan **sedang diparkir**. App ini tidak boleh bergantung padanya. |
| K-2 | Sumber daftar = tempel teks **atau** berkas `.txt` | Keduanya sudah tersedia di Dashboard hari ini. |
| K-3 | Tugas inti = memindahkan RAW terpilih ke folder baru | Ini kebutuhan nyatanya, bukan sekadar mengunduh `.txt` (yang sudah 1 klik di browser). |
| K-4 | Dua tombol Salin/Pindahkan, **tanpa default** | Keputusan pemilik. |
| K-5 | Folder tujuan **rata (flatten)** | Alur seleksi → impor Lightroom. |
| K-6 | Hanya ekstensi RAW yang disalin | Keputusan pemilik. |
| K-7 | Pratinjau wajib sebelum eksekusi | `.txt` hanya berisi basename; tanpa pratinjau, file yang tidak ketemu hilang diam-diam. |
| K-8 | Mode Pindahkan: verifikasi ukuran dulu, baru hapus sumber | RAW adalah aset yang tidak bisa diulang. |
| K-9 | Tabrakan → sufiks ` (2)`, tidak pernah menimpa | — |
| K-10 | `.xmp` pendamping ikut tersalin | Banyak fotografer sudah menerapkan sidecar di folder sumber. |
| K-11 | Stack: **Go + Fyne** | Satu binary, nol dependensi OS untuk pengguna, pemilih folder bawaan. Tampilan identik saat dikembangkan di Linux. |
| K-12 | Windows-first; `.exe` dibangun di CI `windows-latest` + `fyne-cross` lokal | Pengembang hanya punya Linux — CI adalah satu-satunya cara menguji di Windows asli. |

## 8. Di luar cakupan

Sengaja **tidak** dikerjakan. Menambahkannya butuh perubahan PRD ini:

- Mode API / token `plh_live_` / login studio
- Menulis atau mengubah berkas `.xmp` (app hanya menyalin `.xmp` yang sudah ada)
- Mengunduh dari Google Drive
- Pembaruan otomatis
- macOS
- Antarmuka multi-bahasa
- Basis data lokal atau riwayat antar-jalankan
- Mode baris perintah

## 9. Kriteria selesai

Sebuah rilis dianggap selesai bila seluruh butir berikut terbukti:

- [ ] Daftar 120 nama dari berkas `.txt` ber-CRLF → terbaca 120 nama unik
- [ ] Daftar dengan baris duplikat → duplikat dibuang, dilaporkan
- [ ] Daftar yang masih berisi ekstensi dan path → tetap cocok
- [ ] Folder sumber 800 file, 117 cocok, 3 tidak → angka pratinjau tepat
- [ ] Tombol eksekusi nonaktif selama pemindaian berjalan
- [ ] Mode Salin: jumlah file dan `mtime` di folder sumber **tidak berubah** setelah eksekusi
- [ ] Mode Pindahkan: berkas tujuan ada dengan ukuran sama → sumber terhapus; bila verifikasi gagal → sumber **tidak** terhapus
- [ ] Tabrakan nama → berkas lama di tujuan tetap utuh, berkas baru bersufiks ` (2)`
- [ ] `_laporan.txt` ada di folder tujuan dan isinya sesuai §4.7
- [ ] Path sumber > 260 karakter di Windows berhasil
- [ ] Berkas terkunci → dicoba ulang, lalu dilaporkan gagal, tanpa menghentikan sisanya
- [ ] Nama non-ASCII (`Café_0001.CR3`) cocok dan tersalin
- [ ] Peringatan §4.4 muncul untuk folder JPEG-only
- [ ] CI hijau di `ubuntu-latest` **dan** `windows-latest`
- [ ] Tidak ada satu pun panggilan jaringan di seluruh kode

## 10. Pertanyaan terbuka

Belum diputuskan; **jangan** dikerjakan tanpa keputusan eksplisit:

1. **Mengingat folder terakhir?** Saran: ya, satu berkas kecil di
   `%APPDATA%\PilihanmuPicker\config.json` (Windows) dan
   `~/.config/pilihanmu-picker/config.json` (Linux). Dikerjakan **setelah** inti stabil.
2. **Menyalin `.xmp` yang tidak ada di sumber?** Saat ini hanya menyalin yang sudah
   ada (K-10). Menghasilkan `.xmp` baru adalah fitur lain.
3. **Nama berkas laporan** — `_laporan.txt` dipilih agar mudah ditemukan; belum
   ada keputusan final.
4. **Ikon dan nama tampilan aplikasi** — belum ditentukan.
