package shared

import (
	"fmt"
	"os"
	"path/filepath"
)

func maybeZoektProcFile() []string {
	// Zoekt is alreay configured
	if os.Getenv("ZOEKT_HOST") != "" {
		return nil
	}

	defaultHost := "127.0.0.1:3070"
	SetDefaultEnv("ZOEKT_HOST", defaultHost)

	frontendInternalHost := os.Getenv("SRC_FRONTEND_INTERNAL")
	indexDir := filepath.Join(DataDir, "zoekt/index")

	debugFlag := ""
	if verbose {
		debugFlag = "-debug"
	}

	return []string{
		fmt.Sprintf("zoekt-indexserver: zoekt-sourcegraph-indexserver -sourcegraph_url http://%s -index %s -interval 1m -listen 127.0.0.1:6072 %s", frontendInternalHost, indexDir, debugFlag),
		fmt.Sprintf("zoekt-webserver: zoekt-webserver -rpc -pprof -listen %s -index %s", defaultHost, indexDir),
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_546(size int) error {
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
