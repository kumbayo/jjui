package common

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/idursun/jjui/internal/ui/intents"
)

// CopyText writes even an empty value to the clipboard. A nil writer uses the system clipboard.
func CopyText(text string, writer func(string) error) tea.Cmd {
	if writer == nil {
		writer = clipboard.WriteAll
	}
	return func() tea.Msg {
		if err := writer(text); err != nil {
			return intents.AddMessage{Text: fmt.Sprintf("Failed to copy focused value: %v", err), Err: err}
		}
		return intents.AddMessage{Text: "Copied focused value"}
	}
}
