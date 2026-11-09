package view

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/bojack/internal/msg"
	"github.com/henrrrik/bojack/internal/runway"
)

var enterKey = tea.KeyPressMsg{Code: tea.KeyEnter}

func appOnTab(tab int) AppDetailModel {
	m := NewAppDetailModel("terry", testDetailStyles(), 120, 30)
	m.activeTab = tab
	return m
}

func TestAppConfigSetFlow(t *testing.T) {
	m := appOnTab(appConfigTab)
	if !strings.Contains(m.View(), "S: set") || !strings.Contains(m.View(), "D: remove") {
		t.Errorf("config tab help should offer S and D, got:\n%s", m.View())
	}
	m, cmd := m.Update(key("S"))
	if cmd != nil || !strings.Contains(m.View(), "KEY=value:") {
		t.Fatalf("S should ask for input, got:\n%s", m.View())
	}
	m, _ = m.Update(key("FOO=bar baz"))
	if !strings.Contains(m.View(), "FOO=bar baz") {
		t.Errorf("typed text should be shown, got:\n%s", m.View())
	}
	m, cmd = m.Update(enterKey)
	if cmd != nil || !strings.Contains(m.View(), "Set FOO on app terry? (y/N)") {
		t.Fatalf("enter should ask for confirmation, got:\n%s", m.View())
	}
	_, cmd = m.Update(key("y"))
	want := msg.ActionRequestMsg{Kind: msg.ActionSetAppConfig, Target: "terry", Arg: "FOO=bar baz"}
	if req := requestOf(t, cmd); req != want {
		t.Errorf("request = %+v, want %+v", req, want)
	}
}

func TestConfigSetMasksSecretInput(t *testing.T) {
	m := appOnTab(appConfigTab)
	m, _ = m.Update(key("S"))
	m, _ = m.Update(key("API_TOKEN=hunter2"))
	if strings.Contains(m.View(), "hunter2") || !strings.Contains(m.View(), "API_TOKEN=") {
		t.Errorf("a secret value should be masked while typing, got:\n%s", m.View())
	}
	m, _ = m.Update(enterKey)
	if v := m.View(); strings.Contains(v, "hunter2") || !strings.Contains(v, "Set API_TOKEN on app terry?") {
		t.Errorf("the question should name only the key, got:\n%s", v)
	}
}

func TestConfigSetRejectsBadInput(t *testing.T) {
	for _, in := range []string{"novalue", "=x", "-flag=1", "BAD KEY=1"} {
		m := appOnTab(appConfigTab)
		m, _ = m.Update(key("S"))
		m, _ = m.Update(key(in))
		m, cmd := m.Update(enterKey)
		if cmd != nil || strings.Contains(m.View(), "(y/N)") || !strings.Contains(m.View(), "KEY=value") {
			t.Errorf("%q: should stay in input with an error, got:\n%s", in, m.View())
		}
	}
}

func TestInputEditingAndCancel(t *testing.T) {
	m := appOnTab(appConfigTab)
	m, _ = m.Update(key("S"))
	if !m.ClaimsKey("q") || !m.ClaimsKey("esc") {
		t.Error("while typing the view should claim every key")
	}
	m, _ = m.Update(key("FOOX"))
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	m, _ = m.Update(key("R=1"))
	m, _ = m.Update(enterKey)
	if !strings.Contains(m.View(), "Set FOOR on app terry?") {
		t.Errorf("backspace should edit, and R should type, got:\n%s", m.View())
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	m, _ = m.Update(key("S"))
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if strings.Contains(m.View(), "KEY=value:") || m.ClaimsKey("q") {
		t.Errorf("esc should cancel the input, got:\n%s", m.View())
	}
}

func TestAppConfigRemove(t *testing.T) {
	m := appOnTab(appConfigTab)
	m, _ = m.Update(key("D"))
	m, _ = m.Update(key("FOO"))
	m, _ = m.Update(enterKey)
	if !strings.Contains(m.View(), "Remove FOO from app terry? (y/N)") {
		t.Fatalf("got:\n%s", m.View())
	}
	_, cmd := m.Update(key("y"))
	want := msg.ActionRequestMsg{Kind: msg.ActionRemoveAppConfig, Target: "terry", Arg: "FOO"}
	if req := requestOf(t, cmd); req != want {
		t.Errorf("request = %+v, want %+v", req, want)
	}
}

func TestConfigActionsOnlyOnConfigTab(t *testing.T) {
	m := appOnTab(0)
	if strings.Contains(m.View(), "S: set") {
		t.Error("info tab should not offer config actions")
	}
	if m, _ = m.Update(key("S")); strings.Contains(m.View(), "KEY=value") {
		t.Error("S on the info tab should do nothing")
	}
}

func TestAppPlanChange(t *testing.T) {
	m := appOnTab(0)
	if !strings.Contains(m.View(), "P: plan") {
		t.Errorf("info tab should offer P: plan, got:\n%s", m.View())
	}
	m, _ = m.Update(key("P"))
	m, _ = m.Update(key("starter"))
	m, _ = m.Update(enterKey)
	if !strings.Contains(m.View(), "Move app terry to plan starter?") {
		t.Fatalf("got:\n%s", m.View())
	}
	_, cmd := m.Update(key("y"))
	want := msg.ActionRequestMsg{Kind: msg.ActionSetAppPlan, Target: "terry", Arg: "starter"}
	if req := requestOf(t, cmd); req != want {
		t.Errorf("request = %+v, want %+v", req, want)
	}
}

func TestServiceConfigSet(t *testing.T) {
	m := NewServiceDetailModel("db", testDetailStyles(), 120, 30)
	m.activeTab = serviceConfigTab
	m, _ = m.Update(key("S"))
	m, _ = m.Update(key("FOO=1"))
	m, _ = m.Update(enterKey)
	_, cmd := m.Update(key("y"))
	want := msg.ActionRequestMsg{Kind: msg.ActionSetServiceConfig, Target: "db", Arg: "FOO=1"}
	if req := requestOf(t, cmd); req != want {
		t.Errorf("request = %+v, want %+v", req, want)
	}
	m = NewServiceDetailModel("db", testDetailStyles(), 120, 30)
	m.activeTab = serviceConfigTab
	m, _ = m.Update(key("D"))
	m, _ = m.Update(key("FOO"))
	m, _ = m.Update(enterKey)
	_, cmd = m.Update(key("y"))
	want = msg.ActionRequestMsg{Kind: msg.ActionRemoveServiceConfig, Target: "db", Arg: "FOO"}
	if req := requestOf(t, cmd); req != want {
		t.Errorf("request = %+v, want %+v", req, want)
	}
}

func loadedVolume(v runway.Volume) Viewer {
	var m Viewer = NewVolumeDetailModel(v.Name, testDetailStyles(), 120, 30)
	m, _ = m.Update(msg.ItemResultMsg[runway.Volume]{Resource: msg.Resource{Name: v.Name}, Item: &v})
	return m
}

func TestVolumeResize(t *testing.T) {
	m := loadedVolume(runway.Volume{Name: "data-1", SizeMB: 5120, Resizable: true})
	if !strings.Contains(m.View(), "S: resize") {
		t.Errorf("a resizable volume should offer S: resize, got:\n%s", m.View())
	}
	m, _ = m.Update(key("S"))
	m, _ = m.Update(key("4096"))
	m, cmd := m.Update(enterKey)
	if cmd != nil || !strings.Contains(m.View(), "more than 5120") {
		t.Fatalf("a smaller size should be rejected, got:\n%s", m.View())
	}
	for range 4 {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	}
	m, _ = m.Update(key("6144"))
	m, _ = m.Update(enterKey)
	if !strings.Contains(m.View(), "Grow volume data-1 from 5.0 GB to 6.0 GB?") {
		t.Fatalf("got:\n%s", m.View())
	}
	m, cmd = m.Update(key("y"))
	want := msg.ActionRequestMsg{Kind: msg.ActionResizeVolume, Target: "data-1", Arg: "6144"}
	if req := requestOf(t, cmd); req != want {
		t.Errorf("request = %+v, want %+v", req, want)
	}
	m, _ = m.Update(msg.ActionResultMsg{Kind: msg.ActionResizeVolume, Target: "data-1"})
	if !strings.Contains(m.View(), "Resized data-1") {
		t.Errorf("result should be shown, got:\n%s", m.View())
	}
}

func TestVolumeNotResizable(t *testing.T) {
	m := loadedVolume(runway.Volume{Name: "data-1", SizeMB: 5120})
	if strings.Contains(m.View(), "S: resize") {
		t.Error("a volume that is not resizable should not offer resize")
	}
	if m, _ = m.Update(key("S")); strings.Contains(m.View(), "MiB") {
		t.Errorf("S should do nothing, got:\n%s", m.View())
	}
}
