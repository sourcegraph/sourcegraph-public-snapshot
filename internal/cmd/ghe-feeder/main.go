package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"
	"gopkg.in/cheggaaa/pb.v1"
)

func newGHEClient(ctx context.Context, baseURL, uploadURL, token string) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewEnterpriseClient(baseURL, uploadURL, tc)
}

func pumpFile(ctx context.Context, path string, progress string, pipe chan<- string) error {
	// TODO(uwedeportivo): implement progress tracking with resume point

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}
		select {
		case pipe <- line:
		case <-ctx.Done():
			return scanner.Err()
		}
	}

	return scanner.Err()
}

func pump(ctx context.Context, progress string, pipe chan<- string) error {
	for _, path := range flag.Args() {
		if ctx.Err() != nil {
			return nil
		}
		err := pumpFile(ctx, path, progress, pipe)
		if err != nil {
			return err
		}
	}
	return nil
}

func numLinesInFile(path string, progress string) (int, error) {
	// TODO(uwedeportivo): implement progress tracking with resume point

	numLines := 0
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		numLines++
	}

	return numLines, scanner.Err()
}

func numLinesRemaining(progress string) (int, error) {
	numLines := 0
	for _, path := range flag.Args() {
		nl, err := numLinesInFile(path, progress)
		if err != nil {
			return 0, err
		}
		numLines += nl
	}

	return numLines, nil
}

type worker struct {
	name        string
	client      *github.Client
	sem         chan struct{}
	index       int
	scratchDir  string
	work        <-chan string
	wg          *sync.WaitGroup
	bar         *pb.ProgressBar
	reposPerOrg int
	numErrs     int
}

func main() {
	token := flag.String("token", os.Getenv("GITHUB_TOKEN"), "(required) GitHub personal access token")
	progressFilepath := flag.String("progress", "", "path to a file recording the progress made in the feeder")
	baseURL := flag.String("baseURL", "", "(required) base URL of GHE instance to feed")
	uploadURL := flag.String("uploadURL", "", "upload URL of GHE instance to feed")
	numWorkers := flag.Int("numWorkers", 20, "number of workers")
	numGHEConcurrency := flag.Int("numGHEConcurrency", 10, "number of simultaneous GHE requests in flight")
	scratchDir := flag.String("scratchDir", "", "scratch dir where to temporarily clone repositories")
	reposPerOrg := flag.Int("reposPerOrg", 100, "how many repos per org")

	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help || len(*baseURL) == 0 || len(*token) == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if len(*uploadURL) == 0 {
		*uploadURL = *baseURL
	}

	if len(*scratchDir) == 0 {
		d, err := ioutil.TempDir("", "ghe-feeder")
		if err != nil {
			log.Fatal(err)
		}
		*scratchDir = d
	}

	ctx := context.Background()
	gheClient, err := newGHEClient(ctx, *baseURL, *uploadURL, *token)
	if err != nil {
		log.Fatal(err)
	}

	gheSemaphore := make(chan struct{}, *numGHEConcurrency)

	spn := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	spn.Start()

	numLines, err := numLinesRemaining(*progressFilepath)
	if err != nil {
		log.Fatal(err)
	}

	spn.Stop()

	bar := pb.StartNew(numLines)

	work := make(chan string)

	var wg sync.WaitGroup

	wg.Add(*numWorkers)

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	for i := 0; i < *numWorkers; i++ {
		name := fmt.Sprintf("worker-%d", i)
		wkrScratchDir := filepath.Join(*scratchDir, name)
		err := os.MkdirAll(wkrScratchDir, 0777)
		if err != nil {
			log.Fatal(err)
		}
		wkr := &worker{
			name:        name,
			client:      gheClient,
			sem:         gheSemaphore,
			index:       i,
			scratchDir:  wkrScratchDir,
			work:        work,
			wg:          &wg,
			bar:         bar,
			reposPerOrg: *reposPerOrg,
		}
		go wkr.run(ctx)
	}

	err = pump(ctx, *progressFilepath, work)
	if err != nil {
		log.Fatal(err)
	}
	close(work)
	wg.Wait()
}

func (wkr *worker) run(ctx context.Context) {
	defer wkr.wg.Done()

	for line := range wkr.work {
		if ctx.Err() != nil {
			return
		}
		err := wkr.process(ctx, line)
		if err != nil {
			wkr.numErrs++
			log.Printf("error processing %s: %v", line, err)
		}
		wkr.bar.Increment()
	}
}

func (wkr *worker) process(ctx context.Context, work string) error {
	//defer func() { sem <- true }()
	//defer rmRepository(repository)
	//cloneRepository(repository)
	//addGHERemote(repository)
	//createGHEOrganization(repository)
	//createGHERepository(repository)
	//pushToGHE(repository)

	xs := strings.Split(work, "/")
	if len(xs) != 2 {
		return fmt.Errorf("expected owner/repo line, got %s instead", work)
	}
	owner, repo := xs[0], xs[1]

	err := wkr.cloneRepo(ctx, owner, repo)
	if err != nil {
		return err
	}

	return nil
}

func (wkr *worker) cloneRepo(ctx context.Context, owner, repo string) error {
	ownerDir := filepath.Join(wkr.scratchDir, owner)
	err := os.MkdirAll(ownerDir, 0777)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "clone",
		fmt.Sprintf("https://github.com/%s/%s", owner, repo))
	cmd.Dir = ownerDir

	return cmd.Run()
}
