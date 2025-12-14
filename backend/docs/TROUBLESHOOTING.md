# Troubleshooting Guide

Common issues and solutions for Universal Downloader Pro.

## Table of Contents

1. [Installation Issues](#installation-issues)
2. [Backend Issues](#backend-issues)
3. [Frontend Issues](#frontend-issues)
4. [Download Issues](#download-issues)
5. [Streaming Issues](#streaming-issues)
6. [Performance Issues](#performance-issues)
7. [Database Issues](#database-issues)
8. [Platform-Specific Issues](#platform-specific-issues)

## Installation Issues

### Go Installation Problems

**Issue:** `go: command not found`

**Solution:**
```bash
# Verify Go is installed
go version

# If not found, check PATH
echo $PATH  # Linux/macOS
echo %PATH% # Windows

# Add Go to PATH (Linux/macOS)
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Windows: Add to System Environment Variables
# Go to System Properties > Environment Variables > Path
```

### PostgreSQL Connection Failed

**Issue:** `Failed to connect to database`

**Solution:**
```bash
# Check if PostgreSQL is running
sudo systemctl status postgresql  # Linux
brew services list                # macOS
services.msc                      # Windows - check postgresql service

# Start PostgreSQL if stopped
sudo systemctl start postgresql   # Linux
brew services start postgresql    # macOS

# Test connection manually
psql -U postgres -h localhost

# Check port is not blocked
netstat -an | grep 5432
```

**Issue:** `password authentication failed for user "postgres"`

**Solution:**
```bash
# Reset password
sudo -u postgres psql
ALTER USER postgres PASSWORD 'newpassword';
\q

# Update connection string in main.go
dsn := "host=localhost port=5432 user=postgres password=newpassword dbname=postgres sslmode=disable"
```

### FFmpeg Not Found

**Issue:** `exec: "ffmpeg": executable file not found in $PATH`

**Solution:**
```bash
# Check FFmpeg installation
ffmpeg -version

# Install if missing
# Linux
sudo apt install ffmpeg

# macOS
brew install ffmpeg

# Windows
# Download from https://www.gyan.dev/ffmpeg/builds/
# Add bin folder to PATH

# Verify installation
which ffmpeg  # Linux/macOS
where ffmpeg  # Windows
```

**Issue:** Windows FFmpeg path not working

**Solution:**
```go
// In main.go, update path with double backslashes
var ffmpegPath = "C:\\ffmpeg\\bin"

// Or use forward slashes
var ffmpegPath = "C:/ffmpeg/bin"

// Or add FFmpeg to system PATH and use:
ffmpegCmd := exec.Command("ffmpeg", ffmpegArgs...)
```

### yt-dlp Not Found

**Issue:** `exec: "yt-dlp": executable file not found`

**Solution:**
```bash
# Install yt-dlp
pip install yt-dlp

# Or using system package manager
# Linux
sudo curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp
sudo chmod a+rx /usr/local/bin/yt-dlp

# macOS
brew install yt-dlp

# Windows
pip install yt-dlp
# Or download .exe from GitHub releases

# Update yt-dlp
yt-dlp -U
```

## Backend Issues

### Port Already in Use

**Issue:** `listen tcp :8081: bind: address already in use`

**Solution:**
```bash
# Find process using port 8081
# Linux/macOS
lsof -i :8081
kill -9 <PID>

# Windows
netstat -ano | findstr :8081
taskkill /PID <PID> /F

# Or change port in main.go
port := ":8082"  // Use different port
```

### Go Module Issues

**Issue:** `package XXX is not in GOROOT`

**Solution:**
```bash
# Clean module cache
go clean -modcache

# Download dependencies
go mod download

# Tidy modules
go mod tidy

# If go.mod doesn't exist
go mod init universal-downloader
```

### CORS Errors in Browser

**Issue:** `Access to fetch at 'http://localhost:8081' blocked by CORS policy`

**Solution:**
```go
// In main.go, ensure CORS middleware is properly configured
app.Use(cors.New(cors.Config{
    AllowOrigins:     "*",  // Or specific origin
    AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
    AllowHeaders:     "Content-Type, Authorization, Accept, Origin",
    ExposeHeaders:    "Content-Length, Content-Disposition",
    AllowCredentials: false,
}))

// Make sure middleware is before routes
```

### Database Migration Failed

**Issue:** `Failed to migrate database`

**Solution:**
```bash
# Check database exists
psql -U postgres
\l  # List databases
CREATE DATABASE postgres;  # If needed

# Check user permissions
GRANT ALL PRIVILEGES ON DATABASE postgres TO postgres;

# Try manual table creation
CREATE TABLE downloads (
    id SERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    title TEXT,
    format TEXT,
    quality TEXT,
    extension TEXT,
    status TEXT,
    duration DOUBLE PRECISION,
    platform TEXT,
    created_at TIMESTAMP,
    completed_at TIMESTAMP
);
```

## Frontend Issues

### npm Install Fails

**Issue:** `npm ERR! code ERESOLVE`

**Solution:**
```bash
# Clear npm cache
npm cache clean --force

# Delete node_modules and package-lock.json
rm -rf node_modules package-lock.json

# Reinstall with legacy peer deps
npm install --legacy-peer-deps

# Or force install
npm install --force
```

### Frontend Can't Connect to Backend

**Issue:** `Failed to fetch` or `Network Error`

**Solution:**
```javascript
// Check API_URL in App.jsx
const API_URL = 'http://localhost:8081/api';

// Verify backend is running
// Open http://localhost:8081/ in browser

// Check browser console for errors
// Press F12 > Console tab

// Try direct API call
fetch('http://localhost:8081/')
  .then(r => r.json())
  .then(console.log)
```

### Build Fails

**Issue:** `npm run build` fails with errors

**Solution:**
```bash
# Check Node version
node --version  # Should be 18+

# Update Node if needed
# Using nvm
nvm install 18
nvm use 18

# Clear cache and rebuild
rm -rf node_modules package-lock.json dist
npm install
npm run build
```

### Vite Port Already in Use

**Issue:** `Port 5173 is already in use`

**Solution:**
```bash
# Kill process using port
# Linux/macOS
lsof -ti:5173 | xargs kill -9

# Windows
netstat -ano | findstr :5173
taskkill /PID <PID> /F

# Or use different port
npm run dev -- --port 3000
```

## Download Issues

### Download Fails Immediately

**Issue:** Download status goes to "failed" immediately

**Solution:**
```bash
# Check yt-dlp can access URL
yt-dlp --get-title "YOUR_URL"

# Check for error messages in backend logs
# Look for yt-dlp errors

# Update yt-dlp
yt-dlp -U

# Try different quality or format
# Use "best" quality instead of specific resolution
```

### Invalid URL Error

**Issue:** `URL is required` or `Download not found`

**Solution:**
```javascript
// Ensure URL is valid YouTube/TikTok format
// Valid formats:
// YouTube: https://www.youtube.com/watch?v=VIDEO_ID
// YouTube Short: https://youtu.be/VIDEO_ID
// TikTok: https://www.tiktok.com/@user/video/1234567890
// TikTok Short: https://vm.tiktok.com/ZMxxxxxx/

// Check URL encoding
const encodedUrl = encodeURIComponent(url);
```

### Age-Restricted Videos

**Issue:** Cannot download age-restricted content

**Solution:**
```bash
# yt-dlp can handle some age-restricted content
# But some videos may require authentication

# Try with cookies (login to YouTube first)
yt-dlp --cookies-from-browser firefox YOUR_URL

# In code, add cookies parameter
"--cookies-from-browser", "firefox"
```

### Private/Unavailable Videos

**Issue:** Video unavailable error

**Solution:**
- Verify video is publicly accessible
- Check if video requires authentication
- Ensure video isn't region-locked
- Try accessing video in browser first

## Streaming Issues

### Stream Stops Midway

**Issue:** Download stops at 50% or randomly

**Solution:**
```go
// Increase timeouts in main.go
ReadTimeout:  60 * time.Minute,  // Increase from 30
WriteTimeout: 60 * time.Minute,

// Check network stability
// Monitor backend logs for errors

// Increase retry attempts in yt-dlp
"--retries", "20",
"--fragment-retries", "20",
```

### Stream Never Starts

**Issue:** Download button clicked but nothing happens

**Solution:**
```javascript
// Check browser console for errors
// Verify stream URL is correct
console.log(`${API_URL}/stream/${id}`);

// Check if popup blocker is preventing download
// Some browsers block automatic downloads

// Try direct link
window.open(`${API_URL}/stream/${id}`, '_blank');
```

### Slow Streaming Speed

**Issue:** Download is very slow

**Solution:**
```go
// Increase concurrent fragments
"--concurrent-fragments", "16",  // Increase from 4

// Increase buffer sizes
"--buffer-size", "2M",
"--http-chunk-size", "20M",

// Check network bandwidth
// Run speed test

// Try lower quality setting
quality: "720p"  // Instead of "1080p"
```

### Memory Errors During Streaming

**Issue:** `Out of memory` or backend crashes

**Solution:**
```go
// Reduce buffer size
buf := make([]byte, 256*1024)  // Reduce from 512KB

// Limit concurrent downloads
var maxDownloads = make(chan struct{}, 5)

// Increase system swap space (Linux)
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

## Performance Issues

### High CPU Usage

**Issue:** CPU usage spikes to 100%

**Solution:**
```go
// Limit FFmpeg threads
"-threads", "4",  // Instead of "0" (all cores)

// Use faster encoding preset
"-preset", "ultrafast",

// Limit concurrent downloads
// Reduce concurrent fragments
"--concurrent-fragments", "2",
```

### High Memory Usage

**Issue:** Memory usage grows continuously

**Solution:**
```go
// Verify buffer pool is working correctly
// Check for memory leaks with pprof

import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

// Then access http://localhost:6060/debug/pprof/

// Ensure processes are properly closed
defer ytdlpCmd.Wait()
defer ffmpegCmd.Wait()
```

### Database Growing Large

**Issue:** Database size increases rapidly

**Solution:**
```sql
-- Clean old completed downloads
DELETE FROM downloads WHERE completed_at < NOW() - INTERVAL '30 days';

-- Or keep only recent downloads
DELETE FROM downloads WHERE id NOT IN (
    SELECT id FROM downloads ORDER BY created_at DESC LIMIT 1000
);

-- Vacuum database
VACUUM FULL downloads;
```

## Database Issues

### Connection Pool Exhausted

**Issue:** `too many connections`

**Solution:**
```go
// Configure connection pool
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(50)      // Reduce if needed
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(time.Hour)
```

### Slow Queries

**Issue:** Database operations are slow

**Solution:**
```sql
-- Add index on commonly queried fields
CREATE INDEX idx_downloads_status ON downloads(status);
CREATE INDEX idx_downloads_created ON downloads(created_at DESC);

-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM downloads ORDER BY created_at DESC;

-- Update statistics
ANALYZE downloads;
```

## Platform-Specific Issues

### YouTube Issues

**Issue:** YouTube downloads failing

**Solution:**
```bash
# Update yt-dlp (YouTube changes frequently)
yt-dlp -U

# Use Android player client (faster)
"--extractor-args", "youtube:player_client=android"

# Try different format
"-f", "bestaudio/best"  # Instead of specific format

# Check for IP rate limiting
# Wait a few minutes and try again
```

### TikTok Issues

**Issue:** TikTok downloads not working

**Solution:**
```bash
# Update yt-dlp
yt-dlp -U

# TikTok URLs expire quickly - use immediately

# Try vm.tiktok.com short links instead of full URLs

# Some TikTok videos have DRM protection and can't be downloaded
```

### Platform Detection Not Working

**Issue:** Wrong platform detected

**Solution:**
```go
// Check detectPlatform function
func detectPlatform(url string) string {
    url = strings.ToLower(url)
    if strings.Contains(url, "tiktok.com") || 
       strings.Contains(url, "vm.tiktok.com") {
        return "tiktok"
    }
    if strings.Contains(url, "youtube.com") || 
       strings.Contains(url, "youtu.be") {
        return "youtube"
    }
    return "unknown"
}
```

## Error Log Analysis

### Reading Backend Logs

```bash
# Run backend with verbose logging
go run main.go 2>&1 | tee backend.log

# Filter for errors
grep -i "error" backend.log

# Filter for specific download
grep "ID: 123" backend.log
```

### Common Error Messages

**"Failed to start yt-dlp"**
- yt-dlp not installed or not in PATH
- Invalid URL format

**"Failed to start ffmpeg"**
- FFmpeg not installed or path incorrect
- Missing codecs

**"Error reading from ffmpeg"**
- Network interrupted
- Source video unavailable
- Encoding error

**"Database error"**
- PostgreSQL not running
- Connection pool exhausted
- Invalid SQL query

## Getting Help

If issues persist:

1. **Check logs carefully** - most issues have error messages
2. **Test components individually**:
   ```bash
   # Test yt-dlp
   yt-dlp --get-title URL
   
   # Test FFmpeg
   ffmpeg -version
   
   # Test database
   psql -U postgres
   ```
3. **Verify all dependencies** are properly installed and updated
4. **Check GitHub issues** for similar problems
5. **Provide full error logs** when asking for help

## Debug Mode

Enable detailed logging:

```go
// In main.go
log.SetFlags(log.LstdFlags | log.Lshortfile)

// Add debug logging
log.Printf("DEBUG: URL=%s, Format=%s, Quality=%s", url, format, quality)

// Log all yt-dlp output
ytdlpStderr, _ := ytdlpCmd.StderrPipe()
go func() {
    scanner := bufio.NewScanner(ytdlpStderr)
    for scanner.Scan() {
        log.Printf("[yt-dlp] %s", scanner.Text())
    }
}()
```

## Prevention Tips

1. **Keep everything updated**:
   - yt-dlp (most important, YouTube changes frequently)
   - FFmpeg
   - Go dependencies
   - Node packages

2. **Monitor system resources**:
   - CPU usage
   - Memory usage
   - Disk space
   - Network bandwidth

3. **Regular maintenance**:
   - Clean old downloads from database
   - Vacuum database
   - Check log files size
   - Update security patches

4. **Test after changes**:
   - Test both YouTube and TikTok
   - Test video and audio downloads
   - Test different quality settings
   - Verify streaming works end-to-end