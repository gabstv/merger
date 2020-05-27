package merger

import (
	"reflect"
)

type TypeConvFn func(v interface{}) interface{}

type TypeConverter struct {
	SrcZeroType interface{}
	DstZeroType interface{}
	Fn          TypeConvFn
	srcType     reflect.Type
	dstType     reflect.Type
}

type TypeConverters []TypeConverter

func toTC(v []TypeConverter) TypeConverters {
	if v == nil {
		return nil
	}
	for i, j := range v {
		j.srcType = reflect.TypeOf(j.SrcZeroType)
		j.dstType = reflect.TypeOf(j.DstZeroType)
		v[i] = j
	}
	return TypeConverters(v)
}

func (tc TypeConverters) TrySet(dstval reflect.Value, srcval reflect.Value) bool {
	if tc == nil {
		return false
	}
	if !dstval.CanSet() {
		return false
	}
	stype := srcval.Type()
	dtype := dstval.Type()
	for _, v := range tc {
		if v.srcType == stype && v.dstType == dtype {
			newv := reflect.ValueOf(v.Fn(srcval.Interface()))
			dstval.Set(newv)
			return true
		}
	}
	return false
}
