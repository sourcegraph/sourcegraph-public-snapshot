# User Agents
The agent package contains a number of methods for creating user agent strings for popular browsers
and crawlers, and for generating your own user agents.

```go
bow := surf.NewBrowser()

// Use the Google Chrome user agent. The Chrome() method returns:
// "Mozilla/5.0 (Windows NT 6.3; x64) Chrome/37.0.2049.0 Safari/537.36".
bow.SetUserAgent(agent.Chrome())

// The Firefox() method returns:
// "Mozilla/5.0 (Windows NT 6.3; x64; rv:31.0) Gecko/20100101 Firefox/31.0".
bow.SetUserAgent(agent.Firefox())

// The Safari() method returns:
// "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0 Safari/8536.25".
bow.SetUserAgent(agent.Safari())

// There are methods for a number of bows and crawlers. For example
// Opera(), MSIE(), AOL(), GoogleBot(), and many more. You can even choose
// the bow version. This will create:
// "Mozilla/5.0 (Windows NT 6.3; x64) Chrome/35 Safari/537.36".
ua := agent.CreateVersion("chrome", "35")
bow.SetUserAgent(ua)

// Creating your own custom user agent is just as easy. The following code
// generates the user agent:
// "Mybow/1.0 (Windows NT 6.1; WOW64; x64)".
agent.Name = "Mybow"
agent.Version = "1.0"
agent.OSName = "Windows NT"
agent.OSVersion = "6.1"
agent.Comments = []string{"WOW64", "x64"}
bow.SetUserAgent(agent.Create())
```

The agent package has an internal database for many different versions of many different browsers.
See the [agent package API documentation](api/packages/agent) for more information.
