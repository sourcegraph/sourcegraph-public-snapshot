package e2etest

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/auth/idkey"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/sharedsecret"
	"sourcegraph.com/sourcegraph/sourcegraph/e2etest/e2etestuser"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"

	"github.com/nlopes/slack"
	"github.com/rogpeppe/rog-go/parallel"
	"sourcegraph.com/sourcegraph/go-selenium"
)

// T is passed as context into all tests. It provides generic helper methods to
// make life during testing easier.
type T struct {
	selenium.WebDriverT

	// Log is where all errors, warnings, etc. should be written to.
	Log *log.Logger

	// Target is the target Sourcegraph server to test, e.g. https://sourcegraph.com
	Target string

	// TestLogin is a username prefixed with e2etestuser.Prefix which is unique
	// for this test. In specific it is e2etestuser.Prefix + Test.Name.
	TestLogin string

	// TestEmail is a email address prefixed with e2etestuser.Prefix which is
	// unique for this test. In specific it is e2etestuser.Prefix + Test.Name + "@sourcegraph.com".
	TestEmail string

	// WebDriver is the underlying selenium web driver. Useful if you want to
	// handle errors yourself (the embedded WebDriverT handles them for you by
	// calling Fatalf).
	WebDriver selenium.WebDriver

	tr *testRunner
}

type internalError struct {
	err error
}

func (e *internalError) Error() string {
	return e.err.Error()
}

// Fatalf implements the selenium.TestingT interface. Because unlike the testing
// package we are a single process, we instead cause a panic (which is caught
// by the test executor).
func (t *T) Fatalf(fmtStr string, v ...interface{}) {
	panic(fmt.Sprintf(fmtStr, v...))
}

// Endpoint returns an absolute URL given one relative to the target instance
// root. For example, if t.Target == "https://sourcegraph.com", Endpoint("/login")
// will return "https://sourcegraph.com/login"
func (t *T) Endpoint(e string) string {
	u, err := url.Parse(t.Target)
	if err != nil {
		panic(err) // Target is validated in main, always.
	}
	u.Path = path.Join(u.Path, e)
	return u.String()
}

// GRPCClient returns a new authenticated Sourcegraph gRPC client. It uses the
// server's ID key, and thus has 100% unrestricted access. Use with caution!
func (t *T) GRPCClient() (context.Context, *sourcegraph.Client) {
	target, err := url.Parse(t.Target)
	if err != nil {
		panic(err) // Target is validated in main, always.
	}

	// Create context with gRPC endpoint and idKey credentials.
	ctx := context.TODO()
	ctx = sourcegraph.WithGRPCEndpoint(ctx, target)
	ctx = sourcegraph.WithCredentials(ctx, sharedsecret.TokenSource(t.tr.idKey, "internal:e2etest"))

	// Create client.
	c, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		t.Fatalf("could not create gRPC client:", c)
	}
	return ctx, c
}

// WaitForCondition waits up to d for cond() to be true. After each time cond()
// is called, time.Sleep(optimisticD) is called. If timeout occurs, t.Fatalf is
// invoked.
func (t *T) WaitForCondition(d time.Duration, optimisticD time.Duration, cond func() bool, condName string) {
	start := time.Now()
	for time.Now().Sub(start) < d {
		time.Sleep(optimisticD)
		if cond() {
			return
		}
	}
	t.Fatalf("timed out waiting %v for condition %q to be met", d, condName)
}

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
	Func func(t *T) error
}

// Register should be called inside of an init function in order to register a
// new test as part of the testsuite.
func Register(t *Test) {
	tr.tests = append(tr.tests, t)
}

// testRunner is provided as input to each test and provides generic helper
// methods to make testing easier.
type testRunner struct {
	log      *log.Logger
	target   string
	tests    []*Test
	executor string
	idKey    *idkey.IDKey

	slack              *slack.Client
	slackToken         string
	slackChannel       slack.Channel
	slackLogBuffer     *bytes.Buffer
	slackSkipAtChannel bool
}

func (t *testRunner) slackMessage(msg, quoted string) {
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

// run runs the test suite over and over again against $TARGET, if $TARGET is set,
// otherwise it runs the test suite just once.
func (t *testRunner) run() {
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
func (t *testRunner) runTests(logSuccess bool) bool {
	// Execute the registered tests in parallel.
	var (
		failuresMu sync.Mutex
		failures   int
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
			// This error should not be bubbled up! That will cause the parallel.Run to short circuit,
			// but we want all tests to run regardless.
			err, screenshot := t.runTest(test)
			if _, ok := err.(*internalError); ok {
				t.log.Printf("[warning] [%v] unable to establish a session: %v\n", test.Name, err)
				t.slackMessage(fmt.Sprintf("Test %v failed due to inability to establish a connection: %v", test.Name, err), "")
				return nil
			}

			unitTime := time.Since(unitStart)
			if err != nil {
				failuresMu.Lock()
				failures++
				failuresMu.Unlock()

				t.log.Printf("[failure] [%v] [%v]: %v\n", test.Name, unitTime, err)

				// When running without Slack support, write the screenshot to a file
				// instead.
				if t.slack == nil {
					if e := ioutil.WriteFile(test.Name+".png", screenshot, 0666); e != nil {
						t.log.Printf("[failure] [%v]: could not save screenshot: %v\n", test.Name, e)
					}
				}
				return nil
			}

			t.log.Printf("[success] [%v] [%v]\n", test.Name, unitTime)
			return nil
		})
	}
	run.Wait()

	t.log.Printf("%v tests finished in %v [%v success] [%v failure]\n", total, time.Since(start), total-failures, failures)

	if failures == 0 {
		t.slackSkipAtChannel = false // do @channel on next failure
		if logSuccess {
			t.slackMessage(fmt.Sprintf(":thumbsup: *Success! %v tests successful against %v!*", total, t.target), "")
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
			fmt.Sprintf(":fire: *FAILURE! %v/%v tests failed against %v: *"+atChannel, failures, total, t.target),
			t.slackLogBuffer.String(),
		)

		// emit the alert to monitoring-bot
		err := t.sendAlert()
		if err != nil {
			t.log.Printf("[WARNING] error while sending alert to monitoring-bot %s", err)
		}
	}
	t.slackLogBuffer.Reset()
	return failures == 0
}

func (t *testRunner) sendAlert() error {
	username := os.Getenv("MONITORING_BOT_USERNAME")
	if username == "" {
		return nil
	}

	u, err := url.Parse("https://monitoring-bot.sourcegraph.com/alert?source=e2etest")
	if err != nil {
		return err
	}

	u.User = url.UserPassword(username, os.Getenv("MONITORING_BOT_PASSWORD"))

	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("monitoring bot returned non-200 status code: %d", resp.StatusCode)
	}
	return nil
}

// runTest runs a single test and recovers from panics, should they occur. If
// the test failed for any reason err != nil is returned. If it was possible to
// capture a screenshot of the error, screenshot will be the encoded PNG bytes.
func (t *testRunner) runTest(test *Test) (err error, screenshot []byte) {
	// Create a selenium web driver for the test.
	// Set up webdriver.
	caps := selenium.Capabilities(map[string]interface{}{
		"browserName": "chrome",
		"chromeOptions": map[string]interface{}{
			"args": []string{"user-agent=Sourcegraph e2etest-bot"},
		},
	})
	wd, err := selenium.NewRemote(caps, t.executor)
	if err != nil {
		return &internalError{err: err}, nil
	}

	// Handle things after the test has finished executing.
	defer func() {
		// Handle any panics that might occur because we are a single-process test
		// suite.
		if r := recover(); r != nil {
			if err == nil {
				err = errors.New(fmt.Sprint(r))
			}
		}

		// If there was an error, capture a screenshot of the problem.
		if err != nil {
			var err2 error
			screenshot, err2 = wd.Screenshot()
			if err2 != nil {
				t.log.Println("could not capture screenshot for", test.Name, err2)
			} else if t.slack != nil {
				if err2 = t.slackFileUpload(screenshot, test.Name+".png"); err2 != nil {
					t.log.Println("could not upload screenshot to Slack", test.Name, err2)
				}
			}
		}
		wd.Quit()
	}()

	// Setup the context for the test.
	ctx := &T{
		Log:       t.log,
		Target:    t.target,
		TestLogin: e2etestuser.Prefix + test.Name,
		TestEmail: e2etestuser.Prefix + test.Name + "@sourcegraph.com",
		WebDriver: wd,
		tr:        t,
	}
	ctx.WebDriverT = ctx.WebDriver.T(ctx)

	// Execute the test.
	return test.Func(ctx), nil
}

// slackFileUpload implements slack multipart file upload.
//
// TODO(slimsag): upstream this type of change to github.com/nlopes/slack.
func (t *testRunner) slackFileUpload(f []byte, title string) error {
	b := &bytes.Buffer{}
	w := multipart.NewWriter(b)
	fw, err := w.CreateFormFile("file", title)
	if err != nil {
		return err
	}
	if _, err := io.Copy(fw, bytes.NewReader(f)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	// Write additional fields.
	fields := map[string]string{
		"channels": t.slackChannel.ID,
		"token":    t.slackToken,
	}
	for k, v := range fields {
		fw, err := w.CreateFormField(k)
		if err != nil {
			return err
		}
		if _, err := fw.Write([]byte(v)); err != nil {
			return err
		}
	}

	// Make the request.
	resp, err := http.Post("https://slack.com/api/files.upload", w.FormDataContentType(), b)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Decode the response and check ok field as the Slack API docs say to do.
	var slackResp = struct {
		Ok bool
		// TODO(slimsag): "file" object field
	}{}
	if err := json.Unmarshal(body, &slackResp); err != nil {
		return err
	}
	if !slackResp.Ok {
		return fmt.Errorf("slack file upload failed, response: %s", string(body))
	}
	return nil
}

var tr = &testRunner{
	slackLogBuffer: &bytes.Buffer{},
}

var (
	runOnce = flag.Bool("once", true, "run the tests only once (true) or forever (false)")
	runFlag = flag.String("run", "", "specify an exact test name to run (e.g. 'login_flow', 'register_flow')")
)

func Main() {
	flag.Parse()

	// Prepare logging.
	tr.log = log.New(io.MultiWriter(os.Stderr, tr.slackLogBuffer), "", 0)

	// Determine which Selenium server to connect to.
	serverAddr := os.Getenv("SELENIUM_SERVER_IP")
	serverPort := os.Getenv("SELENIUM_SERVER_PORT")
	if serverAddr == "" {
		log.Fatal("Unable to get SELENIUM_SERVER_IP from environment")
	}
	if serverPort == "" {
		serverPort = "4444" // default to standard Selenium port
	}

	if !strings.Contains(serverAddr, "://") {
		serverAddr = "http://" + serverAddr
	}

	u, err := url.Parse(fmt.Sprintf("%s:%s", serverAddr, serverPort))
	if err != nil {
		log.Fatal(err)
	}

	u.Path = path.Join(u.Path, "wd/hub")

	tr.executor = u.String()

	// Determine the target Sourcegraph instance to test against.
	tr.target = os.Getenv("TARGET")
	if tr.target == "" {
		log.Fatal("Unable to get TARGET Sourcegraph instance from environment")
	}
	tgt, err := url.Parse(tr.target)
	if err != nil {
		log.Fatal(err)
	}
	if tgt.Scheme == "" {
		log.Fatal("TARGET must specify scheme (http or https) prefix")
	}

	// Find server ID key information.
	if key := os.Getenv("ID_KEY_DATA"); key != "" {
		tr.idKey, err = idkey.FromString(key)
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
		tr.idKey, err = idkey.New(data)
		if err != nil {
			log.Fatal(err)
		}
	}

	if token := os.Getenv("SLACK_API_TOKEN"); token != "" {
		tr.slack = slack.New(token)
		tr.slackToken = token

		// Find the channel ID.
		channelName := os.Getenv("SLACK_CHANNEL")
		if channelName == "" {
			channelName = "e2etest"
		}
		channels, err := tr.slack.GetChannels(true)
		if err != nil {
			log.Fatal(err)
		}
		found := false
		for _, c := range channels {
			if c.Name == channelName {
				found = true
				tr.slackChannel = c
			}
		}
		if !found {
			log.Println("could not find slack channel", channelName)
			log.Println("disabling slack notifications")
			tr.slack = nil
		} else {
			registeredTests := &bytes.Buffer{}
			for _, t := range tr.tests {
				fmt.Fprintf(registeredTests, "[%v]: %v\n", t.Name, t.Description)
			}
			tr.slackMessage(":shield: *Ready and reporting for duty!* Registered tests:", registeredTests.String())
		}
	}

	tr.run()
}
