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
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestFilters(t *testing.T) {

	c := NewTestConn()
	PopulateTestDB(t, c)
	defer TearDownTestDB(c)

	Convey("Exists filter", t, func() {
		qry := Search("oilers").Filter(
			Filter().Exists("goals"),
		)
		out, err := qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits, ShouldNotBeNil)
		So(out.Hits.Len(), ShouldEqual, 10)
		So(out.Hits.Total, ShouldEqual, 12)
	})

	Convey("Missing filter", t, func() {
		qry := Search("oilers").Filter(
			Filter().Missing("goals"),
		)
		out, err := qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits, ShouldNotBeNil)
		So(out.Hits.Total, ShouldEqual, 2)
	})

	Convey("Terms filter", t, func() {
		qry := Search("oilers").Filter(
			Filter().Terms("pos", "RW", "LW"),
		)
		out, err := qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits, ShouldNotBeNil)
		So(out.Hits.Total, ShouldEqual, 6)
	})

	Convey("Filter involving an AND", t, func() {
		qry := Search("oilers").Filter(
			Filter().Terms("pos", "LW"),
			Filter().Exists("PIM"),
		)
		out, err := qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits, ShouldNotBeNil)
		So(out.Hits.Total, ShouldEqual, 2)
	})

	Convey("Filterng filter results", t, func() {
		qry := Search("oilers").Filter(
			Filter().Terms("pos", "LW"),
		)
		qry.Filter(
			Filter().Exists("PIM"),
		)
		out, err := qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits, ShouldNotBeNil)
		So(out.Hits.Total, ShouldEqual, 2)
	})

	Convey("Filter involving OR", t, func() {
		qry := Search("oilers").Filter(
			"or",
			Filter().Terms("pos", "G"),
			Range().Field("goals").Gt(80),
		)
		out, err := qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits, ShouldNotBeNil)
		So(out.Hits.Total, ShouldEqual, 3)
	})
}
