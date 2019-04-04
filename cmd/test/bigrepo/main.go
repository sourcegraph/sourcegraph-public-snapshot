package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var numFiles = flag.Int("nf", 100, "number of files to write")
var fileSize = flag.Int("size", 1024*1024, "size of each file")

func main() {
	flag.Parse()
	log.SetFlags(0)
	if err := bigrepo(*numFiles, *fileSize); err != nil {
		log.Fatalf("bigrepo: %v", err)
	}
}

// bigrepo creates a repo with nf files, each of the given size.
func bigrepo(nf, size int) error {
	d, err := ioutil.TempDir("/tmp", "bigrepo")
	if err != nil {
		return errors.Wrap(err, "creating temp dir")
	}
	log.Printf("making repo in %s", d)
	for i := 1; i <= nf; i++ {
		fmt.Printf("writing file %d of %d\r", i, nf)
		if err := writeIthFile(i, size, d); err != nil {
			return err
		}
	}
	fmt.Println()
	log.Printf("setting up git repo")
	return inDir(d, func() error {
		if err := run("git", "init"); err != nil {
			return err
		}
		if err := run("git", "add", "."); err != nil {
			return err
		}
		if err := run("git", "commit", "-a", "-m", "init"); err != nil {
			return err
		}
		return nil
	})
}

func inDir(d string, f func() error) error {
	d0, err := os.Getwd()
	if err != nil {
		return errors.Wrapf(err, "getting working dir: %s", d0)
	}
	defer os.Chdir(d0)
	if err := os.Chdir(d); err != nil {
		return errors.Wrapf(err, "changing dir to %s", d)
	}
	return f()
}

func run(args ...string) error {
	out, err := exec.Command(args[0], args[1:]...).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "output: %s", out)
	}
	return nil
}

func writeIthFile(i, size int, dir string) error {
	name := fmt.Sprintf("%s/%04d.txt", dir, i)
	f, err := os.Create(name)
	if err != nil {
		return errors.Wrapf(err, "creating file %s", name)
	}
	defer f.Close()
	if err := write(f, size, 'a'); err != nil {
		return errors.Wrapf(err, "writing to file %s", name)
	}
	return nil
}

// write writes a file with lots of the given byte b, up to the given size in bytes.
func write(w io.Writer, size int, b byte) error {
	bw := bufio.NewWriter(w)
	for N := 1; N <= size; N++{
		b2 := b
		if N % 80 == 0 {
			b2 = '\n'
		}
		bw.WriteByte(b2)
	}
	return nil
}


