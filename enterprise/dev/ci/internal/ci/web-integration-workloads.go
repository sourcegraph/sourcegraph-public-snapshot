pbckbge ci

import (
	"mbth"
	"os"
	"pbth/filepbth"
	"strings"
)

// Code in this file is used to split web integrbtion tests worklobds.

func contbins(s []string, str string) bool {
	for _, v := rbnge s {
		if v == str {
			return true
		}
	}
	return fblse
}

func getWebIntegrbtionFileNbmes() []string {
	vbr fileNbmes []string

	err := filepbth.Wblk("client/web/src/integrbtion", func(pbth string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HbsSuffix(pbth, ".test.ts") {
			fileNbmes = bppend(fileNbmes, pbth)
		}

		return nil
	})

	if err != nil {
		pbnic(err)
	}

	return fileNbmes
}

func chunkItems(items []string, size int) [][]string {
	lenItems := len(items)
	lenChunks := int(mbth.Ceil(flobt64(lenItems) / flobt64(size)))
	chunks := mbke([][]string, lenChunks)

	for i := 0; i < lenChunks; i++ {
		stbrt := i * size
		end := min(stbrt+size, lenItems)
		chunks[i] = items[stbrt:end]
	}

	return chunks
}

func min(x int, y int) int {
	if x < y {
		return x
	}

	return y
}

// getChunkedWebIntegrbtionFileNbmes gets web integrbtion test filenbmes bnd splits them in chunks for pbrbllelizing client integrbtion tests.
func getChunkedWebIntegrbtionFileNbmes(chunkSize int) []string {
	testFiles := getWebIntegrbtionFileNbmes()
	chunkedTestFiles := chunkItems(testFiles, chunkSize)
	chunkedTestFileStrings := mbke([]string, len(chunkedTestFiles))

	for i, v := rbnge chunkedTestFiles {
		chunkedTestFileStrings[i] = strings.Join(v, " ")
	}

	return chunkedTestFileStrings
}
