package dependency

import (
	"os/exec"
	"runtime"
	"strings"
)

// Dependency represents a required external tool
type Dependency struct {
	Name        string
	Description string
	CheckCmd    string            // Command to check if installed (e.g., "ffmpeg -version")
	InstallCmds map[string]string // OS -> install command
	Required    bool
}

// Status represents the installation status of a dependency
type Status struct {
	Dep       Dependency
	Installed bool
	Version   string
	Error     error
}

// Manager handles dependency checking and installation
type Manager struct {
	dependencies []Dependency
}

// NewManager creates a new dependency manager with default dependencies
func NewManager() *Manager {
	return &Manager{
		dependencies: []Dependency{
			{
				Name:        "ffmpeg",
				Description: "Video/audio processing (required for playback)",
				CheckCmd:    "ffmpeg",
				InstallCmds: map[string]string{
					"windows": "winget install --id Gyan.FFmpeg -e --source winget",
					"darwin":  "brew install ffmpeg",
					"linux":   "sudo apt-get install -y ffmpeg",
				},
				Required: true,
			},
			{
				Name:        "yt-dlp",
				Description: "YouTube video downloading",
				CheckCmd:    "yt-dlp",
				InstallCmds: map[string]string{
					"windows": "winget install --id yt-dlp.yt-dlp -e --source winget",
					"darwin":  "brew install yt-dlp",
					"linux":   "sudo apt-get install -y yt-dlp",
				},
				Required: false,
			},
		},
	}
}

// GetOS returns the current operating system identifier
func GetOS() string {
	return runtime.GOOS
}

// IsWindows returns true if running on Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsMacOS returns true if running on macOS
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux returns true if running on Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// CheckDependency checks if a single dependency is installed
func CheckDependency(dep Dependency) Status {
	status := Status{Dep: dep, Installed: false}

	// Try to run the check command
	var cmd *exec.Cmd
	if IsWindows() {
		cmd = exec.Command("where", dep.CheckCmd)
	} else {
		cmd = exec.Command("which", dep.CheckCmd)
	}

	output, err := cmd.Output()
	if err != nil {
		status.Error = err
		return status
	}

	path := strings.TrimSpace(string(output))
	if path != "" {
		status.Installed = true
		// Try to get version
		verCmd := exec.Command(dep.CheckCmd, "-version")
		verOutput, _ := verCmd.Output()
		if len(verOutput) > 0 {
			// Extract first line as version
			lines := strings.Split(string(verOutput), "\n")
			if len(lines) > 0 {
				status.Version = strings.TrimSpace(lines[0])
			}
		}
	}

	return status
}

// CheckAll checks all dependencies and returns their statuses
func (m *Manager) CheckAll() []Status {
	var statuses []Status
	for _, dep := range m.dependencies {
		statuses = append(statuses, CheckDependency(dep))
	}
	return statuses
}

// GetMissing returns only missing dependencies
func (m *Manager) GetMissing() []Dependency {
	var missing []Dependency
	for _, dep := range m.dependencies {
		status := CheckDependency(dep)
		if !status.Installed {
			missing = append(missing, dep)
		}
	}
	return missing
}

// GetInstallCommand returns the install command for a dependency on current OS
func GetInstallCommand(dep Dependency) string {
	os := GetOS()
	if cmd, ok := dep.InstallCmds[os]; ok {
		return cmd
	}
	return ""
}

// HasMissingRequired returns true if any required dependency is missing
func (m *Manager) HasMissingRequired() bool {
	for _, dep := range m.dependencies {
		if dep.Required {
			status := CheckDependency(dep)
			if !status.Installed {
				return true
			}
		}
	}
	return false
}
