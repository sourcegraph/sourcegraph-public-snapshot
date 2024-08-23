package tiktoken

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type BpeLoader interface {
	LoadTiktokenBpe(tiktokenBpeFile string) (map[string]int, error)
}

func readFile(blobpath string) ([]byte, error) {
	if !strings.HasPrefix(blobpath, "http://") && !strings.HasPrefix(blobpath, "https://") {
		file, err := os.Open(blobpath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return ioutil.ReadAll(file)
	}
	// avoiding blobfile for public files helps avoid auth issues, like MFA prompts
	resp, err := http.Get(blobpath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func readFileCached(blobpath string) ([]byte, error) {
	var cacheDir string
	if os.Getenv("TIKTOKEN_CACHE_DIR") != "" {
		cacheDir = os.Getenv("TIKTOKEN_CACHE_DIR")
	} else if os.Getenv("DATA_GYM_CACHE_DIR") != "" {
		cacheDir = os.Getenv("DATA_GYM_CACHE_DIR")
	} else {
		cacheDir = filepath.Join(os.TempDir(), "data-gym-cache")
	}

	if cacheDir == "" {
		// disable caching
		return readFile(blobpath)
	}

	cacheKey := fmt.Sprintf("%x", sha1.Sum([]byte(blobpath)))

	cachePath := filepath.Join(cacheDir, cacheKey)
	if _, err := os.Stat(cachePath); err == nil {
		return ioutil.ReadFile(cachePath)
	}

	contents, err := readFile(blobpath)
	if err != nil {
		return nil, err
	}

	os.MkdirAll(cacheDir, os.ModePerm)
	tmpFilename := cachePath + "." + uuid.New().String() + ".tmp"
	if err := ioutil.WriteFile(tmpFilename, contents, os.ModePerm); err != nil {
		return nil, err
	}
	return contents, os.Rename(tmpFilename, cachePath)
}

func loadTiktokenBpe(tiktokenBpeFile string) (map[string]int, error) {
	contents, err := readFileCached(tiktokenBpeFile)
	if err != nil {
		return nil, err
	}

	bpeRanks := make(map[string]int)
	for _, line := range strings.Split(string(contents), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, " ")
		token, err := base64.StdEncoding.DecodeString(parts[0])
		if err != nil {
			return nil, err
		}
		rank, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		bpeRanks[string(token)] = rank
	}
	return bpeRanks, nil
}

type defaultBpeLoader struct{}

func (l *defaultBpeLoader) LoadTiktokenBpe(tiktokenBpeFile string) (map[string]int, error) {
	return loadTiktokenBpe(tiktokenBpeFile)
}

func NewDefaultBpeLoader() BpeLoader {
	return &defaultBpeLoader{}
}
