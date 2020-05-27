package merger

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type XString string
type XInt int64

func TestCustomTypeConverters(t *testing.T) {
	tcs := []TypeConverter{
		{
			SrcZeroType: XInt(0),
			DstZeroType: XString(""),
			Fn: func(v interface{}) interface{} {
				if vv, ok := v.(XInt); ok {
					return XString("XInt->" + strconv.FormatInt(int64(vv), 10))
				}
				return XString("")
			},
		},
		{
			SrcZeroType: time.Time{},
			DstZeroType: XString(""),
			Fn: func(v interface{}) interface{} {
				if vv, ok := v.(time.Time); ok {
					return XString("Year->" + strconv.Itoa(vv.Year()))
				}
				return XString("")
			},
		},
	}
	a := struct {
		Bar XInt      `json:"bar"`
		Baz time.Time `json:"baz"`
		Foo uint64    `json:"foo"`
	}{
		Bar: XInt(10),
		Baz: time.Date(2012, 1, 1, 0, 0, 1, 1, time.UTC),
		Foo: 600,
	}
	b := struct {
		Bar XString `json:"bar"`
		Baz XString `json:"baz"`
		Foo uint64  `json:"foo"`
	}{}
	assert.NoError(t, MergeWithOptions(&b, &a, "json", true, tcs))
	assert.Equal(t, XString("XInt->10"), b.Bar)
	assert.Equal(t, XString("Year->2012"), b.Baz)
}
