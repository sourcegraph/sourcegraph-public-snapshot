// Command watchmanwrapper runs watchman subscribe and parses its output to
// trigger another command with the files that have changed. See
// dev/changewatch for how it is invocated.
package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
)

type Response struct {
	IsFreshInstance bool     `json:"is_fresh_instance"`
	Files           []string `json:"files"`
}

func main() {
	cmd := exec.Command("watchman", "-j", "--server-encoding=json", "-p")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		dec := json.NewDecoder(stdout)
		for {
			var r Response
			err := dec.Decode(&r)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Fatal(err)
			}
			if r.IsFreshInstance || len(r.Files) == 0 {
				continue
			}

			cmd := exec.Command(os.Args[1], r.Files...)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			err = cmd.Run()
			if err != nil {
				log.Printf("%s failed to run: %v", os.Args[1], err)
			}
		}
	}()

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_565(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
