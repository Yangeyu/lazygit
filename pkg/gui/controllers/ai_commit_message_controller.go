package controllers

import (
	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/jesseduffield/lazygit/pkg/utils"
)

// AICommitMessageController drives the suggestion panel that sits above the
// commit message inputs. It lets the user accept the generated message,
// regenerate it, or dismiss the panel and go back to writing by hand.
type AICommitMessageController struct {
	baseController
	c *ControllerCommon
}

var _ types.IController = &AICommitMessageController{}

func NewAICommitMessageController(
	c *ControllerCommon,
) *AICommitMessageController {
	return &AICommitMessageController{
		baseController: baseController{},
		c:              c,
	}
}

func (self *AICommitMessageController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{
			Keys:        opts.GetKeys(opts.Config.CommitMessage.AcceptSuggestion),
			Handler:     self.accept,
			Description: self.c.Tr.AcceptCommitMessageSuggestion,
		},
		{
			Keys:        opts.GetKeys(opts.Config.CommitMessage.GenerateMessage),
			Handler:     self.regenerate,
			Description: self.c.Tr.RegenerateCommitMessage,
		},
		{
			Keys:    opts.GetKeys(opts.Config.Universal.TogglePanel),
			Handler: self.switchToCommitMessage,
		},
		{
			Keys:        opts.GetKeys(opts.Config.Universal.Return),
			Handler:     self.dismiss,
			Description: self.c.Tr.DismissCommitMessageSuggestion,
		},
	}
}

func (self *AICommitMessageController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName:    self.Context().GetViewName(),
			FocusedView: self.c.Contexts().CommitMessage.GetViewName(),
			Key:         gocui.MouseLeft,
			Handler:     self.onClick,
		},
	}
}

func (self *AICommitMessageController) GetOnFocus() func(types.OnFocusOpts) {
	return func(types.OnFocusOpts) {
		self.c.Views().AICommitMessage.Footer = utils.ResolvePlaceholderString(
			self.c.Tr.AICommitMessageFooter,
			map[string]string{
				"acceptKeybinding":     self.c.UserConfig().Keybinding.CommitMessage.AcceptSuggestion.String(),
				"regenerateKeybinding": self.c.UserConfig().Keybinding.CommitMessage.GenerateMessage.String(),
				"dismissKeybinding":    self.c.UserConfig().Keybinding.Universal.Return.String(),
			},
		)
	}
}

func (self *AICommitMessageController) Context() types.Context {
	return self.c.Contexts().AICommitMessage
}

func (self *AICommitMessageController) accept() error {
	self.c.Helpers().AICommit.AcceptSuggestion()
	self.c.Context().Replace(self.c.Contexts().CommitMessage)
	return nil
}

func (self *AICommitMessageController) regenerate() error {
	return self.c.Helpers().AICommit.Generate()
}

func (self *AICommitMessageController) switchToCommitMessage() error {
	self.c.Context().Replace(self.c.Contexts().CommitMessage)
	return nil
}

func (self *AICommitMessageController) dismiss() error {
	self.c.Helpers().AICommit.HideSuggestionPanel()
	self.c.Context().Replace(self.c.Contexts().CommitMessage)
	return nil
}

func (self *AICommitMessageController) onClick(opts gocui.ViewMouseBindingOpts) error {
	self.c.Context().Replace(self.c.Contexts().AICommitMessage)
	return nil
}
