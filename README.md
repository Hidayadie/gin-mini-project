# Latihan Gin - Todo API

Aplikasi API sederhana menggunakan framework [Gin](https://github.com/gin-gonic/gin) dan [GORM](https://gorm.io/) untuk mengelola data todo. Aplikasi ini terhubung ke database PostgreSQL.


Berikut keterangan singkat cara kerja kode tersebut:

1. **`main.go`** - Tit utama program. Mengambil `database.Connect()` (lihat bawah) untuk koneksi DB, melakukan `AutoMigrate` agar tabel `todos` dibuat otomatis, lalu setup router Gin (`router.SetupRouter`) dan menjalankan server di `localhost:8080`.

2. **`database/database.go`** - Fungsi `Connect()` membaca konfigurasi dari file `.env` (melalui `godotenv.Load`), lalu menjalankan GORM Open dengan DSN dari variabel `DATABASE_URL`, dan melakukan `PingContext` selama 15 detik untuk memastikan koneksi aktif sebelum dikembalikan.

3. **`router/router.go`** - Membuat group route `/api/v1` dan menambahkan route todo (melalui `todo.RegisterRoutes`).

4. **`todo/model.go`** - Mendefinisikan struct `Todo` (id, title, is_done), serta request structs untuk create/update.

Untuk menjalankan aplikasi, pastikan `DATABASE_URL` didefinisikan di `.env`, lalu `go run main.go`.


## Features

- Feature 1
- Feature 2
- Feature 3

## Getting Started

### Prerequisites

- Go 1.20+ (install from [golang.org](https://golang.org/dl/))

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/yourproject.git
```

2. Build the project:
```bash
go build
```

3. Run the application:
```bash
./yourproject
```

## Usage

```bash
# Example command
./yourproject [flags]
```

## Configuration

Add configuration instructions if applicable

## Contributing

Pull requests are welcome. Please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)

