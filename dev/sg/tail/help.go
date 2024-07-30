package tail

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	CtrlC      key.Binding
	Esc        key.Binding
	Prompt     key.Binding
	PromptSend key.Binding
	Help       key.Binding
	Quit       key.Binding
	Pause      key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
	Search     key.Binding
}

var keys = keyMap{
	CtrlC: key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	Quit:  key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	Esc:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "clear search or exit prompt")),
	Help:  key.NewBinding(key.WithKeys("?", "h"), key.WithHelp("?/h", "toggle help")),

	Prompt:     key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "show command prompt (available commands: drop, only, grep, reset, tabnew tabclose)")),
	PromptSend: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "if prompt is active, execute command prompt,  otherwise resume follow")),
	Search:     key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search prompt")),

	Pause:      key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "toggle pause/following mode")),
	ScrollUp:   key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "scroll up")),
	ScrollDown: key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "scroll down")),
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Pause, k.ScrollUp, k.ScrollDown, k.Esc, k.Search, k.PromptSend, k.Prompt}, // first column
		{k.Help, k.Quit, k.CtrlC}, // second column
	}
}
