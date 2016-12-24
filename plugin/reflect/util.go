package reflect

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	indexRoot     = "\\[(([+|-]*[0-9]+)|((.*)=(.*)))\\]$"
	arrayIndexExp = regexp.MustCompile("(.*)" + indexRoot)
	indexExp      = regexp.MustCompile("^" + indexRoot)
)

func get(object interface{}, path []string) interface{} {

	if len(path) == 0 {
		return object
	}

	key := path[0]
	if key == "" {
		return get(object, path[1:])
	}

	// check if key is an array index of the form <1>[<2>]
	matches := arrayIndexExp.FindStringSubmatch(key)
	if len(matches) > 2 && matches[1] != "" {
		key = matches[1]
		path = append([]string{key, fmt.Sprintf("[%s]", matches[2])}, path[1:]...)
		return get(object, path)
	}

	v := reflect.Indirect(reflect.ValueOf(object))
	switch v.Kind() {
	case reflect.Slice:
		i := 0
		matches = indexExp.FindStringSubmatch(key)
		if len(matches) > 0 {
			if matches[2] != "" {
				// numeric index
				if index, err := strconv.Atoi(matches[1]); err == nil {
					switch {
					case index >= 0 && v.Len() > index:
						i = index
					case index < 0 && v.Len() > -index: // negative index like python
						i = v.Len() + index
					}
				}
				return get(v.Index(i).Interface(), path[1:])

			} else if matches[3] != "" {
				// equality search index for 'field=check'
				lhs := matches[4] // supports another select expression for extractly deeply from the struct
				rhs := matches[5]
				// loop through the array looking for field that matches the check value
				for j := 0; j < v.Len(); j++ {
					if el := get(v.Index(j).Interface(), tokenize(lhs)); el != nil {
						if fmt.Sprintf("%v", el) == rhs {
							return get(v.Index(j).Interface(), path[1:])
						}
					}
				}
			}
		}
	case reflect.Map:
		value := v.MapIndex(reflect.ValueOf(key))
		if value.IsValid() {
			return get(value.Interface(), path[1:])
		}
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

	case "unix":
		// unix: will look for a socket that matches the host name at a
		// directory path set by environment variable.
		c, err := socketClient(u)
		if err != nil {
			return nil, err
		}
		u.Scheme = "http"
		resp, err := c.Get(u.String())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}

	return nil, fmt.Errorf("unsupported url:%s", s)
}

const (
	// EnvUnixSocketDir is the environment variable used by the unix:// client to locate the sockets (=hostname in url)
	EnvUnixSocketDir = "SOCKET_DIR"
)

func socketClient(u *url.URL) (*http.Client, error) {
	socketPath := filepath.Join(os.Getenv(EnvUnixSocketDir), u.Host)
	if f, err := os.Stat(socketPath); err != nil {
		return nil, err
	} else if f.Mode()&os.ModeSocket == 0 {
		return nil, fmt.Errorf("not-a-socket:%v", socketPath)
	}
	return &http.Client{
		Transport: &http.Transport{
			Dial: func(proto, addr string) (conn net.Conn, err error) {
				return net.Dial("unix", socketPath)
			},
		},
	}, nil
}
