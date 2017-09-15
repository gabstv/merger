package merger

import (
	"reflect"
	"time"
)

type tryFunc func(dstval reflect.Value, srcval reflect.Value) bool

func tryMergeAll(dstval reflect.Value, srcval reflect.Value) {
	if tryMergeTimeString(dstval, srcval) {
		return
	}
	if tryMergeNumeric(dstval, srcval) {
		return
	}
}

func tryMergeTimeString(dstval reflect.Value, srcval reflect.Value) bool {
	// not matching types
	srcval = getRealValue(srcval)
	// try to convert
	if srcval.Kind() == reflect.String && dstval.Kind() == reflect.Struct {
		// maybe it's time
		vvv := dstval.Interface()
		if _, ok := vvv.(time.Time); ok {
			t0 := time.Now()
			t1 := &t0
			if t1.UnmarshalJSON([]byte("\""+srcval.String()+"\"")) == nil {
				// set
				t0 = *t1
				newsrc := reflect.ValueOf(t0)
				dstval.Set(newsrc)
				return true
			}
		}
	}
	return false
}

func tryMergeNumeric(dstval reflect.Value, srcval reflect.Value) bool {
	// not matching types
	srcval = getRealValue(srcval)
	switch dstval.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return _mNumericInt(dstval, srcval)
		//TODO: floats, uints
	}
	return false
}

func _mNumericInt(dstval reflect.Value, srcval reflect.Value) bool {
	switch srcval.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		ii0 := srcval.Int()
		dstval.SetInt(ii0)
		return true
	case reflect.Float32, reflect.Float64:
		ff0 := srcval.Float()
		ii0 := int64(ff0)
		dstval.SetInt(ii0)
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		ui0 := srcval.Uint()
		ii0 := int64(ui0)
		dstval.SetInt(ii0)
		return true
	}
	return false
}

func getRealValue(v reflect.Value) reflect.Value {
	n := 25
	kk := v.Kind()
	for (kk == reflect.Interface || kk == reflect.Ptr) && n > 0 {
		v = v.Elem()
		kk = v.Kind()
		n--
	}
	return v
}
