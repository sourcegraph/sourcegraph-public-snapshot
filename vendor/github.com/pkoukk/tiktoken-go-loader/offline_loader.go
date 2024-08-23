package tiktoken_loader

import (
	"encoding/base64"
	"path"
	"strconv"
	"strings"

	"github.com/pkoukk/tiktoken-go-loader/assets"
)

type OfflineLoader struct{}

func (l *OfflineLoader) LoadTiktokenBpe(tiktokenBpeFile string) (map[string]int, error) {
	baseFileName := path.Base(tiktokenBpeFile)
	contents, err := assets.Assets.ReadFile(baseFileName)
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

func NewOfflineLoader() *OfflineLoader {
	return &OfflineLoader{}
}
