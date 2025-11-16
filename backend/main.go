package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
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
	FilePath    string     `json:"file_path"`
	FileSize    int64      `json:"file_size"`
	Duration    float64    `json:"duration"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type DownloadRequest struct {
	URL       string `json:"url"`
	Format    string `json:"format"`    // video, audio
	Quality   string `json:"quality"`   // 144p, 240p, 360p, 480p, 720p, 1080p, 1440p, 2160p, best (video) | 64k, 128k, 192k, 256k, 320k, best (audio)
	Extension string `json:"extension"` // mp4, webm, mkv (video) | mp3, aac, opus, m4a (audio)
}

type FormatInfo struct {
	VideoQualities []string `json:"video_qualities"`
	AudioQualities []string `json:"audio_qualities"`
	VideoFormats   []string `json:"video_formats"`
	AudioFormats   []string `json:"audio_formats"`
}

var db *gorm.DB
var ffmpegPath = "C:\\ffmpeg-8.0-essentials_build\\bin" // Change this to your FFmpeg path

func main() {
	// Database connection with GORM
	dsn := "host=localhost port=5432 user=postgres password=root dbname=postgres sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate
	db.AutoMigrate(&Download{})

	router := mux.NewRouter()
	router.HandleFunc("/api/formats", handleGetFormats).Methods("GET")
	router.HandleFunc("/api/download", handleDownload).Methods("POST")
	router.HandleFunc("/api/downloads", handleGetDownloads).Methods("GET")
	router.HandleFunc("/api/downloads/{id}", handleGetDownload).Methods("GET")
	router.HandleFunc("/api/stream/{id}", handleStreamFile).Methods("GET")

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	})

	handler := c.Handler(router)
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func handleGetFormats(w http.ResponseWriter, r *http.Request) {
	formats := FormatInfo{
		VideoQualities: []string{"144p", "240p", "360p", "480p", "720p", "1080p", "1440p", "2160p", "best"},
		AudioQualities: []string{"64k", "128k", "192k", "256k", "320k", "best"},
		VideoFormats:   []string{"mp4", "webm", "mkv"},
		AudioFormats:   []string{"mp3", "aac", "opus", "m4a"},
	}
	json.NewEncoder(w).Encode(formats)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Quality == "" {
		if req.Format == "audio" {
			req.Quality = "best"
		} else {
			req.Quality = "best"
		}
	}
	if req.Extension == "" {
		if req.Format == "audio" {
			req.Extension = "mp3"
		} else {
			req.Extension = "mp4"
		}
	}

	// Create download record
	download := Download{
		URL:       req.URL,
		Format:    req.Format,
		Quality:   req.Quality,
		Extension: req.Extension,
		Status:    "pending",
	}

	if err := db.Create(&download).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Start download in goroutine
	go processDownload(download.ID, req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      download.ID,
		"message": "Download started",
	})
}

func processDownload(id uint, req DownloadRequest) {
	// Update status to processing
	db.Model(&Download{}).Where("id = ?", id).Update("status", "processing")

	// Build yt-dlp command to stream to FFmpeg
	var ytdlpArgs []string
	var ffmpegArgs []string

	if req.Format == "audio" {
		// Audio download with quality
		ytdlpArgs = []string{
			"--no-playlist",
			"-f", "bestaudio",
			"-o", "-", // Output to stdout
			req.URL,
		}

		// FFmpeg args for audio conversion
		audioCodec := getAudioCodec(req.Extension)
		audioBitrate := req.Quality
		if audioBitrate == "best" {
			audioBitrate = "320k"
		}

		ffmpegArgs = []string{
			"-i", "pipe:0",
			"-vn",
			"-acodec", audioCodec,
			"-b:a", audioBitrate,
			"-f", req.Extension,
			"pipe:1",
		}
	} else {
		// Video download with quality
		formatStr := buildVideoFormat(req.Quality)
		ytdlpArgs = []string{
			"--no-playlist",
			"-f", formatStr,
			"-o", "-", // Output to stdout
			req.URL,
		}

		// FFmpeg args for video conversion
		videoCodec, audioCodec := getVideoCodecs(req.Extension)
		ffmpegArgs = []string{
			"-i", "pipe:0",
			"-c:v", videoCodec,
			"-c:a", audioCodec,
			"-movflags", "+faststart",
			"-f", req.Extension,
			"pipe:1",
		}
	}

	// Get video info for title
	titleCmd := exec.Command("yt-dlp", "--get-title", "--no-playlist", req.URL)
	titleOutput, _ := titleCmd.Output()
	title := sanitizeFilename(string(titleOutput))
	if title == "" {
		title = fmt.Sprintf("download_%d", id)
	}

	// Create output file
	outputPath := fmt.Sprintf("downloads/%d_%s.%s", id, title, req.Extension)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Printf("Failed to create output file: %v", err)
		db.Model(&Download{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status": "failed",
			"title":  title,
		})
		return
	}
	defer outputFile.Close()

	// Create yt-dlp command
	ytdlpCmd := exec.Command("yt-dlp", ytdlpArgs...)
	ytdlpStdout, err := ytdlpCmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to create yt-dlp stdout pipe: %v", err)
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return
	}

	// Create FFmpeg command
	ffmpegCmd := exec.Command(ffmpegPath+"\\ffmpeg.exe", ffmpegArgs...)
	ffmpegCmd.Stdin = ytdlpStdout
	ffmpegStdout, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to create ffmpeg stdout pipe: %v", err)
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return
	}

	// Start both commands
	if err := ytdlpCmd.Start(); err != nil {
		log.Printf("Failed to start yt-dlp: %v", err)
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return
	}

	if err := ffmpegCmd.Start(); err != nil {
		log.Printf("Failed to start ffmpeg: %v", err)
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return
	}

	// Copy FFmpeg output to file
	fileSize, err := io.Copy(outputFile, ffmpegStdout)
	if err != nil {
		log.Printf("Failed to copy output: %v", err)
		db.Model(&Download{}).Where("id = ?", id).Update("status", "failed")
		return
	}

	// Wait for commands to complete
	ytdlpCmd.Wait()
	ffmpegCmd.Wait()

	// Get duration
	duration := getVideoDuration(outputPath)

	// Update database
	now := time.Now()
	db.Model(&Download{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "completed",
		"file_path":    outputPath,
		"title":        title,
		"file_size":    fileSize,
		"duration":     duration,
		"completed_at": now,
	})
}

func buildVideoFormat(quality string) string {
	if quality == "best" {
		return "bestvideo+bestaudio/best"
	}
	height := quality[:len(quality)-1] // Remove 'p'
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
	// Remove invalid characters
	reg := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	name = reg.ReplaceAllString(name, "")
	// Trim spaces and newlines
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	name = regexp.MustCompile(`^\s+|\s+$`).ReplaceAllString(name, "")
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}

func getVideoDuration(filePath string) float64 {
	cmd := exec.Command(ffmpegPath+"\\ffprobe.exe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		filePath,
	)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	duration, _ := strconv.ParseFloat(string(output[:len(output)-1]), 64)
	return duration
}

func handleGetDownloads(w http.ResponseWriter, r *http.Request) {
	var downloads []Download
	db.Order("created_at DESC").Find(&downloads)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(downloads)
}

func handleGetDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var download Download
	if err := db.First(&download, id).Error; err != nil {
		http.Error(w, "Download not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(download)
}

func handleStreamFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var download Download
	if err := db.Where("id = ? AND status = ?", id, "completed").First(&download).Error; err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Set headers for download
	w.Header().Set("Content-Type", getContentType(download.Extension))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.%s\"", download.Title, download.Extension))
	
	http.ServeFile(w, r, download.FilePath)
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