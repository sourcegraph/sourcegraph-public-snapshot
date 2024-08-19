package process

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
)

type consoleMessage struct {
	Stderr []byte `json:"stderr"`
	Stdout []byte `json:"stdout"`
}

// consumeStderr is intended to be used as a goroutine, and will consume this
// process' stderr stream until it reaches EOF. It reads the stream line-by-line
// and will decode any console messages per the jsii wire protocol specification.
// Once EOF has been reached, true will be sent to the done channel, allowing
// other goroutines to check whether the goroutine has reached EOF (and hence
// finished) or not.
func (p *Process) consumeStderr(done chan bool) {
	reader := bufio.NewReader(p.stderr)

	for true {
		line, err := reader.ReadBytes('\n')
		if len(line) == 0 || err == io.EOF {
			done <- true
			return
		}
		var message consoleMessage
		if err := json.Unmarshal(line, &message); err != nil {
			os.Stderr.Write(line)
		} else {
			if message.Stderr != nil {
				os.Stderr.Write(message.Stderr)
			}
			if message.Stdout != nil {
				os.Stdout.Write(message.Stdout)
			}
		}
	}
}
