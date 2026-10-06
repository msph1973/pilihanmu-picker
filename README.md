# Pilihanmu Picker

Aplikasi desktop **offline** untuk fotografer. Mengubah daftar nama foto hasil
seleksi klien menjadi **folder berisi berkas RAW yang benar**, siap diimpor ke
Lightroom.

Daftar nama itu berasal dari Dashboard Pilihanmu (tombol **Unduh (.txt)** atau
**Salin Daftar Foto**). Mencari 120 nama di antara 800 berkas secara manual
melelahkan dan rawan salah — alat ini melakukannya sekaligus, dengan laporan.

**Windows 10+** (utama) dan **Linux x86_64**. macOS di luar cakupan.

## Cara pakai

1. Di Dashboard Pilihanmu, buka galeri yang sudah disubmit klien → **Unduh (.txt)**.
2. Buka Pilihanmu Picker.
3. Tempel daftar ke kotak teks, **atau** tekan **Buka .txt**.
4. Pilih **folder sumber** (tempat RAW berada).
5. Pilih **folder tujuan** (folder baru).
6. Tekan **Pindai** → muncul pratinjau: berapa cocok, mana yang tidak ketemu,
   mana yang bertabrakan. **Pratinjau wajib** — tanpa itu, nama yang tidak
   ketemu akan hilang diam-diam.
7. Tekan **Salin** atau **Pindahkan**. Pindahkan akan memverifikasi ukuran tiap
   berkas sebelum menghapus sumbernya.
8. Selesai. Folder tujuan berisi berkas terpilih + `_laporan.txt`.

Bila Anda memotret JPEG saja, matikan saringan RAW — aplikasi akan mengingatkan
kalau hasil cocok nol padahal ada berkas non-RAW.

## Yang tidak dilakukan aplikasi ini

- **Tidak ada jaringan sama sekali.** Tanpa telemetri, tanpa pemeriksaan
  pembaruan otomatis. Ini properti keamanan, bukan kebetulan (R-P1).
- Tidak menyentuh Google Drive, tidak login, tidak butuh token.
- Tidak menulis atau mengubah berkas `.xmp` (hanya ikut menyalin yang sudah ada).
- Tidak menyimpan riwayat antar-jalankan.

## Unduh

Repo ini **privat**. Aset rilis hanya bisa diunduh oleh pemilik dan kolaborator.
Untuk fotografer, kirim berkas `.zip` hasil rilis secara langsung (mis. lewat
WhatsApp atau Drive).

Tanpa sertifikat code-signing, Windows SmartScreen akan menampilkan peringatan
saat pertama dijalankan. Pilih **More info → Run anyway**.

## Build dari sumber

Butuh Go 1.22+. Di Linux, Fyne juga butuh header dev X11:

```bash
# Debian / Ubuntu
sudo apt install gcc libgl1-mesa-dev xorg-dev libwayland-dev libxkbcommon-dev

# Arch / CachyOS
sudo pacman -S --needed base-devel mesa libxrandr libxcursor libxinerama libxi
```

```bash
# Gerbang kualitas
gofmt -l .          # harus kosong
go vet ./...
go test ./...

# Jalankan
go run ./cmd/pilihanmu-picker

# Build binary Linux
go build -o pilihanmu-picker ./cmd/pilihanmu-picker
```

`.exe` Windows dibangun di CI `windows-latest` (lihat `.github/workflows/ci.yml`),
bukan lewat cross-compile dari Linux.

## Rilis

Rilis dibuat dengan mendorong tag `v*`:

```bash
git tag v0.1.0
git push origin v0.1.0
```

`.github/workflows/release.yml` lalu:

- `windows-latest` → `fyne package -os windows -release` → `Pilihanmu Picker.exe`
  (ikon + metadata tertanam) → `.zip`
- `ubuntu-latest` → `go build` → `.tar.gz`
- membuat GitHub Release + `SHA256SUMS`

Versi disuntik ke binary lewat `-ldflags "-X main.version=<tag>"`, dan ditampilkan
di judul jendela. Build pengembangan menampilkan `dev`.

Metadata paket (nama, ID, versi, ikon) ada di
`cmd/pilihanmu-picker/FyneApp.toml` — sengaja di direktori paket `main`, karena
`fyne package` memuat metadata dan menulis `fyne.syso` dari direktori itu.

## Dokumentasi

| Berkas | Isi |
|---|---|
| `PRD.md` | persyaratan produk + keputusan yang dikunci |
| `RULES.md` | aturan yang tidak boleh dilanggar |
| `AGENTS.md` | panduan kerja untuk agen AI dan manusia |

## Lisensi

Proprieter — lihat [`LICENSE`](LICENSE). Seluruh hak dilindungi.
