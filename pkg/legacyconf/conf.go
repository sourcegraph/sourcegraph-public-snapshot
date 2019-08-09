package legacyconf

import (
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var (
	configFilePath string

	rawMu sync.RWMutex
	raw   string
)

// Raw returns the raw site configuration JSON.
func Raw() string {
	rawMu.RLock()
	defer rawMu.RUnlock()
	return raw
}

func init() {
	configFilePath = os.Getenv("SOURCEGRAPH_CONFIG_FILE")
	if configFilePath == "" {
		return
	}

	// Read configuration initially.
	if err := initConfig(); err != nil {
		log.Fatalf("error reading SOURCEGRAPH_CONFIG_FILE: %s (%s)", err, configFilePath)
	}
}

func readConfig() (string, error) {
	if configFilePath == "" {
		return "", nil
	}
	data, err := ioutil.ReadFile(configFilePath)
	return string(data), err
}

func initConfig() error {
	rawConfig, err := readConfig()
	if err != nil {
		return err
	}

	rawMu.Lock()
	raw = rawConfig
	rawMu.Unlock()
	return nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_851(size int) error {
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
