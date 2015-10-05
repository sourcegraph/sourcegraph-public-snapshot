package elastigo

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

const (
	percIndexName = "test-perc-index"
)

func TestPercolate(t *testing.T) {
	Convey("With a registered percolator", t, func() {
		c := NewTestConn()
		_, createErr := c.CreateIndex(percIndexName)
		So(createErr, ShouldBeNil)
		defer c.DeleteIndex(percIndexName)

		options := `{
      "percType": {
        "properties": {
          "message": {
            "type": "string"
          }
        }
      }
    }`

		err := c.PutMappingFromJSON(percIndexName, "percType", []byte(options))
		So(err, ShouldBeNil)

		data := `{
      "query": {
        "match": {
          "message": "bonsai tree"
        }
      }
    }`

		_, err = c.RegisterPercolate(percIndexName, "PERCID", data)
		So(err, ShouldBeNil)

		Convey("That matches the document", func() {
			// Should return the percolator id (registered query)
			doc := `{"doc": { "message": "A new bonsai tree in the office" }}`

			result, err := c.Percolate(percIndexName, "percType", "", nil, doc)
			So(err, ShouldBeNil)
			So(len(result.Matches), ShouldEqual, 1)
			match := result.Matches[0]
			So(match.Id, ShouldEqual, "PERCID")
			So(match.Index, ShouldEqual, percIndexName)
		})

		Convey("That does not match the document", func() {
			// Should NOT return the percolator id (registered query)
			doc := `{"doc": { "message": "Barren wasteland with no matches" }}`

			result, err := c.Percolate(percIndexName, "percType", "", nil, doc)
			So(err, ShouldBeNil)
			So(len(result.Matches), ShouldEqual, 0)
		})
	})
}
