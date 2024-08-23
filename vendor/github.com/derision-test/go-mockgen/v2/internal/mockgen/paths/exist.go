package paths

import "os"

func DirExists(path string) bool {
	if info, err := os.Stat(path); err == nil {
		return info.IsDir()
	}

	return false
}

func Exists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = nil
		}

		return false, err
	}

	return true, nil
}

func AnyExists(paths []string) (string, error) {
	for _, path := range paths {
		exists, err := Exists(path)
		if err != nil {
			return "", err
		}

		if exists {
			return path, nil
		}
	}

	return "", nil
}

func EnsureDirExists(dirname string) error {
	exists, err := Exists(dirname)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return os.MkdirAll(dirname, os.ModeDir|os.ModePerm)
}
