package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	_ "github.com/lib/pq"
)

type Download struct {
	ID          int       `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Format      string    `json:"format"`
	Status      string    `json:"status"`
	FilePath    string    `json:"file_path"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Duration    int       `json:"duration,omitempty"`
}

type DownloadRequest struct {
	URL    string `json:"url"`
	Format string `json:"format"` // "video" or "audio"
}

var db *sql.DB
// Database connection 
func main() {
	// Database connection
	connStr := "host=localhost port=5432 user=postgres password=root dbname=postgres sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table if not exists
	createTable()

	// Create downloads directory
	os.MkdirAll("downloads", 0755)

	router := mux.NewRouter()
	router.HandleFunc("/api/download", handleDownload).Methods("POST")
	router.HandleFunc("/api/downloads", handleGetDownloads).Methods("GET")
	router.HandleFunc("/api/downloads/{id}", handleGetDownload).Methods("GET")
	router.HandleFunc("/api/file/{id}", handleServeFile).Methods("GET")

	// CORS - Allow all origins in development
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

func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS downloads (
		id SERIAL PRIMARY KEY,
		url TEXT NOT NULL,
		title TEXT,
		format VARCHAR(10),
		status VARCHAR(20),
		file_path TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		completed_at TIMESTAMP
	)`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Insert into database
	var downloadID int
	err := db.QueryRow(
		"INSERT INTO downloads (url, format, status) VALUES ($1, $2, $3) RETURNING id",
		req.URL, req.Format, "pending",
	).Scan(&downloadID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Start download in goroutine
	go processDownload(downloadID, req.URL, req.Format)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      downloadID,
		"message": "Download started",
	})
}

func processDownload(id int, url, format string) {
	// Update status to processing
	db.Exec("UPDATE downloads SET status = $1 WHERE id = $2", "processing", id)

	outputDir := "downloads"
	outputTemplate := filepath.Join(outputDir, fmt.Sprintf("%d_%%(title)s.%%(ext)s", id))

	var cmd *exec.Cmd
	if format == "audio" {
		// Download audio only
		cmd = exec.Command("yt-dlp",
			"-x",
			"--audio-format", "mp3",
			"--audio-quality", "0",
			"-o", outputTemplate,
			url,
		)
	} else {
		// Download video
		cmd = exec.Command("yt-dlp",
			"-f", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best",
			"--merge-output-format", "mp4",
			"-o", outputTemplate,
			url,
		)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Download error: %v, Output: %s", err, string(output))
		db.Exec("UPDATE downloads SET status = $1 WHERE id = $2", "failed", id)
		return
	}

	// Get the downloaded file
	files, _ := filepath.Glob(filepath.Join(outputDir, fmt.Sprintf("%d_*", id)))
	if len(files) == 0 {
		db.Exec("UPDATE downloads SET status = $1 WHERE id = $2", "failed", id)
		return
	}

	filePath := files[0]
	fileName := filepath.Base(filePath)
	// Extract title from filename
	title := fileName[len(fmt.Sprintf("%d_", id)):len(fileName)-len(filepath.Ext(fileName))]

	now := time.Now()
	db.Exec(
		"UPDATE downloads SET status = $1, file_path = $2, title = $3, completed_at = $4 WHERE id = $5",
		"completed", filePath, title, now, id,
	)
}

func handleGetDownloads(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, url, title, format, status, file_path, created_at, completed_at FROM downloads ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var downloads []Download
	for rows.Next() {
		var d Download
		var title, filePath sql.NullString
		var completedAt sql.NullTime
		err := rows.Scan(&d.ID, &d.URL, &title, &d.Format, &d.Status, &filePath, &d.CreatedAt, &completedAt)
		if err != nil {
			continue
		}
		if title.Valid {
			d.Title = title.String
		}
		if filePath.Valid {
			d.FilePath = filePath.String
		}
		if completedAt.Valid {
			d.CompletedAt = &completedAt.Time
		}
		downloads = append(downloads, d)
	}

	json.NewEncoder(w).Encode(downloads)
}

func handleGetDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var d Download
	var title, filePath sql.NullString
	var completedAt sql.NullTime
	err := db.QueryRow(
		"SELECT id, url, title, format, status, file_path, created_at, completed_at FROM downloads WHERE id = $1",
		id,
	).Scan(&d.ID, &d.URL, &title, &d.Format, &d.Status, &filePath, &d.CreatedAt, &completedAt)

	if err != nil {
		http.Error(w, "Download not found", http.StatusNotFound)
		return
	}

	if title.Valid {
		d.Title = title.String
	}
	if filePath.Valid {
		d.FilePath = filePath.String
	}
	if completedAt.Valid {
		d.CompletedAt = &completedAt.Time
	}

	json.NewEncoder(w).Encode(d)
}

func handleServeFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var filePath string
	err := db.QueryRow("SELECT file_path FROM downloads WHERE id = $1 AND status = 'completed'", id).Scan(&filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}