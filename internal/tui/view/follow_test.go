package view

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/henrrrik/runway-tui/internal/msg"
	"github.com/henrrrik/runway-tui/internal/runway"
)

// instantTicks makes follow ticks fire immediately and records their IDs.
func instantTicks(t *testing.T) {
	t.Helper()
	orig := followTick
	followTick = func(id int) tea.Cmd { return func() tea.Msg { return msg.FollowTickMsg{ID: id} } }
	t.Cleanup(func() { followTick = orig })
}

// cmdMsgs runs cmd (unwrapping batches) and returns the messages it produces.
func cmdMsgs(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	m := cmd()
	if b, ok := m.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, c := range b {
			out = append(out, cmdMsgs(c)...)
		}
		return out
	}
	return []tea.Msg{m}
}

func followingStats(t *testing.T) (AppDetailModel, msg.FollowTickMsg) {
	t.Helper()
	m := statsModel(runway.Stats{"cpu": {Title: "CPU", Data: []runway.StatsPoint{{Value: 1}}}})
	m, cmd := m.Update(key("f"))
	msgs := cmdMsgs(cmd)
	if len(msgs) != 1 {
		t.Fatalf("f should schedule one tick, got %#v", msgs)
	}
	tick, ok := msgs[0].(msg.FollowTickMsg)
	if !ok {
		t.Fatalf("f should schedule a FollowTickMsg, got %#v", msgs[0])
	}
	return m, tick
}

func TestFollowTogglesOnStatsTab(t *testing.T) {
	instantTicks(t)
	m, _ := followingStats(t)
	v := m.View()
	if !strings.Contains(v, "Auto-refresh: every 15s") || !strings.Contains(v, "f: stop following") {
		t.Errorf("should show follow state, got:\n%s", v)
	}
	m, cmd := m.Update(key("f"))
	if cmd != nil || strings.Contains(m.View(), "Auto-refresh") || !strings.Contains(m.View(), "f: follow") {
		t.Errorf("second f should stop following, got:\n%s", m.View())
	}
}

func TestFollowTickRefreshesStatsWithoutLoading(t *testing.T) {
	instantTicks(t)
	m, tick := followingStats(t)
	m, cmd := m.Update(tick)
	var gotFetch, gotTick bool
	for _, x := range cmdMsgs(cmd) {
		switch x := x.(type) {
		case msg.StatsIntervalMsg:
			gotFetch = x.Interval == DefaultStatsInterval
		case msg.FollowTickMsg:
			gotTick = x.ID == tick.ID
		}
	}
	if !gotFetch || !gotTick {
		t.Errorf("tick should refetch stats and schedule the next tick (fetch=%v tick=%v)", gotFetch, gotTick)
	}
	if strings.Contains(m.View(), "Loading") {
		t.Error("a refresh should keep showing the current data")
	}
}

func TestFollowTickRefreshesLogs(t *testing.T) {
	instantTicks(t)
	m := logsModel()
	m, cmd := m.Update(key("f"))
	tick := cmdMsgs(cmd)[0].(msg.FollowTickMsg)
	_, cmd = m.Update(tick)
	found := false
	for _, x := range cmdMsgs(cmd) {
		if lo, ok := x.(msg.LogOptionsMsg); ok && lo.Options == DefaultLogOptions {
			found = true
		}
	}
	if !found {
		t.Error("tick on the logs tab should refetch logs with the current options")
	}
}

func TestFollowIgnoresStaleOrStoppedTicks(t *testing.T) {
	instantTicks(t)
	m, tick := followingStats(t)
	if _, cmd := m.Update(msg.FollowTickMsg{ID: tick.ID + 1000}); cmd != nil {
		t.Error("a tick from another follow session should be ignored")
	}
	m, _ = m.Update(key("f")) // stop
	if _, cmd := m.Update(tick); cmd != nil {
		t.Error("a tick after following stopped should be ignored")
	}
}

func TestFollowRestartGetsNewID(t *testing.T) {
	instantTicks(t)
	m, first := followingStats(t)
	m, _ = m.Update(key("f"))
	_, cmd := m.Update(key("f"))
	second := cmdMsgs(cmd)[0].(msg.FollowTickMsg)
	if second.ID == first.ID {
		t.Error("each follow session needs its own ID so old ticks can't double up")
	}
}

func TestFollowTickOnOtherTabOnlyReschedules(t *testing.T) {
	instantTicks(t)
	m, tick := followingStats(t)
	m.activeTab = 0
	_, cmd := m.Update(tick)
	msgs := cmdMsgs(cmd)
	if len(msgs) != 1 {
		t.Fatalf("got %#v", msgs)
	}
	if _, ok := msgs[0].(msg.FollowTickMsg); !ok {
		t.Errorf("on a non-live tab a tick should only reschedule, got %#v", msgs[0])
	}
}

func TestFollowKeyOnlyOnLiveTabs(t *testing.T) {
	m := NewAppDetailModel("my-app", testDetailStyles(), 80, 24)
	if _, cmd := m.Update(key("f")); cmd != nil {
		t.Error("f on the Info tab should do nothing")
	}
}

func TestServiceStatsFollow(t *testing.T) {
	instantTicks(t)
	m := NewServiceDetailModel("my-db", testDetailStyles(), 80, 24)
	m.activeTab = serviceStatsTab
	m, cmd := m.Update(key("f"))
	tick := cmdMsgs(cmd)[0].(msg.FollowTickMsg)
	_, cmd = m.Update(tick)
	found := false
	for _, x := range cmdMsgs(cmd) {
		if _, ok := x.(msg.StatsIntervalMsg); ok {
			found = true
		}
	}
	if !found {
		t.Error("service stats tick should refetch stats")
	}
}
