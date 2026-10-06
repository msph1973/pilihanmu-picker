# AGENTS.md — Pilihanmu Picker (standalone)

Panduan kerja untuk agent AI dan manusia di repo ini.

Aplikasi desktop **offline** (Go + Fyne, Windows-first) yang memindahkan berkas RAW
hasil seleksi klien ke folder baru, berdasarkan daftar `.txt` dari Dashboard
Pilihanmu. **Repo ini berdiri sendiri** — tidak mengimpor apa pun dari repo induk
`msph1973/pilih`.

> **Baca dulu, dalam urutan ini:** `PRD.md` (apa yang dibangun) → `RULES.md`
> (batasan yang tidak boleh dilanggar) → berkas ini (cara bekerja).
>
> **Aturan termahal bila salah:** **(1)** jangan pernah menimpa atau menghapus
> berkas fotografer tanpa verifikasi ukuran — R-P3/R-P5; **(2)** parser `.txt`
> wajib membuang `\r` (CRLF), kalau tidak seluruh daftar akan tampak "tidak ketemu";
> **(3)** `internal/` tidak boleh mengimpor Fyne — R-A2, ini yang menjaga logika
> tetap bisa diuji di CI Windows.

---

## 1. Batas proyek

| | |
|---|---|
| **Repo ini** | `msph1973/pilihanmu-picker` (private) — alat desktop standalone |
| **Repo induk** | `msph1973/pilih` (private) — web + API Pilihanmu. **Jangan diimpor.** |
| **Keterkaitan satu-satunya** | Format `.txt` yang berasal dari `apps/web/src/lib/export.ts` → `listText()` (PRD §6) |
| **Jaringan** | Nol. Tidak ada HTTP, tidak ada telemetri (R-P1) |
| **Platform** | Windows 10+ (utama), Linux x86_64 (sekunder). macOS di luar cakupan. |

Repo ini **bukan** bagian dari monorepo induk, tidak punya `package.json`, dan
tidak boleh ditambahkan sebagai dependensi repo induk.

## 2. Repo map

```text
pilihanmu-picker/
  PRD.md              # product requirements (ID) — sumber kebenaran fitur
  RULES.md            # aturan produk/arsitektur/kode — pelanggaran = gagal review
  AGENTS.md           # berkas ini
  go.mod              # module github.com/msph1973/pilihanmu-picker — terimplementasi
  README.md           # pengantar, cara pakai, build, dan rilis
  LICENSE             # proprieter — semua hak dilindungi
  cmd/pilihanmu-picker/
    main.go           # titik masuk — terimplementasi (menyuntik `version` lewat ldflags)
    FyneApp.toml      # metadata paket Fyne (nama, ID, versi, ikon) — WAJIB di sini, bukan di akar repo
    Icon.png          # ikon aplikasi — monogram Pilihanmu
  internal/
    list/             # parser daftar .txt (PRD §4.2) — terimplementasi, cakupan tes lihat §8
    scan/             # pemindaian folder sumber (PRD §4.3) — terimplementasi, cakupan tes lihat §8
    match/            # pencocokan + empat keranjang (PRD §4.3) — terimplementasi, cakupan tes lihat §8
    fsops/            # salin/pindah/verifikasi — SATU-SATUNYA tempat I/O (R-A6) — terimplementasi, cakupan tes lihat §8
    report/           # penulis _laporan.txt (PRD §4.7) — terimplementasi, cakupan tes lihat §8
    ui/               # lapisan Fyne — satu-satunya tempat impor Fyne — terimplementasi, cakupan tes lihat §8
  .github/workflows/ci.yml       # gerbang ubuntu + windows — pipeline CI #1 (GitHub Actions)
  .github/workflows/release.yml  # rilis: tag v* → .zip Windows (ber-ikon) + .tar.gz Linux + SHA256SUMS
  .circleci/config.yml           # gerbang linux + windows — pipeline CI #2 (CircleCI), paralel dengan GitHub Actions
  .circleci/info.yml             # metadata integrasi proyek CircleCI (id organisasi & proyek)
  .gitignore          # terimplementasi — abaikan binary, .exe, AppImage, *.test, berkas besar
```

`internal/` **tidak boleh** mengimpor `internal/ui` maupun Fyne (R-A2).

Proyek ini punya **dua pipeline CI paralel yang sengaja dipertahankan keduanya**:
GitHub Actions (`.github/workflows/ci.yml`, dipicu push/PR ke `main`) dan CircleCI
(`.circleci/config.yml`, terhubung lewat integrasi proyek CircleCI — lihat
`.circleci/info.yml`). Keduanya menjalankan gerbang kualitas yang sama (`gofmt`,
`go vet`, `go test`, `go build`) pada matriks Linux + Windows; jangan hapus salah
satunya tanpa keputusan eksplisit pemilik.

## 3. Status saat ini

**Implementasi inti sudah ada dan berfungsi.** `go.mod`, `.gitignore`, keenam
paket `internal/*` (`list`, `scan`, `match`, `fsops`, `report`, `ui`) beserta
tesnya, dan `cmd/pilihanmu-picker/main.go` sudah ada. Kedua pipeline CI — GitHub
Actions (`.github/workflows/ci.yml`) dan CircleCI (`.circleci/config.yml`) —
sudah ada dan menjalankan gerbang kualitas (gofmt, go vet, go test, go build) di
matriks Linux + Windows.

**BELUM ADA — jangan klaim ada:**

Artefak rilis yang sudah terbit (belum ada tag `v*` yang didorong, jadi
`release.yml` belum pernah berjalan) · code-signing Windows — tanpa sertifikat,
SmartScreen akan memperingatkan · berkas konfigurasi pengguna (mis. mengingat
folder terakhir — masih pertanyaan terbuka di PRD §10) · mode API/token

## 4. Prasyarat lingkungan

Go 1.22+ dan `gofmt` (keduanya sudah ada di mesin pengembangan).

**Fyne di Linux butuh header dev X11.** Tanpa ini, `go build` gagal dengan pesan
tentang `Xrandr.h`, `Xcursor.h`, atau `XInput2.h`.

```bash
# Debian / Ubuntu
sudo apt install gcc libgl1-mesa-dev xorg-dev

# Arch / CachyOS (mesin pemilik)
sudo pacman -S --needed base-devel mesa libxrandr libxcursor libxinerama libxi

# Verifikasi cepat
pkg-config --exists xrandr xcursor xi xinerama gl x11 && echo "dependensi OK"
```

> Daftar paket di atas **perlu diverifikasi** terhadap dokumentasi Fyne terkini.
> Yang konkret: pada 2026-09-24 di VPS ini, header yang hilang persis
> `X11/extensions/Xrandr.h`, `X11/Xcursor/Xcursor.h`, dan
> `X11/extensions/XInput2.h`.

## 5. Perintah

```bash
# Gerbang kualitas (jalankan sebelum setiap commit)
gofmt -l .                 # harus KOSONG
go vet ./...               # harus BERSIH
go test ./...              # harus LULUS

# Jalankan di Linux (dev loop normal)
go run ./cmd/pilihanmu-picker

# Build binary Linux
go build -o pilihanmu-picker ./cmd/pilihanmu-picker

# Paket rilis Windows (butuh fyne CLI: go install fyne.io/tools/cmd/fyne@v1.7.3;
# CI menyematkan versi itu agar rilis bisa diulang)
cd cmd/pilihanmu-picker
GOFLAGS="-ldflags=-X=main.version=v0.1.0" fyne package -os windows -release \
  --app-id id.web.pilihanmu.picker --app-version 0.1.0 --app-build 1
# → "Pilihanmu Picker.exe" ber-ikon + bermetadata. FyneApp.toml dan Icon.png
#   dibaca dari direktori ini (lihat §7 jebakan 10).
# Perhatikan `-X=` TANPA spasi — lihat §7 jebakan 12.
# Alternatif tanpa Docker: fyne-cross, tapi CI windows-latest tetap cara utama (K-12).

# Rilis resmi = dorong tag, bukan build lokal
git tag v0.1.0 && git push origin v0.1.0    # → .github/workflows/release.yml
```

## 6. Rencana implementasi

Urutan ini disusun supaya **seluruh logika berisiko selesai dan teruji di Linux
lebih dulu**, dan GUI tinggal jadi cangkang tipis.

### Fase 1 — Scaffold
- `go.mod` (module `github.com/msph1973/pilihanmu-picker`, `go 1.22`)
- Struktur direktori §2 (paket kosong dengan `doc.go`)
- `.gitignore` (binary, `.exe`, AppImage, `*.test`, berkas besar)
- `.github/workflows/ci.yml` — matriks `ubuntu-latest` + `windows-latest`,
  menjalankan `gofmt -l .`, `go vet ./...`, `go test ./...`, `go build ./...`

### Fase 2 — Logika murni + tes (tanpa Fyne, tanpa UI)
Ini bagian yang paling menentukan. Semua harus teruji sebelum GUI disentuh.
- `internal/list` — parser PRD §4.2, tujuh aturan P-1..P-7, tiap aturan punya tes
- `internal/scan` — pemindaian rekursif opsional, filter ekstensi RAW (PRD §4.4),
  lewati symlink direktori, kumpulkan direktori tak terbaca sebagai peringatan
- `internal/match` — empat keranjang (cocok / tidak ketemu / tabrakan / dilewati)
- `internal/fsops` — salin streaming, verifikasi ukuran, sufiks tabrakan ` (2)`,
  retry berkas terkunci, path panjang `\\?\`, mode pindah
- `internal/report` — `_laporan.txt` sesuai PRD §4.7

### Fase 3 — UI Fyne
- Kotak tempel + tombol pilih `.txt`
- Dua pemilih folder (sumber, tujuan)
- Checkbox "Sertakan sub-folder" (default aktif)
- Kontrol saringan RAW (PRD §4.4)
- Layar pratinjau: empat keranjang, daftar nama tidak ketemu bisa disalin (R-U2)
- Dua tombol tanpa default: **Salin** / **Pindahkan** (K-4), keduanya berkonfirmasi
- Progres + tombol batal (R-U3)

### Fase 4 — Paket & rilis — SELESAI (2026-10-06)
- `cmd/pilihanmu-picker/FyneApp.toml` + `Icon.png` (monogram Pilihanmu) →
  `.exe` ber-ikon dan bermetadata
- `.github/workflows/release.yml`: tag `v*` → gerbang verifikasi (tag harus di
  `main` + lulus gofmt/vet/test) → `.zip` Windows + `.tar.gz` Linux
  + `SHA256SUMS` + GitHub Release
- **AppImage Linux diganti `.tar.gz`** (keputusan pemilik 2026-10-06; PRD §5
  sudah diperbarui)
- Menjalankan seluruh kriteria selesai PRD §9 tetap tugas **per rilis**, bukan
  sekali saja

**Jangan lompat ke Fase 3 sebelum Fase 2 lulus di CI Windows.**

## 7. Jebakan yang diketahui

1. **CRLF.** Berkas `.txt` yang diunduh di Windows berisi `\r\n`. Kalau `\r` tidak
   dibuang, **setiap** nama akan tampak "tidak ketemu" dan penyebabnya tidak
   terlihat. Ini bug nomor satu yang paling mungkin terjadi (P-1).

2. **Cross-compile ke Windows jangan lewat mingw.** Fyne + mingw dari Linux
   rapuh. Pakai `fyne-cross` (Docker) atau — cara utama — biarkan
   `windows-latest` di CI yang mengompilasi (K-12).

3. **Windows case-insensitive, Linux case-sensitive.** `DSC_0001.CR3` dan
   `dsc_0001.cr3` adalah berkas yang sama di Windows, dua berkas berbeda di Linux.
   Tes tabrakan harus mempertimbangkan kedua perilaku.

4. **Jangan pernah `os.ReadFile` berkas RAW.** Satu berkas bisa 60 MB+; pakai
   `io.Copy` (R-A5).

5. **Jangan hapus tanpa verifikasi.** Ukuran berkas tujuan harus sama dengan
   sumber sebelum `os.Remove` dipanggil (R-P3, E-2). Tidak ada pengecualian.

6. **`\\?\` menonaktifkan normalisasi path.** Resolusi ke path absolut dan bersih
   **sebelum** menambahkan prefiks, dan gunakan pemisah `\` (R-D1).

7. **Fyne hanya boleh diperbarui dari goroutine utama.** Pekerjaan latar
   melaporkan hasil lewat kanal; pembaruan widget dilakukan di goroutine UI
   (R-A4). Jangan menyentuh widget dari goroutine pekerja.

8. **Skill `pilihin-photographer` tidak relevan di sini.** Skill user-scope itu
   membahas web/API/billing repo induk. Memuatnya di repo ini hanya membawa
   konteks yang salah.

9. **Jangan membuka ulang keputusan K-1..K-13** (PRD §7) tanpa persetujuan
   pemilik. Terutama K-1 (offline): permintaan "tambahkan login" berarti PRD
   harus diubah lebih dulu, bukan langsung dikerjakan.

10. **`FyneApp.toml` dan `Icon.png` harus di direktori paket main**
    (`cmd/pilihanmu-picker/`), bukan di akar repo. `fyne package --src <dir>`
    memuat metadata DARI `<dir>` itu, menulis `fyne.syso` ke sana, lalu
    menjalankan `go build` di sana. Di akar repo, ikon dan metadata tidak akan
    ikut tertanam — dan tanpa `Icon.png` perintahnya langsung gagal
    ("Missing application icon").

11. **`fyne package` MENULIS ULANG `FyneApp.toml`.** Setelah packaging, berkas itu
    ditulis kembali dari struct hasil parse: `Build` naik satu, dan **seluruh
    komentar hilang**. Jadi jangan menaruh penjelasan penting sebagai komentar di
    sana — taruh di berkas ini. Di CI berkasnya jadi kotor setelah build; jangan
    commit hasil tulis-ulang itu.

12. **Penyuntikan versi lewat `GOFLAGS` harus memakai `-X=` tanpa spasi.**
    Fyne mengambil `-ldflags` dari `GOFLAGS` dengan `strings.Fields`, yang
    **memecah pada spasi**. Bentuk lazim `-ldflags=-X main.version=v1` terbaca
    sebagai dua token (`-ldflags=-X` + `main.version=v1`) → `go` gagal dengan
    `parsing $GOFLAGS: non-flag "main.version=v1"`. Pakai
    `-ldflags=-X=main.version=v1` (linker menerima bentuk `-X=`).
    Pada `go build` biasa kedua bentuk sama-sama jalan — jebakan ini khusus
    jalur Fyne.

## 8. Definisi selesai

Sebuah pekerjaan dianggap selesai bila:

- [ ] `gofmt -l .` kosong, `go vet ./...` bersih, `go test ./...` lulus
- [ ] CI hijau di **`ubuntu-latest` dan `windows-latest`** — keduanya, bukan salah satu
- [ ] Gerbang review di `RULES.md` §6 terpenuhi
- [ ] Perilaku baru punya tes; perubahan pada `internal/` menaikkan atau
      mempertahankan cakupan ≥ 80%
- [ ] Bila sebuah keputusan berubah, `PRD.md` diperbarui di commit yang sama
- [ ] Tidak ada berkas biner hasil build yang ter-commit

## 9. Protokol kerja

- **Commit kecil, satu maksud.** Format: `tipe(lingkup): ringkasan` —
  `feat(list)`, `fix(fsops)`, `test(match)`, `ci`, `docs(prd)`.
- **Bahasa**: commit dan dokumentasi Bahasa Indonesia; nama fungsi, variabel, dan
  pesan error teknis Bahasa Inggris; teks antarmuka Bahasa Indonesia.
- **Jangan menambah dependensi** tanpa justifikasi tertulis di deskripsi PR (R-A3).
- **Fase 2 tidak boleh bergantung pada Fyne** — kalau sebuah tes butuh GUI untuk
  berjalan, desainnya salah.
- **Bila `listText()` di repo induk berubah**, perbarui parser di sini di commit
  yang sama dan catat di PRD §6.
