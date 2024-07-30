package tail

import (
	"net"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "tail",
	Usage: "Listens for 'sg start' log events and streams them with a nice UI",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "only-name",
			Usage: "--only-name [service_name] Starts with a new tab that display only logs from service named [service_name]",
			Value: "",
		},
	},
	Category: category.Dev,
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
			help:        help.New(),
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
