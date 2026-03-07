package banner

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const bannerText = `██╗  ██╗ ██████╗ ███╗   ███╗███████╗██████╗ ██╗   ██╗███╗   ██╗
██║  ██║██╔═══██╗████╗ ████║██╔════╝██╔══██╗██║   ██║████╗  ██║
███████║██║   ██║██╔████╔██║█████╗  ██████╔╝██║   ██║██╔██╗ ██║
██╔══██║██║   ██║██║╚██╔╝██║██╔══╝  ██╔══██╗██║   ██║██║╚██╗██║
██║  ██║╚██████╔╝██║ ╚═╝ ██║███████╗██║  ██║╚██████╔╝██║ ╚████║
╚═╝  ╚═╝ ╚═════╝ ╚═╝     ╚═╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝`

const superscript2 = `██████╗
     ██║
█████╔╝
╚═══██╗
█████╔╝
╚════╝`

var services = []string{
	"omni-pitcher",
	"cluster-scout",
	"kube-slugger",
	"homerun-ui",
}

// Baseball animation frames — ball traveling across the field.
var baseballFrames = []string{
	"  ⚾                                                          ",
	"       ⚾                                                     ",
	"            ⚾                                                ",
	"                 ⚾                                           ",
	"                      ⚾                                      ",
	"                           ⚾                                 ",
	"                                ⚾                            ",
	"                                     ⚾                       ",
	"                                          ⚾                  ",
	"                                               ⚾             ",
	"                                                    ⚾        ",
	"                                                         ⚾   ",
}

// Glitch characters used during the intro phase.
const glitchChars = "░▒▓█▄▀▐▌╠╣╬═║╗╝╚╔"

var (
	greenGlow = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#39FF14")).
			Bold(true)

	dimGreen = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1a8a0e"))

	brightWhite = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	serviceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true)
)

type tickMsg time.Time

type model struct {
	width        int
	frame        int
	serviceIdx   int
	glitchPhase  bool
	glitchFrames int
	done         bool
}

// Show displays the animated banner for a brief duration, then returns.
// The banner runs for approximately 4 seconds.
func Show() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		// Non-fatal: if banner fails, just skip it.
		fmt.Println(greenGlow.Render(renderStatic()))
	}
}

func initialModel() model {
	return model{
		width:        80,
		glitchPhase:  true,
		glitchFrames: 0,
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.done = true
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil

	case tickMsg:
		_ = msg
		m.frame++

		// Glitch phase: ~1 second (10 frames at 100ms)
		if m.glitchPhase {
			m.glitchFrames++
			if m.glitchFrames >= 10 {
				m.glitchPhase = false
			}
			return m, tickCmd()
		}

		// Rotate services every ~2 seconds (20 frames)
		if m.frame%20 == 0 {
			m.serviceIdx = (m.serviceIdx + 1) % len(services)
		}

		// Auto-quit after ~4 seconds total (40 frames)
		if m.frame >= 40 {
			m.done = true
			return m, tea.Quit
		}

		return m, tickCmd()
	}

	return m, nil
}

func (m model) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder

	// Render banner
	var bannerOutput string
	if m.glitchPhase {
		bannerOutput = glitchText(bannerText, m.glitchFrames)
	} else {
		bannerOutput = greenGlow.Render(bannerText)
	}

	// Render superscript 2
	var sup2Output string
	if m.glitchPhase {
		sup2Output = glitchText(superscript2, m.glitchFrames)
	} else {
		sup2Output = brightWhite.Render(superscript2)
	}

	// Compose banner + superscript side by side
	bannerLines := strings.Split(bannerOutput, "\n")
	sup2Lines := strings.Split(sup2Output, "\n")

	for i, line := range bannerLines {
		b.WriteString(line)
		if i < len(sup2Lines) {
			b.WriteString(" ")
			b.WriteString(sup2Lines[i])
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Baseball animation
	if !m.glitchPhase {
		ballIdx := m.frame % len(baseballFrames)
		b.WriteString(dimGreen.Render(baseballFrames[ballIdx]))
		b.WriteString("\n\n")
	} else {
		b.WriteString("\n\n")
	}

	// Service display
	svc := services[m.serviceIdx]
	serviceLine := fmt.Sprintf("[ %s ]", svc)
	b.WriteString(serviceStyle.Render(serviceLine))
	b.WriteString("\n")

	// Apply CRT scanline effect
	output := applyScanlines(b.String())

	// Center in terminal
	return centerText(output, m.width)
}

// glitchText replaces random characters in the text with glitch chars,
// decreasing over time as glitchFrame increases.
func glitchText(text string, glitchFrame int) string {
	// More glitch early, less as it stabilizes
	glitchProbability := float64(10-glitchFrame) / 10.0
	if glitchProbability < 0 {
		glitchProbability = 0
	}

	runes := []rune(text)
	glitchRunes := []rune(glitchChars)
	result := make([]rune, len(runes))

	for i, r := range runes {
		if r == '\n' || r == ' ' {
			result[i] = r
			continue
		}
		if rand.Float64() < glitchProbability {
			result[i] = glitchRunes[rand.IntN(len(glitchRunes))]
		} else {
			result[i] = r
		}
	}

	return greenGlow.Render(string(result))
}

// applyScanlines dims every other line for a CRT monitor effect.
func applyScanlines(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if i%2 == 1 {
			lines[i] = dimGreen.Render(line)
		}
	}
	return strings.Join(lines, "\n")
}

// centerText centers each line of text within the given width.
func centerText(text string, width int) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		// Strip ANSI for length calculation
		visLen := lipgloss.Width(line)
		if visLen < width {
			pad := (width - visLen) / 2
			lines[i] = strings.Repeat(" ", pad) + line
		}
	}
	return strings.Join(lines, "\n")
}

// renderStatic returns the banner without animation (fallback).
func renderStatic() string {
	bannerLines := strings.Split(bannerText, "\n")
	sup2Lines := strings.Split(superscript2, "\n")

	var b strings.Builder
	for i, line := range bannerLines {
		b.WriteString(line)
		if i < len(sup2Lines) {
			b.WriteString(" ")
			b.WriteString(sup2Lines[i])
		}
		b.WriteString("\n")
	}
	return b.String()
}
