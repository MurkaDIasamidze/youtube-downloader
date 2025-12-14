# Configuration Guide

This guide covers all configuration options for Universal Downloader Pro.

## Table of Contents

1. [Backend Configuration](#backend-configuration)
2. [Frontend Configuration](#frontend-configuration)
3. [Database Configuration](#database-configuration)
4. [FFmpeg Configuration](#ffmpeg-configuration)
5. [yt-dlp Configuration](#yt-dlp-configuration)
6. [Performance Tuning](#performance-tuning)
7. [Environment Variables](#environment-variables)

## Backend Configuration

### Server Settings

Location: `main.go` - Fiber configuration

```go
app := fiber.New(fiber.Config{
    StreamRequestBody: true,              // Enable request body streaming
    BodyLimit:         10 * 1024 * 1024,  // Max body size: 10MB
    ReadTimeout:       30 * time.Minute,   // Read timeout
    WriteTimeout:      30 * time.Minute,   // Write timeout
    IdleTimeout:       30 * time.Minute,   // Idle connection timeout
    DisableKeepalive:  false,              // Enable keep-alive
})
```

**Adjustable Parameters:**

| Parameter | Default | Description | Recommended Range |
|-----------|---------|-------------|-------------------|
| BodyLimit | 10MB | Max request body size | 5-50MB |
| ReadTimeout | 30min | Request read timeout | 10-60min |
| WriteTimeout | 30min | Response write timeout | 10-60min |
| IdleTimeout | 30min | Idle connection timeout | 10-60min |

**When to Adjust:**
- Increase timeouts for very large files (>2GB)
- Decrease timeouts for faster failure detection
- Increase BodyLimit if handling large metadata requests

### Port Configuration

```go
port := ":8081"
```

**To change port:**
```go
port := ":3000" // Your desired port
```

**Considerations:**
- Ensure port is not in use
- Update frontend API URL accordingly
- Configure firewall rules if needed
- Ports below 1024 require root/admin privileges

### CORS Configuration

```go
app.Use(cors.New(cors.Config{
    AllowOrigins:     "*",                                    // Allow all origins
    AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",          // Allowed methods
    AllowHeaders:     "Content-Type, Authorization, Accept, Origin",
    ExposeHeaders:    "Content-Length, Content-Disposition",  // Exposed headers
    AllowCredentials: false,                                  // Don't allow credentials
}))
```

**For Production:**
```go
AllowOrigins: "https://yourdomain.com,https://www.yourdomain.com"
AllowCredentials: true  // If using authentication
```

### Buffer Configuration

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        buf := make([]byte, 512*1024) // 512KB buffer
        return &buf
    },
}
```

**Buffer Size Recommendations:**

| Use Case | Buffer Size | Notes |
|----------|-------------|-------|
| Low bandwidth | 256KB | Faster initial response |
| Standard | 512KB | Balanced (default) |
| High bandwidth | 1MB - 2MB | Better throughput for large files |
| Very large files | 2MB - 4MB | Maximum performance |

**To adjust:**
```go
buf := make([]byte, 1024*1024) // 1MB buffer
```

## Frontend Configuration

### API URL

Location: `src/App.jsx`

```javascript
const API_URL = 'http://localhost:8081/api';
```

**For Production:**
```javascript
// Use environment variable
const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081/api';
```

Create `.env` file:
```env
VITE_API_URL=https://api.yourdomain.com/api
```

### Polling Interval

Downloads list refresh interval:

```javascript
const interval = setInterval(fetchDownloads, 3000); // 3 seconds
```

**Recommendations:**

| Scenario | Interval | Reason |
|----------|----------|--------|
| Development | 3000ms | Fast updates for testing |
| Production (low traffic) | 5000ms | Balanced |
| Production (high traffic) | 10000ms | Reduce server load |
| Background updates | 30000ms | Minimal load |

### Build Configuration

Location: `vite.config.js` (if exists)

```javascript
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    minify: 'terser'
  }
})
```

## Database Configuration

### Connection String

```go
dsn := "host=localhost port=5432 user=postgres password=root dbname=postgres sslmode=disable"
```

**Parameters:**

| Parameter | Description | Example |
|-----------|-------------|---------|
| host | Database host | localhost, 192.168.1.100 |
| port | Database port | 5432 (default) |
| user | Database user | postgres |
| password | User password | your_password |
| dbname | Database name | postgres, downloader_db |
| sslmode | SSL mode | disable, require, verify-full |

**SSL Modes:**
- `disable` - No SSL (development)
- `require` - Require SSL (recommended for production)
- `verify-ca` - Verify CA certificate
- `verify-full` - Full certificate verification

**Production Example:**
```go
dsn := "host=db.example.com port=5432 user=app_user password=strong_pass dbname=downloader sslmode=require"
```

### Connection Pool

GORM handles connection pooling automatically, but you can configure:

```go
sqlDB, err := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

**Recommendations:**

| Parameter | Development | Production | High Load |
|-----------|-------------|------------|-----------|
| MaxIdleConns | 5 | 10 | 20 |
| MaxOpenConns | 50 | 100 | 200 |
| ConnMaxLifetime | 1 hour | 2 hours | 1 hour |

### Database Migrations

Auto-migration is enabled:

```go
db.AutoMigrate(&Download{})
```

**To disable:**
```go
// Comment out or remove:
// db.AutoMigrate(&Download{})
```

Then run migrations manually via migration files.

## FFmpeg Configuration

### FFmpeg Path

```go
var ffmpegPath = "C:\\ffmpeg-8.0-essentials_build\\bin"
```

**Platform-specific:**

```go
// Windows
ffmpegPath = "C:\\ffmpeg\\bin"

// Linux/macOS (if in PATH)
ffmpegPath = "" // Use system PATH
// Then call "ffmpeg" directly

// Linux/macOS (specific path)
ffmpegPath = "/usr/local/bin"
```

### Video Encoding Settings

#### MP4 (H.264)
```go
"-c:v", "libx264",
"-preset", "veryfast",      // Encoding speed
"-tune", "zerolatency",     // Optimize for streaming
"-crf", "23",               // Quality (0-51, lower = better)
"-c:a", "aac",
"-b:a", "192k",             // Audio bitrate
```

**Preset Options (speed vs compression):**
- `ultrafast` - Fastest, largest files
- `veryfast` - Fast, good for streaming (default)
- `faster` - Balanced
- `medium` - Better compression
- `slow` - Best compression, slowest

**CRF Values (Constant Rate Factor):**
- `18` - Visually lossless
- `23` - High quality (default)
- `28` - Good quality
- `35` - Lower quality, smaller files

#### WebM (VP8/VP9)
```go
"-c:v", "libvpx",
"-deadline", "realtime",    // Encoding speed
"-cpu-used", "8",           // Max speed (0-16)
"-crf", "23",
"-b:v", "2M",               // Video bitrate
```

#### Audio MP3
```go
"-acodec", "libmp3lame",
"-b:a", audioBitrate,       // e.g., "320k"
"-q:a", "0",                // Quality (0 = best)
"-ar", "44100",             // Sample rate
"-ac", "2",                 // Channels (stereo)
```

**Audio Bitrate Recommendations:**

| Quality | Bitrate | Use Case |
|---------|---------|----------|
| Low | 64k-128k | Voice, podcasts |
| Standard | 192k | Music, general use |
| High | 256k | High quality music |
| Best | 320k | Audiophile quality |

### Threading

```go
"-threads", "0"  // Use all available CPU cores
```

**Custom threading:**
```go
"-threads", "4"  // Use 4 threads
```

### Fragmented MP4

For better streaming:
```go
"-movflags", "+frag_keyframe+empty_moov+faststart+default_base_moof"
```

**Flags explained:**
- `frag_keyframe` - Fragment at keyframes
- `empty_moov` - Allow streaming start before complete
- `faststart` - Move metadata to beginning
- `default_base_moof` - Better seeking support

## yt-dlp Configuration

### Video Downloads

```go
"--concurrent-fragments", "4",    // Parallel fragment downloads
"--buffer-size", "512K",          // Buffer size
"--http-chunk-size", "10M",       // HTTP chunk size
"--throttled-rate", "100K",       // Minimum rate before throttling
```

**Performance Tuning:**

| Parameter | Low Bandwidth | Standard | High Bandwidth |
|-----------|---------------|----------|----------------|
| concurrent-fragments | 2 | 4 | 8 |
| buffer-size | 256K | 512K | 2M |
| http-chunk-size | 5M | 10M | 20M |

### Audio Downloads

```go
"--concurrent-fragments", "16",   // More fragments for faster audio
"--buffer-size", "2M",            // Larger buffer
"--retries", "10",                // Retry attempts
"--fragment-retries", "10",       // Fragment retry attempts
```

### Platform-Specific Settings

#### YouTube
```go
"--extractor-args", "youtube:player_client=android"  // Faster extraction
```

#### TikTok
```go
"--no-warnings",     // Suppress warnings
"-f", "best",        // Best available quality
```

### Format Selection

```go
// Video
"-f", "bestvideo[height<=720]+bestaudio/best[height<=720]"

// Audio
"-f", "bestaudio[ext=m4a]/bestaudio/best"
```

**Custom format strings:**
```go
// 4K video
"bestvideo[height<=2160]+bestaudio"

// HD only
"bestvideo[height<=1080]+bestaudio"

// Specific codec
"bestvideo[vcodec^=avc]+bestaudio[acodec^=mp4a]"
```

## Performance Tuning

### Streaming Optimization

Current flush interval:
```go
if totalBytes%(512*1024) == 0 { // Every 512KB
    w.Flush()
}
```

**Recommendations:**

| Scenario | Flush Interval | Latency | Memory |
|----------|----------------|---------|--------|
| Low latency | 256KB | Lower | Higher CPU |
| Balanced | 512KB | Medium | Balanced |
| High throughput | 1MB | Higher | Lower CPU |

### System Resources

**CPU Usage:**
- FFmpeg threading controls CPU usage
- More concurrent downloads = more CPU
- Consider limiting concurrent requests

**Memory Usage:**
- Buffer size affects memory per download
- Pool buffers to reuse memory
- Monitor with `go tool pprof`

**Disk I/O:**
- No temp files created (streaming only)
- Database writes are minimal
- Consider SSD for database

### Concurrent Downloads

Currently unlimited. To limit:

```go
// Add semaphore
var downloadSem = make(chan struct{}, 5) // Max 5 concurrent

func handleStreamFile(c *fiber.Ctx) error {
    downloadSem <- struct{}{} // Acquire
    defer func() { <-downloadSem }() // Release
    
    // ... rest of function
}
```

## Environment Variables

Create a `.env` file for configuration:

### Backend `.env`

```env
# Server
PORT=8081
HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=downloader
DB_SSLMODE=disable

# Paths
FFMPEG_PATH=/usr/local/bin
YTDLP_PATH=/usr/local/bin/yt-dlp

# Performance
MAX_CONCURRENT_DOWNLOADS=10
BUFFER_SIZE=524288
FLUSH_INTERVAL=524288

# Timeouts (in minutes)
READ_TIMEOUT=30
WRITE_TIMEOUT=30
IDLE_TIMEOUT=30
```

### Load environment variables in Go:

```go
import (
    "os"
    "github.com/joho/godotenv"
)

func init() {
    godotenv.Load()
}

port := os.Getenv("PORT")
if port == "" {
    port = "8081"
}
```

### Frontend `.env`

```env
VITE_API_URL=http://localhost:8081/api
VITE_POLLING_INTERVAL=3000
VITE_MAX_FILE_SIZE=5368709120
```

## Security Configuration

### Production Checklist

- [ ] Change default database password
- [ ] Enable SSL for database connection
- [ ] Configure specific CORS origins
- [ ] Set up HTTPS with SSL certificates
- [ ] Implement rate limiting
- [ ] Add authentication if needed
- [ ] Enable request logging
- [ ] Set up monitoring and alerts
- [ ] Configure firewall rules
- [ ] Disable debug logging

### Rate Limiting Example

```go
import "github.com/gofiber/fiber/v2/middleware/limiter"

app.Use(limiter.New(limiter.Config{
    Max:        20,
    Expiration: 30 * time.Second,
    LimitReached: func(c *fiber.Ctx) error {
        return c.Status(429).JSON(fiber.Map{
            "error": "Too many requests",
        })
    },
}))
```

## Monitoring Configuration

### Logging

Current setup uses Fiber's logger middleware:

```go
app.Use(logger.New(logger.Config{
    Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
}))
```

**Enhanced logging:**
```go
app.Use(logger.New(logger.Config{
    Format: "[${time}] ${status} - ${latency} ${method} ${path} ${ip} ${ua}\n",
    TimeFormat: "2006-01-02 15:04:05",
    TimeZone: "Local",
    Output: logFile, // io.Writer
}))
```

### Metrics

Consider adding:
- Prometheus metrics
- Response time tracking
- Active downloads counter
- Success/failure rates
- Storage usage

## Backup Configuration

### Database Backups

```bash
# Daily backup script
pg_dump -U postgres downloader > backup_$(date +%Y%m%d).sql

# Automated with cron
0 2 * * * /usr/bin/pg_dump -U postgres downloader > /backups/db_$(date +\%Y\%m\%d).sql
```

### Restore

```bash
psql -U postgres downloader < backup_20241214.sql
```

## Testing Configuration

### Development Mode

```go
if os.Getenv("ENVIRONMENT") == "development" {
    // Enable verbose logging
    // Disable rate limiting
    // Use shorter timeouts
}
```

### Test Database

```go
if os.Getenv("ENVIRONMENT") == "test" {
    dsn = "host=localhost port=5432 user=postgres password=test dbname=test_db sslmode=disable"
}
```

## Summary

Key configuration files:
- `main.go` - Backend settings
- `App.jsx` - Frontend API URL
- `.env` - Environment variables
- `vite.config.js` - Build configuration

Most common adjustments:
1. Change port numbers
2. Adjust timeouts for large files
3. Tune buffer sizes for performance
4. Configure CORS for production
5. Set up database connection properly