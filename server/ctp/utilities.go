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
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"
)

var (
	AnybodyAccess = NewTags("access:anybody")
	UserAccess    = NewTags("access:user")
	AgentAccess   = NewTags("access:agent")
	AdminAccess   = NewTags("access:admin")
)

var idRegex = regexp.MustCompile("^[-_a-zA-Z0-9=]+$")

func IsValidIdentifier(id string) bool {
	return idRegex.MatchString(id)
}

func RequestSignature(prefix string, r *http.Request) (signature string, params []string, querystring_x string) {

	path := path.Clean(r.URL.Path)

	n := len(r.URL.Path)
	if n > 0 && r.URL.Path[n-1] == '/' {
		path = path + "/"
	}

	if !strings.HasPrefix(path, prefix) {
		log.Printf("%s, pref=%s", path, prefix)
		return "", nil, ""
	}

	path = strings.TrimPrefix(path, prefix)
	if path == "" {
		signature = r.Method + ":/"
	} else {
		pathItems := strings.Split(path, "/")

		signature = r.Method + ":"

		for i, v := range pathItems {
			if v != "" {
				if (i & 1) == 1 {
					if !IsValidIdentifier(v) {
						log.Printf("identifier '%s' is not in base64url format", v)
						return "", nil, ""
					}
					signature = signature + "/$"
				} else {
					signature = signature + "/" + v
				}
				params = append(params, v)
			} else {
				signature = signature + "/"
			}
		}
	}
	querystring_x = r.URL.Query().Get("x")
	if querystring_x != "" {
		signature += "?" + querystring_x
	}
	return /* signature, params, x */
}
