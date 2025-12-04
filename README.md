# KBBI API

KBBI API adalah aplikasi RESTful API yang dibangun menggunakan Go dengan framework Fiber untuk menyediakan akses ke Kamus Besar Bahasa Indonesia (KBBI).

## Fitur

- Pencarian kata dalam kamus KBBI
- Mendapatkan definisi kata spesifik
- Pagination untuk hasil pencarian
- Statistik database
- Database dengan 100,000+ entri kata

## Tech Stack

- **Go** - Programming language
- **Fiber** - Web framework
- **PostgreSQL** - Database
- **SQLx** - Database library
- **Docker** - Containerization

## Prerequisites

- Go 1.19+
- Docker & Docker Compose
- PostgreSQL 16 (jika tidak menggunakan Docker)

## Setup

### 1. Clone Repository

```bash
git clone https://github.com/alifdwt/kbbi-api.git
cd kbbi-api
```

### 2. Setup Database dengan Docker

```bash
# Jalankan PostgreSQL dan pgAdmin
docker-compose up -d

# Cek status containers
docker-compose ps
```

Database akan tersedia di:
- PostgreSQL: `localhost:5436`
- pgAdmin: `http://localhost:5050`
  - Email: `admin@kbbi.local`
  - Password: `admin123`

### 3. Konfigurasi Environment

```bash
# Copy file environment example
cp .env.example .env

# Edit file .env sesuai kebutuhan
nano .env
```

### 4. Install Dependencies

```bash
go mod download
```

### 5. Jalankan Aplikasi

```bash
# Development
go run main.go

# Atau build terlebih dahulu
go build -o kbbi-api
./kbbi-api
```

Server akan berjalan di `http://localhost:8080`

## API Endpoints

### 1. Root Endpoint
```
GET /
```

Response:
```json
{
  "message": "KBBI API",
  "version": "1.0.0",
  "endpoints": {
    "search": "/api/v1/search?query=<word>&page=1&limit=20",
    "word": "/api/v1/word/<word>",
    "stats": "/api/v1/stats",
    "health": "/api/v1/health"
  }
}
```

### 2. Search Words
```
GET /api/v1/search?query=kata&page=1&limit=20
```

Parameters:
- `query` (required): Kata yang dicari
- `page` (optional): Halaman, default 1
- `limit` (optional): Jumlah hasil per halaman, max 100, default 20

Response:
```json
{
  "data": [
    {
      "word": "abadi",
      "arti": "<b>aba·di</b> <i>a</i> kekal; tidak berkesudahan: <i>di dunia ini tidak ada yg --;</i><br></i>",
      "type": 1
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20,
  "total_pages": 1
}
```

### 3. Get Specific Word
```
GET /api/v1/word/abadi
```

Response:
```json
{
  "word": "abadi",
  "arti": "<b>aba·di</b> <i>a</i> kekal; tidak berkesudahan: <i>di dunia ini tidak ada yg --;</i><br></i>",
  "type": 1
}
```

### 4. Get Statistics
```
GET /api/v1/stats
```

Response:
```json
{
  "total_words": 100000
}
```

### 5. Health Check
```
GET /api/v1/health
```

Response:
```json
{
  "status": "healthy"
}
```

## Database Schema

```sql
CREATE TABLE dictionary (
    id SERIAL PRIMARY KEY,
    word VARCHAR(255) NOT NULL,
    arti TEXT NOT NULL,
    type INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Project Structure

```
kbbi-api/
├── main.go              # Main application entry point
├── go.mod              # Go module file
├── docker-compose.yml  # Docker configuration
├── .env.example        # Environment variables example
├── config/             # Configuration files
│   └── database.go
├── models/             # Data models
│   └── dictionary.go
├── database/           # Database queries
│   └── queries.go
├── handlers/           # HTTP handlers
│   └── dictionary.go
└── routes/             # Route definitions
    └── routes.go
```

## Development

### Migrate Database

Jika menggunakan Docker, data akan otomatis di-migrate saat container pertama kali dijalankan. Untuk setup manual:

```bash
# Import data ke database
psql -h localhost -p 5436 -U kbbi_user -d kbbi -f dictionary_PostgreSQL.sql
```

### Environment Variables

- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5436)
- `DB_USER`: Database username (default: kbbi_user)
- `DB_PASSWORD`: Database password (default: kbbi_password)
- `DB_NAME`: Database name (default: kbbi)
- `PORT`: API server port (default: 8080)

## Contributing

1. Fork repository
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.