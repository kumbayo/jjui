package details

import (
	"testing"

	"github.com/idursun/jjui/internal/ui/intents"
	"github.com/idursun/jjui/test"
	"github.com/stretchr/testify/require"
)

func TestCopyFocusedFile(t *testing.T) {
	runner := test.NewTestCommandRunner(t)
	defer runner.Verify()
	operation := loadOperation(t, runner, "<<file status_char::M>><<file:: >><<file display_diff_path::dir/file.txt>>\n<<file status_char::A>><<file:: >><<file display_diff_path::other.txt>>\n")
	var copied string
	operation.clipboardWriter = func(value string) error { copied = value; return nil }
	cmd, handled := operation.HandleIntent(intents.CopyFocused{})
	require.True(t, handled)
	require.Equal(t, "Copied focused value", cmd().(intents.AddMessage).Text)
	require.Equal(t, "dir/file.txt", copied)
	operation.LeaveFiles()
	cmd, handled = operation.HandleIntent(intents.CopyFocused{})
	require.True(t, handled)
	require.Equal(t, "Nothing to copy", cmd().(intents.AddMessage).Text)
}
