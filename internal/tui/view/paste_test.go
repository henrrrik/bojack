package view

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/bojack/internal/msg"
	"github.com/henrrrik/bojack/internal/runway"
)

func TestPasteIntoFilter(t *testing.T) {
	var v Viewer = NewTemplateListModel(testDetailStyles(), 100, 30)
	v, _ = v.Update(msg.ListResultMsg[runway.Template]{Items: []runway.Template{{Name: "valkey"}, {Name: "mongodb"}}})
	v, _ = v.Update(key("/"))
	v, _ = v.Update(tea.PasteMsg{Content: "mongo"})
	if out := v.View(); !strings.Contains(out, "Filter: mongo") || strings.Contains(out, "valkey") {
		t.Errorf("pasted text should go into the filter, got:\n%s", out)
	}
}

func TestPasteIgnoredWhenNotFiltering(t *testing.T) {
	var v Viewer = NewTemplateListModel(testDetailStyles(), 100, 30)
	v, _ = v.Update(msg.ListResultMsg[runway.Template]{Items: []runway.Template{{Name: "valkey"}}})
	v, _ = v.Update(tea.PasteMsg{Content: "mongo"})
	if out := v.View(); strings.Contains(out, "Filter:") {
		t.Errorf("paste outside the filter should do nothing, got:\n%s", out)
	}
}

func TestPasteIntoActionInput(t *testing.T) {
	m := appOnTab(appConfigTab)
	m, _ = m.Update(key("S"))
	m, _ = m.Update(tea.PasteMsg{Content: "FOO=bar"})
	m, _ = m.Update(enterKey)
	if !strings.Contains(m.View(), "Set FOO on app terry?") {
		t.Errorf("pasted text should go into the action input, got:\n%s", m.View())
	}
	v := loadedVolume(runway.Volume{Name: "data-1", SizeMB: 100, Resizable: true})
	v, _ = v.Update(key("S"))
	v, _ = v.Update(tea.PasteMsg{Content: "200"})
	v, _ = v.Update(enterKey)
	if !strings.Contains(v.View(), "to 200 MB?") {
		t.Errorf("pasted size should go into the volume input, got:\n%s", v.View())
	}
	s := NewServiceDetailModel("db", testDetailStyles(), 120, 30)
	s.activeTab = serviceConfigTab
	s, _ = s.Update(key("D"))
	s, _ = s.Update(tea.PasteMsg{Content: "FOO"})
	s, _ = s.Update(enterKey)
	if !strings.Contains(s.View(), "Remove FOO from service db?") {
		t.Errorf("pasted key should go into the service input, got:\n%s", s.View())
	}
}
