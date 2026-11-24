package split

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	uv "github.com/charmbracelet/ultraviolet"
	keybindings "github.com/idursun/jjui/internal/ui/bindings"
	"github.com/idursun/jjui/internal/ui/common"
	"github.com/idursun/jjui/internal/ui/intents"
	"github.com/idursun/jjui/internal/ui/layout"
	"github.com/idursun/jjui/internal/ui/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeContent struct {
	scope   keybindings.ScopeName
	shows   int
	hides   int
	updates int
	focused bool
}

func (f *fakeContent) Init() tea.Cmd { return nil }

func (f *fakeContent) Update(tea.Msg) tea.Cmd {
	f.updates++
	return nil
}

func (f *fakeContent) ViewRect(*render.DisplayContext, layout.Box) {}

func (f *fakeContent) OnShow() tea.Cmd {
	f.shows++
	return nil
}

func (f *fakeContent) OnHide() {
	f.hides++
}

func (f *fakeContent) SetFocused(focused bool) {
	f.focused = focused
}

func (f *fakeContent) Scopes() []common.Scope {
	return []common.Scope{{Name: f.scope, Handler: f}}
}

func (f *fakeContent) HandleIntent(intents.Intent) (tea.Cmd, bool) {
	return nil, false
}

func TestSplitContainerShowToggleAndClose(t *testing.T) {
	sc := NewSplitContainer(NewSplitState(50))
	a := &fakeContent{scope: "a"}
	b := &fakeContent{scope: "b"}
	sc.RegisterContent("a", a)
	sc.RegisterContent("b", b)

	_, shown := sc.ShowContent("a")
	require.True(t, shown)
	assert.Equal(t, 1, a.shows)
	_, shown = sc.ShowContent("a")
	assert.False(t, shown)

	sc.Update(struct{}{})
	assert.Equal(t, 1, a.updates)
	assert.Equal(t, 0, b.updates)

	_, shown = sc.ShowContent("b")
	require.True(t, shown)
	assert.Equal(t, 1, a.hides)
	assert.Equal(t, 1, b.shows)

	_, changed := sc.ToggleContent("b")
	require.True(t, changed)
	assert.Equal(t, 1, b.hides)
	assert.Empty(t, sc.activeContentID)

	assert.False(t, sc.Close())
}

func TestSplitContainerShowContentUnknownIDReturnsFalseWithoutChangingActiveContent(t *testing.T) {
	sc := NewSplitContainer(NewSplitState(50))
	a := &fakeContent{scope: "a"}
	sc.RegisterContent("a", a)

	_, shown := sc.ShowContent("a")
	require.True(t, shown)

	_, shown = sc.ShowContent("missing")
	assert.False(t, shown)
	assert.Equal(t, "a", sc.activeContentID)
	assert.Equal(t, 0, a.hides)
	assert.Equal(t, 1, a.shows)
}

func TestSplitContainerTracksOnlyWhetherContentIsFocused(t *testing.T) {
	sc := NewSplitContainer(NewSplitState(50))
	content := &fakeContent{scope: "content"}
	sc.RegisterContent("content", content)
	_, _ = sc.ShowContent("content")

	primary := []common.Scope{{Name: "primary"}}
	assert.False(t, sc.ContentFocused())
	assert.Equal(t, []keybindings.ScopeName{"primary", "content"}, scopeNames(sc.Scopes(primary)))

	sc.setContentFocused(true)
	assert.True(t, sc.ContentFocused())
	assert.True(t, content.focused)
	assert.Equal(t, []keybindings.ScopeName{"content"}, scopeNames(sc.Scopes(primary)))

	sc.setContentFocused(!sc.contentFocused)
	assert.False(t, sc.ContentFocused())
	assert.False(t, content.focused)

	sc.setContentFocused(!sc.contentFocused)
	assert.True(t, sc.ContentFocused())
	assert.True(t, content.focused)
	sc.Close()
	assert.False(t, sc.ContentFocused())
	assert.False(t, content.focused)
}

func TestSplitContainerResizeAndPosition(t *testing.T) {
	state := NewSplitState(50)
	state.SetPlacement(PlacementAuto)
	sc := NewSplitContainer(state)
	sc.RegisterContent("preview", &fakeContent{scope: "preview"})
	_, _ = sc.ShowContent("preview")
	assert.True(t, sc.CanResize(1))
	assert.True(t, sc.CanResize(-1))

	sc.Resize(10)
	assert.Equal(t, 60.0, state.Percent)
	sc.Resize(-20)
	assert.Equal(t, 40.0, state.Percent)
	state.Percent = state.MaxPercent
	assert.False(t, sc.CanResize(1))
	assert.True(t, sc.CanResize(-1))
	state.Percent = state.MinPercent
	assert.True(t, sc.CanResize(1))
	assert.False(t, sc.CanResize(-1))

	sc.SetAutoPosition(true)
	assert.True(t, state.AtBottom)
	sc.TogglePosition()
	assert.False(t, state.AutoPosition)
	assert.False(t, state.AtBottom)
	sc.SetAutoPosition(true)
	assert.False(t, state.AtBottom)
}

func TestSplitContainerDividerFollowsMouseWithoutJumping(t *testing.T) {
	for _, tc := range []struct {
		name      string
		atBottom  bool
		width     int
		height    int
		start     layout.Position
		direction layout.Position
		separator string
	}{
		{"right", false, 204, 8, layout.Pos(105, 8), layout.Pos(1, 0), "│"},
		{"bottom", true, 8, 20, layout.Pos(7, 16), layout.Pos(0, 1), "─"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := NewSplitState(50)
			state.AtBottom = tc.atBottom
			sc := NewSplitContainer(state)
			sc.RegisterContent("preview", &fakeContent{})
			_, _ = sc.ShowContent("preview")
			box := layout.NewBox(layout.Rect(4, 7, tc.width, tc.height))
			primary := &fakeContent{}

			renderDividerAt := func(pos layout.Position) *render.DisplayContext {
				t.Helper()
				dl := render.NewDisplayContext()
				sc.Render(dl, box, primary)
				buf := uv.NewScreenBuffer(box.R.Max.X, box.R.Max.Y)
				dl.Render(buf)
				require.Equal(t, tc.separator, buf.CellAt(pos.X, pos.Y).Content)
				return dl
			}

			dl := renderDividerAt(tc.start)
			msg, handled := dl.ProcessMouseEvent(tea.MouseClickMsg{X: tc.start.X, Y: tc.start.Y, Button: tea.MouseLeft})
			require.True(t, handled)
			drag, ok := msg.(SplitDragMsg)
			require.True(t, ok)
			sc.StartDrag(drag)
			assert.Equal(t, 50.0, state.Percent)
			renderDividerAt(tc.start)

			require.True(t, sc.DragTo(tc.start.X, tc.start.Y))
			assert.Equal(t, 50.0, state.Percent)
			renderDividerAt(tc.start)

			for _, offset := range []int{1, 0, -1} {
				pos := tc.start.Add(tc.direction.Mul(offset))
				require.True(t, sc.DragTo(pos.X, pos.Y))
				renderDividerAt(pos)
			}

			sc.EndDrag()
			assert.False(t, sc.DragTo(tc.start.X, tc.start.Y))
			renderDividerAt(tc.start.Sub(tc.direction))
		})
	}
}

func scopeNames(scopes []common.Scope) []keybindings.ScopeName {
	names := make([]keybindings.ScopeName, len(scopes))
	for i, scope := range scopes {
		names[i] = scope.Name
	}
	return names
}

func TestSplitContainerSetSizeClampsBetweenMinAndFull(t *testing.T) {
	state := NewSplitState(50)
	sc := NewSplitContainer(state)

	sc.SetSize(100)
	assert.Equal(t, 100.0, state.Percent)
	sc.SetSize(250)
	assert.Equal(t, 100.0, state.Percent)
	sc.SetSize(0)
	assert.Equal(t, state.MinPercent, state.Percent)
	sc.SetSize(30)
	assert.Equal(t, 30.0, state.Percent)
}
