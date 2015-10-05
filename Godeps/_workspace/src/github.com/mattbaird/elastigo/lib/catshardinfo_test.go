package elastigo

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCatShardInfo(t *testing.T) {
	Convey("Create cat shard from started shard", t, func() {
		c, err := NewCatShardInfo("foo-2000-01-01-bar	0 p STARTED	1234	121	127.0.0.1	Ultra Man")
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
		So(c.IndexName, ShouldEqual, "foo-2000-01-01-bar")
		So(c.Primary, ShouldEqual, "p")
		So(c.State, ShouldEqual, "STARTED")
		So(c.Docs, ShouldEqual, 1234)
		So(c.Store, ShouldEqual, 121)
		So(c.NodeIP, ShouldEqual, "127.0.0.1")
		So(c.NodeName, ShouldEqual, "Ultra Man")

	})
	Convey("Create cat shard from realocating shard", t, func() {
		c, err := NewCatShardInfo("foo-2000-01-01-bar	0 p RELOCATING	1234	121	127.0.0.1	Ultra Man -> 10.0.0.1 Super Man")
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
		So(c.IndexName, ShouldEqual, "foo-2000-01-01-bar")
		So(c.Primary, ShouldEqual, "p")
		So(c.State, ShouldEqual, "RELOCATING")
		So(c.Docs, ShouldEqual, 1234)
		So(c.Store, ShouldEqual, 121)
		So(c.NodeIP, ShouldEqual, "127.0.0.1")
		So(c.NodeName, ShouldEqual, "Ultra Man")
	})
	Convey("Create cat shard from unallocated shard", t, func() {
		c, err := NewCatShardInfo("foo-2000-01-01-bar	0 p UNASSIGNED")
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
		So(c.IndexName, ShouldEqual, "foo-2000-01-01-bar")
		So(c.Primary, ShouldEqual, "p")
		So(c.State, ShouldEqual, "UNASSIGNED")
		So(c.Docs, ShouldEqual, 0)
		So(c.Store, ShouldEqual, 0)
		So(c.NodeIP, ShouldEqual, "")
		So(c.NodeName, ShouldEqual, "")
	})
	Convey("Create cat shard from invalid shard", t, func() {
		c, err := NewCatShardInfo("foo-2000-01-01-bar	0 p")
		So(err, ShouldEqual, ErrInvalidShardLine)
		So(c, ShouldBeNil)
	})
	Convey("Create cat shard from garbled shard", t, func() {
		c, err := NewCatShardInfo("foo-2000-01-01-bar	a p STARTED	abc	121	127.0.0.1	Ultra Man")
		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
		So(c.Shard, ShouldEqual, -1)
		So(c.IndexName, ShouldEqual, "foo-2000-01-01-bar")
		So(c.Primary, ShouldEqual, "p")
		So(c.State, ShouldEqual, "STARTED")
		So(c.Docs, ShouldEqual, 0)
		So(c.Store, ShouldEqual, 121)
		So(c.NodeIP, ShouldEqual, "127.0.0.1")
		So(c.NodeName, ShouldEqual, "Ultra Man")
	})
	Convey("Print cat shard from started shard", t, func() {
		c, _ := NewCatShardInfo("foo-2000-01-01-bar	0 p STARTED	1234	121	127.0.0.1	Ultra Man")
		s := c.String()
		So(s, ShouldContainSubstring, "foo-2000-01-01-bar:")
		So(s, ShouldContainSubstring, ":Ultra Man")
		c = nil
		s = c.String()
		So(s, ShouldEqual, ":::::::")
	})
	Convey("Print cat shard from short shard", t, func() {
		c, _ := NewCatShardInfo("foo-2000-01-01-bar	0 p STARTED	1234")
		s := c.String()
		So(s, ShouldContainSubstring, "foo-2000-01-01-bar:0:p:STARTED:1234")
		c, _ = NewCatShardInfo("foo-2000-01-01-bar	0 p STARTED	1234	121")
		s = c.String()
		So(s, ShouldContainSubstring, "oo-2000-01-01-bar:0:p:STARTED:1234:121")
		c, _ = NewCatShardInfo("foo-2000-01-01-bar	0 p STARTED	1234	121	127.0.0.1")
		s = c.String()
		So(s, ShouldContainSubstring, "oo-2000-01-01-bar:0:p:STARTED:1234:121:127.0.0.1")
	})

}
