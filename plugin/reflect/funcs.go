package reflect

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (t *Template) DefaultFuncs() map[string]interface{} {
	return map[string]interface{}{
		"ref": func(p string, o interface{}) interface{} {
			return get(o, tokenize(p))
		},

		"jsonMarshal": func(o interface{}) (string, error) {
			buff, err := json.MarshalIndent(o, "", "  ")
			return string(buff), err
		},

		"jsonUnmarshal": func(o interface{}) (interface{}, error) {
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

		"include": func(p string, o interface{}) (string, error) {
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
