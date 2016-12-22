package reflect

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
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

	v := reflect.Indirect(reflect.ValueOf(object))
	switch v.Kind() {
	case reflect.Slice:
		i := 0
		matches = indexExp.FindStringSubmatch(key)
		if len(matches) == 2 {
			if index, err := strconv.Atoi(matches[1]); err == nil {
				switch {
				case index >= 0 && v.Len() > index:
					i = index
				case index < 0 && v.Len() > -index: // negative index like python
					i = v.Len() + index
				}
			}
		}
		return get(v.Index(i).Interface(), path[1:])
	case reflect.Map:
		return get(v.MapIndex(reflect.ValueOf(key)).Interface(), path[1:])
	case reflect.Struct:
		return get(v.FieldByName(key).Interface(), path[1:])
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

// returns a url string of the base and a relative path.
// e.g. http://host/foo/bar/baz, ./boo.tpl gives http://host/foo/bar/boo.tpl
func getURL(root, rel string) (string, error) {

	// handle the case when rel is actually a full url
	if strings.Index(rel, "://") > 0 {
		u, err := url.Parse(rel)
		if err != nil {
			return "", err
		}
		return u.String(), nil
	}

	u, err := url.Parse(root)
	if err != nil {
		return "", err
	}
	u.Path = filepath.Clean(filepath.Join(filepath.Dir(u.Path), rel))
	return u.String(), nil
}

func fetch(s string) ([]byte, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "file":
		return ioutil.ReadFile(u.Path)

	case "http", "https":
		resp, err := http.Get(u.String())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}

	return nil, fmt.Errorf("unsupported url:%s", s)
}
