package merger

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyValue(t *testing.T) {
	tvf := func(ts *testing.T, val interface{}, expected bool) {
		vv := getRealValue(reflect.ValueOf(val))
		assert.Equal(t, expected, isEmptyValue(vv))
	}
	tvf(t, int16(0), true)
	tvf(t, int16(2), false)
	tvf(t, []int{}, true)
	tvf(t, []int{1}, false)
	tvf(t, "", true)
	tvf(t, "a", false)
	tvf(t, true, false)
	tvf(t, uint8(2), false)
	tvf(t, uint8(0), true)
	tvf(t, float64(0), true)
	tvf(t, float64(1), false)
	// empty ptr
	type Strc struct {
		Name string
	}
	var if0 *Strc
	assert.Equal(t, true, isEmptyValue(reflect.ValueOf(if0)))
	// invalid
	var invalid interface{}
	assert.Equal(t, "invalid", reflect.ValueOf(invalid).Kind().String())
	assert.Equal(t, false, isEmptyValue(reflect.ValueOf(invalid)))
}
