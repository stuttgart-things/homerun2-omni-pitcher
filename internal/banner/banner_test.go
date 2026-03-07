package banner

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func updateModel(m model, msg tea.Msg) model {
	updated, _ := m.Update(msg)
	return updated.(model)
}

func TestRenderStatic(t *testing.T) {
	output := renderStatic()
	if !strings.Contains(output, "██╗  ██╗") {
		t.Error("expected banner to contain HOMERUN block characters")
	}
	if !strings.Contains(output, "█████╔╝") {
		t.Error("expected banner to contain superscript 2")
	}
}

func TestModelGlitchPhase(t *testing.T) {
	m := initialModel()
	if !m.glitchPhase {
		t.Error("expected initial model to be in glitch phase")
	}

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view during glitch phase")
	}
}

func TestModelTransitionsOutOfGlitch(t *testing.T) {
	m := initialModel()

	for range 10 {
		m = updateModel(m, tickMsg{})
	}

	if m.glitchPhase {
		t.Error("expected model to exit glitch phase after 10 ticks")
	}
}

func TestModelAutoQuits(t *testing.T) {
	m := initialModel()

	for range 40 {
		m = updateModel(m, tickMsg{})
	}

	if !m.done {
		t.Error("expected model to be done after 40 ticks")
	}
}

func TestServiceRotation(t *testing.T) {
	m := initialModel()
	m.glitchPhase = false

	initial := m.serviceIdx

	for range 20 {
		m = updateModel(m, tickMsg{})
	}

	if m.serviceIdx == initial {
		t.Error("expected service index to rotate after 20 ticks")
	}
}

func TestApplyScanlines(t *testing.T) {
	input := "line0\nline1\nline2\nline3"
	output := applyScanlines(input)

	lines := strings.Split(output, "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(lines))
	}

	if lines[0] != "line0" {
		t.Errorf("expected even line unchanged, got '%s'", lines[0])
	}
}

func TestCenterText(t *testing.T) {
	output := centerText("hi", 20)
	if !strings.HasPrefix(output, "         ") {
		t.Error("expected centered text to have leading spaces")
	}
}

func TestKeyQuits(t *testing.T) {
	m := initialModel()
	m = updateModel(m, tea.KeyMsg{Type: tea.KeyEnter})

	if !m.done {
		t.Error("expected key press to set done=true")
	}
}
