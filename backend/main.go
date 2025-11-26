package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Download struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	URL         string     `json:"url" gorm:"not null"`
	Title       string     `json:"title"`
	Format      string     `json:"format"` // video, audio
	Quality     string     `json:"quality"`
	Extension   string     `json:"extension"`
	Status      string     `json:"status"` // ready, streaming, completed, failed
	Duration    float64    `json:"duration"`
	Platform    string     `json:"platform"` // youtube, tiktok
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type DownloadRequest struct {
	URL       string `json:"url"`
	Format    string `json:"format"`
	Quality   string `json:"quality"`
	Extension string `json:"extension"`
}

type FormatInfo struct {
	VideoQualities []string `json:"video_qualities"`
	AudioQualities []string `json:"audio_qualities"`
	VideoFormats   []string `json:"video_formats"`
	AudioFormats   []string `json:"audio_formats"`
}

var (
	db         *gorm.DB
	ffmpegPath = "C:\\ffmpeg-8.0-essentials_build\\bin" // Change to your FFmpeg path
	bufferPool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, 512*1024) // Increased to 512KB
			return &buf
		},
	}
)

func main() {
	// Database connection
	dsn := "host=localhost port=5432 user=postgres password=root dbname=postgres sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Connected to database successfully")

	// Auto migrate
	if err := db.AutoMigrate(&Download{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed")

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		StreamRequestBody: true,
		BodyLimit:         10 * 1024 * 1024, // 10MB
		ReadTimeout:       30 * time.Minute,
		WriteTimeout:      30 * time.Minute,
		IdleTimeout:       30 * time.Minute,
		DisableKeepalive:  false,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Content-Type, Authorization, Accept, Origin",
		ExposeHeaders:    "Content-Length, Content-Disposition",
		AllowCredentials: false,
	}))

	// Routes
	app.Get("/", handleRoot)
	app.Get("/api/formats", handleGetFormats)
	app.Post("/api/download", handleDownload)
	app.Get("/api/downloads", handleGetDownloads)
	app.Get("/api/downloads/:id", handleGetDownload)
	app.Get("/api/stream/:id", handleStreamFile)

	port := ":8080" // Changed from 8080
	log.Printf("Server starting on %s", port)
	log.Println("Available routes:")
	log.Println("  GET  /")
	log.Println("  GET  /api/formats")
	log.Println("  POST /api/download")
	log.Println("  GET  /api/downloads")
	log.Println("  GET  /api/downloads/:id")
	log.Println("  GET  /api/stream/:id")

	if err := app.Listen(port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func handleRoot(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "YouTube & TikTok Downloader API is running",
		"version": "2.0 (Fiber + Optimized Streaming)",
	})
}

func detectPlatform(url string) string {
	if strings.Contains(url, "tiktok.com") || strings.Contains(url, "vm.tiktok.com") {
		return "tiktok"
	}
	return "youtube"
}

func handleGetFormats(c *fiber.Ctx) error {
	formats := FormatInfo{
		VideoQualities: []string{"144p", "240p", "360p", "480p", "720p", "1080p", "1440p", "2160p", "best"},
		AudioQualities: []string{"64k", "128k", "192k", "256k", "320k", "best"},
		VideoFormats:   []string{"mp4", "webm", "mkv"},
		AudioFormats:   []string{"mp3", "opus", "m4a"},
	}
	return c.JSON(formats)
}

func handleDownload(c *fiber.Ctx) error {
	var req DownloadRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate URL
	if req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "URL is required",
		})
	}

	// Detect platform
	platform := detectPlatform(req.URL)
	log.Printf("Detected platform: %s for URL: %s", platform, req.URL)

	// Set defaults
	if req.Quality == "" {
		req.Quality = "best"
	}
	if req.Extension == "" {
		if req.Format == "audio" {
			req.Extension = "mp3"
		} else {
			req.Extension = "mp4"
		}
	}

	// Get video title
	title := getVideoTitle(req.URL, platform)
	if title == "" {
		title = fmt.Sprintf("download_%d", time.Now().Unix())
	}

	log.Printf("Creating download: %s (%s, %s, %s, %s)", title, platform, req.Format, req.Quality, req.Extension)

	// Create download record
	download := Download{
		URL:       req.URL,
		Format:    req.Format,
		Quality:   req.Quality,
		Extension: req.Extension,
		Status:    "ready",
		Title:     title,
		Platform:  platform,
	}

	if err := db.Create(&download).Error; err != nil {
		log.Printf("Database error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create download record",
		})
	}

	log.Printf("Download created with ID: %d", download.ID)

	return c.JSON(fiber.Map{
		"id":       download.ID,
		"message":  fmt.Sprintf("Download ready, use /api/stream/%d to download", download.ID),
		"title":    title,
		"platform": platform,
	})
}

func handleGetDownloads(c *fiber.Ctx) error {
	var downloads []Download
	if err := db.Order("created_at DESC").Find(&downloads).Error; err != nil {
		log.Printf("Database error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch downloads",
		})
	}

	log.Printf("Found %d downloads", len(downloads))
	return c.JSON(downloads)
}

func handleGetDownload(c *fiber.Ctx) error {
	id := c.Params("id")
	log.Printf("Fetching download: %s", id)

	var download Download
	if err := db.First(&download, id).Error; err != nil {
		log.Printf("Download not found: %s", id)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Download not found",
		})
	}

	return c.JSON(download)
}

func handleStreamFile(c *fiber.Ctx) error {
	id := c.Params("id")
	log.Printf("Starting stream for ID: %s", id)

	var download Download
	if err := db.First(&download, id).Error; err != nil {
		log.Printf("Download not found: %s", id)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Download not found",
		})
	}

	// Update status
	db.Model(&Download{}).Where("id = ?", id).Update("status", "streaming")

	// Build commands
	var ytdlpArgs []string
	var ffmpegArgs []string

	if download.Format == "audio" {
		ytdlpArgs = buildAudioDownloadArgs(download.URL, download.Platform)
		ffmpegArgs = buildAudioFFmpegArgs(download.Extension, download.Quality)
	} else {
		ytdlpArgs = buildVideoDownloadArgs(download.URL, download.Quality, download.Platform)
		ffmpegArgs = buildVideoFFmpegArgs(download.Extension)
	}

	// Start yt-dlp
	log.Printf("Starting yt-dlp: %v", ytdlpArgs)
	ytdlpCmd := exec.Command("yt-dlp", ytdlpArgs...)
	ytdlpStdout, err := ytdlpCmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to create yt-dlp pipe: %v", err)
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to start download")
	}

	ytdlpStderr, _ := ytdlpCmd.StderrPipe()

	// Start ffmpeg
	log.Printf("Starting ffmpeg: %v", ffmpegArgs)
	ffmpegCmd := exec.Command(ffmpegPath+"\\ffmpeg.exe", ffmpegArgs...)
	ffmpegCmd.Stdin = ytdlpStdout
	ffmpegStdout, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to create ffmpeg pipe: %v", err)
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to start conversion")
	}

	ffmpegStderr, _ := ffmpegCmd.StderrPipe()

	// Start processes
	if err := ytdlpCmd.Start(); err != nil {
		log.Printf("Failed to start yt-dlp: %v", err)
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to start download")
	}

	if err := ffmpegCmd.Start(); err != nil {
		log.Printf("Failed to start ffmpeg: %v", err)
		ytdlpCmd.Process.Kill()
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to start conversion")
	}

	// Log errors in background
	go logErrors(ytdlpStderr, "yt-dlp")
	go logErrors(ffmpegStderr, "ffmpeg")

	// Set response headers
	c.Set("Content-Type", getContentType(download.Extension))
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.%s\"", download.Title, download.Extension))
	c.Set("Transfer-Encoding", "chunked")
	c.Set("Cache-Control", "no-cache")

	// Stream with optimized buffer
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		buf := bufferPool.Get().(*[]byte)
		defer bufferPool.Put(buf)

		totalBytes := int64(0)
		for {
			n, err := ffmpegStdout.Read(*buf)
			if n > 0 {
				written, writeErr := w.Write((*buf)[:n])
				if writeErr != nil {
					log.Printf("Error writing to response: %v", writeErr)
					break
				}
				totalBytes += int64(written)

				// Flush more frequently for better streaming
				if totalBytes%(512*1024) == 0 { // Every 512KB
					if err := w.Flush(); err != nil {
						log.Printf("Error flushing: %v", err)
						break
					}
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("Error reading from ffmpeg: %v", err)
				break
			}
		}

		// Final flush
		w.Flush()

		log.Printf("Streamed %d bytes for download ID %s", totalBytes, id)

		// Wait for processes to complete
		ytdlpCmd.Wait()
		ffmpegCmd.Wait()

		// Update database
		now := time.Now()
		db.Model(&Download{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":       "completed",
			"completed_at": now,
		})
	})

	return nil
}

func buildVideoDownloadArgs(url, quality, platform string) []string {
	if platform == "tiktok" {
		return []string{
			"--no-warnings",
			"-f", "best",
			"--concurrent-fragments", "4",
			"--buffer-size", "512K",
			"--http-chunk-size", "10M",
			"-o", "-",
			url,
		}
	}

	formatStr := buildVideoFormat(quality)
	return []string{
		"--no-playlist",
		"-f", formatStr,
		"--no-check-certificate",
		"--concurrent-fragments", "4",
		"--buffer-size", "512K",
		"--http-chunk-size", "10M",
		"--throttled-rate", "100K",
		"-o", "-",
		url,
	}
}

func buildAudioDownloadArgs(url, platform string) []string {
	if platform == "tiktok" {
		// TikTok: download best quality with optimized settings
		return []string{
			"--no-warnings",
			"-f", "best",
			"--concurrent-fragments", "16", // Increased to 16
			"--buffer-size", "2M", // Increased to 2MB
			"--http-chunk-size", "10M",
			"--retries", "10",
			"--fragment-retries", "10",
			"--no-part", // Don't create .part files
			"-o", "-",
			url,
		}
	}

	// YouTube: optimize for audio streams
	return []string{
		"--no-playlist",
		"-f", "bestaudio[ext=m4a]/bestaudio/best", // Prefer m4a for AAC
		"--no-check-certificate",
		"--concurrent-fragments", "16", // Increased to 16
		"--buffer-size", "2M", // Increased to 2MB
		"--http-chunk-size", "10M",
		"--retries", "10",
		"--fragment-retries", "10",
		"--no-part", // Don't create .part files
		"--extractor-args", "youtube:player_client=android", // Faster YouTube extraction
		"-o", "-",
		url,
	}
}

func buildVideoFFmpegArgs(extension string) []string {
	switch extension {
	case "mp4":
		return []string{
			"-hide_banner",
			"-loglevel", "error",
			"-i", "pipe:0",
			"-c:v", "libx264",
			"-preset", "veryfast", // Changed from ultrafast for better compression
			"-tune", "zerolatency", // Optimize for streaming
			"-crf", "23",
			"-c:a", "aac",
			"-b:a", "192k",
			"-movflags", "+frag_keyframe+empty_moov+faststart+default_base_moof",
			"-max_muxing_queue_size", "9999",
			"-f", "mp4",
			"-threads", "0",
			"pipe:1",
		}
	case "webm":
		return []string{
			"-hide_banner",
			"-loglevel", "error",
			"-i", "pipe:0",
			"-c:v", "libvpx",
			"-deadline", "realtime", // Fastest VP8 encoding
			"-cpu-used", "8", // Maximum speed (0-16, higher = faster)
			"-crf", "23",
			"-b:v", "2M",
			"-c:a", "libopus",
			"-b:a", "192k",
			"-max_muxing_queue_size", "9999",
			"-f", "webm",
			"-threads", "0",
			"pipe:1",
		}
	case "mkv":
		return []string{
			"-hide_banner",
			"-loglevel", "error",
			"-i", "pipe:0",
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-tune", "zerolatency",
			"-crf", "23",
			"-c:a", "aac",
			"-b:a", "192k",
			"-max_muxing_queue_size", "9999",
			"-f", "matroska",
			"-threads", "0",
			"pipe:1",
		}
	default:
		return []string{
			"-hide_banner",
			"-loglevel", "error",
			"-i", "pipe:0",
			"-c:v", "libx264",
			"-preset", "veryfast",
			"-tune", "zerolatency",
			"-crf", "23",
			"-c:a", "aac",
			"-b:a", "192k",
			"-movflags", "+frag_keyframe+empty_moov+faststart",
			"-max_muxing_queue_size", "9999",
			"-f", "mp4",
			"-threads", "0",
			"pipe:1",
		}
	}
}

func buildAudioFFmpegArgs(extension, quality string) []string {
	audioBitrate := quality
	if audioBitrate == "best" {
		audioBitrate = "320k"
	}

	switch extension {
	case "mp3":
		return []string{
			"-hide_banner",
			"-loglevel", "error",
			"-i", "pipe:0",
			"-vn",
			"-acodec", "libmp3lame",
			"-b:a", audioBitrate,
			"-q:a", "0",
			"-compression_level", "0",
			"-ar", "44100",
			"-ac", "2",
			"-f", "mp3",
			"-threads", "0",
			"-max_muxing_queue_size", "9999",
			"pipe:1",
		}
	case "m4a":
		return []string{
			"-hide_banner",
			"-loglevel", "error",
			"-i", "pipe:0",
			"-vn",
			"-c:a", "aac",
			"-b:a", audioBitrate,
			"-aac_coder", "fast",
			"-profile:a", "aac_low",
			"-ar", "44100",
			"-ac", "2",
			"-movflags", "+frag_keyframe+empty_moov+faststart",
			"-f", "ipod",
			"-threads", "0",
			"-max_muxing_queue_size", "9999",
			"pipe:1",
		}
	case "opus":
		return []string{
			"-hide_banner",
			"-loglevel", "error",
			"-i", "pipe:0",
			"-vn",
			"-acodec", "libopus",
			"-b:a", audioBitrate,
			"-compression_level", "0",
			"-ar", "48000",
			"-ac", "2",
			"-f", "opus",
			"-threads", "0",
			"-max_muxing_queue_size", "9999",
			"pipe:1",
		}
	default:
		return []string{
			"-hide_banner",
			"-loglevel", "error",
			"-i", "pipe:0",
			"-vn",
			"-acodec", "libmp3lame",
			"-b:a", audioBitrate,
			"-ar", "44100",
			"-ac", "2",
			"-f", "mp3",
			"-threads", "0",
			"-max_muxing_queue_size", "9999",
			"pipe:1",
		}
	}
}

func buildVideoFormat(quality string) string {
	if quality == "best" {
		return "bestvideo+bestaudio/best"
	}
	height := quality[:len(quality)-1]
	return fmt.Sprintf("bestvideo[height<=%s]+bestaudio/best[height<=%s]", height, height)
}

func getAudioCodec(extension string) string {
	switch extension {
	case "mp3":
		return "libmp3lame"
	case "aac", "m4a":
		return "aac"
	case "opus":
		return "libopus"
	default:
		return "libmp3lame"
	}
}

func getVideoCodecs(extension string) (string, string) {
	switch extension {
	case "mp4":
		return "libx264", "aac"
	case "webm":
		return "libvpx-vp9", "libopus"
	case "mkv":
		return "libx264", "aac"
	default:
		return "libx264", "aac"
	}
}

func getVideoTitle(url, platform string) string {
	var cmd *exec.Cmd
	if platform == "tiktok" {
		cmd = exec.Command("yt-dlp", "--get-title", "--no-warnings", url)
	} else {
		cmd = exec.Command("yt-dlp", "--get-title", "--no-playlist", "--skip-download", url)
	}

	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to get title: %v", err)
		return ""
	}

	title := strings.TrimSpace(string(output))
	title = fixUTF8(sanitizeFilename(title))
	return title
}

func fixUTF8(s string) string {
	return string([]rune(s))
}

func sanitizeFilename(name string) string {
	reg := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	name = reg.ReplaceAllString(name, "")
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}

func getContentType(extension string) string {
	switch extension {
	case "mp4":
		return "video/mp4"
	case "webm":
		return "video/webm"
	case "mkv":
		return "video/x-matroska"
	case "mp3":
		return "audio/mpeg"
	case "m4a":
		return "audio/mp4"
	case "opus":
		return "audio/opus"
	default:
		return "application/octet-stream"
	}
}

func logErrors(stderr io.ReadCloser, source string) {
	if stderr == nil {
		return
	}
	buf := make([]byte, 4096)
	for {
		n, err := stderr.Read(buf)
		if n > 0 {
			msg := strings.TrimSpace(string(buf[:n]))
			// Filter out FFmpeg progress messages (frame=, size=, time=)
			if !strings.Contains(msg, "frame=") && 
			   !strings.Contains(msg, "size=") && 
			   !strings.Contains(msg, "speed=") &&
			   msg != "" {
				log.Printf("[%s] %s", source, msg)
			}
		}
		if err != nil {
			break
		}
	}
}

func getVideoDuration(url string) float64 {
	cmd := exec.Command("yt-dlp", "--get-duration", "--no-playlist", url)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	duration, _ := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	return duration
}