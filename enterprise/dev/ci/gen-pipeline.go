// gen-pipeline.go generates a Buildkite YAML file that tests the entire
// Sourcegraph application and writes it to stdout.
package main

import (
	"os"

	ci "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/ci"
)

func main() {
	config := ci.ComputeConfig()
	pipeline, err := ci.GeneratePipeline(config)
	if err != nil {
		panic(err)
	}
	_, err = pipeline.WriteTo(os.Stdout)
	if err != nil {
		panic(err)
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_701(size int) error {
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
