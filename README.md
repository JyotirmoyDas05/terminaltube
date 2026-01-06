# TerminalTube

A powerful, cross-platform terminal media player built with Go and Bubble Tea. Renders images, GIFs, and videos using SIXEL graphics, Unicode blocks, and ANSI colors with synchronized audio playback.

## 🚀 Recent Updates

- **Automated Setup**: New TUI-based dependency installer for `ffmpeg` and `yt-dlp`.
- **Enhanced TUI**: Modernize interface with animations, better spacing, and footer hints.
- **GIF URL Support**: Play animated GIFs directly from the web.
- **Visual Testing**: Built-in rendering test suite for terminal capability verification.
- **Hover/Click Support**: Clickable links in the About section (OSC 8).

## ✨ Features

- **Multi-format Support**: Images (JPG, PNG, BMP), GIFs, Videos (MP4, AVI, MOV, MKV).
- **YouTube Support**: Stream and play YouTube videos directly using `yt-dlp`.
- **Advanced Rendering Modes**:
  - **SIXEL**: High-performance, high-quality graphics for compatible terminals.
  - **Unicode True-color**: Beautiful 24-bit color rendering using half-block characters.
  - **ASCII Color/Grayscale**: Reliable fallbacks for all terminal environments.
- **Intelligent Dependency Management**: Automatically detects missing tools and offers to install them via `winget`, `brew`, or `apt`.
- **Audio-Video Sync**: Precise synchronization for a full media experience.
- **Dynamic Resizing**: Adapts the rendering resolution in real-time as you resize your terminal.

## 🛠️ Installation

TerminalTube is designed to be easy to set up.

### 1. Build from Source

```bash
git clone https://github.com/JyotirmoyDas05/terminaltube.git
cd terminaltube
go mod tidy
go build -o terminaltube.exe .
```

### 2. Automated Setup

On first launch, TerminalTube will detect if you have the necessary external tools:

- **FFmpeg**: Required for media decoding and audio playback.
- **yt-dlp**: Required for YouTube streaming.

If missing, the **Setup Tool** will launch automatically to help you install them with a single click.

## 🎮 Usage

Run the executable to enter the interactive TUI:

```bash
./terminaltube.exe
```

### Main Menu Options:

1.  **🖼️ Display Image**: Show static images with high-fidelity rendering.
2.  **🎞️ Play GIF Animation**: smooth, timed GIF playback.
3.  **🔗 Play GIF from URL**: Download and play GIFs from any web link.
4.  **🌐 Play Video from URL**: Stream videos or YouTube links directly.
5.  **📁 Play Video from File**: Play local video files with full audio.
6.  **🧪 Rendering Tests**: Verify your terminal's color and graphics support.
7.  **🧹 Clear Cache**: Clean up temporary downloaded media files.
8.  **💡 About**: Learn about the project and view developer credits.

## 🖥️ Terminal Compatibility

| Mode          | Supported Terminals                                                 |
| :------------ | :------------------------------------------------------------------ |
| **SIXEL**     | Windows Terminal, iTerm2, WezTerm, Foot, Alacritty (recent), Mintty |
| **TrueColor** | Most modern terminals (VS Code, GNOME, Konsole, etc.)               |
| **Unicode**   | Any terminal with UTF-8 support                                     |

## 🏗️ Architecture

```
terminaltube/
├── main.go                 # Modern Looping TUI Entry Point
├── internal/
│   ├── dependency/        # OS-aware Dependency Installer
│   ├── tui/               # Bubble Tea UI Components & Themes
│   ├── renderer/          # SIXEL/Unicode/ASCII Render Engines
│   ├── decoder/           # FFmpeg-based Media Decoding
│   ├── audio/             # Oto-based Audio Playback
│   └── fetcher/           # Progressive Media Downloader
└── pkg/types/             # Core Shared Types
```

## 🤝 Credits

Developed with ❤️ by **[@JyotirmoyDas05](https://github.com/JyotirmoyDas05)** using:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI framework.
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) for beautiful terminal styling.
- [FFmpeg](https://ffmpeg.org/) for powerful media processing.
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) for YouTube integration.

---

_Inspired by the synergy of modern Go and classic terminal graphics._
