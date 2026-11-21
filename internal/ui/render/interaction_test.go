package render

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/idursun/jjui/internal/ui/layout"
	"github.com/stretchr/testify/assert"
)

type testScrollMsg struct {
	Delta      int
	Horizontal bool
}

func TestProcessMouseEvent_WheelCallsMsgFn(t *testing.T) {
	dl := NewDisplayContext()
	dl.AddInteractionFn(layout.Rect(0, 0, 10, 5), func(msg tea.MouseMsg) tea.Msg {
		delta, horizontal := WheelDelta(msg)
		return testScrollMsg{Delta: delta, Horizontal: horizontal}
	}, InteractionScroll, 0)

	tests := []struct {
		button tea.MouseButton
		mod    tea.KeyMod
		want   testScrollMsg
	}{
		{tea.MouseWheelUp, 0, testScrollMsg{Delta: -3}},
		{tea.MouseWheelDown, 0, testScrollMsg{Delta: 3}},
		{tea.MouseWheelLeft, 0, testScrollMsg{Delta: -3, Horizontal: true}},
		{tea.MouseWheelRight, 0, testScrollMsg{Delta: 3, Horizontal: true}},
		{tea.MouseWheelUp, tea.ModCtrl, testScrollMsg{Delta: -9}},
		{tea.MouseWheelDown, tea.ModCtrl, testScrollMsg{Delta: 9}},
		{tea.MouseWheelLeft, tea.ModCtrl, testScrollMsg{Delta: -9, Horizontal: true}},
		{tea.MouseWheelRight, tea.ModCtrl, testScrollMsg{Delta: 9, Horizontal: true}},
	}
	for _, tt := range tests {
		msg, handled := dl.ProcessMouseEvent(tea.MouseWheelMsg{X: 2, Y: 2, Button: tt.button, Mod: tt.mod})
		assert.True(t, handled)
		assert.Equal(t, tt.want, msg)
	}

	_, handled := dl.ProcessMouseEvent(tea.MouseWheelMsg{X: 20, Y: 2, Button: tea.MouseWheelDown})
	assert.False(t, handled, "wheel outside the region is not handled")
}
