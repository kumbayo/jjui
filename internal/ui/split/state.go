package split

import "github.com/idursun/jjui/internal/ui/layout"

type Placement int

const (
	PlacementAuto Placement = iota
	PlacementBottom
	PlacementRight
)

type SplitState struct {
	Percent      float64
	MinPercent   float64
	MaxPercent   float64
	AtBottom     bool
	AutoPosition bool
}

func NewSplitState(percent float64) *SplitState {
	s := &SplitState{
		Percent:    percent,
		MinPercent: 10,
		MaxPercent: 95,
	}
	s.clamp()
	return s
}

func (s *SplitState) SetPlacement(placement Placement) {
	s.AutoPosition = placement == PlacementAuto
	s.AtBottom = placement == PlacementBottom
}

func (s *SplitState) TogglePosition() {
	s.AutoPosition = false
	s.AtBottom = !s.AtBottom
}

func (s *SplitState) clamp() {
	if s.Percent < s.MinPercent {
		s.Percent = s.MinPercent
	}
	if s.Percent > s.MaxPercent {
		s.Percent = s.MaxPercent
	}
}

func (s *SplitState) primarySize(total int) int {
	return int(float64(total) * (100 - s.Percent) / 100)
}

func (s *SplitState) DragTo(box layout.Box, vertical bool, x, y int) bool {
	total, target := box.R.Dx(), x-box.R.Min.X
	if vertical {
		total, target = box.R.Dy(), y-box.R.Min.Y
	}
	if total <= 0 || target == s.primarySize(total) {
		return false
	}
	old := s.Percent
	// Aim inside the target cell so rounding cannot move the divider back a cell.
	s.Percent = (float64(total-target) - 0.5) * 100 / float64(total)
	s.clamp()
	return s.Percent != old
}

func (s *SplitState) Expand(delta float64) {
	s.Percent += delta
	s.clamp()
}

func (s *SplitState) Shrink(delta float64) {
	s.Percent -= delta
	s.clamp()
}

// SetPercent allows up to 100 so the secondary pane can take the full size,
// unlike expand/drag which stop at MaxPercent.
func (s *SplitState) SetPercent(percent float64) {
	s.Percent = min(max(percent, s.MinPercent), 100)
}
