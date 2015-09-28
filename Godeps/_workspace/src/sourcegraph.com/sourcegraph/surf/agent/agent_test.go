package agent

import (
	"fmt"
	"github.com/headzoo/ut"
	"runtime"
	"testing"
)

func TestCreate(t *testing.T) {
	ut.Run(t)

	ut.AssertEquals(fmt.Sprintf("Surf/%s (%s %s; %s)", Version, osName(), osVersion(), runtime.Version()), Create())

	Name = "Chrome"
	Version = "37.0.2049.0"
	OSName = "Ubuntu"
	OSVersion = "14.04"
	Comments = []string{"X11", "like Gecko"}
	ut.AssertEquals("Mozilla/5.0 (Ubuntu 14.04; X11; like Gecko) Chrome/37.0.2049.0 Safari/537.36", Create())

	Name = "Chrome"
	Version = ""
	OSName = "Ubuntu"
	OSVersion = "14.04"
	Comments = []string{}
	ut.AssertEquals("Mozilla/5.0 (Ubuntu 14.04) Chrome/37.0.2049.0 Safari/537.36", Create())

	Name = "Firefox"
	Version = "31.0"
	Comments = []string{"x64"}
	ut.AssertEquals("Mozilla/5.0 (Ubuntu 14.04; x64; rv:31.0) Gecko/20100101 Firefox/31.0", Create())

	Name = "MSIE"
	Version = "10.6"
	OSName = "Windows NT"
	OSVersion = "6.1"
	Comments = []string{"WOW64"}
	ut.AssertEquals("Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; WOW64; Trident/5.0; .NET CLR 3.5.30729)", Create())

	Name = "Opera"
	Version = "12.14"
	OSName = "Windows NT"
	OSVersion = "6.1"
	Comments = []string{"en"}
	ut.AssertEquals("Opera/9.80 (Windows NT 6.1; U; en) Presto/2.9.181 Version/12.14", Create())

	Name = "Safari"
	Version = "6.0"
	OSName = "Intel Mac OS X"
	OSVersion = "10_6_8"
	Comments = []string{}
	ut.AssertEquals("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0 Safari/8536.25", Create())

	Name = "AOL"
	Version = "9.7"
	OSName = "Windows NT"
	OSVersion = "6.1"
	Comments = []string{}
	ut.AssertEquals("Mozilla/5.0 (compatible; MSIE 9.0; AOL 9.7; AOLBuild 4343.19; Windows NT 6.1; WOW64; Trident/5.0; FunWebProducts)", Create())

	Name = "Konqueror"
	Version = "4.9"
	OSName = "Linux"
	OSVersion = "3.5"
	Comments = []string{}
	ut.AssertEquals("Mozilla/5.0 (compatible; Konqueror/4.0; Linux) KHTML/4.0.3 (like Gecko)", Create())

	Name = "Netscape"
	Version = "9.0.0.6"
	OSName = "Windows NT"
	OSVersion = "6.1"
	Comments = []string{"en-US"}
	ut.AssertEquals("Mozilla/5.0 (Windows NT; U; Windows NT 6.1; rv:1.9.2.4; en-US) Gecko/20070321 Netscape/9.0.0.6", Create())

	Name = "Lynx"
	Version = "2.8.8dev.3"
	OSName = "Linux"
	OSVersion = "3.5"
	Comments = []string{}
	ut.AssertEquals("Lynx/2.8.8dev.3 libwww-FM/2.14 SSL-MM/1.4.1", Create())
}

func TestChrome(t *testing.T) {
	ut.Run(t)
	ut.AssertEquals("Mozilla/5.0 (Windows NT 6.3; x64) Chrome/37.0.2049.0 Safari/537.36", Chrome())
}

func TestFirefox(t *testing.T) {
	ut.Run(t)
	ut.AssertEquals("Mozilla/5.0 (Windows NT 6.3; x64; rv:31.0) Gecko/20100101 Firefox/31.0", Firefox())
}

func TestMSIE(t *testing.T) {
	ut.Run(t)
	ut.AssertEquals("Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.3; x64; Trident/5.0; .NET CLR 3.5.30729)", MSIE())
}

func TestOpera(t *testing.T) {
	ut.Run(t)
	ut.AssertEquals("Opera/9.80 (Windows NT 6.3; U; x64) Presto/2.9.181 Version/12.14", Opera())
}

func TestSafari(t *testing.T) {
	ut.Run(t)
	ut.AssertEquals("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0 Safari/8536.25", Safari())
}

func TestAOL(t *testing.T) {
	ut.Run(t)
	ut.AssertEquals("Mozilla/5.0 (compatible; MSIE 9.0; AOL 9.7; AOLBuild 4343.19; Windows NT 6.3; WOW64; Trident/5.0; FunWebProducts; x64)", AOL())
}
