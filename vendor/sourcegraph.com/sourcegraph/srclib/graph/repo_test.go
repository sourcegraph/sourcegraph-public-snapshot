package graph

import "testing"

// Ensures TryMakeURI works properly and supports different URLs
func TestTryMakeURI(t *testing.T) {

	tests := []struct {
		source   string
		expected string
		error    bool
	}{
		{"http://sourcegraph.com/srclib", "sourcegraph.com/srclib", false},
		{"git://sourcegraph.com/srclib", "sourcegraph.com/srclib", false},
		{"hg://sourcegraph.com/srclib", "sourcegraph.com/srclib", false},
		{"svn://sourcegraph.com/srclib", "sourcegraph.com/srclib", false},
		{"ssh://sourcegraph.com/srclib", "sourcegraph.com/srclib", false},
		{"https://github.com/foo/bar", "github.com/foo/bar", false},
		{"https://bitbucket.org/foo/bar", "bitbucket.org/foo/bar", false},

		{"scm:git:git://github.com/path_to_repository", "github.com/path_to_repository", false},
		{"scm:git:http://github.com/path_to_repository", "github.com/path_to_repository", false},
		{"scm:git:https://github.com/path_to_repository", "github.com/path_to_repository", false},
		{"scm:git:ssh://github.com/path_to_repository", "github.com/path_to_repository", false},
		{"scm:git:file://localhost/path_to_repository", "localhost/path_to_repository", false},

		{"scm:hg:http://host/v3", "host/v3", false},
		{"scm:hg:file://C:/dev/project/v3", "c:/dev/project/v3", false},
		{"scm:hg:file:///home/smorgrav/dev/project/v3", "", true},
		{"scm:hg:/home/smorgrav/dev/project/v3", "hg/home/smorgrav/dev/project/v3", false},

		{"scm:svn:file:///svn/root/module", "", true},
		{"scm:svn:file://localhost/path_to_repository", "localhost/path_to_repository", false},
		{"scm:svn:file://my_server/path_to_repository", "my_server/path_to_repository", false},
		{"scm:svn:http://svn.apache.org/svn/root/module", "svn.apache.org/svn/root/module", false},
		{"scm:svn:https://username@svn.apache.org/svn/root/module", "svn.apache.org/svn/root/module", false},
		{"scm:svn:https://username:password@svn.apache.org/svn/root/module", "svn.apache.org/svn/root/module", false},

		{"scm:perforce://depot/modules/myproject", "depot/modules/myproject", false},
		{"scm:perforce:host:path_to_repository", "host/path_to_repository", false},

		{"user@host:path", "host/path", false},
		{"host:path", "host/path", false},
	}

	for _, test := range tests {
		actual, err := TryMakeURI(test.source)
		if err != nil && !test.error {
			t.Errorf("Got unexpected error %s on %s", err, test.source)
		} else if err == nil && test.error {
			t.Errorf("Expected to have an error on %s", test.source)
		} else if actual != test.expected {
			t.Errorf("Got `%s` while expected `%s`", actual, test.expected)
		}
	}
}
