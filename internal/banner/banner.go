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

const serviceText = ` ██████╗ ███╗   ███╗███╗   ██╗██╗      ██████╗ ██╗████████╗ ██████╗██╗  ██╗███████╗██████╗
██╔═══██╗████╗ ████║████╗  ██║██║      ██╔══██╗██║╚══██╔══╝██╔════╝██║  ██║██╔════╝██╔══██╗
██║   ██║██╔████╔██║██╔██╗ ██║██║█████╗██████╔╝██║   ██║   ██║     ███████║█████╗  ██████╔╝
██║   ██║██║╚██╔╝██║██║╚██╗██║██║╚════╝██╔═══╝ ██║   ██║   ██║     ██╔══██║██╔══╝  ██╔══██╗
╚██████╔╝██║ ╚═╝ ██║██║ ╚████║██║      ██║     ██║   ██║   ╚██████╗██║  ██║███████╗██║  ██║
 ╚═════╝ ╚═╝     ╚═╝╚═╝  ╚═══╝╚═╝      ╚═╝     ╚═╝   ╚═╝    ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝`

// Bigger baseball art for the animation.
var baseballArt = []string{
	"    ___     ",
	"  /     \\   ",
	" | () () |  ",
	"  \\ ___ /   ",
	"    ~~~     ",
}

const fieldWidth = 70

// Glitch characters used during the intro phase.
const glitchChars = "░▒▓█▄▀▐▌╠╣╬═║╗╝╚╔"

var (
	// Orange gradient — Street Fighter style
	orangeHot = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6600")).
			Bold(true)

	orangeBright = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF9900")).
			Bold(true)

	dimOrange = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#994400"))

	// The "2" in a contrasting yellow/white
	accentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFCC00")).
			Bold(true)

	serviceBlockStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF4400")).
				Bold(true)

	ballStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
)

type tickMsg time.Time

type model struct {
	width        int
	frame        int
	glitchPhase  bool
	glitchFrames int
	done         bool
}

// Show displays the animated banner for a brief duration, then prints
// the final frame as a persistent header before returning.
func Show() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		// Non-fatal: if banner fails, just print static.
	}
	// Print the final banner as a persistent header for the running program
	fmt.Println(renderHeader())
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

	// Render HOMERUN banner
	var bannerOutput string
	if m.glitchPhase {
		bannerOutput = glitchText(bannerText, m.glitchFrames)
	} else {
		bannerOutput = orangeHot.Render(bannerText)
	}

	b.WriteString(bannerOutput)

	// Append "2" on the same last line, in accent color
	the2 := accentStyle.Render("2")
	if m.glitchPhase && m.glitchFrames < 8 {
		glitchRunes := []rune(glitchChars)
		the2 = accentStyle.Render(string(glitchRunes[rand.IntN(len(glitchRunes))]))
	}
	b.WriteString(the2)
	b.WriteString("\n\n")

	// Baseball animation — bigger ball sliding across
	if !m.glitchPhase {
		ballPos := (m.frame * 3) % fieldWidth
		for _, artLine := range baseballArt {
			pad := strings.Repeat(" ", ballPos)
			b.WriteString(ballStyle.Render(pad + artLine))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	} else {
		b.WriteString("\n\n")
	}

	// OMNI-PITCHER in big block letters
	var svcOutput string
	if m.glitchPhase {
		svcOutput = glitchText(serviceText, m.glitchFrames)
	} else {
		svcOutput = serviceBlockStyle.Render(serviceText)
	}
	b.WriteString(svcOutput)
	b.WriteString("\n")

	// Apply CRT scanline effect
	output := applyScanlines(b.String())

	// Center in terminal
	return centerText(output, m.width)
}

// glitchText replaces random characters in the text with glitch chars,
// decreasing over time as glitchFrame increases.
func glitchText(text string, glitchFrame int) string {
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

	return orangeBright.Render(string(result))
}

// applyScanlines dims every other line for a CRT monitor effect.
func applyScanlines(text string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if i%2 == 1 {
			lines[i] = dimOrange.Render(line)
		}
	}
	return strings.Join(lines, "\n")
}

// centerText centers each line of text within the given width.
func centerText(text string, width int) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		visLen := lipgloss.Width(line)
		if visLen < width {
			pad := (width - visLen) / 2
			lines[i] = strings.Repeat(" ", pad) + line
		}
	}
	return strings.Join(lines, "\n")
}

// renderHeader returns the colored banner as a persistent header for the terminal.
func renderHeader() string {
	var b strings.Builder
	b.WriteString(orangeHot.Render(bannerText))
	b.WriteString(accentStyle.Render("2"))
	b.WriteString("\n")
	b.WriteString(serviceBlockStyle.Render(serviceText))
	b.WriteString("\n")
	return b.String()
}

// renderStatic returns the banner without colors (fallback).
func renderStatic() string {
	var b strings.Builder
	b.WriteString(bannerText)
	b.WriteString("2\n")
	b.WriteString(serviceText)
	b.WriteString("\n")
	return b.String()
}
