package decoder

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"terminaltube/pkg/types"
	"time"
)

// VideoDecoder handles video decoding using FFmpeg
type VideoDecoder struct {
	filename     string
	width        int
	height       int
	renderWidth  int // Target render width for scaling
	renderHeight int // Target render height for scaling
	fps          float64
	frameCount   int
	hasAudio     bool
	audioCodec   string
	videoCodec   string
	duration     float64
	currentFrame int
	ffmpegCmd    *exec.Cmd
	frameReader  io.ReadCloser
	stopChan     chan struct{}
}

// FFProbeOutput represents the JSON output from ffprobe
type FFProbeOutput struct {
	Streams []FFProbeStream `json:"streams"`
	Format  FFProbeFormat   `json:"format"`
}

// FFProbeStream represents a stream in ffprobe output
type FFProbeStream struct {
	CodecType    string `json:"codec_type"`
	CodecName    string `json:"codec_name"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	RFrameRate   string `json:"r_frame_rate"`
	AvgFrameRate string `json:"avg_frame_rate"`
	NbFrames     string `json:"nb_frames"`
	Duration     string `json:"duration"`
}

// FFProbeFormat represents format info in ffprobe output
type FFProbeFormat struct {
	Duration string `json:"duration"`
	BitRate  string `json:"bit_rate"`
}

// NewVideoDecoder creates a new video decoder
func NewVideoDecoder() *VideoDecoder {
	return &VideoDecoder{
		stopChan: make(chan struct{}),
	}
}

// IsSupported checks if the file is a supported video format
// Uses extension check first, then falls back to FFmpeg probe for temp files
func (d *VideoDecoder) IsSupported(filename string) bool {
	supportedExts := []string{".mp4", ".avi", ".mov", ".mkv", ".webm", ".flv", ".wmv", ".m4v"}
	lowerName := strings.ToLower(filename)

	// Check known extensions first
	for _, ext := range supportedExts {
		if strings.HasSuffix(lowerName, ext) {
			return true
		}
	}

	// For temp files or unknown extensions, try FFmpeg probe
	// This handles downloaded YouTube videos with unusual names
	if strings.Contains(lowerName, "terminaltube") || strings.Contains(lowerName, "temp") {
		cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries", "stream=codec_type", "-of", "csv=p=0", filename)
		output, err := cmd.Output()
		if err == nil && strings.Contains(string(output), "video") {
			return true
		}
	}

	return false
}

// checkFFmpeg verifies that ffmpeg and ffprobe are available
func (d *VideoDecoder) checkFFmpeg() error {
	// Check ffprobe
	cmd := exec.Command("ffprobe", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffprobe not found. Please install FFmpeg and add it to your PATH")
	}

	// Check ffmpeg
	cmd = exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg not found. Please install FFmpeg and add it to your PATH")
	}

	return nil
}

// LoadVideo loads a video file and extracts metadata using ffprobe
func (d *VideoDecoder) LoadVideo(filename string) (*types.MediaInfo, error) {
	// Check FFmpeg availability
	if err := d.checkFFmpeg(); err != nil {
		return nil, err
	}

	d.filename = filename

	// Use ffprobe to get video information
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filename,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	var probeOutput FFProbeOutput
	if err := json.Unmarshal(output, &probeOutput); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Extract video and audio stream info
	for _, stream := range probeOutput.Streams {
		switch stream.CodecType {
		case "video":
			d.width = stream.Width
			d.height = stream.Height
			d.videoCodec = stream.CodecName

			// Parse frame rate (format: "num/den")
			d.fps = parseFrameRate(stream.RFrameRate)
			if d.fps == 0 {
				d.fps = parseFrameRate(stream.AvgFrameRate)
			}
			if d.fps == 0 {
				d.fps = 30.0 // Default fallback
			}

			// Parse frame count
			if stream.NbFrames != "" {
				d.frameCount, _ = strconv.Atoi(stream.NbFrames)
			}

			// Parse duration from stream
			if stream.Duration != "" {
				d.duration, _ = strconv.ParseFloat(stream.Duration, 64)
			}

		case "audio":
			d.hasAudio = true
			d.audioCodec = stream.CodecName
		}
	}

	// Get duration from format if not set
	if d.duration == 0 && probeOutput.Format.Duration != "" {
		d.duration, _ = strconv.ParseFloat(probeOutput.Format.Duration, 64)
	}

	// Estimate frame count if not available
	if d.frameCount == 0 && d.duration > 0 && d.fps > 0 {
		d.frameCount = int(d.duration * d.fps)
	}

	d.currentFrame = 0

	mediaInfo := &types.MediaInfo{
		Type:       types.VIDEO,
		Width:      d.width,
		Height:     d.height,
		FPS:        d.fps,
		Duration:   d.duration,
		HasAudio:   d.hasAudio,
		AudioCodec: d.audioCodec,
		VideoCodec: d.videoCodec,
		FrameCount: d.frameCount,
	}

	return mediaInfo, nil
}

// SetRenderSize sets the target render size for scaled frame extraction
func (d *VideoDecoder) SetRenderSize(width, height int) {
	d.renderWidth = width
	d.renderHeight = height
}

// parseFrameRate parses a frame rate string like "30/1" or "29.97"
func parseFrameRate(rateStr string) float64 {
	if rateStr == "" || rateStr == "0/0" {
		return 0
	}

	// Try parsing as fraction (e.g., "30/1", "30000/1001")
	if strings.Contains(rateStr, "/") {
		parts := strings.Split(rateStr, "/")
		if len(parts) == 2 {
			num, err1 := strconv.ParseFloat(parts[0], 64)
			den, err2 := strconv.ParseFloat(parts[1], 64)
			if err1 == nil && err2 == nil && den != 0 {
				return num / den
			}
		}
	}

	// Try parsing as decimal
	rate, err := strconv.ParseFloat(rateStr, 64)
	if err == nil {
		return rate
	}

	return 0
}

// GetFrame returns a specific frame from the video
func (d *VideoDecoder) GetFrame(frameIndex int) (*types.Frame, error) {
	if frameIndex < 0 || (d.frameCount > 0 && frameIndex >= d.frameCount) {
		return nil, fmt.Errorf("frame index out of range: %d", frameIndex)
	}

	// Calculate timestamp for this frame
	timestamp := float64(frameIndex) / d.fps

	// Determine output dimensions
	outWidth := d.width
	outHeight := d.height
	if d.renderWidth > 0 && d.renderHeight > 0 {
		outWidth = d.renderWidth
		outHeight = d.renderHeight
	}

	// Use ffmpeg to extract a single frame with scaling
	cmd := exec.Command("ffmpeg",
		"-ss", fmt.Sprintf("%.3f", timestamp),
		"-i", d.filename,
		"-vframes", "1",
		"-vf", fmt.Sprintf("scale=%d:%d", outWidth, outHeight),
		"-f", "rawvideo",
		"-pix_fmt", "rgb24",
		"-v", "quiet",
		"pipe:1",
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to extract frame %d: %w", frameIndex, err)
	}

	// Convert raw RGB data to image
	img := d.rawRGBToImage(output, outWidth, outHeight)
	if img == nil {
		return nil, fmt.Errorf("failed to decode frame data")
	}

	frameDuration := 1.0 / d.fps

	frame := &types.Frame{
		Image:     img,
		Timestamp: timestamp,
		Duration:  frameDuration,
	}

	return frame, nil
}

// rawRGBToImage converts raw RGB24 bytes to an image.RGBA
func (d *VideoDecoder) rawRGBToImage(data []byte, width, height int) image.Image {
	expectedSize := width * height * 3
	if len(data) < expectedSize {
		return nil
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := (y*width + x) * 3
			if i+2 < len(data) {
				img.Set(x, y, color.RGBA{
					R: data[i],
					G: data[i+1],
					B: data[i+2],
					A: 255,
				})
			}
		}
	}

	return img
}

// GetFrameChannel returns a channel that yields video frames with proper timing
func (d *VideoDecoder) GetFrameChannel() (<-chan *types.Frame, error) {
	frameChan := make(chan *types.Frame, 5) // Larger buffer for smoother playback

	// Determine output dimensions - scale down for performance
	outWidth := d.renderWidth
	outHeight := d.renderHeight
	if outWidth <= 0 || outHeight <= 0 {
		// Default to smaller size for performance
		outWidth = 160
		outHeight = 90
	}

	// Start ffmpeg process to stream scaled frames
	// Using -an to ignore audio, scale filter with lanczos for high quality
	cmd := exec.Command("ffmpeg",
		"-i", d.filename,
		"-an", // No audio
		"-vf", fmt.Sprintf("scale=%d:%d:flags=lanczos", outWidth, outHeight),
		"-f", "rawvideo",
		"-pix_fmt", "rgb24",
		"-v", "quiet",
		"pipe:1",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	d.ffmpegCmd = cmd
	d.frameReader = stdout
	d.stopChan = make(chan struct{})

	go func() {
		defer close(frameChan)
		defer cmd.Wait()

		frameSize := outWidth * outHeight * 3
		frameBuffer := make([]byte, frameSize)
		reader := bufio.NewReaderSize(stdout, frameSize*4)

		frameDuration := time.Duration(float64(time.Second) / d.fps)
		startTime := time.Now()
		frameIndex := 0

		for {
			select {
			case <-d.stopChan:
				return
			default:
			}

			// Read one frame of raw RGB data
			n, err := io.ReadFull(reader, frameBuffer)
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					return // End of video
				}
				return
			}

			if n != frameSize {
				continue // Skip incomplete frames
			}

			// Convert to image
			img := d.rawRGBToImage(frameBuffer, outWidth, outHeight)
			if img == nil {
				continue
			}

			timestamp := float64(frameIndex) / d.fps
			frame := &types.Frame{
				Image:     img,
				Timestamp: timestamp,
				Duration:  1.0 / d.fps,
			}

			// Calculate when this frame should be displayed
			targetTime := startTime.Add(time.Duration(frameIndex) * frameDuration)
			now := time.Now()

			// If we're ahead of schedule, wait
			if now.Before(targetTime) {
				time.Sleep(targetTime.Sub(now))
			}

			// Send frame without blocking for too long
			select {
			case frameChan <- frame:
				frameIndex++
			case <-d.stopChan:
				return
			default:
				// Channel full, skip frame and move on
				frameIndex++
			}
		}
	}()

	return frameChan, nil
}

// Seek moves to a specific time position in the video
func (d *VideoDecoder) Seek(timestamp float64) error {
	if timestamp < 0 || timestamp > d.duration {
		return fmt.Errorf("timestamp out of range: %f", timestamp)
	}

	frameIndex := int(timestamp * d.fps)
	d.currentFrame = frameIndex

	return nil
}

// GetCurrentPosition returns the current playback position in seconds
func (d *VideoDecoder) GetCurrentPosition() float64 {
	return float64(d.currentFrame) / d.fps
}

// Close cleans up the decoder
func (d *VideoDecoder) Close() error {
	// Signal stop
	if d.stopChan != nil {
		select {
		case <-d.stopChan:
			// Already closed
		default:
			close(d.stopChan)
		}
	}

	// Kill ffmpeg process if running
	if d.ffmpegCmd != nil && d.ffmpegCmd.Process != nil {
		d.ffmpegCmd.Process.Kill()
		d.ffmpegCmd.Wait()
	}

	// Close frame reader
	if d.frameReader != nil {
		d.frameReader.Close()
	}

	d.filename = ""
	d.currentFrame = 0
	return nil
}

// GetVideoInfo returns detailed video information
func (d *VideoDecoder) GetVideoInfo() map[string]interface{} {
	info := make(map[string]interface{})
	info["filename"] = d.filename
	info["width"] = d.width
	info["height"] = d.height
	info["fps"] = d.fps
	info["frame_count"] = d.frameCount
	info["duration"] = d.duration
	info["has_audio"] = d.hasAudio
	info["audio_codec"] = d.audioCodec
	info["video_codec"] = d.videoCodec
	info["current_frame"] = d.currentFrame

	return info
}
