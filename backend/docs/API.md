# API Documentation

Complete API reference for Universal Downloader Pro backend.

## Base URL

```
http://localhost:8081/api
```

## Table of Contents

1. [Authentication](#authentication)
2. [Endpoints](#endpoints)
3. [Data Models](#data-models)
4. [Error Handling](#error-handling)
5. [Examples](#examples)

## Authentication

Currently, the API does not require authentication. All endpoints are publicly accessible.

## Endpoints

### 1. Health Check

Check if the API is running.

**Endpoint:** `GET /`

**Response:**
```json
{
  "status": "ok",
  "message": "YouTube & TikTok Downloader API is running",
  "version": "2.0 (Fiber + Optimized Streaming)"
}
```

**Example:**
```bash
curl http://localhost:8081/
```

---

### 2. Get Available Formats

Retrieve all available quality options and formats.

**Endpoint:** `GET /api/formats`

**Response:**
```json
{
  "video_qualities": [
    "144p",
    "240p",
    "360p",
    "480p",
    "720p",
    "1080p",
    "1440p",
    "2160p",
    "best"
  ],
  "audio_qualities": [
    "64k",
    "128k",
    "192k",
    "256k",
    "320k",
    "best"
  ],
  "video_formats": [
    "mp4",
    "webm",
    "mkv"
  ],
  "audio_formats": [
    "mp3",
    "opus",
    "m4a"
  ]
}
```

**Example:**
```bash
curl http://localhost:8081/api/formats
```

---

### 3. Create Download

Initialize a new download task.

**Endpoint:** `POST /api/download`

**Request Body:**
```json
{
  "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
  "format": "video",
  "quality": "720p",
  "extension": "mp4"
}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| url | string | Yes | YouTube or TikTok URL |
| format | string | Yes | "video" or "audio" |
| quality | string | No | Quality setting (default: "best") |
| extension | string | No | Output format (default: "mp4" for video, "mp3" for audio) |

**Response:**
```json
{
  "id": 1,
  "message": "Download ready, use /api/stream/1 to download",
  "title": "Rick Astley - Never Gonna Give You Up",
  "platform": "youtube"
}
```

**Status Codes:**
- `200 OK` - Download created successfully
- `400 Bad Request` - Invalid request body or missing URL
- `500 Internal Server Error` - Server error

**Examples:**

```bash
# Video download
curl -X POST http://localhost:8081/api/download \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "format": "video",
    "quality": "1080p",
    "extension": "mp4"
  }'

# Audio download
curl -X POST http://localhost:8081/api/download \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "format": "audio",
    "quality": "320k",
    "extension": "mp3"
  }'

# TikTok download
curl -X POST http://localhost:8081/api/download \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.tiktok.com/@user/video/1234567890",
    "format": "video",
    "quality": "best",
    "extension": "mp4"
  }'
```

---

### 4. Get All Downloads

Retrieve the complete download history.

**Endpoint:** `GET /api/downloads`

**Response:**
```json
[
  {
    "id": 1,
    "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "title": "Rick Astley - Never Gonna Give You Up",
    "format": "video",
    "quality": "720p",
    "extension": "mp4",
    "status": "completed",
    "duration": 213.5,
    "platform": "youtube",
    "created_at": "2024-12-14T10:30:00Z",
    "completed_at": "2024-12-14T10:32:15Z"
  },
  {
    "id": 2,
    "url": "https://www.tiktok.com/@user/video/1234567890",
    "title": "TikTok Video",
    "format": "audio",
    "quality": "best",
    "extension": "mp3",
    "status": "ready",
    "duration": 0,
    "platform": "tiktok",
    "created_at": "2024-12-14T10:35:00Z",
    "completed_at": null
  }
]
```

**Example:**
```bash
curl http://localhost:8081/api/downloads
```

---

### 5. Get Single Download

Retrieve information about a specific download.

**Endpoint:** `GET /api/downloads/:id`

**Parameters:**
- `id` (path parameter) - Download ID

**Response:**
```json
{
  "id": 1,
  "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
  "title": "Rick Astley - Never Gonna Give You Up",
  "format": "video",
  "quality": "720p",
  "extension": "mp4",
  "status": "completed",
  "duration": 213.5,
  "platform": "youtube",
  "created_at": "2024-12-14T10:30:00Z",
  "completed_at": "2024-12-14T10:32:15Z"
}
```

**Status Codes:**
- `200 OK` - Download found
- `404 Not Found` - Download not found

**Example:**
```bash
curl http://localhost:8081/api/downloads/1
```

---

### 6. Stream Download File

Stream the processed media file directly to the client.

**Endpoint:** `GET /api/stream/:id`

**Parameters:**
- `id` (path parameter) - Download ID

**Response:**
- Binary stream of the media file
- Headers:
  - `Content-Type`: Appropriate MIME type (video/mp4, audio/mpeg, etc.)
  - `Content-Disposition`: `attachment; filename="title.ext"`
  - `Transfer-Encoding`: chunked
  - `Cache-Control`: no-cache

**Status Codes:**
- `200 OK` - Streaming started successfully
- `404 Not Found` - Download not found
- `500 Internal Server Error` - Streaming failed

**Process Flow:**
1. Updates download status to "streaming"
2. Starts yt-dlp to download from platform
3. Pipes output through FFmpeg for conversion
4. Streams converted output directly to client
5. Updates status to "completed" when finished

**Example:**
```bash
# Download directly
curl http://localhost:8081/api/stream/1 -o output.mp4

# Using wget
wget http://localhost:8081/api/stream/1 -O output.mp4
```

**Browser Usage:**
```html
<a href="http://localhost:8081/api/stream/1" download>Download File</a>
```

**JavaScript Usage:**
```javascript
async function downloadFile(id, title, extension) {
  const response = await fetch(`http://localhost:8081/api/stream/${id}`);
  const blob = await response.blob();
  
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `${title}.${extension}`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  window.URL.revokeObjectURL(url);
}
```

---

## Data Models

### Download

```typescript
{
  id: number;              // Unique identifier
  url: string;             // Source URL (YouTube or TikTok)
  title: string;           // Video/audio title
  format: string;          // "video" or "audio"
  quality: string;         // Quality setting (e.g., "720p", "320k")
  extension: string;       // Output format (e.g., "mp4", "mp3")
  status: string;          // "ready", "streaming", "completed", "failed"
  duration: number;        // Duration in seconds (0 if not available)
  platform: string;        // "youtube" or "tiktok"
  created_at: string;      // ISO 8601 timestamp
  completed_at?: string;   // ISO 8601 timestamp (null if not completed)
}
```

### FormatInfo

```typescript
{
  video_qualities: string[];   // Available video quality options
  audio_qualities: string[];   // Available audio quality options
  video_formats: string[];     // Available video format options
  audio_formats: string[];     // Available audio format options
}
```

### DownloadRequest

```typescript
{
  url: string;          // Required: Media URL
  format: string;       // Required: "video" or "audio"
  quality?: string;     // Optional: Quality setting (default: "best")
  extension?: string;   // Optional: Output format (default: "mp4" or "mp3")
}
```

---

## Error Handling

### Error Response Format

```json
{
  "error": "Error message description"
}
```

### Common Error Codes

| Status Code | Description |
|-------------|-------------|
| 400 | Bad Request - Invalid parameters or missing required fields |
| 404 | Not Found - Resource doesn't exist |
| 500 | Internal Server Error - Server-side error |

### Error Examples

**Invalid URL:**
```json
{
  "error": "URL is required"
}
```

**Download Not Found:**
```json
{
  "error": "Download not found"
}
```

**Invalid Request Body:**
```json
{
  "error": "Invalid request body"
}
```

**Database Error:**
```json
{
  "error": "Failed to create download record"
}
```

---

## Examples

### Complete Workflow Example

#### 1. Check Available Formats
```bash
curl http://localhost:8081/api/formats
```

#### 2. Create a Video Download
```bash
curl -X POST http://localhost:8081/api/download \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "format": "video",
    "quality": "1080p",
    "extension": "mp4"
  }'
```

Response:
```json
{
  "id": 1,
  "message": "Download ready, use /api/stream/1 to download",
  "title": "Rick Astley - Never Gonna Give You Up",
  "platform": "youtube"
}
```

#### 3. Check Download Status
```bash
curl http://localhost:8081/api/downloads/1
```

#### 4. Stream the File
```bash
curl http://localhost:8081/api/stream/1 -o "rick-astley.mp4"
```

#### 5. View All Downloads
```bash
curl http://localhost:8081/api/downloads
```

---

### Python Example

```python
import requests

# Base URL
API_URL = "http://localhost:8081/api"

# Create a download
response = requests.post(f"{API_URL}/download", json={
    "url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
    "format": "audio",
    "quality": "320k",
    "extension": "mp3"
})

data = response.json()
download_id = data['id']
title = data['title']

print(f"Download created: {title} (ID: {download_id})")

# Stream and save file
stream_response = requests.get(f"{API_URL}/stream/{download_id}", stream=True)
with open(f"{title}.mp3", 'wb') as f:
    for chunk in stream_response.iter_content(chunk_size=8192):
        f.write(chunk)

print("Download completed!")
```

---

### Node.js Example

```javascript
const axios = require('axios');
const fs = require('fs');

const API_URL = 'http://localhost:8081/api';

async function downloadMedia() {
  try {
    // Create download
    const createResponse = await axios.post(`${API_URL}/download`, {
      url: 'https://www.youtube.com/watch?v=dQw4w9WgXcQ',
      format: 'video',
      quality: '720p',
      extension: 'mp4'
    });

    const { id, title } = createResponse.data;
    console.log(`Download created: ${title} (ID: ${id})`);

    // Stream file
    const streamResponse = await axios({
      method: 'get',
      url: `${API_URL}/stream/${id}`,
      responseType: 'stream'
    });

    const writer = fs.createWriteStream(`${title}.mp4`);
    streamResponse.data.pipe(writer);

    writer.on('finish', () => {
      console.log('Download completed!');
    });

  } catch (error) {
    console.error('Error:', error.message);
  }
}

downloadMedia();
```

---

## Rate Limiting

Currently, there is no rate limiting implemented. However, be aware that:

- yt-dlp may be rate-limited by YouTube/TikTok
- Concurrent downloads are limited by system resources
- Large files require stable connections

## Best Practices

1. **Always check available formats first** before creating downloads
2. **Use appropriate quality settings** to balance file size and quality
3. **Handle streaming errors** gracefully in your client application
4. **Monitor download status** before attempting to stream
5. **Use concurrent fragment downloads** for faster processing (already configured)
6. **Implement retry logic** for failed downloads
7. **Clean up completed downloads** periodically if storing large history

## Supported Platforms

### YouTube
- Standard videos
- Age-restricted content (if accessible)
- Live streams (after completion)
- Quality options: 144p to 2160p (4K)

### TikTok
- Standard videos
- Videos with music
- Quality: Best available (usually 720p or 1080p)

### Not Supported
- Private videos
- DRM-protected content
- Live streams (in progress)
- Playlists (currently)

## Performance Notes

- **Streaming is optimized** with 512KB buffers and buffer pooling
- **Concurrent fragments** speed up downloads (4 for video, 16 for audio)
- **No temporary files** are created on the server
- **Direct piping** from yt-dlp → FFmpeg → Client minimizes latency
- **FFmpeg presets** are optimized for streaming ("veryfast", "zerolatency")

## Troubleshooting API Issues

### Download Fails Immediately
- Check if URL is valid and accessible
- Verify yt-dlp is up to date
- Check FFmpeg installation

### Streaming Stops Midway
- Network connection issue
- Client closed connection
- Source video unavailable

### Slow Downloads
- Increase `concurrent-fragments` setting
- Check network bandwidth
- Consider using lower quality settings

### Database Errors
- Verify PostgreSQL is running
- Check connection string
- Ensure database migrations ran successfully