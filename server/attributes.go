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
	"github.com/cloudsecurityalliance/ctpd/server/ctp"
	"net/http"
)

type Attribute struct {
	ctp.NamedResource `bson:",inline"`
	Measurements      ctp.Link `json:"measurements" bson:"-"`
}

func (attribute *Attribute) BuildLinks(context *ctp.ApiContext) {
	attribute.Self = ctp.NewLink(context, "@/attributes/$", attribute.Id)
	attribute.Scope = ctp.NewLink(context, "@/assets/$", attribute.Parent)
	attribute.Measurements = ctp.NewLink(context, "@/attributes/$/measurements", attribute.Id)
}

func (attribute *Attribute) Load(context *ctp.ApiContext) *ctp.HttpError {
	if !ctp.LoadResource(context, "attributes", ctp.Base64Id(context.Params[1]), attribute) {
		return ctp.NewHttpError(http.StatusNotFound, "Not Found")
	}
	attribute.BuildLinks(context)
	return nil
}

func (attribute *Attribute) Create(context *ctp.ApiContext) *ctp.HttpError {
	attribute.BuildLinks(context)
	if !ctp.CreateResource(context, "attributes", attribute) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not save object")
	}
	return nil
}

func (attribute *Attribute) Delete(context *ctp.ApiContext) *ctp.HttpError {
    if !attributeDelete(context, attribute.Id) {
        return ctp.NewHttpError(http.StatusInternalServerError, "Could not delete attribute")
    }
    return nil
}


////////////////////////////////////////////////////////////////////////////

func HandleGETAttribute(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var attribute Attribute

	handler := ctp.NewGETHandler(ctp.UserAccess)

	handler.Handle(w, r, context, &attribute)
}

func HandlePOSTAttribute(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var attribute Attribute

	handler := ctp.NewPOSTHandler(ctp.AdminAccess)

	handler.Handle(w, r, context, &attribute)
}

func HandleDELETEAttribute(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var attribute Attribute

	handler := ctp.NewDELETEHandler(ctp.AdminAccess)

	handler.Handle(w, r, context, &attribute)
}

