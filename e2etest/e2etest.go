package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/auth/sharedsecret"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"github.com/nlopes/slack"
	"github.com/rogpeppe/rog-go/parallel"
	"sourcegraph.com/sourcegraph/go-selenium"
)

// Test represents a single E2E test.
type Test struct {
	// Name is the name of your test, which should be short and readable, e.g.,
	// "register_and_login".
	Name string

	// Description is a more verbose description of your test, e.g., "Registers a
	// new user account and logs in to it.".
	Description string

	// Func is called to perform the test. If an error is returned, the test is
	// considered failed.
	//
	// Tests must log all output to t.Log instead of via other logging packages.
	Func func(t *TestSuite) error
}

// Register should be called inside of an init function in order to register a
// new test as part of the testsuite.
func Register(t *Test) {
	testSuite.tests = append(testSuite.tests, t)
}

// TestSuite is provided as input to each test and provides generic helper
// methods to make testing easier.
type TestSuite struct {
	// Log is where all errors, warnings, etc. should be written to.
	Log *log.Logger

	// Target is the target Sourcegraph server to test, e.g. https://sourcegraph.com
	Target string

	tests    []*Test
	executor string
	idKey    *idkey.IDKey

	slack              *slack.Client
	slackChannel       slack.Channel
	slackLogBuffer     *bytes.Buffer
	slackSkipAtChannel bool
}

// WebDriver returns a new remote Selenium webdriver.
func (t *TestSuite) WebDriver() selenium.WebDriver {
	caps := selenium.Capabilities(map[string]interface{}{
		"browserName": "chrome",
	})
	d, err := selenium.NewRemote(caps, t.executor)
	if err != nil {
		log.Fatal(err)
	}
	return d
}

// WebDriverT returns a new remote Selenium webdriver which handles failure
// cases automatically for you.
func (t *TestSuite) WebDriverT() selenium.WebDriverT {
	return t.WebDriver().T(t)
}

// Fatalf implements the selenium.TestingT interface. Because unlike the testing
// package we are a single process, we instead cause a panic (which is caught
// by the test executor).
func (t *TestSuite) Fatalf(fmtStr string, v ...interface{}) {
	panic(fmt.Sprintf(fmtStr, v...))
}

// Endpoint returns an absolute URL given one relative to the target isntance
// root. For example, if t.Target == "https://sourcegraph.com", Endpoint("/login")
// will return "https://sourcegraph.com/login"
func (t *TestSuite) Endpoint(e string) string {
	u, err := url.Parse(t.Target)
	if err != nil {
		panic(err) // Target is validated in main, always.
	}
	u.Path = path.Join(u.Path, e)
	return u.String()
}

// GRPCClient returns a new authenticated Sourcegraph gRPC client. It uses the
// server's ID key, and thus has 100% unrestricted access. Use with caution!
func (t *TestSuite) GRPCClient() (context.Context, *sourcegraph.Client) {
	target, err := url.Parse(t.Target)
	if err != nil {
		panic(err) // Target is validated in main, always.
	}

	// Create context with gRPC endpoint and idKey credentials.
	ctx := context.TODO()
	ctx = sourcegraph.WithGRPCEndpoint(ctx, target)
	ctx = sourcegraph.WithCredentials(ctx, sharedsecret.TokenSource(t.idKey, "internal:e2etest"))

	// Create client.
	c, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		t.Fatalf("could not create gRPC client:", c)
	}
	return ctx, c
}

func (t *TestSuite) slackMessage(msg, quoted string) {
	if t.slack == nil {
		return
	}

	params := slack.PostMessageParameters{
		Username:  "e2etest",
		Parse:     "full",
		IconEmoji: ":shield:",
		Attachments: []slack.Attachment{
			slack.Attachment{
				Text: quoted,
			},
		},
	}
	_, _, err := t.slack.PostMessage(t.slackChannel.ID, msg, params)
	if err != nil {
		log.Println(err)
		return
	}
}

// run runs the test suite over and over again against $TARGET, if set,
// otherwise it runs the test suite just once.
func (t *TestSuite) run() {
	shouldLogSuccess := 0
	for {
		if t.runTests(shouldLogSuccess < 5) {
			shouldLogSuccess++
			if shouldLogSuccess == 5 {
				t.slackMessage(":star: *Five consecutive successes!* (silencing output until next failure)", "")
			}
		} else {
			shouldLogSuccess = 0
		}

		if *runOnce {
			break
		}
	}
}

// runTests runs all of the tests and handles failures. It returns whether or
// not all tests were successful.
func (t *TestSuite) runTests(logSuccess bool) bool {
	// Execute the registered tests in parallel.
	var (
		successMu sync.Mutex
		success   int
		start     = time.Now()
		run       = parallel.NewRun(len(t.tests))
		total     = 0
	)
	for _, testToCopy := range t.tests {
		// If they want to run specifically just one test, check for that now.
		if *runFlag != "" {
			if testToCopy.Name != *runFlag {
				continue
			}
		}
		total++

		test := testToCopy
		run.Do(func() error {
			unitStart := time.Now()
			err := t.runTest(test)
			unitTime := time.Since(unitStart)
			if err != nil {
				t.Log.Printf("[failure] [%v] [%v]: %v\n", test.Name, unitTime, err)
				return nil
			}

			t.Log.Printf("[success] [%v] [%v]\n", test.Name, unitTime)
			successMu.Lock()
			success++
			successMu.Unlock()
			return nil
		})
	}
	run.Wait()

	t.Log.Printf("%v tests finished in %v [%v success] [%v failure]\n", total, time.Since(start), success, total-success)

	if total == success {
		t.slackSkipAtChannel = false // do @channel on next failure
		if logSuccess {
			t.slackMessage(fmt.Sprintf(":thumbsup: *Success! %v tests successful against %v!*", total, t.Target), "")
		}
	} else {
		// Only send @channel on the first failure, not all consecutive ones (that
		// would be very annoying).
		atChannel := ""
		if !t.slackSkipAtChannel {
			t.slackSkipAtChannel = true
			atChannel = " @channel"
		}
		t.slackMessage(
			fmt.Sprintf(":fire: *FAILURE! %v/%v tests failed against %v: *"+atChannel, total-success, total, t.Target),
			t.slackLogBuffer.String(),
		)

		// TODO(slimsag): trigger rollback here via #dev-deploy-bot
	}
	t.slackLogBuffer.Reset()
	return total == success
}

// runTest runs a single test and recovers from panics, should they occur.
func (t *TestSuite) runTest(test *Test) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if err == nil {
				err = errors.New(fmt.Sprint(r))
			}
		}
	}()
	err = test.Func(t)
	return
}

var testSuite = &TestSuite{
	slackLogBuffer: &bytes.Buffer{},
}

var (
	runOnce = flag.Bool("once", true, "run the tests only once (true) or forever (false)")
	runFlag = flag.String("run", "", "specify an exact test name to run (e.g. 'login_flow', 'register_flow')")
)

func main() {
	flag.Parse()

	// Prepare logging.
	testSuite.Log = log.New(io.MultiWriter(os.Stderr, testSuite.slackLogBuffer), "", 0)

	// Determine which Selenium server to connect to.
	serverAddr := os.Getenv("SELENIUM_SERVER")
	if serverAddr == "" {
		log.Fatal("Unable to get SELENIUM_SERVER address from environment")
	}
	u, err := url.Parse(serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	u.Path = path.Join(u.Path, "wd/hub")
	testSuite.executor = u.String()

	// Determine the target Sourcegraph instance to test against.
	testSuite.Target = os.Getenv("TARGET")
	if testSuite.Target == "" {
		log.Fatal("Unable to get TARGET Sourcegraph instance from environment")
	}

	// Find server ID key information.
	if key := os.Getenv("ID_KEY_DATA"); key != "" {
		testSuite.idKey, err = idkey.FromString(key)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		sgpath := os.Getenv("SGPATH")
		if sgpath == "" {
			currentUser, err := user.Current()
			if err != nil {
				log.Fatal(err)
			}
			sgpath = filepath.Join(currentUser.HomeDir, ".sourcegraph")
		}
		data, err := ioutil.ReadFile(filepath.Join(sgpath, "id.pem"))
		if err != nil {
			log.Fatal(err)
		}
		testSuite.idKey, err = idkey.New(data)
		if err != nil {
			log.Fatal(err)
		}
	}

	if token := os.Getenv("SLACK_API_TOKEN"); token != "" {
		testSuite.slack = slack.New(token)

		// Find the channel ID.
		channelName := os.Getenv("SLACK_CHANNEL")
		if channelName == "" {
			channelName = "e2etest"
		}
		channels, err := testSuite.slack.GetChannels(true)
		if err != nil {
			log.Fatal(err)
		}
		found := false
		for _, c := range channels {
			if c.Name == "e2etest" {
				found = true
				testSuite.slackChannel = c
			}
		}
		if !found {
			log.Println("could not find slack channel", channelName)
			log.Println("disabling slack notifications")
			testSuite.slack = nil
		} else {
			registeredTests := &bytes.Buffer{}
			for _, t := range testSuite.tests {
				fmt.Fprintf(registeredTests, "[%v]: %v\n", t.Name, t.Description)
			}
			testSuite.slackMessage(":shield: *Ready and reporting for duty!* Registered tests:", registeredTests.String())
		}
	}

	testSuite.run()
}
