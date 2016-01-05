//    Copyright 2015 Cloud Security Alliance EMEA (cloudsecurityalliance.org)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ctp

import (
	"strings"
    "fmt"
	"unicode/utf8"
)

type Link string

func NewLink(baseURI Link, format string, params ...interface{}) Link {
	for _, v := range params {
        switch s := v.(type) {
		case string:
			format = strings.Replace(format, "$", s, 1)
		case Base64Id:
			format = strings.Replace(format, "$", string(s), 1)
		case fmt.Stringer:
			format = strings.Replace(format, "$", s.String(), 1)
		default:
			panic("NewLink called with non-stringify-able argument")
		}
	}
	return Link(strings.Replace(format, "@/", string(baseURI), 1))
}

func IsShortLink(link Link) bool {
	if len(link) == 0 {
		return false
	}
	return link[0] == '@'
}

func ShortenLink(baseURI Link, link Link) Link {
	return Link(strings.Replace(string(link), string(baseURI), "@/", 1))
}

func ExpandLink(baseURI Link, link Link) Link {
	return Link(strings.Replace(string(link), "@/", string(baseURI), 1))
}

func ParseLink(baseURI Link, format string, link Link) ([]string, bool) {
	link = ShortenLink(baseURI, link)

	var r []string

	for _, v := range format {
		if len(link) == 0 {
			return nil, false
		}
		c, size := utf8.DecodeRuneInString(string(link))

		if v == '$' {
			x := ""
			for c != '/' && len(link) > 0 {
				x += string(c)
				link = link[size:]
				c, size = utf8.DecodeRuneInString(string(link))
			}
			r = append(r, x)
		} else {
			if c != v {
				return nil, false
			}
			link = link[size:]
		}
	}
	if len(link) > 0 {
		return nil, false
	}
	return r, true
}
