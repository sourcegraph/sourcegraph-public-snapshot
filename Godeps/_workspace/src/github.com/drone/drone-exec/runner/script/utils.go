package script

import (
	"bytes"
	"encoding/base64"
	"fmt"
)

// Writes the netrc file.
func writeNetrc(machine, login, password string) string {
	var buf bytes.Buffer
	if len(machine) == 0 {
		return buf.String()
	}
	out := fmt.Sprintf(
		netrcScript,
		machine,
		login,
		password,
	)
	buf.WriteString(out)
	return buf.String()
}

// Writes the RSA private key
func writeKey(key string) string {
	var buf bytes.Buffer
	if len(key) == 0 {
		return buf.String()
	}
	buf.WriteString(fmt.Sprintf(keyScript, key))
	buf.WriteString(keyConfScript)
	return buf.String()
}

// writeCmds is a helper fuction that writes a slice
// of bash commands as a single script.
func writeCmds(cmds []string) string {
	var buf bytes.Buffer
	for _, cmd := range cmds {
		buf.WriteString(trace(cmd))
	}
	return buf.String()
}

// trace is a helper function that allows us to echo
// commands back to the console for debugging purposes.
func trace(cmd string) string {
	echo := fmt.Sprintf("$ %s\n", cmd)
	base := base64.StdEncoding.EncodeToString([]byte(echo))
	return fmt.Sprintf(traceScript, base, cmd)
}

// encode is a helper function that base64 encodes
// a shell command (or entire script)
func encode(script []byte) string {
	encoded := base64.StdEncoding.EncodeToString(script)
	return fmt.Sprintf("echo %s | base64 -d | /bin/sh", encoded)
}
