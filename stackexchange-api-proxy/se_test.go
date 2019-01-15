package se

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func Test_IsAllowedURL(t *testing.T) {

	t.Parallel()

	var cases = []struct {
		urlStr  string
		vals    *url.Values
		allowed bool
	}{
		{
			urlStr:  "http://example.com",
			vals:    nil,
			allowed: false,
		},
		{
			urlStr:  "http://stackoverflow.com",
			vals:    &url.Values{"site": []string{"stackoverflow"}},
			allowed: true,
		},
		{
			urlStr:  "http://www.stackoverflow.com",
			vals:    &url.Values{"site": []string{"stackoverflow"}},
			allowed: true,
		},
		{
			urlStr:  "http://maliciousstackoverflow.com",
			vals:    nil,
			allowed: false,
		},
		{
			urlStr:  "http://www.stackexchange.com",
			vals:    nil,
			allowed: false,
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("simple client %s", tc.urlStr), func(t *testing.T) {
			vals, allowed := IsAllowedURL(tc.urlStr)
			if allowed != tc.allowed {
				t.Fatalf("Expected %t to equal %t when checking if %q is allowed.", allowed, tc.allowed, tc.urlStr)
			}
			if cmp.Diff(vals, tc.vals) != "" {
				t.Fatalf("Expected %#v to equal %#v when checking if %q is allowed.", vals, tc.vals, tc.urlStr)
			}
		})
		t.Run(fmt.Sprintf("custom client %s", tc.urlStr), func(t *testing.T) {
			vals, allowed := Client{allowList: DefaultAllowListPatterns}.IsAllowedURL(tc.urlStr)
			if allowed != tc.allowed {
				t.Fatalf("Expected %t to equal %t when checking if %q is allowed.", allowed, tc.allowed, tc.urlStr)
			}
			if cmp.Diff(vals, tc.vals) != "" {
				t.Fatalf("Expected %#v to equal %#v when checking if %q is allowed.", vals, tc.vals, tc.urlStr)
			}
		})
	}
}

func Test_FetchUpdate(t *testing.T) {

	t.Run("allow-list", func(t *testing.T) {
		var c, _ = NewClient(SpecifyAllowList(AllowList{})) // empty allow-list
		err := c.FetchUpdate(context.Background(), "http://www.example.com/")
		if err != ErrURLNotAllowed {
			t.Fatalf("Expected all urls to be disallowed and return ErrURLNotAllowed, got %q", err)
		}
	})

	t.Run("unparsable URL", func(t *testing.T) {
		t.Skip("can't find anything that Go won't parse!")
		var c, _ = NewClient(SpecifyAllowList(AllowList{"helloworld": regexp.MustCompile(".*")}))
		err := c.FetchUpdate(context.Background(), "hello::::world") // non-integer ports are disallowed by http spec
		if err == nil {
			t.Fatalf("Expected string to return an error propagated from url.Parse, got nil")
		}
		if seErr, ok := err.(Error); !ok {
			t.Fatalf("Expected string to return an error propagated from url.Parse, got nil")
		} else {
			if seErr.Op != "parse-url" {
				t.Fatalf("Expected string to return an error propagated from url.Parse with .Op set to 'parse-url'")
			}
		}
	})

	t.Run("locking", func(t *testing.T) {
		t.Run("errs with a lock contention error on time", func(t *testing.T) {

			// Pre-lock a mutex, guarantees no race conditions
			// in trying to get and hold a lock whilst firing a second
			// request to the same URL resource.
			var alreadyLockedMutex = &sync.Mutex{}
			alreadyLockedMutex.Lock()
			defer alreadyLockedMutex.Unlock()

			var artificallyShortLockWaitTimeout = 50 * time.Millisecond
			var c, _ = NewClient(
				SpecifyLockWaitTimeout(artificallyShortLockWaitTimeout),
				SpecifyLockMechanism(func(u url.URL) sync.Locker {
					return alreadyLockedMutex
				}),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			// start one we care about, and time it..
			start := time.Now()
			err := c.FetchUpdate(ctx, "http://stackoverflow.com/...")
			elapsed := time.Since(start)

			if err != ErrTimeoutLocking {
				t.Fatalf("Expected err to be %s got %s", ErrTimeoutLocking, err)
			}

			// Grace period of double the wait timeout. Observed was 5-7ms
			// overhead through, presumably the locking and scheduling of
			// goroutines and the use of select{}
			if elapsed > artificallyShortLockWaitTimeout*2 {
				t.Fatalf("Lock wait time was configured at %s but elapsed time was %s, must be a mistake in the locking code", artificallyShortLockWaitTimeout, elapsed)
			}

		})
	})

	// this test uses a spy to check we mutate the URL correctly
	t.Run("injecting URL query params", func(t *testing.T) {

		t.Skip("URL mutation doesn't yet turn stackoverflow.com/... into api.stackexchange.com/....&site=stackoverflow")

		var expectedURL = "https://api.stackexchange.com/2.2/questions/18390852?site=stackoverflow&filter=" + FilterID
		var requestedURL string
		var httpDoFnStub = func(req *http.Request) (*http.Response, error) {
			requestedURL = req.URL.String()
			return &http.Response{
				Status:     http.StatusText(201),
				StatusCode: 201,
				Body:       ioutil.NopCloser(bytes.NewBuffer([]byte("no content"))),
			}, nil
		}
		var c, _ = NewClient(SpecifyHTTPDoFn(httpDoFnStub))
		c.FetchUpdate(context.Background(), "https://stackoverflow.com/questions/18390852/go-concurrency-and-channel-confusion")

		if requestedURL != expectedURL {
			t.Fatalf("Expected requested URL %q to match %q", requestedURL, expectedURL)
		}
	})

	t.Run("parsing a response", func(t *testing.T) {

		t.Skip("no asserertions in place yet, will pick this up again soon")

		var b bytes.Buffer
		var respJSON = `{"items":[{"answers":[{"last_activity_date":1377212032,"answer_id":18392073,"body_markdown":"The exact output of your program is not defined and depends on the scheduler. The scheduler can choose freely between all goroutines that are currently not blocked. It tries to run those goroutines concurrently by switching the current goroutine in very short time intervals so that the user gets the feeling that everything happens simultanously. In addition to that, it can also execute more than one goroutine in parallel on different CPUs (if you happen to have a multicore system and increase [runtime.GOMAXPROCS][1]). One situation that might lead to your output is:\r\n\r\n1. ` + "`main`" + `creates two goroutines\r\n2. the scheduler chooses to switch to one of the new goroutines immediately and chooses ` + "`display`" + `\r\n3.` + "`display`" + `prints out the message and is blocked by the channel send (` + "`c &lt;- true`" + `) since there isn&#39;t a receiver yet.\r\n4. the scheduler chooses to run ` + "`sum`" + `next\r\n5. the sum is computed and printed on the screen\r\n6. the scheduler chooses to not resume the ` + "`sum`" + ` goroutine (it has already used a fair amount of time) and continues with ` + "`display`" + `\r\n7. ` + "`display`" + `sends the value to the channel\r\n8. the scheduler chooses to run main next\r\n9. main quits and all goroutines are destroyed\r\n\r\nBut that is just one possible execution order. There are many others and some of them will lead to a different output. If you want to print just the first result and quit the program afterwards, you should probably use a ` + "`result chan string`" + `and change your ` + "`main`" + ` function to print ` + "`fmt.Println(&lt;-result)`" + `.\r\n\r\n\r\n  [1]: http://golang.org/pkg/runtime/#GOMAXPROCS"}],"last_activity_date":1377229506,"question_id":18390852,"body_markdown":"I&#39;m new to Go and have a problem understanding the concurrency and channel.\r\n\r\n    package main\r\n\r\n    import &quot;fmt&quot;\r\n\r\n    func display(msg string, c chan bool){\r\n        fmt.Println(&quot;display first message:&quot;, msg)\r\n        c &lt;- true\r\n    }\r\n\r\n    func sum(c chan bool){\r\n        sum := 0\r\n        for i:=0; i &lt; 10000000000; i++ {\r\n            sum++\r\n        }\r\n        fmt.Println(sum)\r\n        c &lt;- true\r\n    }\r\n\r\n    func main(){\r\n        c := make(chan bool)\r\n\r\n        go display(&quot;hello&quot;, c)\r\n        go sum(c)\r\n        &lt;-c\r\n    }\r\n\r\nThe output of the program is:\r\n\r\n    display first message: hello\r\n    10000000000 \r\n\r\nBut I thought it should be only one line:\r\n\r\n    display first message: hello\r\n\r\nSo in the main function, &lt;-c is blocking it and waits for the other two go rountines to send data to the channel. Once the main function receives the data from c, it should proceed and exit.\r\n\r\ndisplay and sum run simultaneously and sum takes longer so display should send true to c and the program should exit before sum finishes...\r\n\r\nI&#39;m not sure I understand it clearly. Could someone help me with this? Thank you!  "}],"quota_max":300,"quota_remaining":210}`
		gzip.NewWriter(&b).Write([]byte(respJSON))

		type fakeHTTP struct {
			responseCode     int
			responseRawBytes bytes.Buffer
			Do               func(req *http.Request) (*http.Response, error)
		}

		var httpDoFnStub = func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				Status:     http.StatusText(200),
				StatusCode: 200,
				Body:       ioutil.NopCloser(&b),
			}, nil
		}

		var c, _ = NewClient(SpecifyHTTPDoFn(httpDoFnStub))
		c.FetchUpdate(context.Background(), "http://www.stackoverflow.com/")

	})

}

// TODO: This test could be much more comprehensive, this is just a
// big sanity check.
func Test_NewClient(t *testing.T) {
	var c1, _ = NewClient()
	var c2, _ = NewClient(SpecifyLockWaitTimeout(c1.lockWaitTimeout * 2))
	if c1.lockWaitTimeout == c2.lockWaitTimeout {
		t.Fatalf("Expected %s to equal %s when checking whether client optionsFns were applied correctly.", c1.lockWaitTimeout, c2.lockWaitTimeout)
	}
}
