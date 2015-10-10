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
	"net/http"
	"strings"
)

type Access struct {
	NamedResource `bson:",inline"`
	Token         string `json:"token" bson:"token"`
}

func BearerAuth(r *http.Request) (token string, ok bool) {
	const prefix = "Bearer "

	auth := r.Header.Get("Authorization")

	if auth == "" {
		return
	}
	if !strings.HasPrefix(auth, prefix) {
		return
	}
	return auth[len(prefix):], true
}
