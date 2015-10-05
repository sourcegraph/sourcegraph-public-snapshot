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
	"github.com/araddon/gou"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestFacetRegex(t *testing.T) {

	c := NewTestConn()
	PopulateTestDB(t, c)
	defer TearDownTestDB(c)

	Convey("Facted regex query", t, func() {

		// This is a possible solution for auto-complete
		out, err := Search("oilers").Size("0").Facet(
			Facet().Regex("name", "[jk].*").Size("8"),
		).Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)

		// Debug(string(out.Facets))
		fh := gou.NewJsonHelper([]byte(out.Facets))
		facets := fh.Helpers("/name/terms")
		So(err, ShouldBeNil)
		So(facets, ShouldNotBeNil)
		So(len(facets), ShouldEqual, 4)
	})
}
