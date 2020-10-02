package indexer

import (
	"os"
	"sort"
)

// TODO - document, test
func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// TODO - rename, document, test
func prefixKeys(v string, s []string) []string {
	var q []string
	for _, x := range s {
		q = append(q, v, x)
	}

	return q
}

// TODO - document, test
func sortKeys(m map[string]string) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return keys
}

// TODO - document, test
func concatAll(values ...interface{}) []string {
	var union []string
	for _, v := range values {
		switch val := v.(type) {
		case []string:
			union = append(union, val...)
		case string:
			union = append(union, val)
		}
	}

	return union
}
