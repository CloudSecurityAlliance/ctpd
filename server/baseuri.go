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

package server

import (
	"net/http"
	"github.com/cloudsecurityalliance/ctpd/server/ctp"
)

type BaseURI struct {
    ctp.NamedResource     `bson:",inline"`
	Version      string   `json:"version"`
	Provider     string   `json:"provider"`
	ServiceViews ctp.Link `json:"serviceViews"`
	Metrics      ctp.Link `json:"metrics"`
}

func (base *BaseURI) BuildLinks(context *ctp.ApiContext) {
	base.Self = ctp.NewLink(context, "@/")
	base.ServiceViews = ctp.NewLink(context, "@/serviceViews")
	base.Metrics = ctp.NewLink(context, "@/metrics")
}

// HandleGETBaseURI handles a request to the baseURI.
// It proceeds differently from other resources, which user the 'handler' paradigm.
func HandleGETBaseURI(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {

	if !context.AuthenticateClient(w, r) {
		ctp.Log(context, ctp.WARNING, "Missing access tags")
		return
	}

	if !context.VerifyAccessTags(w, ctp.UserAccess) {
		ctp.Log(context, ctp.WARNING, "Mismatched access tags for API signature")
		return
	}

    base := new(BaseURI)

    if !ctp.LoadResource(context, "baseuri", ctp.Base64Id("0"), base) {
        base.Version = "0"
        base.Annotation = "Unconfigured ctpd prototype server"
    }
    base.BuildLinks(context)

	ctp.RenderJsonResponse(w, context, 200, base)
}
