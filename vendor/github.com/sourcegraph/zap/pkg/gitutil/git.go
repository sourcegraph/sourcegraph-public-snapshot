package gitutil

import (
	"bytes"
	"fmt"
	"os"
)

const EmptyTreeSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904" // `git hash-object -t tree /dev/null`

func AbbrevRev(rev string) string {
	const n = 6
	if len(rev) == 40 {
		return rev[:n]
	}
	return rev
}

const RegularFileNonExecutableMode = "100644"

func objectTypeForMode(modeStr string) (string, error) {
	switch modeStr {
	case "040000": // directory
		return "tree", nil
	case RegularFileNonExecutableMode, "100755": // regular file
		return "blob", nil
	case "120000": // symlink
		return "blob", nil
	case "160000": // submodule
		return "commit", nil
	default:
		return "", fmt.Errorf("unrecognized git mode %q", modeStr)
	}
}

func modeForOSFileMode(mode os.FileMode) (string, error) {
	switch {
	case mode&os.ModeSymlink != 0:
		return "120000", nil
	case mode.IsDir():
		return "040000", nil
	case mode.IsRegular():
		p := mode.Perm()
		switch p {
		case 0644:
			return "100644", nil
		case 0755:
			return "100755", nil
		}
		if p&0111 != 0 {
			return "100755", nil // maintain +x
		}
		return "100644", nil
	}
	return "", fmt.Errorf("unable to map OS file mode %s to git mode", mode)
}

func parseLsTreeLine(line []byte) (mode, typ, oid, path string, err error) {
	partsTab := bytes.SplitN(line, []byte("\t"), 2)
	if len(partsTab) != 2 {
		err = fmt.Errorf("bad ls-tree line: %q", line)
		return
	}

	path = string(partsTab[1])
	partsFirst := bytes.Split(partsTab[0], []byte(" "))
	if len(partsFirst) != 3 {
		err = fmt.Errorf("bad ls-tree line section (before first TAB): %q", partsTab[0])
		return
	}
	mode = string(partsFirst[0])
	typ = string(partsFirst[1])
	oid = string(partsFirst[2])
	return
}
