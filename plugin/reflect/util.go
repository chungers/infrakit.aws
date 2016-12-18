package reflect

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	arrayIndexExp = regexp.MustCompile("(.*)\\[([+|-]*[0-9]+)\\]$")
	indexExp      = regexp.MustCompile("^\\[([+|-]*[0-9]+)\\]$")
)

func get(object interface{}, path []string) interface{} {

	if len(path) == 0 {
		return object
	}

	key := path[0]
	if key == "" {
		return get(object, path[1:])
	}

	// check if key is an array index
	matches := arrayIndexExp.FindStringSubmatch(key)
	if len(matches) == 3 && matches[1] != "" {
		key = matches[1]
		path = append([]string{key, fmt.Sprintf("[%s]", matches[2])}, path[1:]...)
		return get(object, path)
	}

	switch object := object.(type) {
	case map[string]interface{}:
		if v, has := object[key]; has {
			return get(v, path[1:])
		}

	case []interface{}:
		matches = indexExp.FindStringSubmatch(key)
		if len(matches) == 2 {
			if index, err := strconv.Atoi(matches[1]); err == nil {
				switch {
				case index > 0 && len(object) > index:
					return get(object[index], path[1:])
				case index < 0 && len(object) > -index: // negative index like python
					return get(object[len(object)+index], path[1:])
				}
			}
		}

	default:
		v := reflect.Indirect(reflect.ValueOf(object))
		if v.Kind() == reflect.Struct {
			return v.FieldByName(key).Interface()
		}
	}
	return nil
}

// With quoting to support azure rm type names: e.g. Microsoft.Network/virtualNetworks
// This will split a sting like /Resources/'Microsoft.Network/virtualNetworks'/managerSubnet/Name" into
// [ , Resources, Microsoft.Network/virtualNetworks, managerSubnet, Name]
func tokenize(s string) []string {
	if len(s) == 0 {
		return []string{}
	}

	a := []string{}
	start := 0
	quoted := false
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '/':
			if !quoted {
				a = append(a, strings.Replace(s[start:i], "'", "", -1))
				start = i + 1
			}
		case '\'':
			quoted = !quoted
		}
	}
	if start < len(s)-1 {
		a = append(a, strings.Replace(s[start:], "'", "", -1))
	}

	return a
}
