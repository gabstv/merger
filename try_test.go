package merger

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRealValueNested(t *testing.T) {
	a := make([]interface{}, 0)
	//
	{
		b := "duck duck"
		c := &b
		d := &c
		a = append(a, d)
	}
	{
		b := "birdie"
		c := &b
		d := &c
		e := &d
		f := &e
		g := &f
		a = append(a, g)
	}
	for _, v := range a {
		assert.Equal(t, "string", getRealValue(reflect.ValueOf(v)).Kind().String())
	}
}
