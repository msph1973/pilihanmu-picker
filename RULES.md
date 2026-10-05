# RULES — Pilihanmu Picker (standalone)

Aturan produk, arsitektur, dan kode. **Pelanggaran = gagal review.**
Pelanggaran aturan yang menyentuh berkas pengguna (R-P3, R-P4, R-P5, R-D1..R-D4)
= gagal tanpa diskusi.

Aturan ini berlaku untuk seluruh isi repo. Bila berbenturan dengan `PRD.md`,
`PRD.md` menang dan `RULES.md` yang harus diperbaiki.

---

## 1. Aturan produk

- **R-P1 — Offline mutlak.** Aplikasi tidak boleh melakukan panggilan jaringan
  dalam bentuk apa pun: tidak ada HTTP client, tidak ada pemeriksaan pembaruan,
  tidak ada telemetri, tidak ada analitik. Kalau sebuah fitur butuh jaringan,
  fitur itu di luar cakupan (PRD §8) dan bukan alasan untuk menambah pengecualian.
  *Uji:* tidak ada impor `net/http`, `net`, atau pustaka HTTP di seluruh modul.

- **R-P2 — `.txt` adalah kontrak beku.** Parser wajib toleran terhadap seluruh
  variasi di PRD §4.2 (CRLF, baris kosong, duplikat, ekstensi, prefiks path) dan
  tidak boleh mengubah arti urutan. Mengubah bentuk keluaran `listText()` di repo
  induk tanpa memperbarui parser di sini = bug, bukan sekadar ketidaknyamanan.

- **R-P3 — Berkas RAW pengguna tidak boleh diubah atau dihapus**, kecuali dalam
  mode **Pindahkan**, dan hanya setelah verifikasi ukuran berhasil (R-P5 / E-2).
  Mode Salin tidak boleh menyentuh folder sumber sama sekali — tidak `Remove`,
  tidak `Rename`, tidak `Chtimes`, tidak membuat berkas di sana.

- **R-P4 — Pratinjau wajib sebelum eksekusi.** Tidak boleh ada jalur, flag,
  pintasan, atau "mode ahli" yang mengeksekusi tanpa pratinjau. Tombol eksekusi
  harus nonaktif selama pemindaian berjalan.

- **R-P5 — Tidak pernah menimpa berkas di folder tujuan.** Tabrakan nama
  diselesaikan dengan sufiks ` (2)`, lalu ` (3)`, dst. Menimpa berkas fotografer
  adalah kegagalan paling mahal yang bisa dilakukan aplikasi ini.

- **R-P6 — Nama berkas asli dipertahankan.** Satu-satunya perubahan yang
  diizinkan adalah sufiks tabrakan dari R-P5. Tidak ada sanitasi, tidak ada
  penggantian spasi, tidak ada perubahan huruf besar-kecil.

- **R-P7 — Laporan selalu ditulis.** `_laporan.txt` ditulis di folder tujuan pada
  setiap eksekusi yang dimulai — termasuk saat gagal sebagian, dibatalkan, atau
  ketika nol berkas berhasil disalin.

---

## 2. Aturan arsitektur

- **R-A1 — Stack: Go + Fyne.** Tidak ada toolkit GUI lain. Tidak ada Electron,
  Tauri, Wails, atau server web lokal.

- **R-A2 — Logika inti bebas Fyne.** Seluruh logika (parsing, pemindaian,
  pencocokan, penyalinan, pelaporan) hidup di `internal/` dan **tidak boleh**
  mengimpor Fyne. Lapisan UI hanya boleh memanggil `internal/`, tidak pernah
  sebaliknya. *Alasan:* ini yang membuat logika bisa diuji tanpa tampilan, dan
  yang membuat pengujian di CI Windows mungkin.

- **R-A3 — Tidak ada dependensi baru tanpa alasan tertulis.** Setiap dependensi
  harus (a) punya justifikasi di deskripsi PR, dan (b) dapat dikompilasi untuk
  `windows/amd64` **dan** `linux/amd64`. Pustaka yang hanya jalan di satu platform
  ditolak.

- **R-A4 — Pekerjaan berat tidak boleh memblokir UI.** Pemindaian folder dan
  penyalinan berjalan di goroutine sendiri dan melaporkan progres lewat kanal.
  UI hanya merender. Folder 5.000 berkas tidak boleh membuat UI membeku.

- **R-A5 — Penyalinan streaming.** Selalu `io.Copy` (atau `io.CopyBuffer`).
  **Dilarang** `os.ReadFile` / `os.WriteFile` untuk berkas RAW — satu berkas bisa
  60 MB+, dan seluruh berkas tidak boleh berada di memori sekaligus.

- **R-A6 — Semua operasi berkas lewat satu paket.** Baca/tulis/pindah/hapus hanya
  boleh terjadi di `internal/fsops`. Paket lain memanggilnya. *Alasan:* supaya
  seluruh perilaku destruktif bisa diuji dengan `t.TempDir()` dan diaudit di satu
  tempat.

- **R-A7 — Tidak ada keadaan global.** Tidak ada variabel paket yang bisa berubah.
  Konfigurasi dan hasil pemindaian dilewatkan sebagai parameter.

---

## 3. Aturan kode (Go)

- **R-C1 — `gofmt` dan `go vet` wajib lulus.** `gofmt -l .` harus kosong;
  `go vet ./...` harus bersih. Keduanya adalah gerbang CI.

- **R-C2 — Kesalahan dibungkus, tidak ditelan.** Pakai
  `fmt.Errorf("konteks: %w", err)`. `_ = err` dan blok `if err != nil {}` kosong
  dilarang. Setiap `error` yang diabaikan butuh komentar yang menjelaskan alasannya.

- **R-C3 — Cakupan tes ≥ 80% untuk `internal/`.** Setiap fungsi publik punya tes.
  Tes wajib mencakup minimal: parser CRLF, duplikat, tabrakan nama, sufiks ` (2)`,
  verifikasi gagal (sumber tidak dihapus), dan nama non-ASCII.

- **R-C4 — Tidak ada `panic` di jalur runtime.** `panic` hanya boleh muncul di
  `main` saat inisialisasi gagal total (mis. UI tidak bisa dibuka), dan tidak
  pernah di `internal/`.

- **R-C5 — Teks antarmuka Bahasa Indonesia, terkumpul di satu tempat.** Tidak ada
  string yang dilihat pengguna yang tersebar di logika. Kumpulkan sebagai
  konstanta agar bisa ditinjau sekali jalan.

- **R-C6 — Tidak ada berkas besar yang di-commit.** Binary hasil build, `.exe`,
  AppImage, dan berkas uji berukuran besar tidak masuk repo. Lihat `.gitignore`.

---

## 4. Aturan data & lintas-platform

- **R-D1 — Path panjang Windows.** Untuk path absolut dengan panjang ≥ 260
  karakter, gunakan prefiks `\\?\`. Normalisasi **hanya** untuk keperluan
  perbandingan — path yang ditulis ke disk harus tetap bentuk aslinya.

- **R-D2 — Perbandingan case-insensitive, penulisan case-preserving.** Windows
  tidak peduli huruf besar-kecil, Linux peduli. Bandingkan setelah `ToLower`,
  tetapi tulis ke disk dengan nama asli.

- **R-D3 — Nama dan karakter terlarang di Windows dideteksi, bukan diperbaiki
  diam-diam.** Nama reserved (`CON`, `PRN`, `AUX`, `NUL`, `COM1`–`COM9`,
  `LPT1`–`LPT9`) dan karakter `<>:"/\|?*` → laporkan berkas itu sebagai gagal
  dengan alasan yang jelas. Jangan mengganti nama berkas fotografer tanpa izin.

- **R-D4 — Berkas terkunci dicoba ulang.** Retry 3× dengan jeda 250 ms / 500 ms /
  1000 ms, lalu tandai gagal dan lanjut ke berkas berikutnya. Berkas yang gagal
  **tidak** dihapus dari sumber.

- **R-D5 — Daftar ekstensi RAW adalah daftar tertutup.** Isinya ada di PRD §4.4.
  Mengubahnya butuh perubahan PRD, bukan sekadar perubahan konstanta.

- **R-D6 — Symlink direktori tidak diikuti** saat pemindaian. Symlink berkas boleh
  disalin isinya (bukan symlink-nya).

- **R-D7 — Encoding berkas `.txt` adalah UTF-8**, toleran terhadap BOM. Bila
  berkas tidak dapat didekode sebagai UTF-8, laporkan dengan jelas dan jangan
  diam-diam mengganti encoding.

---

## 5. Aturan UX

- **R-U1 — Setiap tindakan yang memindahkan atau menghapus berkas butuh
  konfirmasi eksplisit.** Tidak ada eksekusi otomatis, tidak ada "jangan tanya lagi".

- **R-U2 — Pratinjau tidak boleh diringkas jadi satu angka.** Empat keranjang
  (cocok / tidak ketemu / tabrakan / dilewati) harus terlihat, dan
  daftar nama yang tidak ketemu harus bisa dibaca dan disalin.

- **R-U3 — Progres terlihat dan bisa dibatalkan.** Pembatalan menghentikan setelah
  berkas yang sedang berjalan selesai; tidak boleh meninggalkan berkas setengah
  tertulis.

- **R-U4 — Jaring pengaman JPEG.** Bila hasil cocok = 0 tetapi ada nama yang punya
  berkas non-RAW, tampilkan saran untuk mematikan saringan RAW. Tanpa ini,
  fotografer JPEG-only melihat "0 cocok" tanpa tahu sebabnya.

- **R-U5 — Teks antarmuka Bahasa Indonesia.**

- **R-U6 — Kesalahan harus bisa ditindaklanjuti.** "Gagal menyalin DSC_0001.CR3:
  berkas sedang dipakai aplikasi lain" — bukan "Terjadi kesalahan".

---

## 6. Gerbang review (wajib lulus)

- [ ] `gofmt -l .` kosong dan `go vet ./...` bersih
- [ ] `go test ./...` lulus di Linux
- [ ] CI `windows-latest` lulus (build + test) — bukan hanya `ubuntu-latest`
- [ ] Tidak ada impor `net`, `net/http`, atau pustaka HTTP apa pun (R-P1)
- [ ] Tidak ada `os.Remove` yang tidak didahului verifikasi ukuran (R-P3)
- [ ] Tidak ada `os.ReadFile` / `os.WriteFile` pada berkas RAW (R-A5)
- [ ] `internal/` tidak mengimpor Fyne (R-A2)
- [ ] Tes tersedia untuk: CRLF, duplikat, tabrakan + sufiks, verifikasi gagal,
      path panjang, nama non-ASCII
- [ ] `_laporan.txt` tertulis pada jalur gagal sebagian dan jalur dibatalkan (R-P7)
- [ ] Tidak ada berkas biner hasil build yang ter-commit (R-C6)
