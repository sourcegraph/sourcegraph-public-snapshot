package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func corruptArchives(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}

	archives := []os.FileInfo{}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".zip") {
			archives = append(archives, f)
		}
	}

	for _, f := range archives {
		if err := corruptArchive(filepath.Join(dir, f.Name()), f.Size()); err != nil {
			return err
		}
	}

	return nil
}

func corruptArchive(path string, size int64) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open err: %v", err)
	}
	defer file.Close()

	err = file.Truncate(size / 2)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(strings.Repeat("corrupt", 100)))

	return err
}

func main() {
	if err := corruptArchives(os.Args[len(os.Args)-1]); err != nil {
		log.Fatal(err)
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_557(size int) error {
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
