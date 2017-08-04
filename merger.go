package merger

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

func Merge(dst, src interface{}) error {
	return merge(dst, src, false)
}

func merge(dst, src interface{}, overwrite bool) error {
	if dst == nil {
		return fmt.Errorf("dst cannot be nil")
	}
	if src == nil {
		return fmt.Errorf("src cannot be nil")
	}
	dstKind := reflect.ValueOf(dst).Kind()
	srcKind := reflect.ValueOf(src).Kind()
	switch dstKind {
	case reflect.Ptr:
		if reflect.ValueOf(dst).Elem().Kind() != reflect.Struct && reflect.ValueOf(dst).Elem().Kind() != reflect.Map {
			return fmt.Errorf("invalid dst kind %v", reflect.ValueOf(dst).Elem().Kind().String())
		}
	//	return merge(reflect.ValueOf(dst).Elem().Interface(), src, overwrite)
	case reflect.Struct, reflect.Map:
		return fmt.Errorf("dst needs to be a pointer")
	default:
		return fmt.Errorf("invalid destination kind %v", dstKind.String())
	}
	switch srcKind {
	case reflect.Ptr:
		return merge(dst, reflect.ValueOf(src).Elem().Interface(), overwrite)
	case reflect.Struct, reflect.Map:
		// okay
	default:
		return fmt.Errorf("invalid source kind %v", srcKind.String())
	}
	return mergeStep(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src), overwrite)
}

func mergeStep(dst, src reflect.Value, overwrite bool) error {
	// get all "keys"
	if dst.Kind() == reflect.Struct {
		return mergeStepStruct(dst, src, overwrite)
	}
	return fmt.Errorf("not implemented")
}

func mergeStepStruct(dst, src reflect.Value, overwrite bool) error {
	n := dst.NumField()
	fieldNames := make(map[string]reflect.Value)
	fieldJSONNames := make(map[string]reflect.Value)
	dstType := dst.Type()
	for i := 0; i < n; i++ {
		sfield := dstType.Field(i)
		fieldNames[sfield.Name] = dst.Field(i)
		if tt, ok := sfield.Tag.Lookup("json"); ok {
			parts := strings.Split(tt, ",")
			if strings.TrimSpace(parts[0]) != "-" {
				fieldJSONNames[strings.TrimSpace(parts[0])] = dst.Field(i)
			}
		}
	}
	// loop through src
	if src.Kind() == reflect.Map {
		keys := src.MapKeys()
		if len(keys) > 0 && keys[0].Kind() == reflect.String {
			for _, srckey := range keys {
				if dstval, ok := fieldNames[srckey.String()]; ok {
					srcval := src.MapIndex(srckey)
					if dstval.Kind() == srcval.Kind() {
						if isEmptyValue(dstval) || overwrite {
							// copy
							if dstval.CanSet() {
								dstval.Set(srcval)
							}
							//TODO: recursive struct/map set
						}
					}
				} else if dstval, ok := fieldJSONNames[srckey.String()]; ok {
					srcval := src.MapIndex(srckey)
					if dstval.Kind() == srcval.Kind() {
						if isEmptyValue(dstval) || overwrite {
							// copy
							if dstval.CanSet() {
								dstval.Set(srcval)
							} else {
								log.Println("CANNOT SET", srckey.String(), dstval.Interface(), dstval.CanAddr())
							}
							//TODO: recursive struct/map set
						}
					}
				}
			}
		}
	} else {
		// struct
		srcn := src.NumField()
		srcType := src.Type()
		for i := 0; i < srcn; i++ {
			if dstval, ok := fieldNames[srcType.Field(i).Name]; ok {
				srcval := src.Field(i)
				if dstval.Kind() == srcval.Kind() {
					if dstval.CanSet() {
						dstval.Set(srcval)
					}
				}
			}
		}
	}
	return nil
}
