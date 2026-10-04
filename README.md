# TCP Chat Room (Go)

Aplikasi chat berbasis terminal yang dibuat dengan Go menggunakan TCP socket. Banyak client dapat terhubung ke satu server, memilih room, lalu saling berkirim pesan secara real-time di dalam room yang sama.

## Fitur

- Banyak client terhubung secara bersamaan (setiap koneksi ditangani dengan goroutine)
- Nama pengguna harus unik (tidak case-sensitive, jadi `Budi` dan `budi` dianggap sama)
- 6 room siap pakai: `general`, `games`, `hobby`, `study`, `chill`, `qna`
- Pesan hanya dikirim ke anggota room yang sama (tidak dikirim balik ke pengirim)
- Notifikasi otomatis saat pengguna masuk atau keluar room
- Perintah `/leave` untuk keluar dari room dan pindah ke room lain
- Pemilihan nama room tidak case-sensitive (`Games` = `games`)
- Akses ke data bersama (`clients`) diamankan dengan `sync.Mutex`

## Struktur Proyek

```
.
├── server.go   # Server: menerima koneksi, mengelola client, room, dan broadcast pesan
├── client.go   # Client: mengirim input pengguna dan menampilkan pesan dari server
└── README.md
```

## Persyaratan

- [Go](https://go.dev/dl/) 1.18 atau lebih baru
- Hanya memakai standard library, tidak ada dependensi eksternal

## Cara Menjalankan

### 1. Jalankan server

```bash
go run server.go
```

Output:

```
Listening on port 9090
Available rooms: [general games hobby study chill qna]
```

### 2. Jalankan client (di terminal terpisah)

```bash
go run client.go
```

Anda bisa membuka beberapa terminal untuk mensimulasikan banyak pengguna.

> Karena `server.go` dan `client.go` sama-sama berisi `package main` dan fungsi `main()`, keduanya harus dijalankan secara terpisah dengan `go run <nama file>`, bukan dengan `go run .`.

### Opsional: build binary

```bash
go build -o server server.go
go build -o client client.go

./server
./client
```

## Cara Penggunaan

1. Setelah terhubung, masukkan **nama** Anda. Jika nama sudah dipakai, Anda akan diminta memasukkan nama lain.
2. Pilih salah satu **room** dari daftar yang ditampilkan dengan mengetik namanya.
3. Ketik pesan lalu tekan Enter untuk mengirim ke semua anggota room.
4. Ketik `/leave` untuk keluar dari room dan kembali ke daftar room.

### Contoh sesi

```
Connected to server!
Enter name: Andi

Available rooms [/leave to exit room]:
. general
. games
. hobby
. study
. chill
. qna
Choose room:
games

Succesfully joined games.

New chat:
Budi just joined.
Budi: halo semua!
```

### Perintah

| Perintah | Fungsi |
|----------|--------|
| `/leave` | Keluar dari room saat ini dan kembali ke pemilihan room |

## Cara Kerja

### Alur koneksi

```
Client                          Server
  |  --- TCP connect :9090 --->   |
  |  --- nama --------------->    |  cek unik (isNameTaken)
  |  <-- "Name unavailable" ---   |  (diulang sampai unik)
  |  <-- daftar room ---------    |
  |  --- nama room ---------->    |  validasi room
  |  <-- "Successfully joined" -  |
  |  --- pesan -------------->    |  broadcastMessage ke room yang sama
  |  --- /leave ------------->    |  kembali ke pemilihan room
```

### Pembagian tanggung jawab

- **Server** menyimpan semua state dan logika: validasi nama, pemilihan room, perintah `/leave`, dan broadcast.
- **Client** sengaja dibuat sederhana: hanya meneruskan input pengguna ke server dan mencetak apa pun yang dikirim server. Client tidak mengetahui konteks pesan.

### Komponen utama (`server.go`)

| Komponen | Deskripsi |
|----------|-----------|
| `Client` | Struct berisi `conn`, `name`, dan `room` milik satu pengguna |
| `clients` | Map `net.Conn -> *Client` berisi semua pengguna yang sedang terhubung |
| `rooms` | Daftar room yang tersedia |
| `mutex` | Mengamankan akses konkuren ke `clients` |
| `listRooms()` | Membuat teks daftar room untuk dikirim ke client |
| `broadcastMessage()` | Mengirim pesan ke semua client di room yang sama, kecuali pengirim |
| `isNameTaken()` | Mengecek apakah nama sudah dipakai (case-insensitive) |
| `handleClient()` | Menangani seluruh siklus hidup satu koneksi (nama, pilih room, chat, disconnect) |

### Komponen utama (`client.go`)

| Komponen | Deskripsi |
|----------|-----------|
| `receiveMessages()` | Goroutine yang terus membaca pesan dari server dan mencetaknya |
| `main()` | Menghubungkan ke server dan membaca input dari stdin lalu mengirimkannya |

## Konfigurasi

Alamat dan port saat ini ditulis langsung (hardcoded):

- **Server**: mendengarkan di `:9090` (`net.Listen("tcp", ":9090")` di `server.go`)
- **Client**: terhubung ke `:9090`, yaitu localhost (`net.Dial("tcp", ":9090")` di `client.go`)

Untuk terhubung ke server di komputer lain, ubah alamat di `client.go`, misalnya:

```go
conn, err := net.Dial("tcp", "192.168.1.10:9090")
```

Untuk menambah atau mengubah room, edit slice `rooms` di `server.go`:

```go
rooms = []string{"general", "games", "hobby", "study", "chill", "qna"}
```

## Keterbatasan dan Ide Pengembangan

- Komunikasi tidak terenkripsi (plain TCP) dan tidak ada autentikasi
- Pengecekan nama dan pendaftaran client belum atomik, sehingga dua client dengan nama sama yang masuk hampir bersamaan secara teori bisa lolos
- Field `client.room` dibaca dan ditulis dari beberapa goroutine tanpa mutex, sehingga berpotensi data race (bisa dicek dengan `go run -race server.go`)
- Nama kosong masih diterima
- Port dan daftar room belum bisa dikonfigurasi lewat flag atau environment variable
- Belum ada fitur pesan pribadi (DM), daftar pengguna online, atau riwayat chat
