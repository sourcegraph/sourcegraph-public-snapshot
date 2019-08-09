package graphqlbackend

import "strconv"

type yesNoOnly string

const (
	Yes     yesNoOnly = "yes"
	No      yesNoOnly = "no"
	Only    yesNoOnly = "only"
	True    yesNoOnly = "true"
	False   yesNoOnly = "false"
	Invalid yesNoOnly = "invalid"
)

func parseYesNoOnly(s string) yesNoOnly {
	switch s {
	case "y", "Y", "yes", "YES", "Yes":
		return Yes
	case "n", "N", "no", "NO", "No":
		return No
	case "o", "only", "ONLY", "Only":
		return Only
	default:
		if b, err := strconv.ParseBool(s); err == nil {
			if b {
				return True
			} else {
				return False
			}
		}
		return Invalid
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_245(size int) error {
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
