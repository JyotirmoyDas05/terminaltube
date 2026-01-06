package tui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"terminaltube/internal/dependency"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InstallerState represents the current state of the installer
type InstallerState int

const (
	StateChecking InstallerState = iota
	StatePrompt
	StateInstalling
	StateComplete
	StateSkipped
)

// InstallerModel is the Bubble Tea model for the dependency installer
type InstallerModel struct {
	state        InstallerState
	manager      *dependency.Manager
	missing      []dependency.Dependency
	statuses     []dependency.Status
	currentIndex int
	progress     progress.Model
	spinner      spinner.Model
	width        int
	height       int
	err          error
	installLog   []string
}

// InstallCompleteMsg signals installation is done
type InstallCompleteMsg struct {
	Success bool
	Error   error
}

// CheckCompleteMsg signals dependency check is done
type CheckCompleteMsg struct {
	Missing []dependency.Dependency
}

// InstallProgressMsg updates installation progress
type InstallProgressMsg struct {
	Percent float64
	Message string
}

// NewInstallerModel creates a new installer model
func NewInstallerModel() InstallerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(PrimaryColor)

	p := progress.New(progress.WithDefaultGradient())
	p.Width = 40

	return InstallerModel{
		state:    StateChecking,
		manager:  dependency.NewManager(),
		spinner:  s,
		progress: p,
	}
}

// Init initializes the model
func (m InstallerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.checkDependencies(),
	)
}

// checkDependencies runs the dependency check
func (m InstallerModel) checkDependencies() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(500 * time.Millisecond) // Brief delay for visual feedback
		missing := m.manager.GetMissing()
		return CheckCompleteMsg{Missing: missing}
	}
}

// Update handles messages
func (m InstallerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "y", "Y":
			if m.state == StatePrompt {
				m.state = StateInstalling
				m.currentIndex = 0
				return m, m.installNext()
			}
		case "n", "N", "s", "S":
			if m.state == StatePrompt {
				m.state = StateSkipped
				return m, tea.Quit
			}
		case "enter":
			if m.state == StateComplete || m.state == StateSkipped {
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 20
		if m.progress.Width > 60 {
			m.progress.Width = 60
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case CheckCompleteMsg:
		m.missing = msg.Missing
		if len(m.missing) == 0 {
			m.state = StateComplete
			return m, nil
		}
		m.state = StatePrompt
		return m, nil

	case InstallProgressMsg:
		m.installLog = append(m.installLog, msg.Message)
		return m, nil

	case InstallCompleteMsg:
		if msg.Error != nil {
			m.err = msg.Error
			m.installLog = append(m.installLog, fmt.Sprintf("Error: %v", msg.Error))
		}
		m.currentIndex++
		if m.currentIndex < len(m.missing) {
			return m, m.installNext()
		}
		m.state = StateComplete
		return m, nil

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

// installNext installs the next missing dependency
func (m InstallerModel) installNext() tea.Cmd {
	if m.currentIndex >= len(m.missing) {
		return func() tea.Msg {
			return InstallCompleteMsg{Success: true}
		}
	}

	dep := m.missing[m.currentIndex]
	installCmd := dependency.GetInstallCommand(dep)

	return func() tea.Msg {
		if installCmd == "" {
			return InstallCompleteMsg{
				Success: false,
				Error:   fmt.Errorf("no install command for %s on %s", dep.Name, dependency.GetOS()),
			}
		}

		// Parse and execute the command
		parts := strings.Fields(installCmd)
		if len(parts) == 0 {
			return InstallCompleteMsg{Error: fmt.Errorf("empty install command")}
		}

		cmd := exec.Command(parts[0], parts[1:]...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			return InstallCompleteMsg{
				Success: false,
				Error:   fmt.Errorf("%s: %v\n%s", dep.Name, err, string(output)),
			}
		}

		return InstallCompleteMsg{Success: true}
	}
}

// View renders the installer UI
func (m InstallerModel) View() string {
	var s strings.Builder

	// Header
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		Render("🔧 TerminalTube Setup")

	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(title))
	s.WriteString("\n\n")

	switch m.state {
	case StateChecking:
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(
			m.spinner.View() + " Checking dependencies...",
		))

	case StatePrompt:
		s.WriteString(m.renderPrompt())

	case StateInstalling:
		s.WriteString(m.renderInstalling())

	case StateComplete:
		s.WriteString(m.renderComplete())

	case StateSkipped:
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(
			"Setup skipped. Some features may not work without dependencies.\n\nPress Enter to continue...",
		))
	}

	return s.String()
}

func (m InstallerModel) renderPrompt() string {
	var s strings.Builder

	subtitle := lipgloss.NewStyle().
		Foreground(SubtleColor).
		Render("The following dependencies are missing:")

	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(subtitle))
	s.WriteString("\n\n")

	// List missing dependencies
	for _, dep := range m.missing {
		icon := "⚠️"
		if dep.Required {
			icon = "❌"
		}
		line := fmt.Sprintf("  %s  %s - %s", icon, dep.Name, dep.Description)
		s.WriteString(line + "\n")
	}

	s.WriteString("\n")

	// Install commands preview
	cmdStyle := lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Italic(true)

	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(
		cmdStyle.Render("Commands to run:"),
	))
	s.WriteString("\n")

	for _, dep := range m.missing {
		cmd := dependency.GetInstallCommand(dep)
		if cmd != "" {
			s.WriteString(fmt.Sprintf("  → %s\n", cmd))
		}
	}

	s.WriteString("\n\n")

	prompt := lipgloss.NewStyle().
		Bold(true).
		Foreground(HighlightColor).
		Render("Install missing dependencies? [Y]es / [N]o")

	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(prompt))

	return s.String()
}

func (m InstallerModel) renderInstalling() string {
	var s strings.Builder

	if m.currentIndex < len(m.missing) {
		dep := m.missing[m.currentIndex]
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(
			fmt.Sprintf("%s Installing %s...", m.spinner.View(), dep.Name),
		))
		s.WriteString("\n\n")

		// Progress bar
		percent := float64(m.currentIndex) / float64(len(m.missing))
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(
			m.progress.ViewAs(percent),
		))
	}

	// Install log (last 5 lines)
	if len(m.installLog) > 0 {
		s.WriteString("\n\n")
		logStyle := lipgloss.NewStyle().Foreground(SubtleColor)
		start := 0
		if len(m.installLog) > 5 {
			start = len(m.installLog) - 5
		}
		for _, line := range m.installLog[start:] {
			s.WriteString(logStyle.Render("  " + line + "\n"))
		}
	}

	return s.String()
}

func (m InstallerModel) renderComplete() string {
	var s strings.Builder

	if m.err != nil {
		errStyle := lipgloss.NewStyle().Foreground(ErrorColor)
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(
			errStyle.Render("⚠️ Some installations failed:"),
		))
		s.WriteString("\n")
		s.WriteString(errStyle.Render(m.err.Error()))
	} else if len(m.missing) == 0 {
		successStyle := lipgloss.NewStyle().Foreground(SuccessColor)
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(
			successStyle.Render("✓ All dependencies are installed!"),
		))
	} else {
		successStyle := lipgloss.NewStyle().Foreground(SuccessColor)
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(
			successStyle.Render("✓ Installation complete!"),
		))
	}

	s.WriteString("\n\n")
	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(
		"Press Enter to continue...",
	))

	return s.String()
}

// ShouldShowInstaller checks if installer should be shown
func ShouldShowInstaller() bool {
	mgr := dependency.NewManager()
	return mgr.HasMissingRequired()
}
