// +build dev

package graphqlbackend

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"runtime"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/graphqlfile"
)

var Schema = readSchemaFromDisk()

func readSchemaFromDisk() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("No caller information")
	}
	path := filepath.Join(filepath.Dir(filename), "schema.graphql")
	out, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	out, err = graphqlfile.StripInternalComments(out)
	if err != nil {
		log.Fatal(err)
	}
	return string(out)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_190(size int) error {
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
