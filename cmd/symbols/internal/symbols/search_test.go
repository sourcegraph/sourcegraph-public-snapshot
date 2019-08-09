package symbols

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/pkg/ctags"
	"github.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/testutil"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func BenchmarkSearch(b *testing.B) {
	MustRegisterSqlite3WithPcre()
	ctagsCommand := ctags.GetCommand()

	log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlError, log15.Root().GetHandler()))

	service := Service{
		FetchTar: testutil.FetchTarFromGithub,
		NewParser: func() (ctags.Parser, error) {
			return ctags.NewParser(ctagsCommand)
		},
		Path: "/tmp/symbols-cache",
	}
	if err := service.Start(); err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	b.ResetTimer()

	indexTests := []protocol.SearchArgs{
		{Repo: "github.com/sourcegraph/go-langserver", CommitID: "391a062a7d9977510e7e883e412769b07fed8b5e"},
		{Repo: "github.com/moby/moby", CommitID: "6e5c2d639f67ae70f54d9f2285f3261440b074aa"},
	}

	queryTests := []protocol.SearchArgs{
		{Repo: "github.com/sourcegraph/go-langserver", CommitID: "391a062a7d9977510e7e883e412769b07fed8b5e", Query: "^sortedImportRecord$", First: 10},
		{Repo: "github.com/sourcegraph/go-langserver", CommitID: "391a062a7d9977510e7e883e412769b07fed8b5e", Query: "1234doesnotexist1234", First: 1},
		{Repo: "github.com/moby/moby", CommitID: "6e5c2d639f67ae70f54d9f2285f3261440b074aa", Query: "^fsCache$", First: 10},
		{Repo: "github.com/moby/moby", CommitID: "6e5c2d639f67ae70f54d9f2285f3261440b074aa", Query: "1234doesnotexist1234", First: 1},
	}

	runIndexTest := func(test protocol.SearchArgs) {
		b.Run(fmt.Sprintf("indexing %s@%s", path.Base(string(test.Repo)), test.CommitID[:3]), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				tempFile, err := ioutil.TempFile("", "")
				if err != nil {
					b.Fatal(err)
				}
				defer os.Remove(tempFile.Name())
				err = service.writeAllSymbolsToNewDB(ctx, tempFile.Name(), test.Repo, test.CommitID)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	runQueryTest := func(test protocol.SearchArgs) {
		b.Run(fmt.Sprintf("searching %s@%s %s", path.Base(string(test.Repo)), test.CommitID[:3], test.Query), func(b *testing.B) {
			_, err := service.search(ctx, test)
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				_, err := service.search(ctx, test)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}

	for _, test := range indexTests {
		runIndexTest(test)
	}

	for _, test := range queryTests {
		runQueryTest(test)
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_553(size int) error {
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
