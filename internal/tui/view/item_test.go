package view

import (
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/bojack/internal/msg"
	"github.com/henrrrik/bojack/internal/runway"
)

func TestTemplateDetailShowsFields(t *testing.T) {
	var v Viewer = NewTemplateDetailModel("valkey", testDetailStyles(), 100, 30)
	if out := v.View(); !strings.Contains(out, "Fetching template") {
		t.Errorf("before the result it should show a spinner, got:\n%s", out)
	}
	v, _ = v.Update(msg.ItemResultMsg[runway.Template]{Resource: msg.Resource{Name: "valkey"}, Item: &runway.Template{
		Name: "valkey", DisplayName: "Valkey", Category: "cache", Maintainer: "runway",
		MinPlan: "free-launch", Latest: "0.2.0", Versions: []string{"0.1.1", "0.2.0"},
		Description: "In-memory key-value store.\nNot reachable from the internet.\n",
	}})
	out := v.View()
	for _, want := range []string{"Template: valkey", "Valkey", "cache", "free-launch", "0.1.1, 0.2.0",
		"Not reachable from the internet."} {
		if !strings.Contains(out, want) {
			t.Errorf("view should contain %q, got:\n%s", want, out)
		}
	}
}

func TestVolumeDetailShowsFields(t *testing.T) {
	var v Viewer = NewVolumeDetailModel("data-1", testDetailStyles(), 100, 30)
	v, _ = v.Update(msg.ItemResultMsg[runway.Volume]{Resource: msg.Resource{Name: "data-1"}, Item: &runway.Volume{
		Name: "data-1", App: "my-db", Mountpoint: "/data", SizeMB: 5120, Resizable: true,
	}})
	out := v.View()
	for _, want := range []string{"Volume: data-1", "my-db", "/data", "5.0 GB", "Resizable:"} {
		if !strings.Contains(out, want) {
			t.Errorf("view should contain %q, got:\n%s", want, out)
		}
	}
}

func TestItemDetailShowsError(t *testing.T) {
	var v Viewer = NewVolumeDetailModel("nope", testDetailStyles(), 100, 30)
	v, _ = v.Update(msg.ItemResultMsg[runway.Volume]{Resource: msg.Resource{Name: "nope"}, Err: errors.New("volume does not exist")})
	if out := v.View(); !strings.Contains(out, "volume does not exist") {
		t.Errorf("view should show the error, got:\n%s", out)
	}
}

func TestTemplateListEnterOpensDetail(t *testing.T) {
	var v Viewer = NewTemplateListModel(testDetailStyles(), 100, 30)
	v, _ = v.Update(msg.ListResultMsg[runway.Template]{Items: []runway.Template{{Name: "valkey"}}})
	_, cmd := v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should send a navigate command")
	}
	if got := cmd(); got != (msg.NavigateMsg{View: msg.ViewTemplateDetail, Param: "valkey"}) {
		t.Errorf("cmd() = %#v, want navigate to the template detail", got)
	}
}

func TestVolumeListEnterOpensDetail(t *testing.T) {
	var v Viewer = NewVolumeListModel(testDetailStyles(), 100, 30)
	v, _ = v.Update(msg.ListResultMsg[runway.Volume]{Items: []runway.Volume{{Name: "data-1"}}})
	_, cmd := v.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter should send a navigate command")
	}
	if got := cmd(); got != (msg.NavigateMsg{View: msg.ViewVolumeDetail, Param: "data-1"}) {
		t.Errorf("cmd() = %#v, want navigate to the volume detail", got)
	}
}

func TestVolumeFieldsExtra(t *testing.T) {
	extra := func(v runway.Volume) string {
		for _, f := range volumeFields(v) {
			if f.label == "Extra:" {
				return f.value
			}
		}
		return "missing"
	}
	if got := extra(runway.Volume{}); got != "-" {
		t.Errorf("no extra storage should show -, got %q", got)
	}
	if got := extra(runway.Volume{SizeExtraMB: 1024}); got != "1.0 GB" {
		t.Errorf("extra = %q, want 1.0 GB", got)
	}
}
