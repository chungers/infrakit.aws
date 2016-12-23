package reflect

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (t *Template) DefaultFuncs() map[string]interface{} {
	return map[string]interface{}{
		"var": func(name, doc string) interface{} {
			return t.binds[name]
		},
		"alias": func(name string, v interface{}) interface{} {
			t.binds[name] = v
			return ""
		},
		"ref": func(p string, o interface{}) interface{} {
			return get(o, tokenize(p))
		},

		"jsonEncode": func(o interface{}) (string, error) {
			buff, err := json.MarshalIndent(o, "", "  ")
			return string(buff), err
		},

		"jsonDecode": func(o interface{}) (interface{}, error) {
			ret := map[string]interface{}{}
			switch o := o.(type) {
			case string:
				err := json.Unmarshal([]byte(o), &ret)
				return ret, err
			case []byte:
				err := json.Unmarshal(o, &ret)
				return ret, err
			}
			return ret, fmt.Errorf("not-supported-value-type")
		},

		"include": func(p string, opt ...interface{}) (string, error) {
			var o interface{}
			if len(opt) > 0 {
				o = opt[0]
			}

			loc, err := getURL(t.url, p)
			if err != nil {
				return "", err
			}
			included, err := NewTemplate(loc)
			if err != nil {
				return "", err
			}
			// inherit the functions defined for this template
			for k, v := range t.funcs {
				included.AddFunc(k, v)
			}
			return included.Render(o)
		},

		"lines": func(o interface{}) ([]string, error) {
			ret := []string{}
			switch o := o.(type) {
			case string:
				return strings.Split(o, "\n"), nil
			case []byte:
				return strings.Split(string(o), "\n"), nil
			}
			return ret, fmt.Errorf("not-supported-value-type")
		},
	}
}
