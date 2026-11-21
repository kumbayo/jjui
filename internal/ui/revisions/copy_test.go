package revisions

import (
	tea "charm.land/bubbletea/v2"
	"errors"
	"github.com/idursun/jjui/internal/config"
	"github.com/idursun/jjui/internal/ui/actions"
	"github.com/idursun/jjui/internal/ui/bindings"
	"github.com/idursun/jjui/internal/ui/common"
	"github.com/idursun/jjui/internal/ui/dispatch"
	"testing"

	"github.com/idursun/jjui/internal/jj"
	"github.com/idursun/jjui/internal/screen"
	"github.com/idursun/jjui/internal/ui/intents"
	"github.com/idursun/jjui/test"
	"github.com/stretchr/testify/require"
)

func TestCopyFocused(t *testing.T) {
	for _, tc := range []struct {
		kind, value string
		index       int
		template    string
	}{
		{textObjectNode, "full-change-id", 0, "change_id"},
		{textObjectChangeId, "full-change-id", 0, "change_id"},
		{textObjectCommitId, "full-commit-id", 0, "commit_id"},
		{textObjectDescription, "Title\n\nBody\n", 0, "description"},
		{textObjectDescription, "", 0, "description"},
		{textObjectAuthor, "someone@example.com", 0, ""},
		{textObjectTimestamp, "2025-05-22 21:59:38", 0, ""},
		{textObjectBookmark, "second", 0, ""},
		{textObjectBookmark, "second@origin", 1, ""},
	} {
		t.Run(tc.kind+"/"+tc.value, func(t *testing.T) {
			runner := test.NewTestCommandRunner(t)
			defer runner.Verify()
			model := New(test.NewTestContext(runner))
			model.rows = textObjectRows(t)
			model.focusedObjectKind, model.focusedObjectIndex = tc.kind, tc.index
			model.checked = []checkedObject{checkedRevision(model.rows[1].Commit)}
			if tc.template != "" {
				args := jj.GetDescription(model.SelectedRevision().CommitId)
				args[4] = tc.template
				runner.Expect(args).SetOutput([]byte(tc.value))
			}
			called := false
			model.clipboardWriter = func(value string) error { called = true; require.Equal(t, tc.value, value); return nil }
			cmd, handled := model.HandleIntent(intents.CopyFocused{})
			require.True(t, handled)
			model.cursor = 1 // A pending copy keeps its original target.
			require.Equal(t, "Copied focused value", cmd().(intents.AddMessage).Text)
			require.True(t, called)
		})
	}
}

func TestCopyFocusedFailures(t *testing.T) {
	runner := test.NewTestCommandRunner(t)
	defer runner.Verify()
	model := New(test.NewTestContext(runner))
	model.clipboardWriter = func(string) error { t.Fatal("unexpected clipboard write"); return nil }
	require.Equal(t, "Nothing to copy", model.copyFocused()().(intents.AddMessage).Text)
	model.rows = textObjectRows(t)
	model.focusedObjectKind = textObjectDescription
	expected := errors.New("lookup failed")
	runner.Expect(jj.GetDescription(model.SelectedRevision().CommitId)).SetError(expected)
	require.ErrorIs(t, model.copyFocused()().(intents.AddMessage).Err, expected)
	model.focusedObjectKind = textObjectAuthor
	model.clipboardWriter = func(string) error { return expected }
	require.ErrorIs(t, model.copyFocused()().(intents.AddMessage).Err, expected)
}

func TestCopyFocusedBindingScopes(t *testing.T) {
	for _, tc := range []struct {
		name, action string
		scopes       []common.Scope
	}{
		{"revision", "revisions.copy_focused", []common.Scope{{Name: "revisions"}}},
		{"file", "revisions.details.copy_focused", []common.Scope{{Name: "revisions.details", Leak: common.LeakGlobal}, {Name: "revisions"}}},
		{"modal", "", []common.Scope{{Name: "revisions.inline_describe", Leak: common.LeakNone}, {Name: "revisions"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dispatcher, err := dispatch.NewDispatcher(config.BindingsToRuntime(config.Current.Bindings))
			require.NoError(t, err)
			result := dispatcher.Resolve(tea.KeyPressMsg{Code: 'y', Text: "Y", Mod: tea.ModShift}, tc.scopes)
			require.Equal(t, bindings.Action(tc.action), result.Action)
			if tc.action != "" {
				intent, ok := actions.ResolveIntent(string(result.Scope), result.Action, result.Args)
				require.True(t, ok)
				require.IsType(t, intents.CopyFocused{}, intent)
			}
		})
	}
}

func TestCopyElisionLeavesClipboardUnchanged(t *testing.T) {
	model := New(test.NewTestContext(test.NewTestCommandRunner(t)))
	model.rows = textObjectRows(t)
	model.rows[0].Lines[0].Segments = append(model.rows[0].Lines[0].Segments, &screen.Segment{Text: "...", Labels: []string{"elided"}})
	model.focusedObjectKind = textObjectElided
	model.clipboardWriter = func(string) error { t.Fatal("unexpected clipboard write"); return nil }
	require.Equal(t, "Nothing to copy", model.copyFocused()().(intents.AddMessage).Text)
}
