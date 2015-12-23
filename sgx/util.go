package sgx

import (
	"bufio"
	"os"
)

func getLine() (string, error) {
	var line string
	s := bufio.NewScanner(os.Stdin)
	if s.Scan() {
		line = s.Text()
	}
	return line, s.Err()
}
