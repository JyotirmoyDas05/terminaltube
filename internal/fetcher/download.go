package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Downloader handles downloading media from URLs
type Downloader struct {
	client    *http.Client
	userAgent string
}

// NewDownloader creates a new downloader
func NewDownloader() *Downloader {
	return &Downloader{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "TerminalTube/1.0",
	}
}

// SetTimeout sets the HTTP client timeout
func (d *Downloader) SetTimeout(timeout time.Duration) {
	d.client.Timeout = timeout
}

// SetUserAgent sets the user agent string
func (d *Downloader) SetUserAgent(userAgent string) {
	d.userAgent = userAgent
}

// IsValidURL checks if the provided string is a valid URL
func (d *Downloader) IsValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	return u.Scheme == "http" || u.Scheme == "https"
}

// IsYouTubeURL checks if the URL is a YouTube video
func (d *Downloader) IsYouTubeURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	hostname := strings.ToLower(u.Hostname())
	return hostname == "www.youtube.com" ||
		hostname == "youtube.com" ||
		hostname == "youtu.be" ||
		hostname == "m.youtube.com"
}

// ExtractYouTubeVideoID extracts video ID from YouTube URL
func (d *Downloader) ExtractYouTubeVideoID(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	hostname := strings.ToLower(u.Hostname())

	// Handle youtu.be format
	if hostname == "youtu.be" {
		return strings.TrimPrefix(u.Path, "/")
	}

	// Handle youtube.com format
	if strings.Contains(hostname, "youtube.com") {
		query := u.Query()
		return query.Get("v")
	}

	return ""
}

// GetContentType returns the content type of the URL without downloading
func (d *Downloader) GetContentType(urlStr string) (string, int64, error) {
	req, err := http.NewRequest("HEAD", urlStr, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", d.userAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to make HEAD request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	contentLength := int64(0)

	if lengthStr := resp.Header.Get("Content-Length"); lengthStr != "" {
		if length, err := strconv.ParseInt(lengthStr, 10, 64); err == nil {
			contentLength = length
		}
	}

	return contentType, contentLength, nil
}

// ProgressCallback is called during download to report progress
type ProgressCallback func(downloaded, total int64, percent float64)

// DownloadToFile downloads a URL to a local file with progress reporting
func (d *Downloader) DownloadToFile(urlStr, filename string, progress ProgressCallback) error {
	// Create request
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", d.userAgent)

	// Make request
	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Get content length for progress reporting
	contentLength := int64(0)
	if lengthStr := resp.Header.Get("Content-Length"); lengthStr != "" {
		if length, err := strconv.ParseInt(lengthStr, 10, 64); err == nil {
			contentLength = length
		}
	}

	// Create output file
	outFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Create progress reader if callback provided
	var reader io.Reader = resp.Body
	if progress != nil {
		reader = &progressReader{
			Reader:   resp.Body,
			total:    contentLength,
			callback: progress,
		}
	}

	// Copy data
	_, err = io.Copy(outFile, reader)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// DownloadToTemp downloads a URL to a temporary file
func (d *Downloader) DownloadToTemp(urlStr string, progress ProgressCallback) (string, error) {
	// Extract filename from URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Get filename from URL path
	filename := filepath.Base(u.Path)
	if filename == "." || filename == "/" {
		filename = "download"
	}

	// Create temp file
	tempFile, err := os.CreateTemp("", "terminaltube_*_"+filename)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempFile.Close()

	tempPath := tempFile.Name()

	// Download to temp file
	err = d.DownloadToFile(urlStr, tempPath, progress)
	if err != nil {
		os.Remove(tempPath) // Clean up on error
		return "", err
	}

	return tempPath, nil
}

// GetFileExtensionFromURL tries to determine file extension from URL
func (d *Downloader) GetFileExtensionFromURL(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	ext := filepath.Ext(u.Path)
	if ext != "" {
		return ext
	}

	// Try to get from content type
	contentType, _, err := d.GetContentType(urlStr)
	if err != nil {
		return ""
	}

	// Map common content types to extensions
	contentTypeMap := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/bmp":  ".bmp",
		"video/mp4":  ".mp4",
		"video/avi":  ".avi",
		"video/mov":  ".mov",
		"video/mkv":  ".mkv",
		"video/webm": ".webm",
		"audio/mpeg": ".mp3",
		"audio/wav":  ".wav",
		"audio/ogg":  ".ogg",
		"audio/flac": ".flac",
	}

	if ext, exists := contentTypeMap[strings.ToLower(contentType)]; exists {
		return ext
	}

	return ""
}

// IsMediaURL checks if the URL points to a media file
func (d *Downloader) IsMediaURL(urlStr string) bool {
	// Check for YouTube URLs first
	if d.IsYouTubeURL(urlStr) {
		return true
	}

	ext := d.GetFileExtensionFromURL(urlStr)
	if ext == "" {
		return false
	}

	mediaExtensions := []string{
		".jpg", ".jpeg", ".png", ".gif", ".bmp",
		".mp4", ".avi", ".mov", ".mkv", ".webm", ".flv",
		".mp3", ".wav", ".ogg", ".flac",
	}

	ext = strings.ToLower(ext)
	for _, mediaExt := range mediaExtensions {
		if ext == mediaExt {
			return true
		}
	}

	return false
}

// GetYouTubeVideoInfo attempts to get video information from YouTube
func (d *Downloader) GetYouTubeVideoInfo(urlStr string) (map[string]string, error) {
	if !d.IsYouTubeURL(urlStr) {
		return nil, fmt.Errorf("not a YouTube URL")
	}

	videoID := d.ExtractYouTubeVideoID(urlStr)
	if videoID == "" {
		return nil, fmt.Errorf("could not extract video ID")
	}

	info := make(map[string]string)
	info["video_id"] = videoID
	info["url"] = urlStr
	info["title"] = "YouTube Video " + videoID
	info["notice"] = "YouTube videos require yt-dlp or similar tool for direct streaming"

	return info, nil
}

// DownloadYouTubeVideo uses yt-dlp to get a direct stream URL or download the video
func (d *Downloader) DownloadYouTubeVideo(urlStr string) (string, error) {
	if !d.IsYouTubeURL(urlStr) {
		return "", fmt.Errorf("not a YouTube URL")
	}

	videoID := d.ExtractYouTubeVideoID(urlStr)
	if videoID == "" {
		return "", fmt.Errorf("could not extract video ID from URL")
	}

	// Try to use yt-dlp to get direct stream URL
	streamURL, err := d.getYTDLPStreamURL(urlStr)
	if err == nil && streamURL != "" {
		return streamURL, nil
	}

	// If stream URL extraction failed, try downloading to temp file
	tempPath, err := d.downloadWithYTDLP(urlStr)
	if err != nil {
		// Provide helpful instructions if yt-dlp is not available
		instructions := fmt.Sprintf(`
YouTube video detected (ID: %s)

yt-dlp not found or failed. To enable YouTube support:

1. Install yt-dlp:
   Windows: winget install yt-dlp
   Or download from: https://github.com/yt-dlp/yt-dlp/releases

2. Restart TerminalTube and try again.

Alternative: Download manually and use option 4 (Play Video from File):
   yt-dlp -f "best[height<=1080]" "%s"
`, videoID, urlStr)
		return "", fmt.Errorf("yt-dlp required: %s", instructions)
	}

	return tempPath, nil
}

// getYTDLPStreamURL uses yt-dlp to extract a direct stream URL
func (d *Downloader) getYTDLPStreamURL(urlStr string) (string, error) {
	// Try yt-dlp to get direct URL - prefer high FPS (60fps > 30fps), limit to 1080p for performance
	// Format: best quality up to 1080p with highest fps available
	cmd := exec.Command("yt-dlp", "-f", "best[height<=1080][fps>=30]/best[height<=1080]/best", "--get-url", "--no-warnings", urlStr)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w", err)
	}

	streamURL := strings.TrimSpace(string(output))
	if streamURL == "" {
		return "", fmt.Errorf("no stream URL returned")
	}

	return streamURL, nil
}

// downloadWithYTDLP downloads video using yt-dlp to a temp file
func (d *Downloader) downloadWithYTDLP(urlStr string) (string, error) {
	// Create temp file
	tempFile, err := os.CreateTemp("", "terminaltube_yt_*.mp4")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()

	// Download with yt-dlp (1080p max for performance, mp4 format)
	cmd := exec.Command("yt-dlp",
		"-f", "best[height<=1080][ext=mp4]/best[height<=1080]/best",
		"-o", tempPath,
		"--no-warnings",
		"--progress",
		urlStr)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Downloading with yt-dlp...")
	err = cmd.Run()
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("yt-dlp download failed: %w", err)
	}

	return tempPath, nil
}

// progressReader wraps an io.Reader to provide download progress
type progressReader struct {
	io.Reader
	downloaded int64
	total      int64
	callback   ProgressCallback
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.downloaded += int64(n)

	if pr.callback != nil {
		percent := 0.0
		if pr.total > 0 {
			percent = float64(pr.downloaded) / float64(pr.total) * 100.0
		}
		pr.callback(pr.downloaded, pr.total, percent)
	}

	return n, err
}

// Close cleans up the downloader
func (d *Downloader) Close() {
	// Close any open connections
	d.client.CloseIdleConnections()
}
