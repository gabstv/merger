package merger

import (
	"reflect"
	"time"
)

func tryMergeTimeString(dstval reflect.Value, srcval reflect.Value) bool {
	// not matching types
	n := 10
	kk := srcval.Kind()
	for (kk == reflect.Interface || kk == reflect.Ptr) && n > 0 {
		srcval = srcval.Elem()
		kk = srcval.Kind()
		n--
	}
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
