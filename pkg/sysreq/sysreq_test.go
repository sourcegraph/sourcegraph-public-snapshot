package sysreq

import (
	"errors"
	"reflect"
	"testing"

	"context"
)

func TestCheck(t *testing.T) {
	checks = []check{
		{
			Name: "a",
			Check: func(ctx context.Context) (problem, fix string, err error) {
				return "", "", errors.New("foo")
			},
		},
	}
	st := Check(context.Background(), nil)
	want := []Status{{Name: "a", Err: errors.New("foo")}}
	if !reflect.DeepEqual(st, want) {
		t.Errorf("got %v, want %v", st, want)
	}
}

func TestCheck_skip(t *testing.T) {
	checks = []check{
		{
			Name: "a",
			Check: func(ctx context.Context) (problem, fix string, err error) {
				return "", "", errors.New("foo")
			},
		},
	}
	st := Check(context.Background(), []string{"A"})
	want := []Status{{Name: "a", Skipped: true}}
	if !reflect.DeepEqual(st, want) {
		t.Errorf("got %v, want %v", st, want)
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_914(size int) error {
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
