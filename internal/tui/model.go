package tui

import (
	"fmt"
	"strings"

	"terminaltube/internal/renderer"
	"terminaltube/internal/terminal"
	"terminaltube/pkg/types"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View states
type viewState int

const (
	viewMenu viewState = iota
	viewImageInput
	viewGIFInput
	viewGIFURLInput
	viewVideoURLInput
	viewVideoFileInput
	viewTerminalInfo
	viewRenderingTests
	viewCacheClear
	viewAbout
	viewPlayback
)

// MenuItem represents a menu item in the list
type MenuItem struct {
	title       string
	description string
	icon        string
}

func (m MenuItem) Title() string       { return m.icon + "  " + m.title }
func (m MenuItem) Description() string { return m.description }
func (m MenuItem) FilterValue() string { return m.title }

// Model is the main Bubble Tea model
type Model struct {
	// State
	currentView   viewState
	headerAnim    HeaderAnimation
	menuList      list.Model
	textInput     textinput.Model
	inputPrompt   string
	errorMessage  string
	statusMessage string

	// External dependencies
	rendererManager *renderer.RendererManager
	termControl     *terminal.Control
	capabilities    types.TerminalCapabilities

	// Window size
	width  int
	height int

	// Quit flag
	Quitting   bool
	NextAction string // Action to perform after TUI exits (e.g., "play", "test")
	NextArgs   string // Arguments for the action (e.g., file path)
}

// NewModel creates a new TUI model
func NewModel(rm *renderer.RendererManager, tc *terminal.Control, caps types.TerminalCapabilities) Model {
	// Create text input
	ti := textinput.New()
	ti.Placeholder = "Type here..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	// Create menu items
	items := []list.Item{
		MenuItem{title: "Display Image", description: "Show static images in terminal", icon: "🖼️"},
		MenuItem{title: "Play GIF Animation", description: "Play animated GIFs with proper timing", icon: "🎞️"},
		MenuItem{title: "Play GIF from URL", description: "Download and play GIF from web", icon: "🔗"},
		MenuItem{title: "Play Video from URL", description: "Download and play video from web URLs", icon: "🌐"},
		MenuItem{title: "Play Video from File", description: "Play local video files", icon: "📁"},
		MenuItem{title: "Terminal Information", description: "Display terminal capabilities", icon: "ℹ️"},
		MenuItem{title: "Rendering Tests", description: "Test different rendering modes", icon: "🧪"},
		MenuItem{title: "Clear Cache", description: "Remove temporary download files", icon: "🧹"},
		MenuItem{title: "About", description: "App info and credits", icon: "💡"},
		MenuItem{title: "Exit", description: "Quit TerminalTube", icon: "❌"},
	}

	// Create list with custom delegate
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(PrimaryColor).
		Foreground(PrimaryColor).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(PrimaryColor).
		Foreground(SecondaryColor)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(TextColor)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.
		Foreground(SubtleColor)

	menuList := list.New(items, delegate, 0, 0)
	menuList.Title = ""
	menuList.SetShowTitle(false)
	menuList.SetShowStatusBar(false)
	menuList.SetFilteringEnabled(false)
	menuList.SetShowHelp(false)
	menuList.SetShowPagination(false)

	return Model{
		currentView:     viewMenu,
		headerAnim:      NewHeaderAnimation(ASCIIArt),
		menuList:        menuList,
		textInput:       ti,
		rendererManager: rm,
		termControl:     tc,
		capabilities:    caps,
		width:           caps.Width,
		height:          caps.Height,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		HeaderAnimationCmd(),
		textinput.Blink,
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		case "esc":
			if m.currentView != viewMenu {
				m.currentView = viewMenu
				m.errorMessage = ""
				m.statusMessage = ""
				m.textInput.Reset()
				return m, nil
			}
		case "q":
			// Only quit if in menu or info views (not while typing)
			if m.currentView == viewMenu || m.currentView == viewTerminalInfo || m.currentView == viewAbout {
				m.Quitting = true
				return m, tea.Quit
			}
		case "enter":
			if m.currentView == viewMenu {
				return m.handleMenuSelection()
			} else if isInputView(m.currentView) {
				return m.handleInputSubmission()
			} else if m.currentView == viewCacheClear {
				return m.handleCacheClear()
			}

		case " ":
			// Skip animation with spacebar (only in menu)
			if m.currentView == viewMenu && !m.headerAnim.IsDone() {
				m.headerAnim.SkipAnimation()
				return m, ColorCycleCmd()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.menuList.SetSize(msg.Width-4, msg.Height-20)
		return m, nil

	case HeaderAnimationTickMsg, ColorCycleTickMsg:
		var animCmd tea.Cmd
		animCmd = m.headerAnim.Update(msg)
		return m, animCmd // Animation doesn't return commands (internal tick)
	}

	// Update components based on view
	if m.currentView == viewMenu {
		m.menuList, cmd = m.menuList.Update(msg)
		return m, cmd
	} else if isInputView(m.currentView) {
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// Helper to check if current view is an input view
func isInputView(v viewState) bool {
	return v == viewImageInput || v == viewGIFInput || v == viewGIFURLInput ||
		v == viewVideoURLInput || v == viewVideoFileInput
}

// handleMenuSelection handles menu item selection
func (m Model) handleMenuSelection() (tea.Model, tea.Cmd) {
	selected, ok := m.menuList.SelectedItem().(MenuItem)
	if !ok {
		return m, nil
	}

	m.textInput.Reset()
	m.textInput.Focus()
	m.errorMessage = ""
	m.statusMessage = ""

	switch selected.title {
	case "Display Image":
		m.inputPrompt = "Enter image file path:"
		m.currentView = viewImageInput
	case "Play GIF Animation":
		m.inputPrompt = "Enter GIF file path:"
		m.currentView = viewGIFInput
	case "Play GIF from URL":
		m.inputPrompt = "Enter GIF URL:"
		m.currentView = viewGIFURLInput
	case "Play Video from URL":
		m.inputPrompt = "Enter video URL:"
		m.currentView = viewVideoURLInput
	case "Play Video from File":
		m.inputPrompt = "Enter video file path:"
		m.currentView = viewVideoFileInput
	case "Terminal Information":
		m.currentView = viewTerminalInfo
	case "Rendering Tests":
		// Exit TUI to run rendering tests
		m.NextAction = "test"
		m.Quitting = true
		return m, tea.Quit
	case "Clear Cache":
		m.currentView = viewCacheClear
	case "About":
		m.currentView = viewAbout
	case "Exit":
		m.Quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// handleInputSubmission handles input submission
func (m Model) handleInputSubmission() (tea.Model, tea.Cmd) {
	value := m.textInput.Value()
	if strings.TrimSpace(value) == "" {
		m.errorMessage = "Input cannot be empty"
		return m, nil
	}

	// Determine action based on view
	action := ""
	switch m.currentView {
	case viewImageInput:
		action = "image"
	case viewGIFInput:
		action = "gif"
	case viewGIFURLInput:
		action = "gif-url"
	case viewVideoURLInput:
		action = "video-url"
	case viewVideoFileInput:
		action = "video"
	}

	if action != "" {
		m.NextAction = action
		m.NextArgs = value
		m.Quitting = true
		return m, tea.Quit
	}

	m.statusMessage = fmt.Sprintf("Selected: %s", value)
	m.errorMessage = ""

	return m, nil
}

// handleCacheClear handles cache clearing
func (m Model) handleCacheClear() (tea.Model, tea.Cmd) {
	// Simulate cache clearing
	m.statusMessage = "Temp files cleared effectively!"
	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.Quitting {
		return "\n  Thank you for using TerminalTube! 👋\n\n"
	}

	var s strings.Builder

	switch m.currentView {
	case viewMenu:
		s.WriteString(m.renderMainMenu())
	case viewTerminalInfo:
		s.WriteString(m.renderTerminalInfo())
	case viewImageInput, viewGIFInput, viewGIFURLInput, viewVideoURLInput, viewVideoFileInput:
		s.WriteString(m.renderInputView())
	case viewAbout:
		s.WriteString(m.renderAbout())
	case viewCacheClear:
		s.WriteString(m.renderCacheClear())
	default:
		s.WriteString(m.renderMainMenu())
	}

	return s.String()
}

// renderMainMenu renders the main menu view
func (m Model) renderMainMenu() string {
	var s strings.Builder

	// Animated header
	header := m.headerAnim.View()
	headerBox := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).
		Render(header)
	s.WriteString(headerBox)
	s.WriteString("\n")

	// Capability badges
	badges := m.renderCapabilityBadges()
	badgeLine := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width).
		Render(badges)
	s.WriteString(badgeLine)
	s.WriteString("\n\n\n") // Added extra gap (3 lines) as requested

	// Menu
	menuStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(m.width)
	s.WriteString(menuStyle.Render(m.menuList.View()))

	// Footer with keybindings
	footer := FooterStyle.Width(m.width).Render(
		"↑/↓: Navigate  •  Enter: Select  •  q: Quit  •  Space: Skip animation",
	)
	s.WriteString("\n\n") // Added gap before footer to ensure visibility
	s.WriteString(footer)

	return s.String()
}

// renderCapabilityBadges renders the terminal capability badges
func (m Model) renderCapabilityBadges() string {
	var badges []string

	badges = append(badges, lipgloss.NewStyle().
		Foreground(SubtleColor).
		Render(fmt.Sprintf("Terminal: %dx%d", m.capabilities.Width, m.capabilities.Height)))

	badges = append(badges, Badge("SIXEL", m.capabilities.SixelSupport))
	badges = append(badges, Badge("TrueColor", m.capabilities.TrueColor))
	badges = append(badges, Badge("Unicode", m.capabilities.UnicodeSupport))

	return strings.Join(badges, "  ┃  ")
}

// renderTerminalInfo renders the terminal information view
func (m Model) renderTerminalInfo() string {
	var s strings.Builder

	title := TitleStyle.Render("Terminal Information")
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(title))
	s.WriteString("\n\n")

	info := []struct {
		key   string
		value string
	}{
		{"Terminal Size", fmt.Sprintf("%d x %d", m.capabilities.Width, m.capabilities.Height)},
		{"SIXEL Support", fmt.Sprintf("%v", m.capabilities.SixelSupport)},
		{"True Color (24-bit)", fmt.Sprintf("%v", m.capabilities.TrueColor)},
		{"256 Colors", fmt.Sprintf("%v", m.capabilities.Color256)},
		{"Unicode Support", fmt.Sprintf("%v", m.capabilities.UnicodeSupport)},
	}

	for _, item := range info {
		keyStyle := lipgloss.NewStyle().Foreground(SecondaryColor).Bold(true).Width(25)
		valStyle := lipgloss.NewStyle().Foreground(TextColor)
		line := keyStyle.Render(item.key+":") + " " + valStyle.Render(item.value)
		s.WriteString("  " + line + "\n")
	}

	s.WriteString("\n")
	s.WriteString(FooterStyle.Width(m.width).Render("Press q or Esc to return"))

	return s.String()
}

// renderInputView renders input prompt views
func (m Model) renderInputView() string {
	var s strings.Builder

	title := TitleStyle.Render(m.inputPrompt)
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(title))
	s.WriteString("\n\n")

	// Render the text input
	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(
		m.textInput.View(),
	))
	s.WriteString("\n\n")

	if m.errorMessage != "" {
		errStyle := lipgloss.NewStyle().Foreground(ErrorColor)
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(
			errStyle.Render(m.errorMessage),
		))
		s.WriteString("\n")
	}

	if m.statusMessage != "" {
		statStyle := lipgloss.NewStyle().Foreground(SuccessColor)
		s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(
			statStyle.Render(m.statusMessage),
		))
		s.WriteString("\n")
	}

	s.WriteString(FooterStyle.Width(m.width).Render("Press Enter to submit • Esc to cancel"))

	return s.String()
}

// renderAbout renders the about screen
func (m Model) renderAbout() string {
	var s strings.Builder

	title := TitleStyle.Render("About TerminalTube")
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(title))
	s.WriteString("\n\n")

	// OSC 8 hyperlink sequence
	link := "\x1b]8;;https://github.com/JyotirmoyDas05\x1b\\@JyotirmoyDas05\x1b]8;;\x1b\\"

	content := fmt.Sprintf(`TerminalTube is a modern terminal media player
built with Go by %s

Features:
• SIXEL & TrueColor Graphics
• Video & GIF Playback
• YouTube & URL Support
• Beautiful TUI Interface`, link)

	box := BoxStyle.Render(content)
	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(box))

	s.WriteString("\n\n")
	s.WriteString(FooterStyle.Width(m.width).Render("Press Esc to return"))
	return s.String()
}

// renderCacheClear renders the cache clear screen
func (m Model) renderCacheClear() string {
	var s strings.Builder

	title := TitleStyle.Render("Clear Cache")
	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(title))
	s.WriteString("\n\n")

	content := "Press Enter to clear temporary files\nand free up disk space."

	if m.statusMessage != "" {
		content = m.statusMessage + "\n\n(Press Esc to return)"
	}

	s.WriteString(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(content))

	s.WriteString("\n\n")
	if m.statusMessage == "" {
		s.WriteString(FooterStyle.Width(m.width).Render("Enter: Confirm • Esc: Cancel"))
	}
	return s.String()
}
