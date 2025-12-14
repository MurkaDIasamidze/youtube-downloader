# Installation Guide

This guide will walk you through the complete installation process for Universal Downloader Pro.

## Table of Contents

1. [System Requirements](#system-requirements)
2. [Installing Dependencies](#installing-dependencies)
3. [Database Setup](#database-setup)
4. [Backend Setup](#backend-setup)
5. [Frontend Setup](#frontend-setup)
6. [Verification](#verification)
7. [Production Deployment](#production-deployment)

## System Requirements

### Minimum Requirements
- **OS**: Windows 10/11, Linux (Ubuntu 20.04+), macOS 11+
- **CPU**: 2 cores
- **RAM**: 4GB
- **Storage**: 2GB free space
- **Network**: Stable internet connection

### Recommended Requirements
- **CPU**: 4+ cores
- **RAM**: 8GB+
- **Storage**: 10GB+ free space
- **Network**: High-speed connection

## Installing Dependencies

### 1. Install Go (Golang)

**Windows:**
1. Download Go from https://go.dev/dl/
2. Run the installer
3. Verify installation:
```bash
go version
```

**Linux:**
```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
go version
```

**macOS:**
```bash
brew install go
go version
```

### 2. Install Node.js and npm

**Windows:**
1. Download from https://nodejs.org/
2. Run the installer
3. Verify:
```bash
node --version
npm --version
```

**Linux:**
```bash
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
node --version
npm --version
```

**macOS:**
```bash
brew install node
node --version
npm --version
```

### 3. Install PostgreSQL

**Windows:**
1. Download from https://www.postgresql.org/download/windows/
2. Run the installer
3. Remember your postgres password
4. Verify:
```bash
psql --version
```

**Linux:**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
psql --version
```

**macOS:**
```bash
brew install postgresql@14
brew services start postgresql@14
psql --version
```

### 4. Install FFmpeg

**Windows:**
1. Download FFmpeg from https://www.gyan.dev/ffmpeg/builds/
2. Extract to `C:\ffmpeg-8.0-essentials_build`
3. Add `C:\ffmpeg-8.0-essentials_build\bin` to PATH
4. Verify:
```bash
ffmpeg -version
```

**Linux:**
```bash
sudo apt update
sudo apt install ffmpeg
ffmpeg -version
```

**macOS:**
```bash
brew install ffmpeg
ffmpeg -version
```

### 5. Install yt-dlp

**Windows:**
```bash
# Using pip
pip install yt-dlp

# Or download executable from https://github.com/yt-dlp/yt-dlp/releases
# Add to PATH
```

**Linux:**
```bash
sudo curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
sudo chmod a+rx /usr/local/bin/yt-dlp
yt-dlp --version
```

**macOS:**
```bash
brew install yt-dlp
yt-dlp --version
```

## Database Setup

### 1. Create Database

**Connect to PostgreSQL:**
```bash
# Windows/Linux/macOS
psql -U postgres
```

**Create database and configure:**
```sql
-- Create database (if needed)
CREATE DATABASE postgres;

-- Verify connection settings
\conninfo

-- Exit
\q
```

### 2. Configure Connection

The default connection string in the code is:
```
host=localhost port=5432 user=postgres password=root dbname=postgres sslmode=disable
```

**To change the password:**
```sql
psql -U postgres
ALTER USER postgres PASSWORD 'your_new_password';
```

Then update `main.go`:
```go
dsn := "host=localhost port=5432 user=postgres password=your_new_password dbname=postgres sslmode=disable"
```

## Backend Setup

### 1. Navigate to Project Directory
```bash
cd path/to/universal-downloader-pro/backend
```

### 2. Initialize Go Module
```bash
go mod init universal-downloader
```

### 3. Install Dependencies
```bash
go get github.com/gofiber/fiber/v2
go get github.com/gofiber/fiber/v2/middleware/cors
go get github.com/gofiber/fiber/v2/middleware/logger
go get github.com/gofiber/fiber/v2/middleware/recover
go get gorm.io/driver/postgres
go get gorm.io/gorm
```

Or simply:
```bash
go mod tidy
```

### 4. Configure FFmpeg Path

**Edit `main.go` line 44:**

**Windows:**
```go
ffmpegPath = "C:\\ffmpeg-8.0-essentials_build\\bin"
```

**Linux/macOS:**
```go
ffmpegPath = "/usr/local/bin" // or wherever FFmpeg is installed
```

Or use the system PATH:
```go
// Just use "ffmpeg" if it's in PATH
ffmpegPath = ""
// Then in commands, use "ffmpeg" instead of ffmpegPath+"\\ffmpeg.exe"
```

### 5. Run the Backend
```bash
go run main.go
```

You should see:
```
Connected to database successfully
Database migration completed
Server starting on :8081
```

### 6. Test Backend API
```bash
curl http://localhost:8081/
```

Expected response:
```json
{
  "status": "ok",
  "message": "YouTube & TikTok Downloader API is running",
  "version": "2.0 (Fiber + Optimized Streaming)"
}
```

## Frontend Setup

### 1. Navigate to Frontend Directory
```bash
cd path/to/universal-downloader-pro/frontend
```

### 2. Install Dependencies
```bash
npm install
```

### 3. Configure API URL

**Edit `src/App.jsx` line 12:**
```javascript
const API_URL = 'http://localhost:8081/api';
```

If deploying to production, change to your backend URL:
```javascript
const API_URL = 'https://your-api-domain.com/api';
```

### 4. Start Development Server
```bash
npm run dev
```

You should see:
```
VITE v5.x.x ready in xxx ms

➜  Local:   http://localhost:5173/
➜  Network: use --host to expose
```

### 5. Access the Application

Open your browser and navigate to:
```
http://localhost:5173
```

## Verification

### 1. Test Video Download
1. Open the application at `http://localhost:5173`
2. Paste a YouTube URL (e.g., a short video)
3. Select "Video" format
4. Click "Start Download"
5. After processing, click "Download"

### 2. Test Audio Download
1. Paste a YouTube or TikTok URL
2. Select "Audio" format
3. Choose quality and format (MP3)
4. Click "Start Download"
5. Download the audio file

### 3. Check Database
```bash
psql -U postgres
SELECT * FROM downloads;
```

You should see your download records.

## Production Deployment

### Backend Production Build

1. **Build the binary:**
```bash
go build -o downloader-server main.go
```

2. **Run in production:**
```bash
./downloader-server
```

3. **Use a process manager (systemd on Linux):**

Create `/etc/systemd/system/downloader.service`:
```ini
[Unit]
Description=Universal Downloader Backend
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/downloader
ExecStart=/opt/downloader/downloader-server
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable downloader
sudo systemctl start downloader
sudo systemctl status downloader
```

### Frontend Production Build

1. **Build for production:**
```bash
npm run build
```

2. **Serve with nginx:**

Install nginx:
```bash
sudo apt install nginx  # Linux
brew install nginx      # macOS
```

Configure `/etc/nginx/sites-available/downloader`:
```nginx
server {
    listen 80;
    server_name your-domain.com;

    root /var/www/downloader/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://localhost:8081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

Enable and restart:
```bash
sudo ln -s /etc/nginx/sites-available/downloader /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

## Troubleshooting Installation

### Go Module Issues
```bash
go clean -modcache
go mod tidy
```

### PostgreSQL Connection Failed
```bash
# Check if PostgreSQL is running
sudo systemctl status postgresql  # Linux
brew services list                 # macOS

# Test connection
psql -U postgres -h localhost
```

### FFmpeg Not Found
```bash
# Check if FFmpeg is in PATH
which ffmpeg       # Linux/macOS
where ffmpeg       # Windows

# Test FFmpeg
ffmpeg -version
```

### yt-dlp Not Found
```bash
# Update yt-dlp
pip install -U yt-dlp
# or
sudo yt-dlp -U
```

### Port Already in Use
```bash
# Find process using port 8081
lsof -i :8081        # Linux/macOS
netstat -ano | findstr :8081  # Windows

# Kill the process or change port in main.go
```

### CORS Issues
Check that the backend CORS middleware allows your frontend origin:
```go
app.Use(cors.New(cors.Config{
    AllowOrigins: "*", // or "http://localhost:5173"
}))
```

## Next Steps

After successful installation:

1. Read the [API Documentation](API.md)
2. Configure advanced settings in [Configuration Guide](CONFIGURATION.md)
3. Review [Troubleshooting](TROUBLESHOOTING.md) for common issues

