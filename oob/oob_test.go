package oob

import (
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBySeqASC(t *testing.T) {
	Convey("带外事件日志切片倒序", t, func() {
		items := []*EventLog{
			{SeqNumber: "3"},
			{SeqNumber: "1"},
			{SeqNumber: "2"},
		}

		sort.Sort(BySeqASC(items))

		So(len(items), ShouldEqual, 3)
		So(items[0].SeqNumber, ShouldEqual, "1")
		So(items[1].SeqNumber, ShouldEqual, "2")
		So(items[2].SeqNumber, ShouldEqual, "3")
	})
}
