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

type Asset struct {
	ctp.NamedResource `bson:",inline"`
	Attributes        ctp.Link `json:"attributes" bson:"-"`
	AssetClass        *string  `json:"assetClass" bson:"assetClass"`
}

func (asset *Asset) BuildLinks(context *ctp.ApiContext) {
	asset.Self = ctp.NewLink(context.CtpBase, "@/assets/$", asset.Id)
	asset.Scope = ctp.NewLink(context.CtpBase, "@/serviceViews/$", asset.Parent[0])
	asset.Attributes = ctp.NewLink(context.CtpBase, "@/assets/$/attributes", asset.Id)
}

func (asset *Asset) Load(context *ctp.ApiContext) *ctp.HttpError {
	if !ctp.LoadResource(context, "assets", ctp.Base64Id(context.Params[1]), asset) {
		return ctp.NewHttpError(http.StatusNotFound, "Not Found")
	}
	asset.BuildLinks(context)
	return nil
}

func (asset *Asset) Create(context *ctp.ApiContext) *ctp.HttpError {
	asset.BuildLinks(context)
	if !ctp.CreateResource(context, "assets", asset) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not save asset")
	}
	return nil
}

func (asset *Asset) Delete(context *ctp.ApiContext) *ctp.HttpError {
	if !assetDelete(context, asset.Id) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not delete asset")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////

func HandleGETAsset(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var asset Asset

	handler := ctp.NewGETHandler(ctp.UserRoleTag)

	handler.Handle(w, r, context, &asset)
}

func HandlePOSTAsset(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var asset Asset

	handler := ctp.NewPOSTHandler(ctp.AdminRoleTag)

	handler.Handle(w, r, context, &asset)
}

func HandleDELETEAsset(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var asset Asset

	handler := ctp.NewDELETEHandler(ctp.AdminRoleTag)

	handler.Handle(w, r, context, &asset)
}
