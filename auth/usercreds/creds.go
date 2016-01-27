package usercreds

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/bgentry/go-netrc/netrc"
	"github.com/howeyc/gopass"
)

type LoginCredentials struct {
	Login, Password string
}

var netrcPath = "$HOME/.netrc"

// FromNetRC returns saved credentials for the endpoint. If it can't find any,
// it returns nil. This is different to userAuth which stores tokens. Instead
// this searches common places creds are stored for domains.
func FromNetRC(endpointURL *url.URL) *LoginCredentials {
	rc, err := netrc.ParseFile(os.ExpandEnv(netrcPath))
	if err != nil {
		return nil
	}
	host := endpointURL.Host
	if i := strings.Index(host, ":"); i != -1 {
		host = host[:i]
	}
	m := rc.FindMachine(host)
	if m == nil || m.Login == "" || m.Password == "" {
		return nil
	}
	return &LoginCredentials{Login: m.Login, Password: m.Password}
}

// FromTTY asks the user via the cmd line what the credentials are
func FromTTY() *LoginCredentials {
	fmt.Print("Username: ")
	username, err := getLine()
	if err != nil {
		return nil
	}
	fmt.Print("Password: ")
	password := string(gopass.GetPasswd())
	if username == "" || password == "" {
		return nil
	}
	return &LoginCredentials{Login: username, Password: password}
}

func getLine() (string, error) {
	var line string
	s := bufio.NewScanner(os.Stdin)
	if s.Scan() {
		line = s.Text()
	}
	return line, s.Err()
}
