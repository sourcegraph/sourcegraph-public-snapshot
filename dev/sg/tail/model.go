package tail

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v2"
)

var backlogSize = 100 * 1024

type model struct {
	// l is the unix socket we're listening on for incoming activities.
	l net.Listener
	// content is the rendered lines for the pager.
	content []string
	// activities are a collection of received activity messages. They get truncated
	// once they go over backlogSize.
	activities []*activityMsg
	// ch is the channel from which we're receiving activity messages.
	ch chan string
	// ready is set to true once we've received the window size.
	ready bool
	// pause stores the following or paused state of the viewport.
	pause bool
	// tabs are a list of predicates to apply to the pager's content, allowing
	// to filter activities.
	tabs []*tab
	// tabIndex stores the current tab index.
	tabIndex int
	// visiblePrompt is set to true when the prompt is visible.
	visiblePrompt bool
	// search stores the search query used to highlight activities.
	search string

	// viewport is the model implementing the pager.
	viewport viewport.Model
	// promptInput is the model implementing the prompt (: or /)
	promptInput textinput.Model
	// statusMsg holds the error if any, after inputting a command
	statusMsg string
}

// refreshContent goes through all activities and applies predicates to filter out
// unwanted activities, before rendering them into a slice of strings.
func (m *model) refreshContent() {
	t := m.tabs[m.tabIndex]
	m.content = []string{}
	for _, a := range m.activities {
		if t.preds.Apply(a) != nil {
			m.content = append(m.content, a.render(m.viewport.Width, m.search))
		}
	}
	m.viewport.SetContent(strings.Join(m.content, "\n"))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	m.promptInput, cmd = m.promptInput.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case commandMsg:
		switch msg.name {
		case "drop", "only", "grep":
			t := m.tabs[m.tabIndex]
			t.preds = append(t.preds, msg.toPred())
			m.refreshContent()
		case "reset":
			t := m.tabs[m.tabIndex]
			t.preds = activityPreds{}
			m.refreshContent()
		case "tabnew":
			m.tabs = append(m.tabs, &tab{title: fmt.Sprintf("%d", len(m.tabs))})
			m.tabIndex = len(m.tabs) - 1
			m.refreshContent()
		case "tabclose":
			if m.tabIndex == 0 {
				// TODO print something
				break
			}
			old := m.tabs
			m.tabs = make([]*tab, 0, len(old))
			for i, t := range old {
				if i != m.tabIndex {
					m.tabs = append(m.tabs, t)
				}
			}
			m.tabIndex = len(m.tabs) - 1
			m.refreshContent()
		}
	case tea.KeyMsg:
		if m.visiblePrompt {
			switch k := msg.String(); k {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.visiblePrompt = false
				m.statusMsg = ""
				m.promptInput.Blur()
			case "enter":
				value := m.promptInput.Value()
				if m.promptInput.Prompt == ":" {
					cmd, err := evalPrompt(value)
					if err != nil {
						m.statusMsg = err.Error()
					} else {
						cmds = append(cmds, cmd)
					}
				} else {
					// It's a search
					m.search = value
					m.refreshContent()
				}
				m.promptInput.SetValue("")
				m.visiblePrompt = false
				m.promptInput.Blur()
			}
		} else {
			m.statusMsg = ""
			switch k := msg.String(); k {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				if m.search != "" {
					m.search = ""
					m.refreshContent()
				}
			case "p":
				m.pause = !m.pause
			case "up", "down":
				// When user scrolls, we want to pause
				m.pause = true
			case "enter":
				// When user presses enter, we want to unpause and go to the bottom
				m.pause = false
				m.viewport.GotoBottom()
			case "tab":
				m.tabIndex = (m.tabIndex + 1) % len(m.tabs)
				m.refreshContent()
			case ":":
				m.visiblePrompt = true
				m.promptInput.Prompt = ":"
				m.promptInput.Focus()
				cmds = append(cmds, textinput.Blink)
			case "/":
				m.visiblePrompt = true
				m.promptInput.Focus()
				m.promptInput.Prompt = "/"
				cmds = append(cmds, textinput.Blink)
			}
		}
	case activityMsg:
		if msg.data != "" {
			// If we've hit the backlog size limit, remove the oldest activities.
			if len(m.activities) >= backlogSize {
				m.activities = m.activities[100:]
			}

			m.activities = append(m.activities, &msg)
			m.refreshContent()
			if !m.pause {
				m.viewport.GotoBottom()
			}
		}
		cmds = append(cmds, waitForActivity(m.ch))
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		statusHeight := lipgloss.Height(m.promptView())

		verticalMarginHeight := headerHeight + footerHeight + statusHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		listenUnixSocket(m.l, m.ch),
		waitForActivity(m.ch),
	)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	var promptView string
	if m.statusMsg != "" {
		promptView = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render(m.statusMsg)
	} else {
		promptView = m.promptView()
	}

	return fmt.Sprintf("%s\n%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView(), promptView)
}

func (m model) headerView() string {
	var tabsStr string
	for i, t := range m.tabs {
		var s string
		if i == m.tabIndex {
			s = activeTabStyle.Render(t.title)
		} else {
			s = inactiveTabStyle.Render(t.title)
		}
		tabsStr = lipgloss.JoinHorizontal(lipgloss.Left, tabsStr, s)
	}
	title := titleStyle.Render("sg")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)-lipgloss.Width(tabsStr)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, tabsStr, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	status := titleStyle.Render("FOLLOW")
	if m.pause {
		status = titleStyle.Render("PAUSED")
	}
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(status)-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, status, line, info)
}

func evalPrompt(value string) (tea.Cmd, error) {
	parts := strings.Split(value, " ")
	switch cmd := parts[0]; cmd {
	case "drop":
		if len(parts[1:]) < 2 {
			return nil, fmt.Errorf("drop requires at least two arguments")
		}
		return func() tea.Msg {
			return commandMsg{
				name: "drop",
				args: parts[1:],
			}
		}, nil
	case "only":
		if len(parts[1:]) < 2 {
			return nil, fmt.Errorf("only requires at least two arguments")
		}
		return func() tea.Msg {
			return commandMsg{
				name: "only",
				args: parts[1:],
			}
		}, nil
	case "grep":
		if len(parts[1:]) < 1 {
			return nil, fmt.Errorf("drop requires at least one arguments")
		}
		return func() tea.Msg {
			return commandMsg{
				name: "grep",
				args: parts[1:],
			}
		}, nil
	case "reset":
		return func() tea.Msg {
			return commandMsg{
				name: "reset",
			}
		}, nil
	case "tabnew":
		return func() tea.Msg {
			return commandMsg{
				name: "tabnew",
			}
		}, nil
	case "tabclose":
		return func() tea.Msg {
			return commandMsg{
				name: "tabclose",
			}
		}, nil

	}
	return nil, nil
}

func (m model) promptView() string {
	if m.visiblePrompt {
		return m.promptInput.View()
	}
	return ""
}

func main() {
	app := &cli.App{
		Name:  "sgtail",
		Usage: "Listens for 'sg start' activity and streams it with a nice UI.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "only-name",
				Usage: "--only-name [service_name] Starts with a new tab that display only logs from service named [service_name]",
				Value: "",
			},
		},
		Action: func(cctx *cli.Context) error {
			l, err := net.Listen("unix", "/tmp/sg.sock")
			if err != nil {
				panic(err)
			}
			defer func() {
				_ = os.Remove("/tmp/sg.sock")
			}()

			m := model{
				ch: make(chan string, 10),
				l:  l,
				tabs: []*tab{
					{title: "all", preds: []activityPred{}},
				},
				promptInput: textinput.New(),
			}

			if cctx.String("only-name") != "" {
				onlyCmd := commandMsg{
					name: "only",
					args: []string{"name", cctx.String("only-name")},
				}
				m.tabs = append(m.tabs, &tab{title: "^" + cctx.String("only-name"), preds: []activityPred{onlyCmd.toPred()}})
				m.tabIndex = len(m.tabs) - 1
			}

			p := tea.NewProgram(
				m,
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)

			_, err = p.Run()
			return err
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
