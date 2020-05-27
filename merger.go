// Package merger provides the ability to merge the content of structs and maps.
package merger

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

// Merge objects with the default configuration.
//
// The default configuration will look for `json` tags and use them for
// merging members when available. This merge function does not overwrite
// empty values. For that, use MergeOverwrite.
//
// You can merge structs and objects. Example:
//     package main
//
//     import (
//         "fmt"
//         "github.com/gabstv/merger"
//     )
//     type A struct{
//         Name string `json:"name"`
//     }
//     type C struct{
//	       Name string
//	       Info string
//     }
//     func main(){
//         a := A{Name:"Test A"}
//         b := make(map[string]interface{})
//         c := C{}
//         merger.Merge(&c, &a)
//         merger.Merge(&b, &a)
//         fmt.Println(b["name"])
//         fmt.Println(c.NameWithSameTag)
//     }
func Merge(dst, src interface{}) error {
	return merge(dst, src, false, "json", nil)
}

func MergeWithTag(dst, src interface{}, tag string) error {
	return merge(dst, src, false, tag, nil)
}

func MergeOverwrite(dst, src interface{}) error {
	return merge(dst, src, true, "json", nil)
}

func MergeOverwriteWithTag(dst, src interface{}, tag string) error {
	return merge(dst, src, true, tag, nil)
}

func MergeWithOptions(dst, src interface{}, tag string, overwrite bool, tc []TypeConverter) error {
	if tag == "" {
		tag = "json"
	}
	return merge(dst, src, overwrite, tag, tc)
}

func merge(dst, src interface{}, overwrite bool, tag string, tc []TypeConverter) error {
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
		return merge(dst, reflect.ValueOf(src).Elem().Interface(), overwrite, tag, tc)
	case reflect.Struct, reflect.Map:
		// okay
	default:
		return fmt.Errorf("invalid source kind %v", srcKind.String())
	}
	return mergeStep(reflect.ValueOf(dst).Elem(), reflect.ValueOf(src), overwrite, tag, toTC(tc))
}

func mergeStep(dst, src reflect.Value, overwrite bool, tag string, tc TypeConverters) error {
	// get all "keys"
	if dst.Kind() == reflect.Struct {
		return mergeStepStruct(dst, src, overwrite, tag, tc)
	}
	return mergeStepMap(dst, src, overwrite, tag, tc)
}

func mergeStepStruct(dst, src reflect.Value, overwrite bool, tag string, tc TypeConverters) error {
	n := dst.NumField()
	fieldNames := make(map[string]reflect.Value)
	fieldJSONNames := make(map[string]reflect.Value)
	dstType := dst.Type()
	for i := 0; i < n; i++ {
		sfield := dstType.Field(i)
		if sfield.Name[0] >= 'A' && sfield.Name[0] <= 'Z' {
			fieldNames[sfield.Name] = dst.Field(i)
			if tt, ok := sfield.Tag.Lookup(tag); ok {
				parts := strings.Split(tt, ",")
				if strings.TrimSpace(parts[0]) != "-" {
					fieldJSONNames[strings.TrimSpace(parts[0])] = dst.Field(i)
				}
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
					if dstval.Kind() != reflect.Interface && srcval.Kind() == reflect.Interface {
						n := 10
						for srcval.Kind() == reflect.Interface && n > 0 {
							n--
							srcval = srcval.Elem()
						}
					}
					if dstval.Kind() == srcval.Kind() {
						if isEmptyValue(dstval) || overwrite {
							// copy
							if dstval.CanSet() && dstval.Type() == srcval.Type() {
								dstval.Set(srcval)
							} else {
								//err?
							}
							//TODO: recursive struct/map set
						}
					} else if dstval.CanSet() {
						tryMergeAll(dstval, srcval, tc)
					}
				} else if dstval, ok := fieldJSONNames[srckey.String()]; ok {
					srcval := src.MapIndex(srckey)
					if dstval.Kind() != reflect.Interface && srcval.Kind() == reflect.Interface {
						n := 10
						for srcval.Kind() == reflect.Interface && n > 0 {
							n--
							srcval = srcval.Elem()
						}
					}
					if dstval.Kind() == srcval.Kind() {
						if isEmptyValue(dstval) || overwrite {
							// copy
							if dstval.CanSet() && dstval.Type() == srcval.Type() {
								dstval.Set(srcval)
							} else {
								log.Println("CANNOT SET", srckey.String(), dstval.Interface(), dstval.CanAddr())
							}
							//TODO: recursive struct/map set
						}
					} else if dstval.CanSet() {
						tryMergeAll(dstval, srcval, tc)
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
					if dstval.CanSet() && dstval.Type() == srcval.Type() {
						dstval.Set(srcval)
					}
				} else {
					if dstval.CanSet() {
						tryMergeAll(dstval, srcval, tc)
					}
				}
			}
		}
	}
	return nil
}

func mergeStepMap(dst, src reflect.Value, overwrite bool, tag string, tc TypeConverters) error {
	if src.Kind() == reflect.Map {
		srckeys := src.MapKeys()
		for _, v := range srckeys {
			dv := dst.MapIndex(v)
			if !dv.IsValid() || (dv.IsValid() && overwrite) {
				// set
				dst.SetMapIndex(v, src.MapIndex(v))
			}
		}
		return nil
	}
	// src is struct
	n := src.NumField()
	srcType := src.Type()
	for i := 0; i < n; i++ {
		sfield := srcType.Field(i)
		if jt, ok := sfield.Tag.Lookup(tag); ok {
			parts := strings.Split(jt, ",")
			if part0 := strings.TrimSpace(parts[0]); part0 != "-" {
				dv := dst.MapIndex(reflect.ValueOf(part0))
				if !dv.IsValid() || (dv.IsValid() && overwrite) {
					// check for omitempty
					omitempty := false
					if len(parts) > 1 {
						for _, v := range parts[1:] {
							if v == "omitempty" {
								omitempty = true
								break
							}
						}
					}
					if omitempty && isEmptyValue(src.Field(i)) {
						continue
					}
					dst.SetMapIndex(reflect.ValueOf(part0), src.Field(i))
				}
			}
		} else {
			vv := sfield.Name
			// don't try to import unexported fields
			if vv[0] >= 'A' && vv[0] <= 'Z' {
				dv := dst.MapIndex(reflect.ValueOf(vv))
				if !dv.IsValid() || (dv.IsValid() && overwrite) {
					dst.SetMapIndex(reflect.ValueOf(vv), src.Field(i))
				}
			}
		}
	}
	return nil
}
