package usercreds

import (
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"testing"
)

func TestFromNetRC(t *testing.T) {
	netrc := []byte(`machine src.foobar.com
        login foo
        password bar
`)
	f, err := ioutil.TempFile("", "netrc")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	_, err = f.Write(netrc)
	if err != nil {
		t.Fatal(err)
	}
	netrcPath = f.Name()

	u, _ := url.Parse("http://src.foobar.com:3080")
	c := FromNetRC(u)
	want := &LoginCredentials{Login: "foo", Password: "bar"}
	if !reflect.DeepEqual(c, want) {
		t.Errorf("Endpoint in netrc did not find creds: %#v", c)
	}

	u, _ = url.Parse("http://src.baz.com")
	c = FromNetRC(u)
	if c != nil {
		t.Errorf("Got credentials for endpoint not in netrc: %#v", c)
	}
}
