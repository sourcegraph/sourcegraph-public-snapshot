package graph

import "testing"

func Test_TreePath(t *testing.T) {
	treePaths := []string{
		"foo",
		"foo/bar",
		"foo/-/bar",

		"commonjs/lib/async.js",
		"commonjs/lib/async.js/-/all",
		"commonjs/test/test-async.js/-/all alias",
		"commonjs/test/test-async.js/-/queue",
		"commonjs/test/test-async.js/-/queue too many callbacks",
		"file/lib/async.js/-/@231/@local/@231/@local/@1011/@local/arr/<i>",
		"file/lib/async.js/-/@231/@local/@231/@local/@22540/@local/_insert",
		"global/-/console.<i>",
		"global/-/Function.prototype.bind",
		"global/-/process.nextTick",

		"flask",
		"flask/app",
		"flask/app/Flask",
		"flask/app/Flask/add_template_filter",

		".",
		"./foo/bar",
	}
	notTreePaths := []string{"", "/", "///", "foo//bar", "/foo/bar"}

	for _, tp := range treePaths {
		if !IsValidTreePath(tp) {
			t.Errorf("%s should be valid, but was invalid", tp)
		}
	}

	for _, ntp := range notTreePaths {
		if IsValidTreePath(ntp) {
			t.Errorf("%s should be invalid, but was valid", ntp)
		}
	}
}
