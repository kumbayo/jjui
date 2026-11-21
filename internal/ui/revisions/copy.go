package revisions

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/ui/common"
	"github.com/idursun/jjui/internal/ui/intents"
)

func (m *Model) copyFocused() tea.Cmd {
	obj := m.focusedTextObject()
	if obj == nil || obj.Kind == textObjectElided {
		return intents.Invoke(intents.AddMessage{Text: "Nothing to copy"})
	}
	value := obj.Value
	template := ""
	switch obj.Kind {
	case textObjectNode, textObjectChangeId:
		template = "change_id"
	case textObjectCommitId:
		template = "commit_id"
	case textObjectDescription:
		template = "description"
	case textObjectBookmark, textObjectTag:
		if obj.Remote != "" {
			value += "@" + obj.Remote
		}
	}
	writer := m.clipboardWriter
	if template == "" {
		return common.CopyText(value, writer)
	}
	revision := m.SelectedRevision()
	target := revision.CommitId
	if target == "" {
		target = revision.GetChangeId()
	}
	args := jj.GetRevisionText(target, template)
	return func() tea.Msg {
		output, err := m.context.RunCommandImmediate(args)
		if err != nil {
			return intents.AddMessage{Text: fmt.Sprintf("Failed to copy focused value: %v", err), Err: err}
		}
		return common.CopyText(string(output), writer)()
	}
}
