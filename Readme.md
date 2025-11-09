# YouTube Downloader - Setup Instructions

A full-stack application to download YouTube videos and audio files using Go, PostgreSQL, React, yt-dlp, and ffmpeg.

## Prerequisites

Install the following on your system:

1. **Go** (1.21 or higher)
   ```bash
   # Download from https://golang.org/dl/
   ```

2. **Node.js** (18 or higher) and npm
   ```bash
   # Download from https://nodejs.org/
   ```

3. **PostgreSQL** (14 or higher)
   ```bash
   # Ubuntu/Debian
   sudo apt update
   sudo apt install postgresql postgresql-contrib
   
   # macOS
   brew install postgresql@14
   brew services start postgresql@14
   
   # Windows: Download from https://www.postgresql.org/download/
   ```

4. **yt-dlp** (YouTube downloader)
   ```bash
   # Using pip
   pip install yt-dlp
   
   # Or download binary from https://github.com/yt-dlp/yt-dlp/releases
   ```

5. **ffmpeg** (for audio extraction and video merging)
   ```bash
   # Ubuntu/Debian
   sudo apt install ffmpeg
   
   # macOS
   brew install ffmpeg
   
   # Windows: Download from https://ffmpeg.org/download.html
   ```

## Database Setup

1. **Create PostgreSQL database and user:**
   ```bash
   # Login to PostgreSQL
   sudo -u postgres psql
   
   # Or on macOS
   psql postgres
   ```

2. **Run these SQL commands:**
   ```sql
   CREATE DATABASE ytdownloader;
   CREATE USER postgres WITH PASSWORD 'postgres';
   GRANT ALL PRIVILEGES ON DATABASE ytdownloader TO postgres;
   \q
   ```

3. **Test connection:**
   ```bash
   psql -h localhost -U postgres -d ytdownloader
   # Enter password: postgres
   ```

## Backend Setup

1. **Create project directory:**
   ```bash
   mkdir youtube-downloader
   cd youtube-downloader
   ```

2. **Create the Go files:**
   - Save the `main.go` file from the artifact
   - Save the `go.mod` file from the artifact

3. **Install Go dependencies:**
   ```bash
   go mod tidy
   ```

4. **Update database connection (if needed):**
   Edit `main.go` line 34 if your PostgreSQL credentials differ:
   ```go
   connStr := "host=localhost port=5432 user=YOUR_USER password=YOUR_PASSWORD dbname=ytdownloader sslmode=disable"
   ```

5. **Create downloads directory:**
   ```bash
   mkdir downloads
   ```

6. **Run the backend:**
   ```bash
   go run main.go
   ```
   
   You should see: `Server starting on :8080`

## Frontend Setup

1. **Create React app (in a new terminal):**
   ```bash
   # Using Vite (recommended)
   npm create vite@latest youtube-downloader-frontend -- --template react
   cd youtube-downloader-frontend
   npm install
   ```

2. **Install dependencies:**
   ```bash
   # Install Tailwind CSS 3.x and related packages
   npm install -D tailwindcss@3 postcss autoprefixer
   
   # Install Lucide icons
   npm install lucide-react
   
   # Initialize Tailwind (after installation)
   npx tailwindcss init -p
   ```

3. **Configure Tailwind:**
   
   Edit `tailwind.config.js`:
   ```js
   export default {
     content: [
       "./index.html",
       "./src/**/*.{js,ts,jsx,tsx}",
     ],
     theme: {
       extend: {},
     },
     plugins: [],
   }
   ```

4. **Update CSS:**
   
   Replace `src/index.css` with:
   ```css
   @tailwind base;
   @tailwind components;
   @tailwind utilities;
   ```

5. **Replace App.jsx:**
   - Copy the React component from the artifact to `src/App.jsx`

6. **Run the frontend:**
   ```bash
   npm run dev
   ```
   
   Open http://localhost:5173 in your browser

## Usage

1. **Start the backend** (terminal 1):
   ```bash
   cd youtube-downloader
   go run main.go
   ```

2. **Start the frontend** (terminal 2):
   ```bash
   cd youtube-downloader-frontend
   npm run dev
   ```

3. **Use the application:**
   - Open http://localhost:5173
   - Paste a YouTube URL
   - Select Video or Audio format
   - Click "Start Download"
   - Monitor progress in the download history
   - Click "Download" when complete

## Troubleshooting

### yt-dlp not found
```bash
# Check if installed
which yt-dlp

# Add to PATH if needed (Linux/macOS)
export PATH=$PATH:/usr/local/bin
```

### PostgreSQL connection error
```bash
# Check if PostgreSQL is running
sudo systemctl status postgresql  # Linux
brew services list                # macOS

# Start if stopped
sudo systemctl start postgresql   # Linux
brew services start postgresql    # macOS
```

### CORS errors
- Ensure backend is running on port 8080
- Check that the API_URL in App.jsx matches your backend URL

### Download failures
- Verify yt-dlp and ffmpeg are installed: `yt-dlp --version` and `ffmpeg -version`
- Check that the YouTube URL is valid
- Some videos may be restricted by region or age

## Project Structure

```
youtube-downloader/
├── main.go           # Backend server
├── go.mod            # Go dependencies
└── downloads/        # Downloaded files

youtube-downloader-frontend/
├── src/
│   ├── App.jsx       # Main React component
│   └── index.css     # Tailwind CSS
├── package.json      # Node dependencies
└── vite.config.js    # Vite configuration
```

## API Endpoints

- `POST /api/download` - Start a new download
- `GET /api/downloads` - Get all downloads
- `GET /api/downloads/:id` - Get specific download
- `GET /api/file/:id` - Download the file

## Features

- ✅ Download YouTube videos (MP4)
- ✅ Extract audio only (MP3)
- ✅ Real-time download status
- ✅ Download history tracking
- ✅ PostgreSQL database storage
- ✅ Clean, modern UI
- ✅ Concurrent downloads support

## Notes

- Downloads are stored in the `downloads/` directory
- The database table is created automatically on first run
- Status updates every 3 seconds automatically
- yt-dlp handles YouTube's changing API automatically