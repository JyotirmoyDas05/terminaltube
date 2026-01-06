# TerminalTube

A cross-platform terminal media player that renders images, GIFs, and videos using SIXEL graphics with synchronized audio playback.

## Features

- **Multi-format Support**: Images (JPG, PNG, BMP), GIFs, Videos (MP4, AVI, MOV, MKV)
- **Four Rendering Modes**:
  - SIXEL: High-quality graphics for compatible terminals
  - ASCII Color: Block-based color rendering using ANSI codes
  - ASCII Grayscale: Character-based monochrome rendering
  - Unicode True-color: High-quality Unicode blocks with 24-bit color
- **Cross-platform**: Windows, macOS, Linux support
- **Audio-Video Synchronization**: Synchronized audio playback with video
- **URL Download Support**: Stream media directly from URLs
- **Adaptive Scaling**: Automatic scaling to terminal dimensions with aspect ratio preservation
- **Performance Optimized**: High-precision frame timing with adaptive frame skipping

## Installation

### Prerequisites

#### For Full Functionality (Optional CGO Dependencies)

- **OpenCV**: For video processing (gocv)
- **libsixel**: For high-quality SIXEL rendering
- **Audio libraries**: For cross-platform audio support

#### Go Dependencies

All Go dependencies will be automatically downloaded via `go mod`:

- `gocv.io/x/gocv` - OpenCV bindings for video/image processing
- `github.com/ebitengine/oto/v3` - Cross-platform audio playback
- `github.com/faiface/beep` - Audio processing and format support
- `github.com/nfnt/resize` - Image resizing
- `github.com/gdamore/tcell/v2` - Terminal handling and ANSI colors
- `golang.org/x/term` - Terminal size detection

### Quick Install

1. **Clone the repository**:

   ```bash
   git clone https://github.com/yourusername/terminaltube.git
   cd terminaltube
   ```

2. **Build and run**:
   ```bash
   go mod tidy
   go build -o terminaltube .
   ./terminaltube
   ```

### Platform-Specific Installation

#### Windows

```powershell
# Install Go if not already installed
# Download from https://golang.org/dl/

# Clone and build
git clone https://github.com/yourusername/terminaltube.git
cd terminaltube
go mod tidy
go build -o terminaltube.exe .
.\terminaltube.exe
```

#### macOS

```bash
# Install Go via Homebrew
brew install go

# For full video support (optional)
brew install opencv pkg-config

# Clone and build
git clone https://github.com/yourusername/terminaltube.git
cd terminaltube
go mod tidy
go build -o terminaltube .
./terminaltube
```

#### Linux (Ubuntu/Debian)

```bash
# Install Go
sudo apt update
sudo apt install golang-go

# For full video support (optional)
sudo apt install libopencv-dev pkg-config

# For SIXEL support (optional)
sudo apt install libsixel-dev

# Clone and build
git clone https://github.com/yourusername/terminaltube.git
cd terminaltube
go mod tidy
go build -o terminaltube .
./terminaltube
```

#### Linux (CentOS/RHEL/Fedora)

```bash
# Install Go
sudo dnf install golang

# For full video support (optional)
sudo dnf install opencv-devel pkgconfig

# Clone and build
git clone https://github.com/yourusername/terminaltube.git
cd terminaltube
go mod tidy
go build -o terminaltube .
./terminaltube
```

## Usage

### Interactive Mode

Run `terminaltube` without arguments to enter interactive mode:

```bash
./terminaltube
```

The application will display a menu with the following options:

1. **Display Image** - Show static images
2. **Play GIF Animation** - Play animated GIFs with proper timing
3. **Play Video from URL** - Download and play video from web URLs
4. **Play Video from File** - Play local video files
5. **Terminal Information** - Display terminal capabilities
6. **Rendering Tests** - Test different rendering modes

### Command Line Usage (Future Enhancement)

```bash
# Display an image
./terminaltube image.jpg

# Play a GIF
./terminaltube animation.gif

# Play a video
./terminaltube movie.mp4

# Play from URL
./terminaltube https://example.com/video.mp4

# Specify rendering mode
./terminaltube --mode sixel image.jpg
./terminaltube --mode ascii-color video.mp4
```

## Terminal Compatibility

### SIXEL Support

- **xterm** (with SIXEL enabled)
- **mlterm**
- **wezterm**
- **foot**
- **mintty** (Git Bash on Windows)
- **iTerm2** (macOS)
- **Windows Terminal** (with SIXEL support)

### True Color (24-bit) Support

- **Windows Terminal**
- **iTerm2**
- **Alacritty**
- **wezterm**
- **GNOME Terminal**
- **Konsole**
- **Kitty**

### Fallback Support

All terminals support ASCII color (256-color) and grayscale fallback modes.

## Architecture

### Project Structure

```
terminaltube/
├── main.go                 # Main application entry point
├── go.mod                 # Go module definition
├── internal/              # Internal packages
│   ├── renderer/          # Rendering backends
│   │   ├── renderer.go    # Common renderer interface
│   │   ├── sixel.go       # SIXEL graphics rendering
│   │   ├── ascii.go       # ASCII color/grayscale rendering
│   │   └── unicode.go     # True-color Unicode rendering
│   ├── decoder/           # Media decoders
│   │   ├── image.go       # Static image decoding
│   │   ├── gif.go         # GIF animation handling
│   │   └── video.go       # Video decoding with OpenCV
│   ├── audio/             # Audio playback
│   │   └── player.go      # Audio player using oto/beep
│   ├── fetcher/           # URL downloading
│   │   └── download.go    # HTTP download with progress
│   └── terminal/          # Terminal control
│       ├── capability.go  # Terminal capability detection
│       └── control.go     # ANSI escape sequences
└── pkg/
    └── types/
        └── options.go     # Configuration types
```

### Rendering Pipeline

1. **Media Detection**: Identify file type and capabilities
2. **Terminal Detection**: Detect SIXEL, true-color, and Unicode support
3. **Renderer Selection**: Choose best available renderer
4. **Adaptive Scaling**: Calculate optimal dimensions based on FPS and terminal size
5. **Frame Processing**: Decode, resize, and render frames
6. **Synchronized Playback**: Maintain precise timing with audio

## Performance

### Optimization Features

- **Adaptive Frame Skipping**: Automatically skip frames when behind schedule
- **Buffered I/O**: Efficient terminal output
- **Memory Management**: Frame pooling to reduce GC pressure
- **Fast Interpolation**: Use nearest-neighbor scaling for real-time video

### Performance Targets

- Frame drop rate < 10% for smooth playback
- Support for videos up to 60 FPS
- Memory usage proportional to video resolution
- Startup time < 1 second

## Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/yourusername/terminaltube.git
cd terminaltube

# Download dependencies
go mod tidy

# Build for current platform
go build -o terminaltube .

# Cross-compile for different platforms
GOOS=windows GOARCH=amd64 go build -o terminaltube.exe .
GOOS=darwin GOARCH=arm64 go build -o terminaltube-macos .
GOOS=linux GOARCH=amd64 go build -o terminaltube-linux .
```

### Testing

```bash
# Run tests
go test ./...

# Run with race detection
go test -race ./...

# Benchmark tests
go test -bench=. ./...
```

### Using the Makefile

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Clean build artifacts
make clean

# Install to $GOPATH/bin
make install
```

## Configuration

### Environment Variables

- `TERM`: Terminal type (affects capability detection)
- `COLORTERM`: Color support indicator (`truecolor`, `24bit`)
- `TERM_PROGRAM`: Terminal application name
- `COLUMNS` / `LINES`: Terminal dimensions (fallback)

### Runtime Options

Rendering options can be configured through the interactive menu:

- **Rendering Mode**: SIXEL, ASCII Color, ASCII Grayscale, Unicode True-color
- **Aspect Ratio**: Preserve original proportions
- **Brightness/Contrast**: Image adjustments
- **Palette Size**: For SIXEL rendering (default: 256 colors)

## Troubleshooting

### Common Issues

#### "No suitable renderer found"

- **Cause**: Terminal doesn't support any advanced rendering modes
- **Solution**: Use ASCII grayscale mode or upgrade terminal

#### "Failed to decode video"

- **Cause**: Missing OpenCV or unsupported codec
- **Solution**: Install OpenCV (`libopencv-dev`) or use supported formats

#### "Audio playback failed"

- **Cause**: Missing audio libraries or unsupported format
- **Solution**: Install audio development libraries

#### Poor video performance

- **Solutions**:
  - Reduce terminal size
  - Use ASCII modes instead of SIXEL
  - Close other applications
  - Use videos with lower resolution/FPS

#### SIXEL not working

- **Check**: Terminal SIXEL support with `echo -e '\ePq"1;1;100;100#0;2;0;0;0#0~~@@vv@@~~@@~~$#1;2;100;100;0#1!14@\e\\'`
- **Solution**: Use compatible terminal or fallback to ASCII modes

### Debug Mode

Enable verbose logging:

```bash
export TERMINALTUBE_DEBUG=1
./terminaltube
```

## Contributing

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Install development dependencies
4. Run tests before submitting PR

### Code Style

- Follow Go idioms and best practices
- Use `gofmt` for formatting
- Add comprehensive comments
- Write unit tests for new features

### Submitting Issues

Please include:

- Operating system and version
- Terminal application and version
- Go version
- Complete error messages
- Steps to reproduce

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

- Terminal graphics techniques from impulse-gerp-2025 SIXEL demo
- Uses [OpenCV](https://opencv.org/) for video processing
- Terminal graphics techniques from [libsixel](https://github.com/saitoha/libsixel)
- Cross-platform audio via [oto](https://github.com/ebitengine/oto) and [beep](https://github.com/faiface/beep)

## Roadmap

### Version 1.1

- [ ] Complete gocv integration for video processing
- [ ] CGO bindings for libsixel
- [ ] Command-line argument support
- [ ] Configuration file support

### Version 1.2

- [ ] Streaming support for online videos
- [ ] Playlist support
- [ ] Subtitle rendering
- [ ] Video filters and effects

### Version 2.0

- [ ] Interactive controls (pause, seek, volume)
- [ ] Multiple video format support
- [ ] Hardware acceleration
- [ ] Network streaming protocols
