package git

import (
	"os"
	"strconv"
	"strings"
)

func isFile(filePath string) bool {
	f, e := os.Stat(filePath)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

func RefEndName(refStr string) string {
	index := strings.LastIndex(refStr, "/")
	if index != -1 {
		return refStr[index+1:]
	}
	return refStr
}

func StrToInt(str string) (int, error) {
	n, err := strconv.ParseInt(str, 10, 64)
	nn := int(n)
	return nn, err
}

func IntToStr(n int) string {
	return strconv.FormatInt(int64(n), 10)
}
