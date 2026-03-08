package banner

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
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
	if !strings.Contains(output, "2") {
		t.Error("expected banner to contain '2'")
	}
	if !strings.Contains(output, "██████╗") {
		t.Error("expected banner to contain OMNI-PITCHER block text")
	}
}

func TestRenderHeader(t *testing.T) {
	output := renderHeader()
	if !strings.Contains(output, "██╗  ██╗") {
		t.Error("expected header to contain HOMERUN block characters")
	}
	if output == "" {
		t.Error("expected non-empty header")
	}
}

func TestModelGlitchPhase(t *testing.T) {
	m := initialModel()
	if !m.glitchPhase {
		t.Error("expected initial model to be in glitch phase")
	}

	view := m.View()
	if view.Content == "" {
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

func TestBaseballAnimation(t *testing.T) {
	m := initialModel()
	m.glitchPhase = false

	view := m.View()
	// Should contain the baseball art characters
	if !strings.Contains(view.Content, "()") {
		t.Error("expected view to contain baseball art")
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
	m = updateModel(m, tea.KeyPressMsg{Code: tea.KeyEnter})

	if !m.done {
		t.Error("expected key press to set done=true")
	}
}
