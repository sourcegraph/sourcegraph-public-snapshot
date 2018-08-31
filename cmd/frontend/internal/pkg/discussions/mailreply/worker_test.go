package mailreply

import "testing"

func TestTrimGmailReplyQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "trim",
			input: "\n\nWow, cool!?\r\n\r\nOn Wed, Aug 29, 2018 at 3:22 AM, Stephen Gutekanst <stephen@sourcegraph.com>\r\nwrote:\r\n\r\n> ok\r\n>\r\n> On Wed, Aug 29, 2018 at 3:20 AM, Stephen Gutekanst <\r\n> stephen@sourcegraph.com> wrote:\r\n>\r\n>> nice\r\n>>\r\n>> On Wed, Aug 29, 2018 at 3:07 AM, Stephen Gutekanst <\r\n>> stephen@sourcegraph.com> wrote:\r\n>>\r\n>>> Hmm, ok\r\n>>>\r\n>>> On Wed, Aug 29, 2018 at 2:43 AM, Stephen Gutekanst <\r\n>>> stephen@sourcegraph.com> wrote:\r\n>>>\r\n>>>> Ok, sure\r\n>>>>\r\n>>>> On Wed, Aug 29, 2018 at 2:43 AM, Stephen Gutekanst <\r\n>>>> noreply+stephen@sourcegraph.com> wrote:\r\n>>>>\r\n>>>>> *@slimsag* commented on *mux.go*:\r\n>>>>>\r\n>>>>> Hello 2!\r\n>>>>> 585\r\n>>>>> 586 // methodNotAllowedHandler returns a simple request handler\r\n>>>>> 587 // that replies to each request with a status code 405.\r\n>>>>> 588 func methodNotAllowedHandler() http.Handler { return http.\r\n>>>>> HandlerFunc(methodNotAllowed) }\r\n>>>>>\r\n>>>>> View and reply on Sourcegraph\r\n>>>>> <http://localhost:3080/github.com/gorilla/mux/-/blob/mux.go?utm_source=email#tab=discussions&threadID=38&commentID=197>\r\n>>>>> 197\r\n>>>>>\r\n>>>>>\r\n>>>>>\r\n>>>>>\r\n>>>>>\r\n>>>>> This email was sent to stephen@sourcegraph.com    unsubscribe from\r\n>>>>> this list\r\n>>>>> <http://mandrillapp.com/track/unsub.php?u=30284645&id=d24a1411e7a74377aedd10086eaf2a1e.k%2BCDVltHXctYkEaPVkt5Oo76cI4%3D&r=https%3A%2F%2Fmandrillapp.com%2Funsub%3Fmd_email%3Dstephen%2540sourcegraph.com>\r\n>>>>>\r\n>>>>\r\n>>>>\r\n>>>>\r\n>>>> --\r\n>>>>\r\n>>>> Stephen Gutekanst\r\n>>>>\r\n>>>> ✱slimsag <https://sourcegraph.com/slimsag> | @slimsag\r\n>>>> <https://twitter.com/slimsag>\r\n>>>>\r\n>>>> ✱Sourcegraph <https://sourcegraph.com/> | Connecting people and code |\r\n>>>> @srcgraph <https://twitter.com/srcgraph>\r\n>>>>\r\n>>>\r\n>>>\r\n>>>\r\n>>> --\r\n>>>\r\n>>> Stephen Gutekanst\r\n>>>\r\n>>> ✱slimsag <https://sourcegraph.com/slimsag> | @slimsag\r\n>>> <https://twitter.com/slimsag>\r\n>>>\r\n>>> ✱Sourcegraph <https://sourcegraph.com/> | Connecting people and code |\r\n>>> @srcgraph <https://twitter.com/srcgraph>\r\n>>>\r\n>>\r\n>>\r\n>>\r\n>> --\r\n>>\r\n>> Stephen Gutekanst\r\n>>\r\n>> ✱slimsag <https://sourcegraph.com/slimsag> | @slimsag\r\n>> <https://twitter.com/slimsag>\r\n>>\r\n>> ✱Sourcegraph <https://sourcegraph.com/> | Connecting people and code |\r\n>> @srcgraph <https://twitter.com/srcgraph>\r\n>>\r\n>\r\n>\r\n>\r\n> --\r\n>\r\n> Stephen Gutekanst\r\n>\r\n> ✱slimsag <https://sourcegraph.com/slimsag> | @slimsag\r\n> <https://twitter.com/slimsag>\r\n>\r\n> ✱Sourcegraph <https://sourcegraph.com/> | Connecting people and code |\r\n> @srcgraph <https://twitter.com/srcgraph>\r\n>\r\n\r\n\r\n\r\n-- \r\n\r\nStephen Gutekanst\r\n\r\n✱slimsag <https://sourcegraph.com/slimsag> | @slimsag\r\n<https://twitter.com/slimsag>\r\n\r\n✱Sourcegraph <https://sourcegraph.com/> | Connecting people and code |\r\n@srcgraph <https://twitter.com/srcgraph>\r\n",
			want:  "\n\nWow, cool!?\r\n",
		},
		{
			name:  "noop",
			input: "Wow, cool!?",
			want:  "Wow, cool!?",
		},
		{
			name:  "trim2",
			input: "Another reply!\r\n\r\nOn Thu, Aug 30, 2018 at 3:42 PM, Stephen Gutekanst <stephen@sourcegraph.com>\r\nwrote:\r\n\r\n> test reply\r\n>\r\n> On Thu, Aug 30, 2018 at 1:45 PM, Stephen Gutekanst <\r\n> notifications@sourcegraph.com> wrote:\r\n>\r\n>> *@slimsag* commented on *mux.go*:\r\n>>\r\n>> Ah, he is. Cool!\r\n>> 169\r\n>> 170 // GetRoute returns a route registered with the given name. This\r\n>> method\r\n>> 171 // was renamed to Get() and remains here for backwards compatibility.\r\n>> 172 func (r *Router) GetRoute(name string) *Route {\r\n>> 173 return r.getNamedRoutes()[name]\r\n>> 174 }\r\n>> 175\r\n>>\r\n>> View and reply on Sourcegraph\r\n>> <http://localhost:3080/github.com/gorilla/mux/-/blob/mux.go?utm_source=email#tab=discussions&threadID=44&commentID=274>\r\n>> 274\r\n>>\r\n>\r\n>\r\n>\r\n> --\r\n>\r\n> Stephen Gutekanst\r\n>\r\n> ✱slimsag <https://sourcegraph.com/slimsag> | @slimsag\r\n> <https://twitter.com/slimsag>\r\n>\r\n> ✱Sourcegraph <https://sourcegraph.com/> | Connecting people and code |\r\n> @srcgraph <https://twitter.com/srcgraph>\r\n>\r\n\r\n\r\n\r\n-- \r\n\r\nStephen Gutekanst\r\n\r\n✱slimsag <https://sourcegraph.com/slimsag> | @slimsag\r\n<https://twitter.com/slimsag>\r\n\r\n✱Sourcegraph <https://sourcegraph.com/> | Connecting people and code |\r\n@srcgraph <https://twitter.com/srcgraph>\r\n",
			want:  "Another reply!\r\n",
		},
		{
			name:  "multi-line",
			input: "test\r\n\r\nreply?\r\n\r\nOn Thu, Aug 30, 2018 at 3:49 PM, Stephen Gutekanst <\r\nstephen.gutekanst@gmail.com> wrote:\r\n\r\n> test reply?\r\n>\r\n> On Thu, Aug 30, 2018 at 3:46 PM, stephen <notifications@sourcegraph.com>\r\n> wrote:\r\n>\r\n>> *@stephen* commented on *mux.go*:\r\n>>\r\n>> Another reply!\r\n>> 169\r\n>> 170 // GetRoute returns a route registered with the given name. This\r\n>> method\r\n>> 171 // was renamed to Get() and remains here for backwards compatibility.\r\n>> 172 func (r *Router) GetRoute(name string) *Route {\r\n>> 173 return r.getNamedRoutes()[name]\r\n>> 174 }\r\n>> 175\r\n>>\r\n>> View and reply on Sourcegraph\r\n>> <http://localhost:3080/github.com/gorilla/mux/-/blob/mux.go?utm_source=email#tab=discussions&threadID=44&commentID=276>\r\n>> 276\r\n>>\r\n>\r\n>\r\n>\r\n> --\r\n> Follow me on twitter @slimsag <https://twitter.com/slimsag>.\r\n>\r\n\r\n\r\n\r\n-- \r\nFollow me on twitter @slimsag <https://twitter.com/slimsag>.\r\n",
			want:  "test\r\n\r\nreply?\r\n",
		},
		{
			name:  "best-effort",
			input: "test\r\n\r\nreply?\r\n\r\nOn Thu, Aug 30, 2018 at 3:49 PM, JUNK wrote:\r\n> JUNK\r\nA whole lot of\r\nA whole lot of\r\nA whole lot of\r\nA whole lot of\r\ngarbage\r\n",
			want:  "test\r\n\r\nreply?\r\n",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := string(trimGmailReplyQuote([]byte(tst.input)))
			if got != tst.want {
				t.Logf("got  %q\n", got)
				t.Logf("want %q\n", tst.want)
				t.Fail()
			}
		})
	}
}
