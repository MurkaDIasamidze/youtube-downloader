# Universal Downloader Pro

A high-performance, full-stack media downloader application that enables users to download videos and audio from YouTube and TikTok with customizable quality and format options.

## 🎯 What This Project Does

Universal Downloader Pro is a web-based application that:

- **Downloads media** from YouTube and TikTok platforms
- **Converts formats** on-the-fly using FFmpeg
- **Streams content** directly to users without server storage
- **Provides quality options** for both video and audio downloads
- **Supports multiple formats**: MP4, WebM, MKV for video; MP3, Opus, M4A for audio
- **Tracks download history** with a beautiful, responsive UI
- **Auto-detects platforms** from pasted URLs

## 🎬 What It's For

This application is designed for:

- **Content creators** who need to download their own content for editing
- **Educators** creating offline educational materials
- **Researchers** archiving media for academic purposes
- **Personal use** for legal media downloading and format conversion
- **Anyone** who needs reliable, high-quality media downloads with custom settings

## ✨ Key Features

### Backend (Go + Fiber)
- High-performance streaming architecture
- Direct piping from yt-dlp to FFmpeg to client
- No temporary file storage required
- PostgreSQL database for download history
- Optimized buffer management with sync.Pool
- Concurrent fragment downloading
- Real-time conversion and streaming

### Frontend (React)
- Modern, gradient-based UI design
- Real-time download status updates
- Platform auto-detection
- Advanced settings for power users
- Responsive design for all devices
- Download history management

## 🛠️ Technology Stack

**Backend:**
- Go (Golang)
- Fiber web framework
- GORM (PostgreSQL ORM)
- yt-dlp (media downloader)
- FFmpeg (media converter)

**Frontend:**
- React 18
- Lucide React (icons)
- Tailwind CSS (styling)
- Vite (build tool)

**Database:**
- PostgreSQL

## 📋 Prerequisites

Before installation, ensure you have:

- Go 1.21 or higher
- Node.js 18+ and npm
- PostgreSQL 14+
- FFmpeg (with all codecs)
- yt-dlp (latest version)

## 🚀 Quick Start

1. **Clone the repository**
```bash
git clone <your-repo-url>
cd universal-downloader-pro
```

2. **Set up the backend**
```bash
cd backend
go mod download
# Configure database connection in main.go
go run main.go
```

3. **Set up the frontend**
```bash
cd frontend
npm install
npm run dev
```

4. **Access the application**
- Frontend: http://localhost:5173
- Backend API: http://localhost:8081

## 📚 Documentation

Detailed documentation is available in the `docs/` folder:

- **[Installation Guide](docs/INSTALLATION.md)** - Complete setup instructions
- **[API Documentation](docs/API.md)** - Full API reference
- **[Configuration Guide](docs/CONFIGURATION.md)** - System configuration options
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions

## ⚖️ Legal Notice

This tool is intended for downloading content that you have the right to download. Users are responsible for complying with:

- YouTube's Terms of Service
- TikTok's Terms of Service
- Copyright laws in their jurisdiction
- Content creator rights

**Use responsibly and legally.**

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is provided as-is for educational and personal use.

## 🐛 Known Issues

- Large files (>2GB) may timeout on slower connections
- Some TikTok videos with special DRM may fail
- Rate limiting may occur with excessive requests

