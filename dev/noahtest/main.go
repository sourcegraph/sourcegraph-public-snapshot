package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/morikuni/aec"
)

func main() {
	wd, _ := os.Getwd()
	f, err := os.Open(path.Join(wd, "package.json"))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b := bufio.NewScanner(f)

	fmt.Print(aec.Hide)
	var i int
	const lastNLines = 10
	for i := 0; i < lastNLines; i++ {
		fmt.Println()
	}
	buffer := make([]string, lastNLines)

	for b.Scan() {
		time.Sleep(time.Millisecond * 10)
		buffer[i] = b.Text()
		i = (i + 1) % lastNLines
		fmt.Print(aec.Up(lastNLines))
		for j := 0; j < lastNLines; j++ {
			fmt.Print(aec.EraseLine(aec.EraseModes.All))
			fmt.Println(buffer[(i+j)%lastNLines])
		}
	}
}
