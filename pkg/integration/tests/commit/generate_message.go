package commit

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var GenerateMessage = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Generate a commit message with an external command, then accept it",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(config *config.AppConfig) {
		// A stand-in for an LLM CLI: it proves the staged diff was piped to it on
		// stdin by grepping for it, then prints a summary and body.
		config.GetUserConfig().Git.Commit.GenerateCommitMessageCommand =
			`if grep -q "diff --git"; then printf "Add myfile\n\nThe diff was received on stdin"; else printf "NO DIFF RECEIVED"; fi`
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFile("myfile", "myfile content")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			PressPrimaryAction(). // stage file
			Press(keys.Files.CommitChanges)

		t.Views().CommitMessage().
			IsFocused().
			Press(keys.CommitMessage.GenerateMessage)

		t.Views().AICommitMessage().
			IsFocused().
			Content(Contains("Add myfile")).
			Content(Contains("The diff was received on stdin")).
			Press(keys.CommitMessage.AcceptSuggestion)

		t.Views().CommitMessage().
			IsFocused().
			Content(Equals("Add myfile"))

		t.Views().CommitDescription().
			Content(Contains("The diff was received on stdin"))

		t.ExpectPopup().CommitMessagePanel().Confirm()

		t.Views().Commits().
			Focus().
			Lines(
				Contains("Add myfile").IsSelected(),
			)
	},
})
