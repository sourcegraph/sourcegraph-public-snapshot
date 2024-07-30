package tail

import (
	"bufio"
	"io"
	"net"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/grafana/regexp"
)

func acceptFromListener(l net.Listener, ch chan string) tea.Cmd {
	return func() tea.Msg {
		for {
			fd, err := l.Accept()
			if err != nil {
				panic(err)
			}
			go reader(fd, ch)
		}
	}
}

var ansiRe = regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")

func reader(r io.Reader, ch chan string) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		str := scanner.Text()
		str = ansiRe.ReplaceAllString(str, "")
		ch <- str
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func waitForActivity(ch chan string) tea.Cmd {
	return func() tea.Msg {
		msg := <-ch
		return parseActivity(msg)
	}
}
