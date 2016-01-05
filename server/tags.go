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

type TaggedResource ctp.Resource

func (res *TaggedResource)Super() *ctp.Resource {
    return (*ctp.Resource)(res)
}

func (res *TaggedResource) BuildLinks(context *ctp.ApiContext) {
    res.Self = ctp.NewLink(context.CtpBase, "@/$/$?x=tags", context.Params[0], res.Id)
}

func (res *TaggedResource) Load(context *ctp.ApiContext) *ctp.HttpError {
	if !ctp.LoadResource(context, context.Params[0], ctp.Base64Id(context.Params[1]), res) {
		return ctp.NewHttpError(http.StatusNotFound, "Not Found")
	}
    res.BuildLinks(context)
	return nil
}

func (res *TaggedResource) Update(context *ctp.ApiContext, update ctp.ResourceUpdater) *ctp.HttpError {
	res.BuildLinks(context)
	if !ctp.UpdateResourcePart(context, context.Params[0], ctp.Base64Id(context.Params[1]), "accessTags", update.Super().AccessTags) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not update object")
	}
    res.AccessTags = update.Super().AccessTags
	return nil
}

////////////////////////////////////////////////////////////////////////////

func HandleGETTags(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var res TaggedResource

	handler := ctp.NewGETHandler(ctp.AdminRoleTag)

    handler.ShowTags = true

	handler.Handle(w, r, context, &res)
}

func HandlePUTTags(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var res TaggedResource
    var update TaggedResource

	handler := ctp.NewPUTHandler(ctp.AdminRoleTag)

    handler.ShowTags = true

	handler.Handle(w, r, context, &res, &update)
}


