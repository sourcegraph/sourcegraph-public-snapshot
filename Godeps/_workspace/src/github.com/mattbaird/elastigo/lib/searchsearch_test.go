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

func TestSearch(t *testing.T) {

	c := NewTestConn()
	PopulateTestDB(t, c)
	defer TearDownTestDB(c)

	Convey("Wildcard request query", t, func() {

		qry := map[string]interface{}{
			"query": map[string]interface{}{
				"wildcard": map[string]string{"name": "*hu*"},
			},
		}
		out, err := c.Search("oilers", "", nil, qry)

		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits, ShouldNotBeNil)
		So(out.Hits.Total, ShouldEqual, 3)
	})

	Convey("Simple search", t, func() {

		// searching without faceting
		qry := Search("oilers").Pretty().Query(
			Query().Search("dave"),
		)

		// how many different docs used the word "dave"
		out, err := qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits, ShouldNotBeNil)
		So(out.Hits.Total, ShouldEqual, 2)

		out, _ = Search("oilers").Search("dave").Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits, ShouldNotBeNil)
		So(out.Hits.Total, ShouldEqual, 2)
	})

	Convey("URL Request query string", t, func() {

		out, err := c.SearchUri("oilers", "", map[string]interface{}{"q": "pos:LW"})

		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits, ShouldNotBeNil)
		So(out.Hits.Total, ShouldEqual, 3)
	})


	//	A faceted search for what "type" of events there are
	//	- since we are not specifying an elasticsearch type it searches all ()
	//
	//	{
	//	    "terms" : {
	//	      "_type" : "terms",
	//	      "missing" : 0,
	//	      "total" : 7561,
	//	      "other" : 0,
	//	      "terms" : [ {
	//	        "term" : "pushevent",
	//	        "count" : 4185
	//	      }, {
	//	        "term" : "createevent",
	//	        "count" : 786
	//	      }.....]
	//	    }
	//	 }

	Convey("Facet search simple", t, func() {

		qry := Search("oilers").Pretty().Facet(
			Facet().Fields("teams").Size("4"),
		).Query(
			Query().All(),
		).Size("1")
		out, err := qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)

		h := gou.NewJsonHelper(out.Facets)
		So(h.Int("teams.total"), ShouldEqual, 37)
		So(h.Int("teams.missing"), ShouldEqual, 0)
		So(len(h.List("teams.terms")), ShouldEqual, 4)

		// change the size
		qry.FacetVal.Size("20")
		out, err = qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)

		h = gou.NewJsonHelper(out.Facets)
		So(h.Int("teams.total"), ShouldEqual, 37)
		So(len(h.List("teams.terms")), ShouldEqual, 11)

	})

	Convey("Facet search with type", t, func() {

		out, err := Search("oilers").Type("heyday").Pretty().Facet(
			Facet().Fields("teams").Size("4"),
		).Query(
			Query().All(),
		).Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)

		h := gou.NewJsonHelper(out.Facets)
		So(h.Int("teams.total"), ShouldEqual, 37)
		So(len(h.List("teams.terms")), ShouldEqual, 4)
	})


	Convey("Facet search with wildcard", t, func() {

		qry := Search("oilers").Pretty().Facet(
			Facet().Fields("teams").Size("20"),
		).Query(
			Query().Search("*w*"),
		)
		out, err := qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)

		h := gou.NewJsonHelper(out.Facets)
		So(h.Int("teams.total"), ShouldEqual, 20)
		So(len(h.List("teams.terms")), ShouldEqual, 7)
	})

	Convey("Facet search with range", t, func() {

		qry := Search("oilers").Pretty().Facet(
			Facet().Fields("teams").Size("20"),
		).Query(
			Query().Range(
				Range().Field("dob").From("19600101").To("19621231"),
			).Search("*w*"),
		)
		out, err := qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)

		h := gou.NewJsonHelper(out.Facets)
		So(h.Int("teams.total"), ShouldEqual, 12)
		So(len(h.List("teams.terms")), ShouldEqual, 5)
	})

	Convey("Search query with terms", t, func() {

		qry := Search("oilers").Query(
			Query().Term("teams", "NYR"),
		)
		out, err := qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits.Len(), ShouldEqual, 4)
		So(out.Hits.Total, ShouldEqual, 4)
	})

	Convey("Search query with fields", t, func() {

		qry := Search("oilers").Query(
			Query().Fields("teams", "NYR", "", ""),
		)
		out, err := qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits.Len(), ShouldEqual, 4)
		So(out.Hits.Total, ShouldEqual, 4)
	})

	Convey("Search query with fields exist and missing", t, func() {

		qry := Search("oilers").Filter(
			Filter().Exists("PIM"),
		)
		out, err := qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits.Len(), ShouldEqual, 2)
		So(out.Hits.Total, ShouldEqual, 2)

		qry = Search("oilers").Filter(
			Filter().Missing("PIM"),
		)
		out, err = qry.Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits.Len(), ShouldEqual, 10)
		So(out.Hits.Total, ShouldEqual, 12)
	})

	Convey("Search with query and filter", t, func() {

		out, err := Search("oilers").Size("25").Query(
			Query().Fields("name", "*d*", "", ""),
		).Filter(
			Filter().Terms("teams", "STL"),
		).Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits.Len(), ShouldEqual, 2)
		So(out.Hits.Total, ShouldEqual, 2)
	})

	Convey("Search with range", t, func() {

		out, err := Search("oilers").Size("25").Query(
			Query().Range(
				Range().Field("dob").From("19600101").To("19621231"),
			).Search("*w*"),
		).Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits.Len(), ShouldEqual, 4)
		So(out.Hits.Total, ShouldEqual, 4)
	})

	Convey("Search with sorting desc", t, func() {

		out, err := Search("oilers").Pretty().Query(
			Query().All(),
		).Sort(
			Sort("dob").Desc(),
		).Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits.Len(), ShouldEqual, 10)
		So(out.Hits.Total, ShouldEqual, 14)

		b, err := out.Hits.Hits[0].Source.MarshalJSON()
		h1 := gou.NewJsonHelper(b)
		So(h1.String("name"), ShouldEqual, "Grant Fuhr")
	})

	Convey("Search with sorting asc", t, func() {

		out, err := Search("oilers").Pretty().Query(
			Query().All(),
		).Sort(
			Sort("dob"),
		).Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits.Len(), ShouldEqual, 10)
		So(out.Hits.Total, ShouldEqual, 14)

		b, err := out.Hits.Hits[0].Source.MarshalJSON()
		h1 := gou.NewJsonHelper(b)
		So(h1.String("name"), ShouldEqual, "Pat Hughes")
	})

	Convey("Search with sorting desc with query", t, func() {

		out, err := Search("oilers").Pretty().Query(
			Query().Search("*w*"),
		).Sort(
			Sort("dob").Desc(),
		).Result(c)
		So(err, ShouldBeNil)
		So(out, ShouldNotBeNil)
		So(out.Hits.Len(), ShouldEqual, 8)
		So(out.Hits.Total, ShouldEqual, 8)

		b, err := out.Hits.Hits[0].Source.MarshalJSON()
		h1 := gou.NewJsonHelper(b)
		So(h1.String("name"), ShouldEqual, "Wayne Gretzky")
	})
}
