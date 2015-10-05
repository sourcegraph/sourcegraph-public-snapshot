// Copyright 2013 Matthew Baird
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package elastigo

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type SuggestTest struct {
	Completion string `json:"completion"`
}

type hash map[string]interface{}

func TestCoreSearch(t *testing.T) {

	c := NewTestConn()
	c.CreateIndex("github")
	waitFor(func() bool { return false }, 5)

	defer func() {
		c.DeleteIndex("github")
	}()

	Convey("Convert a search result to JSON", t, func() {

		qry := map[string]interface{}{
			"query": map[string]interface{}{
				"wildcard": map[string]string{"actor": "a*"},
			},
		}
		var args map[string]interface{}
		out, err := c.Search("github", "", args, qry)
		So(err, ShouldBeNil)

		_, err = json.Marshal(out.Hits.Hits)
		So(err, ShouldBeNil)
	})

	Convey("Update a document and verify that it is reflected", t, func() {
		mappingOpts := MappingOptions{Properties: hash{
			"completion": hash{
				"type": "completion",
			},
		}}
		err := c.PutMapping("github", "SuggestTest", SuggestTest{}, mappingOpts)
		So(err, ShouldBeNil)

		_, err = c.UpdateWithPartialDoc("github", "SuggestTest", "1", nil, SuggestTest{"foobar"}, true)
		So(err, ShouldBeNil)

		query := hash{"completion_completion": hash{
			"text": "foo",
			"completion": hash{
				"size":  10,
				"field": "completion",
			},
		}}

		_, err = c.Refresh("github")
		So(err, ShouldBeNil)

		res, err := c.Suggest("github", nil, query)
		So(err, ShouldBeNil)

		opts, err := res.Result("completion_completion")
		So(err, ShouldBeNil)

		So(len(opts[0].Options), ShouldBeGreaterThan, 0)
		So(opts[0].Options[0].Text, ShouldEqual, "foobar")
	})
}
