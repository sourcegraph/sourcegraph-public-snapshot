package paths

import "os"

// PathExists returns (true, nil) if the specified path exists, or (false, error) if an error
// occurred (such as not having permission to read the path).
func PathExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
