package view

import (
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/henrrrik/runway-tui/internal/msg"
)

const (
	followInterval = 15 * time.Second
	followKey      = "f"
)

// followIDs hands out a unique ID per follow session across all views, so a
// tick from a previous session (or another view) is never mistaken for the
// current one and timers can't double up.
var followIDs atomic.Int64

// followTick schedules the next tick. Tests replace it to fire immediately.
var followTick = func(id int) tea.Cmd {
	return tea.Tick(followInterval, func(time.Time) tea.Msg { return msg.FollowTickMsg{ID: id} })
}

// follower tracks whether a view is auto-refreshing.
type follower struct {
	on bool
	id int
}

func (f follower) toggle() (follower, tea.Cmd) {
	if f.on {
		return follower{}, nil
	}
	f.on, f.id = true, int(followIDs.Add(1))
	return f, followTick(f.id)
}

// accepts reports whether t belongs to the current follow session.
func (f follower) accepts(t msg.FollowTickMsg) bool {
	return f.on && t.ID == f.id
}

// onTick returns the command for an accepted tick: refresh (may be nil, e.g.
// when the active tab has nothing live) and the next tick.
func (f follower) onTick(refresh tea.Cmd) tea.Cmd {
	next := followTick(f.id)
	if refresh == nil {
		return next
	}
	return tea.Batch(refresh, next)
}

func (f follower) help() string {
	if f.on {
		return followKey + ": stop following"
	}
	return followKey + ": follow"
}

func (f follower) banner(styles DetailStyles) string {
	return styles.StatusUp.Render("Auto-refresh: every " + followInterval.String())
}
