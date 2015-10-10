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
	"crypto/rand"
	"encoding/base64"
	"net/http"
)

type Access ctp.Access

func (access *Access) BuildLinks(context *ctp.ApiContext) {
	access.Self = ctp.NewLink(context, "@/access/$", access.Id)
}

func (access *Access) Load(context *ctp.ApiContext) *ctp.HttpError {
	if !ctp.LoadResource(context, "access", ctp.Base64Id(context.Params[1]), access) {
		return ctp.NewHttpError(http.StatusNotFound, "Not Found")
	}
	access.BuildLinks(context)
	return nil
}

func (access *Access) Create(context *ctp.ApiContext) *ctp.HttpError {
	var key [24]byte

	access.BuildLinks(context)

	if access.Token == "" {
		_, err := rand.Read(key[:])
		if err != nil {
			return ctp.NewInternalServerError("Error generating key")
		}
		access.Token = base64.StdEncoding.EncodeToString(key[:])
	}

    if len(access.AccessTags.WithPrefix("id:"))==0 {
        access.AccessTags.Append("id:" + string(access.Id))
    }
    if len(access.AccessTags.WithPrefix("access:"))==0 {
        access.AccessTags.Append("access:user")
    }

	if !ctp.CreateResource(context, "access", access) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not save access")
	}
	return nil
}

func (access *Access) Delete(context *ctp.ApiContext) *ctp.HttpError {
	if !ctp.DeleteResource(context, "access", access.Id) {
		return ctp.NewInternalServerError("Access deletion failed")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////

func HandleGETAccess(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var access Access

	handler := ctp.NewGETHandler(ctp.AdminAccess)

	handler.Handle(w, r, context, &access)
}

func HandlePOSTAccess(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var access Access

	handler := ctp.NewPOSTHandler(ctp.AdminAccess)

	handler.Handle(w, r, context, &access)
}

func HandleDELETEAccess(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var access Access

	handler := ctp.NewDELETEHandler(ctp.AdminAccess)

	handler.Handle(w, r, context, &access)
}
