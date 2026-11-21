package render

import (
	tea "charm.land/bubbletea/v2"
	"github.com/idursun/jjui/internal/ui/layout"
)

// InteractionType defines what kinds of input an interactive region responds to.
// Multiple types can be combined using bitwise OR.
type InteractionType int

const (
	InteractionClick InteractionType = 1 << iota
	InteractionScroll
	InteractionDrag
)

// InteractionOp represents an interactive region that responds to input.
type InteractionOp struct {
	Rect  layout.Rectangle           // The interactive area (absolute coordinates)
	Msg   tea.Msg                    // Message to send (static)
	MsgFn func(tea.MouseMsg) tea.Msg // Optional message factory.
	Type  InteractionType            // What kind of interaction this supports
	Z     int                        // Z-index for overlapping regions (higher = priority)
}

// WheelDelta converts a mouse wheel event into a scroll delta.
// Holding down Ctrl increases the scroll distance.
// It returns a zero delta for events that are not wheel scrolls.
func WheelDelta(msg tea.MouseMsg) (delta int, horizontal bool) {
	mouse := msg.Mouse()
	distance := 3
	if mouse.Mod&tea.ModCtrl != 0 {
		distance *= 3
	}
	switch mouse.Button {
	case tea.MouseWheelUp:
		return -distance, false
	case tea.MouseWheelDown:
		return distance, false
	case tea.MouseWheelLeft:
		return -distance, true
	case tea.MouseWheelRight:
		return distance, true
	}
	return 0, false
}

func processMouseEvent(interactions []interactionOp, msg tea.MouseMsg) (tea.Msg, bool) {
	mouse := msg.Mouse()
	pos := layout.Pos(mouse.X, mouse.Y)
	switch msg.(type) {
	case tea.MouseClickMsg:
		if mouse.Button != tea.MouseLeft {
			return nil, false
		}
		// Find highest-Z draggable region containing this point
		if interaction, ok := findInteraction(interactions, pos, InteractionDrag); ok {
			if interaction.MsgFn != nil {
				return interaction.MsgFn(msg), true
			}
			return interaction.Msg, true
		}
		// Find highest-Z clickable region containing this point
		if interaction, ok := findInteraction(interactions, pos, InteractionClick); ok {
			if interaction.MsgFn != nil {
				return interaction.MsgFn(msg), true
			}
			return interaction.Msg, true
		}
	case tea.MouseWheelMsg:
		if delta, _ := WheelDelta(msg); delta == 0 {
			return nil, false
		}
		if interaction, ok := findInteraction(interactions, pos, InteractionScroll); ok {
			if interaction.MsgFn != nil {
				return interaction.MsgFn(msg), true
			}
			return interaction.Msg, true
		}
	}

	return nil, false
}

// findInteraction returns the first interaction of the given type containing pos.
// Interactions are expected to be sorted by priority.
func findInteraction(interactions []interactionOp, pos layout.Position, typ InteractionType) (interactionOp, bool) {
	for _, interaction := range interactions {
		if interaction.Type&typ != 0 && pos.In(interaction.Rect) {
			return interaction, true
		}
	}
	return interactionOp{}, false
}
