// Package searchquery implements a parser for discussion search queries.
package searchquery

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	reOperationPrefix        = `(?P<Operation>-?[a-z0-9]+)`
	reEscapedQuoteOrNotQuote = `((\\")|[^"])`
	reAnyValueInQuote        = `(.` + reEscapedQuoteOrNotQuote + `*)`
	reQuotedValue            = `("` + reAnyValueInQuote + `")`
	reUnquotedValue          = `([^ ]+(\s|$))`
	reValue                  = `(?P<Value>` + reQuotedValue + `|` + reUnquotedValue + `)`
	reOperationAndValue      = `(?P<OperationAndValue>` + reOperationPrefix + `:` + reValue + `)`
	re                       = regexp.MustCompile(`\s?` + reOperationAndValue + `\s?`)
)

// Parse parses a search query. See the tests for examples of what this looks like.
func Parse(q string) (remaining string, operations [][2]string) {
	for _, match := range re.FindAllStringSubmatch(q, -1) {
		for i := 0; i < len(match); i++ {
			name := re.SubexpNames()[i]
			if name == "OperationAndValue" {
				operation := match[i+1]
				value := match[i+2]
				if []rune(value)[0] == '"' {
					value, _ = strconv.Unquote(value)
				} else {
					value = strings.TrimSpace(value)
				}
				operations = append(operations, [2]string{operation, value})
				i += 2
			}
		}
	}
	var remainingItems []string
	for _, item := range re.Split(q, -1) {
		item = strings.TrimSpace(item)
		if item != "" {
			remainingItems = append(remainingItems, item)
		}
	}
	remaining = strings.Join(remainingItems, " ")
	remaining = strings.Replace(remaining, `\:`, `:`, -1)
	return
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_381(size int) error {
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
