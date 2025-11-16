package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
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
	Status      string     `json:"status"` // pending, processing, completed, failed
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

var db *gorm.DB
var ffmpegPath = "C:\\ffmpeg-8.0-essentials_build\\bin" // Change this to your FFmpeg path

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Disposition")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Database connection with GORM
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

	router := mux.NewRouter()

	// Apply CORS middleware
	router.Use(corsMiddleware)

	// Register routes
	router.HandleFunc("/api/formats", handleGetFormats).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/download", handleDownload).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/downloads", handleGetDownloads).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/downloads/{id}", handleGetDownload).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/stream/{id}", handleStreamFile).Methods("GET", "OPTIONS")

	// Test route
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "YouTube & TikTok Downloader API is running",
		})
	}).Methods("GET")

	log.Println("Server starting on :8080")
	log.Println("Available routes:")
	log.Println("  GET  /")
	log.Println("  GET  /api/formats")
	log.Println("  POST /api/download")
	log.Println("  GET  /api/downloads")
	log.Println("  GET  /api/downloads/{id}")
	log.Println("  GET  /api/stream/{id}")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func detectPlatform(url string) string {
	if strings.Contains(url, "tiktok.com") || strings.Contains(url, "vm.tiktok.com") {
		return "tiktok"
	}
	return "youtube"
}

func handleGetFormats(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling /api/formats")
	w.Header().Set("Content-Type", "application/json")
	formats := FormatInfo{
		VideoQualities: []string{"144p", "240p", "360p", "480p", "720p", "1080p", "1440p", "2160p", "best"},
		AudioQualities: []string{"64k", "128k", "192k", "256k", "320k", "best"},
		VideoFormats:   []string{"mp4", "webm", "mkv"},
		AudioFormats:   []string{"mp3", "aac", "opus", "m4a"},
	}
	json.NewEncoder(w).Encode(formats)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling /api/download")
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Detect platform
	platform := detectPlatform(req.URL)
	log.Printf("Detected platform: %s", platform)

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

	// Get video info for title
	var titleCmd *exec.Cmd
	if platform == "tiktok" {
		titleCmd = exec.Command("yt-dlp", "--get-title", "--no-warnings", req.URL)
	} else {
		titleCmd = exec.Command("yt-dlp", "--get-title", "--no-playlist", "--skip-download", req.URL)
	}

	titleOutput, err := titleCmd.Output()
	title := "download"
	if err == nil {
		title = fixUTF8(sanitizeFilename(string(titleOutput))) // <- FIXED HERE
	}
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Download created with ID: %d", download.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       download.ID,
		"message":  "Download ready, use /api/stream/" + fmt.Sprint(download.ID) + " to download",
		"title":    title,
		"platform": platform,
	})
}

// ------------------- NEW FUNCTION -------------------
func fixUTF8(s string) string {
	r := []rune(s) // преобразует в руны → автоматически убирает невалидные байты
	return string(r)
}
// ----------------------------------------------------

func handleGetDownloads(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling /api/downloads")
	var downloads []Download
	if err := db.Order("created_at DESC").Find(&downloads).Error; err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Found %d downloads", len(downloads))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(downloads)
}

func handleGetDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("Handling /api/downloads/%s", id)

	var download Download
	if err := db.First(&download, id).Error; err != nil {
		log.Printf("Download not found: %s", id)
		http.Error(w, "Download not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(download)
}

func handleStreamFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("Handling /api/stream/%s", id)

	var download Download
	if err := db.First(&download, id).Error; err != nil {
		log.Printf("Download not found: %s", id)
		http.Error(w, "Download not found", http.StatusNotFound)
		return
	}

	// Update status to streaming
	db.Model(&Download{}).Where("id = ?", id).Update("status", "streaming")

	// Build yt-dlp and ffmpeg commands
	var ytdlpArgs []string
	var ffmpegArgs []string

	if download.Format == "audio" {
		ytdlpArgs = buildAudioDownloadArgs(download.URL, download.Platform)

		audioCodec := getAudioCodec(download.Extension)
		audioBitrate := download.Quality
		if audioBitrate == "best" {
			audioBitrate = "320k"
		}

		ffmpegArgs = []string{
			"-i", "pipe:0",
			"-vn",
			"-acodec", audioCodec,
			"-b:a", audioBitrate,
			"-f", download.Extension,
			"-threads", "0",
			"pipe:1",
		}
	} else {
		ytdlpArgs = buildVideoDownloadArgs(download.URL, download.Quality, download.Platform)

		videoCodec, audioCodec := getVideoCodecs(download.Extension)
		ffmpegArgs = []string{
			"-i", "pipe:0",
			"-c:v", videoCodec,
			"-preset", "ultrafast",
			"-c:a", audioCodec,
			"-movflags", "+frag_keyframe+empty_moov+faststart",
			"-f", download.Extension,
			"-threads", "0",
			"pipe:1",
		}
	}

	log.Printf("Starting yt-dlp with args: %v", ytdlpArgs)
	ytdlpCmd := exec.Command("yt-dlp", ytdlpArgs...)
	ytdlpStdout, err := ytdlpCmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to create yt-dlp stdout pipe: %v", err)
		http.Error(w, "Failed to start download", http.StatusInternalServerError)
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return
	}

	log.Printf("Starting ffmpeg with args: %v", ffmpegArgs)
	ffmpegCmd := exec.Command(ffmpegPath+"\\ffmpeg.exe", ffmpegArgs...)
	ffmpegCmd.Stdin = ytdlpStdout
	ffmpegStdout, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to create ffmpeg stdout pipe: %v", err)
		http.Error(w, "Failed to start conversion", http.StatusInternalServerError)
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return
	}

	if err := ytdlpCmd.Start(); err != nil {
		log.Printf("Failed to start yt-dlp: %v", err)
		http.Error(w, "Failed to start download", http.StatusInternalServerError)
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return
	}

	if err := ffmpegCmd.Start(); err != nil {
		log.Printf("Failed to start ffmpeg: %v", err)
		http.Error(w, "Failed to start conversion", http.StatusInternalServerError)
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return
	}

	// Set headers for streaming
	w.Header().Set("Content-Type", getContentType(download.Extension))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.%s\"", download.Title, download.Extension))
	w.Header().Set("Transfer-Encoding", "chunked")

	log.Printf("Starting stream for download ID %s", id)

	// Stream the output directly to the response
	bytesCopied, err := io.Copy(w, ffmpegStdout)
	if err != nil {
		log.Printf("Error streaming file: %v", err)
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return
	}

	// Wait for commands to complete
	ytdlpCmd.Wait()
	ffmpegCmd.Wait()

	// Update database
	now := time.Now()
	db.Model(&Download{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "completed",
		"completed_at": now,
	})

	log.Printf("Streamed %d bytes for download ID %s", bytesCopied, id)
}

func buildVideoDownloadArgs(url, quality, platform string) []string {
	if platform == "tiktok" {
		// TikTok specific settings
		return []string{
			"--no-warnings",
			"-f", "best",
			"--buffer-size", "16K",
			"-o", "-",
			url,
		}
	}

	// YouTube settings
	formatStr := buildVideoFormat(quality)
	return []string{
		"--no-playlist",
		"-f", formatStr,
		"--no-check-certificate",
		"--buffer-size", "16K",
		"-o", "-",
		url,
	}
}

func buildAudioDownloadArgs(url, platform string) []string {
	if platform == "tiktok" {
		return []string{
			"--no-warnings",
			"-f", "bestaudio",
			"--buffer-size", "16K",
			"-o", "-",
			url,
		}
	}

	return []string{
		"--no-playlist",
		"-f", "bestaudio",
		"--no-check-certificate",
		"--buffer-size", "16K",
		"-o", "-",
		url,
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

func sanitizeFilename(name string) string {
	reg := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	name = reg.ReplaceAllString(name, "")
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	name = regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(name, "")
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}

func getVideoDuration(url string) float64 {
	cmd := exec.Command("yt-dlp",
		"--get-duration",
		"--no-playlist",
		url,
	)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	duration, _ := strconv.ParseFloat(string(output[:len(output)-1]), 64)
	return duration
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
	case "aac", "m4a":
		return "audio/aac"
	case "opus":
		return "audio/opus"
	default:
		return "application/octet-stream"
	}
}
