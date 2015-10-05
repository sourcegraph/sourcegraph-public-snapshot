package elastigo

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCatNode(t *testing.T) {

	c := NewTestConn()

	Convey("Basic cat nodes", t, func() {

		fields := []string{"fm", "fe", "fcm", "fce", "ft", "ftt", "im", "rp", "n"}
		catNodes, err := c.GetCatNodeInfo(fields)

		So(err, ShouldBeNil)
		So(catNodes, ShouldNotBeNil)
		So(len(catNodes), ShouldBeGreaterThan, 0)

		for _, catNode := range catNodes {
			So(catNode.FieldMem, ShouldNotBeEmpty)
			So(catNode.FiltMem, ShouldNotBeEmpty)
			So(catNode.IDCacheMemory, ShouldNotBeEmpty)
			So(catNode.RamPerc, ShouldNotBeEmpty)
			So(catNode.Name, ShouldNotBeEmpty)
		}
	})

	Convey("Cat nodes with default arguments", t, func() {

		fields := []string{}
		catNodes, err := c.GetCatNodeInfo(fields)

		So(err, ShouldBeNil)
		So(catNodes, ShouldNotBeNil)
		So(len(catNodes), ShouldBeGreaterThan, 0)

		for _, catNode := range catNodes {
			So(catNode.Host, ShouldNotBeEmpty)
			So(catNode.IP, ShouldNotBeEmpty)
			So(catNode.NodeRole, ShouldNotBeEmpty)
			So(catNode.Name, ShouldNotBeEmpty)
		}
	})

	Convey("Invalid field error behavior", t, func() {

		fields := []string{"fm", "bogus"}
		catNodes, err := c.GetCatNodeInfo(fields)

		So(err, ShouldNotBeNil)

		for _, catNode := range catNodes {
			So(catNode.FieldMem, ShouldNotBeEmpty)
		}
	})
}
