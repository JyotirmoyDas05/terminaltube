package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"terminaltube/internal/audio"
	"terminaltube/internal/decoder"
	"terminaltube/internal/fetcher"
	"terminaltube/internal/renderer"
	"terminaltube/internal/terminal"
	"terminaltube/pkg/types"
	"time"
)

// calculateOptimalRenderSize calculates the best render dimensions for the current terminal
// For SIXEL mode, this returns character dimensions (will be multiplied by pixel ratio later)
func calculateOptimalRenderSize(imgWidth, imgHeight, termWidth, termHeight int) (int, int) {
	// Use full terminal width, minimal height margin for status
	maxWidth := termWidth
	maxHeight := termHeight - 2 // Leave 2 rows for any UI

	// Ensure minimum available space
	if maxWidth < 10 {
		maxWidth = 10
	}
	if maxHeight < 5 {
		maxHeight = 5
	}

	// Calculate image aspect ratio
	imgAspect := float64(imgWidth) / float64(imgHeight)

	// For terminal characters, they are ~2x tall as wide
	// But for SIXEL (pixel-based), we'll handle this in the pixel conversion
	// Here we just calculate character cell dimensions

	// Terminal character aspect ratio (height/width of a character cell)
	// A character is typically about 2x taller than wide
	charAspect := 2.0

	// Calculate render dimensions to fill available space while maintaining aspect
	// Try fitting to width first
	renderWidth := maxWidth
	renderHeight := int(float64(renderWidth) / imgAspect / charAspect)

	// If height exceeds max, fit to height instead
	if renderHeight > maxHeight {
		renderHeight = maxHeight
		renderWidth = int(float64(renderHeight) * imgAspect * charAspect)
	}

	// Ensure we don't exceed terminal bounds
	if renderWidth > maxWidth {
		renderWidth = maxWidth
	}
	if renderHeight > maxHeight {
		renderHeight = maxHeight
	}

	// Ensure minimum sizes
	if renderWidth < 10 {
		renderWidth = 10
	}
	if renderHeight < 5 {
		renderHeight = 5
	}

	return renderWidth, renderHeight
}

// calculateOptimalRenderSizeSixel calculates dimensions for SIXEL mode
// SIXEL renders in actual pixels - we fill the entire terminal viewport
// Video is stretched to fit (no aspect ratio preservation for maximum coverage)
func calculateOptimalRenderSizeSixel(imgWidth, imgHeight, termWidth, termHeight int) (int, int) {
	// Use full terminal dimensions
	// Width: 10 pixels per char for full width coverage
	// Height: 19 pixels per char to avoid exceeding viewport and causing scroll, got that happening during Windows Terminal Testing
	pixelWidth := termWidth * 10
	pixelHeight := termHeight * 19

	// Make height divisible by 6 (SIXEL requirement form docs)
	pixelHeight = (pixelHeight / 6) * 6

	// Return full terminal pixel dimensions
	return pixelWidth, pixelHeight
}

// getCurrentTerminalSize gets the current terminal size (for dynamic resizing)
func getCurrentTerminalSize() (int, int, error) {
	return terminal.GetTerminalSize()
}

const (
	appName    = "TerminalTube"
	appVersion = "1.0.0"
)

func main() {
	fmt.Printf("%s v%s - Terminal Media Player\n", appName, appVersion)
	fmt.Println("Inspired by Sakura - Cross-platform terminal multimedia player")
	fmt.Println(strings.Repeat("=", 60))

	// Initialize terminal control
	termControl := terminal.NewControl()
	defer termControl.ShowCursor()

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		fmt.Println("\n\nShutting down gracefully...")
		termControl.ShowCursor()
		termControl.Reset()
		os.Exit(0)
	}()

	// Detect terminal capabilities
	capabilities, err := terminal.DetectCapabilities()
	if err != nil {
		fmt.Printf("Warning: Could not detect terminal capabilities: %v\n", err)
		capabilities = types.TerminalCapabilities{
			Width:          80,
			Height:         24,
			Color256:       true,
			TrueColor:      false,
			SixelSupport:   false,
			UnicodeSupport: true,
		}
	}

	// Display terminal information
	displayTerminalInfo(capabilities)

	// Initialize renderer manager
	rendererManager := renderer.NewRendererManager(capabilities)

	// Main application loop
	for {
		if !showMainMenu(rendererManager, termControl, capabilities) {
			break
		}
	}

	fmt.Println("Thank you for using TerminalTube!")
}

// displayTerminalInfo shows detected terminal capabilities
func displayTerminalInfo(capabilities types.TerminalCapabilities) {
	fmt.Printf("Terminal: %dx%d\n", capabilities.Width, capabilities.Height)
	fmt.Printf("SIXEL Support: %v\n", capabilities.SixelSupport)
	fmt.Printf("True Color (24-bit): %v\n", capabilities.TrueColor)
	fmt.Printf("256 Colors: %v\n", capabilities.Color256)
	fmt.Printf("Unicode Support: %v\n", capabilities.UnicodeSupport)

	// Debug: Show environment variables
	fmt.Printf("TERM: %s\n", os.Getenv("TERM"))
	fmt.Printf("TERM_PROGRAM: %s\n", os.Getenv("TERM_PROGRAM"))
	fmt.Printf("COLORTERM: %s\n", os.Getenv("COLORTERM"))
	fmt.Printf("WT_SESSION: %s\n", os.Getenv("WT_SESSION"))
	fmt.Printf("OS: %s\n", os.Getenv("OS"))

	fmt.Println(strings.Repeat("-", 40))
}

// showMainMenu displays the main menu and handles user input
func showMainMenu(rendererManager *renderer.RendererManager, termControl *terminal.Control, capabilities types.TerminalCapabilities) bool {
	fmt.Println("\nMain Menu:")
	fmt.Println("1. Display Image")
	fmt.Println("2. Play GIF Animation")
	fmt.Println("3. Play Video from URL")
	fmt.Println("4. Play Video from File")
	fmt.Println("5. Terminal Information")
	fmt.Println("6. Rendering Tests")
	fmt.Println("0. Exit")
	fmt.Print("\nSelect option: ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}

	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		handleImageDisplay(rendererManager, termControl, capabilities)
	case "2":
		handleGIFPlayback(rendererManager, termControl, capabilities)
	case "3":
		handleVideoFromURL(rendererManager, termControl, capabilities)
	case "4":
		handleVideoFromFile(rendererManager, termControl, capabilities)
	case "5":
		showDetailedTerminalInfo(termControl)
	case "6":
		runRenderingTests(rendererManager, capabilities)
	case "0":
		return false
	default:
		fmt.Println("Invalid option. Please try again.")
	}

	return true
}

// handleImageDisplay handles static image display
func handleImageDisplay(rendererManager *renderer.RendererManager, termControl *terminal.Control, capabilities types.TerminalCapabilities) {
	fmt.Print("Enter image file path: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}

	imagePath := strings.TrimSpace(scanner.Text())
	if imagePath == "" {
		fmt.Println("No path provided.")
		return
	}

	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		fmt.Printf("File not found: %s\n", imagePath)
		return
	}

	// Decode image
	imageDecoder := decoder.NewImageDecoder()
	if !imageDecoder.IsSupported(imagePath) {
		fmt.Println("Unsupported image format.")
		return
	}

	img, mediaInfo, err := imageDecoder.DecodeImage(imagePath)
	if err != nil {
		fmt.Printf("Failed to decode image: %v\n", err)
		return
	}

	fmt.Printf("Image loaded: %dx%d pixels\n", mediaInfo.Width, mediaInfo.Height)

	// Get best renderer
	bestRenderer := rendererManager.GetBestRenderer()
	fmt.Printf("Using renderer: %s\n", bestRenderer.Name())

	// Initialize renderer
	if err := bestRenderer.Initialize(); err != nil {
		fmt.Printf("Failed to initialize renderer: %v\n", err)
		return
	}
	defer bestRenderer.Cleanup()

	// Set up render options for full terminal usage
	options := types.DefaultRenderOptions()

	// Calculate optimal render size using the new function
	options.Width, options.Height = calculateOptimalRenderSize(
		mediaInfo.Width, mediaInfo.Height,
		capabilities.Width, capabilities.Height)

	// Debug: Show terminal size calculation
	fmt.Printf("Terminal size: %dx%d, Available space: %dx%d\n",
		capabilities.Width, capabilities.Height, capabilities.Width-1, capabilities.Height-3)

	fmt.Printf("Rendering at %dx%d (original: %dx%d) - Utilization: %.1f%%\n",
		options.Width, options.Height, mediaInfo.Width, mediaInfo.Height,
		float64(options.Width*options.Height)/float64((capabilities.Width-1)*(capabilities.Height-3))*100)

	// Debug: Show color capabilities
	fmt.Printf("Color capabilities - True Color: %v, 256 Color: %v, SIXEL: %v\n",
		capabilities.TrueColor, capabilities.Color256, capabilities.SixelSupport)

	// Select rendering mode based on capabilities (prefer color modes)
	if capabilities.SixelSupport {
		options.Mode = types.SIXEL
		fmt.Println("Using SIXEL graphics mode")
	} else if capabilities.TrueColor {
		options.Mode = types.EXACT
		fmt.Println("Using Unicode true-color mode")
	} else if capabilities.Color256 {
		options.Mode = types.ASCII_COLOR
		fmt.Println("Using ASCII color mode")
	} else {
		// Force ASCII color even if detection fails - most terminals support it
		options.Mode = types.ASCII_COLOR
		fmt.Println("Using ASCII color mode (forced fallback)")
	}

	// Render image
	fmt.Println("Rendering image...")
	rendered, err := bestRenderer.Render(img, options)
	if err != nil {
		fmt.Printf("Failed to render image: %v\n", err)
		return
	}

	// Display image
	termControl.ClearScreen()
	termControl.HideCursor()
	fmt.Print(rendered)

	fmt.Printf("\nPress Enter to continue...")
	scanner.Scan()

	termControl.ShowCursor()
	termControl.ClearScreen()
}

// handleGIFPlayback handles animated GIF playback
func handleGIFPlayback(rendererManager *renderer.RendererManager, termControl *terminal.Control, capabilities types.TerminalCapabilities) {
	fmt.Print("Enter GIF file path: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}

	gifPath := strings.TrimSpace(scanner.Text())
	if gifPath == "" {
		fmt.Println("No path provided.")
		return
	}

	// Check if file exists
	if _, err := os.Stat(gifPath); os.IsNotExist(err) {
		fmt.Printf("File not found: %s\n", gifPath)
		return
	}

	// Load GIF
	gifDecoder := decoder.NewGIFDecoder()
	if !gifDecoder.IsSupported(gifPath) {
		fmt.Println("Not a valid GIF file.")
		return
	}

	mediaInfo, err := gifDecoder.LoadGIF(gifPath)
	if err != nil {
		fmt.Printf("Failed to load GIF: %v\n", err)
		return
	}

	fmt.Printf("GIF loaded: %dx%d pixels, %d frames, %.1f FPS\n",
		mediaInfo.Width, mediaInfo.Height, mediaInfo.FrameCount, mediaInfo.FPS)

	// Get renderer
	bestRenderer := rendererManager.GetBestRenderer()
	fmt.Printf("Using renderer: %s\n", bestRenderer.Name())

	if err := bestRenderer.Initialize(); err != nil {
		fmt.Printf("Failed to initialize renderer: %v\n", err)
		return
	}
	defer bestRenderer.Cleanup()

	// Set up render options with optimal sizing
	options := types.DefaultRenderOptions()

	// Calculate optimal render size
	options.Width, options.Height = calculateOptimalRenderSize(
		mediaInfo.Width, mediaInfo.Height,
		capabilities.Width, capabilities.Height)

	fmt.Printf("GIF render size: %dx%d (original: %dx%d)\n",
		options.Width, options.Height, mediaInfo.Width, mediaInfo.Height)

	// Select rendering mode
	if capabilities.SixelSupport {
		options.Mode = types.SIXEL
	} else if capabilities.TrueColor {
		options.Mode = types.EXACT
	} else if capabilities.Color256 {
		options.Mode = types.ASCII_COLOR
	} else {
		options.Mode = types.ASCII_GRAY
	}

	fmt.Println("Playing GIF... Press Ctrl+C to stop")
	time.Sleep(1 * time.Second)

	// Play GIF
	termControl.ClearScreen()
	termControl.HideCursor()

	// Get frame channel
	frameChan, err := gifDecoder.GetFrameChannel()
	if err != nil {
		fmt.Printf("Failed to get frame channel: %v\n", err)
		return
	}

	// Play frames
	lastResizeCheck := time.Now()
	for frame := range frameChan {
		// Check for terminal resize every second
		if time.Since(lastResizeCheck) >= time.Second {
			newCapabilities, err := terminal.DetectCapabilities()
			if err == nil {
				newWidth, newHeight := calculateOptimalRenderSize(
					mediaInfo.Width, mediaInfo.Height,
					newCapabilities.Width, newCapabilities.Height)

				// Only update if size changed significantly
				if newWidth != options.Width || newHeight != options.Height {
					fmt.Printf("\nTerminal resized to %dx%d, adjusting GIF size to %dx%d\n",
						newCapabilities.Width, newCapabilities.Height, newWidth, newHeight)
					options.Width = newWidth
					options.Height = newHeight
					capabilities = newCapabilities
				}
			}
			lastResizeCheck = time.Now()
		}

		rendered, err := bestRenderer.Render(frame.Image, options)
		if err != nil {
			fmt.Printf("Failed to render frame: %v\n", err)
			break
		}

		termControl.MoveCursorHome()
		fmt.Print(rendered)

		// The frame timing is handled by the decoder
	}

	termControl.ShowCursor()
	termControl.ClearScreen()
	gifDecoder.Close()
}

// handleVideoFromURL handles video playback from URL
func handleVideoFromURL(rendererManager *renderer.RendererManager, termControl *terminal.Control, capabilities types.TerminalCapabilities) {
	fmt.Print("Enter video URL: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}

	videoURL := strings.TrimSpace(scanner.Text())
	if videoURL == "" {
		fmt.Println("No URL provided.")
		return
	}

	// Download video
	downloader := fetcher.NewDownloader()
	defer downloader.Close()

	if !downloader.IsValidURL(videoURL) {
		fmt.Println("Invalid URL format.")
		return
	}

	// Check for YouTube URLs - stream directly using yt-dlp
	if downloader.IsYouTubeURL(videoURL) {
		fmt.Println("YouTube URL detected! Extracting stream...")

		info, err := downloader.GetYouTubeVideoInfo(videoURL)
		if err != nil {
			fmt.Printf("Error processing YouTube URL: %v\n", err)
			return
		}

		fmt.Printf("Video ID: %s\n", info["video_id"])

		// Get direct stream URL or download file
		streamOrPath, err := downloader.DownloadYouTubeVideo(videoURL)
		if err != nil {
			fmt.Printf("YouTube Error: %v\n", err)
			return
		}

		// If it's a URL (starts with http), download to temp first for reliable playback
		// Otherwise it's already a temp file path
		if strings.HasPrefix(streamOrPath, "http") {
			fmt.Println("Got direct stream URL, downloading for playback...")
			tempPath, err := downloader.DownloadToTemp(streamOrPath, func(downloaded, total int64, percent float64) {
				if total > 0 {
					fmt.Printf("\rProgress: %.1f%% (%d/%d bytes)", percent, downloaded, total)
				} else {
					fmt.Printf("\rDownloaded: %d bytes", downloaded)
				}
			})
			if err != nil {
				fmt.Printf("\nFailed to download stream: %v\n", err)
				return
			}
			defer os.Remove(tempPath)
			fmt.Printf("\nDownload complete!\n")
			playVideoFile(tempPath, rendererManager, termControl, capabilities)
		} else {
			// It's a temp file path from yt-dlp download
			defer os.Remove(streamOrPath)
			playVideoFile(streamOrPath, rendererManager, termControl, capabilities)
		}
		return
	}

	if !downloader.IsMediaURL(videoURL) {
		fmt.Println("URL does not appear to point to a media file.")
		fmt.Println("Supported formats: .mp4, .avi, .mov, .mkv, .webm, .flv")
		fmt.Println("For YouTube videos, please see the guidance above.")
		return
	}

	fmt.Println("Downloading video...")

	// Show download progress
	progressCallback := func(downloaded, total int64, percent float64) {
		if total > 0 {
			fmt.Printf("\rProgress: %.1f%% (%d/%d bytes)", percent, downloaded, total)
		} else {
			fmt.Printf("\rDownloaded: %d bytes", downloaded)
		}
	}

	tempPath, err := downloader.DownloadToTemp(videoURL, progressCallback)
	if err != nil {
		fmt.Printf("\nFailed to download video: %v\n", err)
		return
	}
	defer os.Remove(tempPath) // Clean up temp file

	fmt.Printf("\nDownload complete: %s\n", tempPath)

	// Play the downloaded video
	playVideoFile(tempPath, rendererManager, termControl, capabilities)
}

// handleVideoFromFile handles video playback from local file
func handleVideoFromFile(rendererManager *renderer.RendererManager, termControl *terminal.Control, capabilities types.TerminalCapabilities) {
	fmt.Print("Enter video file path: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}

	videoPath := strings.TrimSpace(scanner.Text())
	if videoPath == "" {
		fmt.Println("No path provided.")
		return
	}

	// Check if file exists
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		fmt.Printf("File not found: %s\n", videoPath)
		return
	}

	playVideoFile(videoPath, rendererManager, termControl, capabilities)
}

// playVideoFile plays a video file with audio synchronization
func playVideoFile(videoPath string, rendererManager *renderer.RendererManager, termControl *terminal.Control, capabilities types.TerminalCapabilities) {
	// Load video
	videoDecoder := decoder.NewVideoDecoder()
	if !videoDecoder.IsSupported(videoPath) {
		fmt.Println("Unsupported video format.")
		return
	}

	mediaInfo, err := videoDecoder.LoadVideo(videoPath)
	if err != nil {
		fmt.Printf("Failed to load video: %v\n", err)
		return
	}

	fmt.Printf("Video loaded: %dx%d pixels, %.1f FPS, %.1fs duration\n",
		mediaInfo.Width, mediaInfo.Height, mediaInfo.FPS, mediaInfo.Duration)
	fmt.Printf("Video codec: %s, Has audio: %v\n", mediaInfo.VideoCodec, mediaInfo.HasAudio)

	// Initialize audio player if video has audio
	var audioPlayer *audio.Player
	if mediaInfo.HasAudio {
		audioPlayer = audio.NewPlayer()
		if err := audioPlayer.LoadAudio(videoPath); err != nil {
			fmt.Printf("Warning: Could not load audio: %v\n", err)
			audioPlayer = nil
		} else {
			fmt.Println("Audio loaded successfully")
		}
	}

	// Get renderer
	bestRenderer := rendererManager.GetBestRenderer()
	fmt.Printf("Using renderer: %s\n", bestRenderer.Name())

	if err := bestRenderer.Initialize(); err != nil {
		fmt.Printf("Failed to initialize renderer: %v\n", err)
		return
	}
	defer bestRenderer.Cleanup()

	// Set up render options with dynamic sizing and adaptive scaling based on FPS
	options := types.DefaultRenderOptions()

	// Calculate optimal render size based on render mode
	var pixelWidth, pixelHeight int
	var optimalWidth, optimalHeight int

	if capabilities.SixelSupport {
		// SIXEL mode: calculate pixel dimensions directly for full terminal coverage
		pixelWidth, pixelHeight = calculateOptimalRenderSizeSixel(
			mediaInfo.Width, mediaInfo.Height,
			capabilities.Width, capabilities.Height)

		// Convert back to character dimensions for options (for renderer info)
		optimalWidth = pixelWidth / 8
		optimalHeight = pixelHeight / 16

		fmt.Printf("SIXEL render size: %dx%d pixels (terminal: %dx%d chars)\n",
			pixelWidth, pixelHeight, capabilities.Width, capabilities.Height)
	} else {
		// Unicode/ASCII mode: calculate character dimensions
		optimalWidth, optimalHeight = calculateOptimalRenderSize(
			mediaInfo.Width, mediaInfo.Height,
			capabilities.Width, capabilities.Height)

		// Unicode half-blocks: 2 pixels per character height
		pixelWidth = optimalWidth
		pixelHeight = optimalHeight * 2

		fmt.Printf("Optimal render size calculated: %dx%d (terminal: %dx%d)\n",
			optimalWidth, optimalHeight, capabilities.Width, capabilities.Height)
	}

	options.Width = optimalWidth
	options.Height = optimalHeight
	videoDecoder.SetRenderSize(pixelWidth, pixelHeight)

	fmt.Printf("Video render size: %dx%d pixels\n", pixelWidth, pixelHeight)

	// Select rendering mode
	if capabilities.SixelSupport {
		options.Mode = types.SIXEL
	} else if capabilities.TrueColor {
		options.Mode = types.EXACT
	} else if capabilities.Color256 {
		options.Mode = types.ASCII_COLOR
	} else {
		options.Mode = types.ASCII_GRAY
	}

	fmt.Println("Playing video... Press Ctrl+C to stop")
	time.Sleep(1 * time.Second)

	// Start audio playback if available
	if audioPlayer != nil {
		if err := audioPlayer.Play(); err != nil {
			fmt.Printf("Warning: Could not start audio: %v\n", err)
		} else {
			defer audioPlayer.Close()
		}
	}

	// Initialize playback statistics
	stats := &types.PlaybackStats{
		StartTime: time.Now().UnixNano(),
	}

	// Play video
	termControl.ClearScreen()
	termControl.HideCursor()

	// Get frame channel
	frameChan, err := videoDecoder.GetFrameChannel()
	if err != nil {
		fmt.Printf("Failed to get frame channel: %v\n", err)
		return
	}

	// Frame timing variables
	_ = time.Duration(1000000000.0 / mediaInfo.FPS) // Frame duration handled by decoder
	lastResizeCheck := time.Now()
	resizeCheckInterval := 1 * time.Second // Check for resize every second

	// Play frames with timing and dynamic resizing
	for frame := range frameChan {
		currentTime := time.Now()

		// Check for terminal resize periodically
		if currentTime.Sub(lastResizeCheck) > resizeCheckInterval {
			if newWidth, newHeight, err := getCurrentTerminalSize(); err == nil {
				if newWidth != capabilities.Width || newHeight != capabilities.Height {
					// Terminal size changed - recalculate render size
					capabilities.Width = newWidth
					capabilities.Height = newHeight

					// Recalculate optimal size with performance scaling
					newRenderWidth, newRenderHeight := calculateOptimalRenderSize(
						mediaInfo.Width, mediaInfo.Height, newWidth, newHeight)

					// Apply performance scaling
					scaleFactor := 1.0
					if mediaInfo.FPS > 50 {
						scaleFactor = 0.85
					} else if mediaInfo.FPS > 30 {
						scaleFactor = 0.90
					}

					options.Width = int(float64(newRenderWidth) * scaleFactor)
					options.Height = int(float64(newRenderHeight) * scaleFactor)

					// Clear screen and update display
					termControl.ClearScreen()
					fmt.Printf("Terminal resized to %dx%d, adjusting render size to %dx%d\n",
						newWidth, newHeight, options.Width, options.Height)
				}
			}
			lastResizeCheck = currentTime
		}

		// Render frame
		rendered, err := bestRenderer.Render(frame.Image, options)
		if err != nil {
			fmt.Printf("Failed to render frame: %v\n", err)
			break
		}

		// Display frame - move cursor to home position
		// For SIXEL, new image overwrites old at same position (no clear needed)
		fmt.Print("\033[H") // Move cursor to home (1,1)
		fmt.Print(rendered)

		stats.FramesRendered++

		// Calculate statistics (but don't display during SIXEL to avoid cursor issues)
		if stats.FramesRendered%30 == 0 {
			elapsed := float64(time.Now().UnixNano()-stats.StartTime) / 1000000000.0
			stats.FPS = float64(stats.FramesRendered) / elapsed
			if stats.FramesRendered+stats.FramesDropped > 0 {
				stats.DropRate = float64(stats.FramesDropped) / float64(stats.FramesRendered+stats.FramesDropped) * 100.0
			}

			// Only show stats for non-SIXEL modes (SIXEL cursor positioning is tricky)
			if options.Mode != types.SIXEL {
				termControl.MoveCursor(capabilities.Height-1, 1)
				fmt.Printf("FPS: %.1f | Frames: %d | Dropped: %d (%.1f%%) | Size: %dx%d",
					stats.FPS, stats.FramesRendered, stats.FramesDropped, stats.DropRate, options.Width, options.Height)
			}
		}
	}

	termControl.ShowCursor()
	termControl.ClearScreen()
	videoDecoder.Close()

	// Display final statistics
	elapsed := float64(time.Now().UnixNano()-stats.StartTime) / 1000000000.0
	stats.FPS = float64(stats.FramesRendered) / elapsed
	if stats.FramesRendered+stats.FramesDropped > 0 {
		stats.DropRate = float64(stats.FramesDropped) / float64(stats.FramesRendered+stats.FramesDropped) * 100.0
	}

	fmt.Println("\nPlayback Statistics:")
	fmt.Printf("Total Time: %.1f seconds\n", elapsed)
	fmt.Printf("Frames Rendered: %d\n", stats.FramesRendered)
	fmt.Printf("Frames Dropped: %d\n", stats.FramesDropped)
	fmt.Printf("Drop Rate: %.1f%%\n", stats.DropRate)
	fmt.Printf("Average FPS: %.1f\n", stats.FPS)
}

// showDetailedTerminalInfo displays detailed terminal information
func showDetailedTerminalInfo(termControl *terminal.Control) {
	fmt.Println("\nDetailed Terminal Information:")
	fmt.Println(strings.Repeat("=", 40))

	info := termControl.GetTerminalInfo()
	for key, value := range info {
		if value != "" {
			fmt.Printf("%-20s: %s\n", key, value)
		}
	}

	// Try to get current terminal size
	if width, height, err := terminal.GetTerminalSize(); err == nil {
		fmt.Printf("%-20s: %dx%d\n", "Current Size", width, height)
	}

	fmt.Println("\nPress Enter to continue...")
	bufio.NewScanner(os.Stdin).Scan()
}

// runRenderingTests runs basic rendering tests
func runRenderingTests(rendererManager *renderer.RendererManager, capabilities types.TerminalCapabilities) {
	fmt.Println("\nRendering Tests:")
	fmt.Println(strings.Repeat("=", 40))

	availableModes := rendererManager.GetAvailableModes()
	fmt.Printf("Available render modes: %d\n", len(availableModes))

	for i, mode := range availableModes {
		fmt.Printf("%d. %s\n", i+1, mode.String())
	}

	fmt.Print("\nSelect mode to test (0 to skip): ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return
	}

	choice := strings.TrimSpace(scanner.Text())
	if choice == "0" || choice == "" {
		return
	}

	modeIndex, err := strconv.Atoi(choice)
	if err != nil || modeIndex < 1 || modeIndex > len(availableModes) {
		fmt.Println("Invalid selection.")
		return
	}

	selectedMode := availableModes[modeIndex-1]
	renderer, exists := rendererManager.GetRenderer(selectedMode)
	if !exists {
		fmt.Println("Renderer not available.")
		return
	}

	fmt.Printf("Testing renderer: %s\n", renderer.Name())

	// TODO: Implement rendering tests with sample patterns
	fmt.Println("Rendering tests not yet implemented.")
	fmt.Println("This would display color gradients, patterns, and test images.")

	fmt.Println("Press Enter to continue...")
	scanner.Scan()
}
