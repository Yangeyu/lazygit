package helpers

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/gocui"
)

// AICommitHelper drives the "generate commit message" feature: it runs a
// user-configured command with the staged diff on stdin and shows the result
// in the AI suggestion panel that sits above the commit message inputs.
type AICommitHelper struct {
	c             *HelperCommon
	commitsHelper *CommitsHelper
}

func NewAICommitHelper(c *HelperCommon, commitsHelper *CommitsHelper) *AICommitHelper {
	return &AICommitHelper{
		c:             c,
		commitsHelper: commitsHelper,
	}
}

// Enabled reports whether a generate-commit-message command has been configured.
func (self *AICommitHelper) Enabled() bool {
	return self.command() != ""
}

func (self *AICommitHelper) command() string {
	return self.c.UserConfig().Git.Commit.GenerateCommitMessageCommand
}

func (self *AICommitHelper) ShowSuggestionPanel() {
	self.c.Views().AICommitMessage.Visible = true
}

func (self *AICommitHelper) HideSuggestionPanel() {
	self.c.Views().AICommitMessage.Visible = false
}

// Generate runs the configured command with the staged diff piped to its stdin
// and renders the generated message into the suggestion panel. The command is
// provider-agnostic: it can be any LLM CLI or a script of the user's own.
func (self *AICommitHelper) Generate() error {
	command := self.command()
	if command == "" {
		return nil
	}

	self.ShowSuggestionPanel()
	self.setSuggestionContent(self.c.Tr.GeneratingCommitMessage)

	return self.c.WithWaitingStatus(self.c.Tr.GeneratingCommitMessageStatus, func(gocui.Task) error {
		diff, err := self.c.Git().Diff.GetDiff(true)
		if err != nil {
			return err
		}

		if strings.TrimSpace(diff) == "" {
			self.c.OnUIThread(func() error {
				self.setSuggestionContent(self.c.Tr.NoStagedChangesToGenerateFrom)
				return nil
			})
			return nil
		}

		output, err := self.c.OS().Cmd.
			NewShell(command, self.c.UserConfig().OS.ShellFunctionsFile).
			SetStdin(diff).
			RunWithOutput()

		self.c.OnUIThread(func() error {
			if err != nil {
				self.setSuggestionContent(fmt.Sprintf("%s\n\n%s", self.c.Tr.GenerateCommitMessageFailed, strings.TrimSpace(output)))
				return nil
			}
			self.setSuggestionContent(strings.TrimSpace(output))
			return nil
		})
		return nil
	})
}

// GetSuggestion returns the message currently shown in the suggestion panel.
func (self *AICommitHelper) GetSuggestion() string {
	return strings.TrimSpace(self.c.Views().AICommitMessage.Buffer())
}

// AcceptSuggestion copies the generated message into the editable summary and
// description fields and hides the suggestion panel.
func (self *AICommitHelper) AcceptSuggestion() {
	self.commitsHelper.SetMessageAndDescriptionInView(self.GetSuggestion())
	self.HideSuggestionPanel()
}

func (self *AICommitHelper) setSuggestionContent(content string) {
	self.c.SetViewContent(self.c.Views().AICommitMessage, content)
}
