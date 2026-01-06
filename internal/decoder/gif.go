package decoder

import (
	"fmt"
	"image/gif"
	"os"
	"terminaltube/pkg/types"
	"time"
)

// GIFDecoder handles animated GIF decoding
type GIFDecoder struct {
	currentGIF *gif.GIF
	filename   string
}

// NewGIFDecoder creates a new GIF decoder
func NewGIFDecoder() *GIFDecoder {
	return &GIFDecoder{}
}

// IsSupported checks if the file is a GIF
func (d *GIFDecoder) IsSupported(filename string) bool {
	// Try to decode as GIF to verify
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	_, err = gif.DecodeAll(file)
	return err == nil
}

// LoadGIF loads a GIF file for frame-by-frame playback
func (d *GIFDecoder) LoadGIF(filename string) (*types.MediaInfo, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open GIF file: %w", err)
	}
	defer file.Close()

	// Decode the entire GIF
	gifData, err := gif.DecodeAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode GIF: %w", err)
	}

	d.currentGIF = gifData
	d.filename = filename

	// Calculate total duration
	var totalDuration time.Duration
	for _, delay := range gifData.Delay {
		// GIF delays are in centiseconds (1/100 second)
		totalDuration += time.Duration(delay) * time.Millisecond * 10
	}

	// Get dimensions from first frame
	var width, height int
	if len(gifData.Image) > 0 {
		bounds := gifData.Image[0].Bounds()
		width = bounds.Dx()
		height = bounds.Dy()
	}

	// Calculate average FPS
	fps := 0.0
	if len(gifData.Delay) > 0 && totalDuration > 0 {
		fps = float64(len(gifData.Delay)) / totalDuration.Seconds()
	}

	// Create media info
	mediaInfo := &types.MediaInfo{
		Type:       types.GIF,
		Width:      width,
		Height:     height,
		FPS:        fps,
		Duration:   totalDuration.Seconds(),
		HasAudio:   false, // GIFs don't have audio
		AudioCodec: "",
		VideoCodec: "GIF",
		FrameCount: len(gifData.Image),
	}

	return mediaInfo, nil
}

// GetFrameCount returns the total number of frames
func (d *GIFDecoder) GetFrameCount() int {
	if d.currentGIF == nil {
		return 0
	}
	return len(d.currentGIF.Image)
}

// GetFrame returns a specific frame from the GIF
func (d *GIFDecoder) GetFrame(frameIndex int) (*types.Frame, error) {
	if d.currentGIF == nil {
		return nil, fmt.Errorf("no GIF loaded")
	}

	if frameIndex < 0 || frameIndex >= len(d.currentGIF.Image) {
		return nil, fmt.Errorf("frame index out of range: %d", frameIndex)
	}

	// Get the frame image
	frameImage := d.currentGIF.Image[frameIndex]

	// Calculate frame duration (GIF delays are in centiseconds)
	var frameDuration float64
	if frameIndex < len(d.currentGIF.Delay) {
		frameDuration = float64(d.currentGIF.Delay[frameIndex]) / 100.0 // Convert to seconds
	}

	// Calculate timestamp based on previous frame delays
	var timestamp float64
	for i := 0; i < frameIndex; i++ {
		if i < len(d.currentGIF.Delay) {
			timestamp += float64(d.currentGIF.Delay[i]) / 100.0
		}
	}

	frame := &types.Frame{
		Image:     frameImage,
		Timestamp: timestamp,
		Duration:  frameDuration,
	}

	return frame, nil
}

// GetFrameChannel returns a channel that yields frames with proper timing
func (d *GIFDecoder) GetFrameChannel() (<-chan *types.Frame, error) {
	if d.currentGIF == nil {
		return nil, fmt.Errorf("no GIF loaded")
	}

	frameChan := make(chan *types.Frame, 1)

	go func() {
		defer close(frameChan)

		for {
			for i := 0; i < len(d.currentGIF.Image); i++ {
				frame, err := d.GetFrame(i)
				if err != nil {
					return
				}

				frameChan <- frame

				// Wait for the frame duration
				if frame.Duration > 0 {
					time.Sleep(time.Duration(frame.Duration * float64(time.Second)))
				} else {
					// Default delay if no delay specified
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	}()

	return frameChan, nil
}

// Close cleans up the decoder
func (d *GIFDecoder) Close() {
	d.currentGIF = nil
	d.filename = ""
}
