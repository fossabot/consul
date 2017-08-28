package config

import (
	"fmt"
	"reflect"
)

// Merge combines multiple config files into a single one. Values of
// simple types like bool, int, string and maps are replaced when
// overwritten. Values of slice types are appended.
func Merge(files []File) File {
	merge := func(out interface{}, val func(File) interface{}) {
		// out is a pointer to a type, e.g. **int or *[]string or *map[string]string
		// outTyp is the element type, e.g. *int or []string or map[string]string
		outTyp := reflect.ValueOf(out).Elem().Type()

		// v is the value that will be assigned to out.
		// It must have the same type as the element of out.
		var v reflect.Value

		for _, f := range files {
			// x is the value from the current file struct
			x := val(f)
			xVal, xTyp := reflect.ValueOf(x), reflect.TypeOf(x)

			// x must be of type elemTyp
			if xTyp != outTyp {
				panic(fmt.Sprintf("out type %v != value type %v", outTyp, xTyp))
			}

			// merge slices by appending and other types by overriding
			switch outTyp.Kind() {
			case reflect.Map:
				if xVal.Len() > 0 {
					v = xVal
				}

			case reflect.Ptr:
				if !reflect.ValueOf(x).IsNil() {
					v = xVal
				}

			case reflect.Slice:
				if !v.IsValid() {
					v = reflect.Zero(outTyp)
				}
				v = reflect.AppendSlice(v, xVal)

			default:
				panic(fmt.Sprintf("unsupported element type: %v", outTyp))
			}
		}

		if v.IsValid() {
			reflect.ValueOf(out).Elem().Set(v)
		}
	}

	var f File

	merge(&f.Bootstrap, func(f File) interface{} { return f.Bootstrap })
	merge(&f.CheckUpdateInterval, func(f File) interface{} { return f.CheckUpdateInterval })
	merge(&f.Datacenter, func(f File) interface{} { return f.Datacenter })
	merge(&f.BindAddr, func(f File) interface{} { return f.BindAddr })
	merge(&f.Ports.DNS, func(f File) interface{} { return f.Ports.DNS })
	merge(&f.JoinAddrsLAN, func(f File) interface{} { return f.JoinAddrsLAN })
	merge(&f.NodeMeta, func(f File) interface{} { return f.NodeMeta })

	return f
}
