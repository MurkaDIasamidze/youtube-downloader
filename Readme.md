# Universal Downloader Pro

Download videos and audio from YouTube and TikTok with custom quality settings.

## Features

- 🎥 Video & 🎵 Audio downloads from YouTube and TikTok
- ⚙️ Multiple quality options (144p-4K, 64k-320k)
- 📝 Format selection (MP4, WebM, MKV, MP3, AAC, Opus)
- 📊 Download history with real-time status
- 🔄 Direct streaming without temp files

## Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 12+
- yt-dlp: `pip install yt-dlp`
- FFmpeg: Download from [ffmpeg.org](https://ffmpeg.org/download.html)

## Quick Install

```bash
# 1. Clone repo
git clone <your-repo-url>
cd universal-downloader-pro

# 2. Setup PostgreSQL
psql -U postgres -c "CREATE DATABASE postgres;"

# 3. Backend setup
go mod tidy
# Update main.go line 49 with FFmpeg path and line 79 with DB password
go run main.go

# 4. Frontend setup (new terminal)
cd frontend
npm install
npm install lucide-react
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
npm run dev
```

## Configuration

**main.go** (line 79):
```go
dsn := "host=localhost port=5432 user=postgres password=YOUR_PASSWORD dbname=postgres sslmode=disable"
```

**main.go** (line 49):
```go
var ffmpegPath = "C:\\ffmpeg-8.0-essentials_build\\bin" // Windows
// var ffmpegPath = "/usr/local/bin" // macOS/Linux
```

**tailwind.config.js**:
```javascript
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: { extend: {} },
  plugins: [],
}
```

**src/index.css**:
```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

## API Endpoints

- `GET /api/formats` - Get available formats
- `POST /api/download` - Create download job
- `GET /api/downloads` - List all downloads
- `GET /api/downloads/{id}` - Get specific download
- `GET /api/stream/{id}` - Stream/download file

## Usage

1. Open `http://localhost:5173`
2. Paste YouTube or TikTok URL
3. Select format (Video/Audio)
4. Click "Start Download"
5. Download from history when ready

## Testing

```bash
# Run tests
go test ./... -v

# With coverage
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Project Structure

```
├── main.go              # Backend server
├── go.mod              # Go dependencies
├── tests/
│   └── main_test.go    # Unit tests
└── frontend/
    ├── src/
    │   ├── App.jsx     # React component
    │   ├── main.jsx
    │   └── index.css
    ├── package.json
    ├── vite.config.js
    └── tailwind.config.js
```

## Troubleshooting

**Database error**: Check PostgreSQL is running and credentials are correct

**yt-dlp not found**: Install with `pip install yt-dlp` and add to PATH

**FFmpeg error**: Update `ffmpegPath` in main.go with correct installation path

**CORS error**: Ensure backend runs on :8080, frontend on :5173

**Download fails**: Update yt-dlp: `pip install -U yt-dlp`

## Tech Stack

**Backend**: Go, Gorilla Mux, GORM, PostgreSQL, yt-dlp, FFmpeg  
**Frontend**: React, Vite, Tailwind CSS, Lucide Icons

## License

MIT License - For educational and personal use only. Respect copyright laws.

## Note

This application is for personal use. Always respect platform terms of service and copyright laws when downloading content.
