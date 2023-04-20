package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
)

func tryClassicDecode(filename string) (*embeddings.RepoEmbeddingIndex, error) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	decoder := gob.NewDecoder(f)
	index := embeddings.RepoEmbeddingIndex{}
	err = decoder.Decode(&index)
	return &index, err
}

func tryModernDecode(filename string) (*embeddings.RepoEmbeddingIndex, error) {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	decoder := gob.NewDecoder(f)
	return embeddings.DecodeRepoEmbeddingIndex(decoder)
}

func decode(filename string) (*embeddings.RepoEmbeddingIndex, error) {
	index, err := tryModernDecode(filename)
	if err == nil {
		return index, nil
	}
	return tryClassicDecode(filename)
}

type MeanVector struct {
       vector []float64
       count uint
}

func makeMeanVector(n int) *MeanVector {
   return &MeanVector {
     make([]float64, n),
     0,
   }
}

func (m *MeanVector) merge(vec []float32) {
   for i, v := range vec {
       m.vector[i] += float64(v)
   }
   m.count += 1
}

func (m *MeanVector) toEmbedding() (result []float32) {
   result = make([]float32, len(m.vector))
   for i, v := range(m.vector) {
       result[i] = float32(v / float64(m.count))
   }
   return
}


func extension(s string) string {
     for {
       j := strings.IndexRune(s, '.')
       if j == -1 {
       	  return s
       }
       s = s[j + 1:]
     }
}


func main() {
	manifest, err := os.Open("/Users/dpc/embeddings/manifest.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer manifest.Close()
	reader := bufio.NewScanner(manifest)
	reader.Split(bufio.ScanLines)

	m := make(map[string]*MeanVector)

	for reader.Scan() {
		filename := fmt.Sprintf("/Users/dpc/embeddings/%s", reader.Text())
		index, err := decode(filename)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for i, metadata := range index.CodeIndex.RowMetadata {
		    ext := extension(metadata.FileName)
		    means, ok := m[ext]
		    if !ok {
		       means = makeMeanVector(index.CodeIndex.ColumnDimension)
		       m[ext] = means
		    }
		    means.merge(index.CodeIndex.Embeddings[i * index.CodeIndex.ColumnDimension:(i+1) * index.CodeIndex.ColumnDimension])
		}
	}

	n := make(map[string][]float32)

	for k, v := range m {
	    if v.count > 1000 {
	            log.Print(k)
	    	    n[k] = v.toEmbedding()
	    }
	}

	f, err := os.OpenFile("means.gob", os.O_RDWR | os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	encoder := gob.NewEncoder(f)
	err = encoder.Encode(n)
	if err != nil {
	   log.Fatal(err)
	}
}
