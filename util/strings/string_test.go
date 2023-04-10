package strings

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestExtractValue(t *testing.T) {
	Convey("提取键值对中的值", t, func() {

		for _, expected := range []struct {
			KV    string
			Sep   string
			Value string
		}{
			{KV: "name : voidint", Sep: ColonSep, Value: "voidint"},
			{KV: "name= tty", Sep: EqSep, Value: "tty"},
			{KV: "age= 22", Sep: EqSep, Value: "22"},
			{KV: "age is 22", Sep: EqSep, Value: "age is 22"},
		} {
			So(ExtractValue(expected.KV, expected.Sep), ShouldEqual, expected.Value)
		}
	})
}
