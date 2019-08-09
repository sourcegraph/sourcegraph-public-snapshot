package db

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

func TestErrorsInterface(t *testing.T) {
	cases := []struct {
		Err       error
		Predicate func(error) bool
	}{
		{&repoNotFoundErr{}, errcode.IsNotFound},
		{userNotFoundErr{}, errcode.IsNotFound},
	}
	for _, c := range cases {
		if !c.Predicate(c.Err) {
			t.Errorf("%s does not match predicate %s", c.Err.Error(), functionName(c.Predicate))
		}
	}
}

func functionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_50(size int) error {
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
