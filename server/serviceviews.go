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

type ServiceView struct {
	ctp.NamedResource `bson:",inline"`
	Provider          string   `json:"provider"         bson:"provider"`
	Dependencies      ctp.Link `json:"dependencies"     bson:"-"`
	Assets            ctp.Link `json:"assets"           bson:"-"`
	ServiceClass      *string  `json:"serviceClass"     bosn:"serviceClass"`
	Logs              ctp.Link `json:"logs"`
	Triggers          ctp.Link `json:"triggers"`
}

func (serviceview *ServiceView) BuildLinks(context *ctp.ApiContext) {
	serviceview.Self = ctp.NewLink(context, "@/serviceViews/$", serviceview.Id)
	serviceview.Scope = ctp.NewLink(context, "@/")
	serviceview.Dependencies = ctp.NewLink(context, "@/serviceViews/$/dependencies", serviceview.Id)
	serviceview.Assets = ctp.NewLink(context, "@/serviceViews/$/assets", serviceview.Id)
	serviceview.Logs = ctp.NewLink(context, "@/serviceViews/$/logs", serviceview.Id)
	serviceview.Triggers = ctp.NewLink(context, "@/serviceViews/$/triggers", serviceview.Id)
}

func (serviceview *ServiceView) Load(context *ctp.ApiContext) *ctp.HttpError {
	if !ctp.LoadResource(context, "serviceViews", ctp.Base64Id(context.Params[1]), serviceview) {
		return ctp.NewHttpError(http.StatusNotFound, "Not Found")
	}
	serviceview.BuildLinks(context)
	return nil
}

func (serviceview *ServiceView) Create(context *ctp.ApiContext) *ctp.HttpError {
	serviceview.BuildLinks(context)

	if !ctp.CreateResource(context, "serviceViews", serviceview) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not save object")
	}
	return nil
}

func (serviceview *ServiceView) Delete(context *ctp.ApiContext) *ctp.HttpError {
	if !serviceViewDelete(context, serviceview.Id) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not delete service-view")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////

func HandleGETServiceView(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var serviceview ServiceView

	handler := ctp.NewGETHandler(ctp.UserRoleTag)

	handler.Handle(w, r, context, &serviceview)
}

func HandlePOSTServiceView(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var serviceview ServiceView

	handler := ctp.NewPOSTHandler(ctp.AdminRoleTag)

	handler.Handle(w, r, context, &serviceview)
}

func HandleDELETEServiceView(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var serviceview ServiceView

	handler := ctp.NewDELETEHandler(ctp.AdminRoleTag)

	handler.Handle(w, r, context, &serviceview)
}
