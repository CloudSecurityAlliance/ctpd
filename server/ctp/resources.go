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
	"encoding/base64"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

func HandleNotImplemented(w http.ResponseWriter, r *http.Request, context *ApiContext) {
	RenderErrorResponse(w, context, NewHttpError(http.StatusNotImplemented, "Not implemented yet"))
}

////////////////////////////////////////////////////////////////////////////

type Base64Id string

func NewBase64Id() Base64Id {
	return Base64Id(base64.URLEncoding.EncodeToString([]byte(bson.NewObjectId())))
}

func (id *Base64Id) String() string {
	return string(*id)
}

type Resource struct {
	Id         Base64Id   `json:"-" bson:"_id,omitempty"`
	Parent     []Base64Id `json:"-" bson:"parent,omitempty"`
	Self       Link       `json:"self" bson:"-"`
	Scope      Link       `json:"scope,omitempty" bson:"-"`
	AccessTags Tags       `json:"accessTags,omitempty" bson:"accessTags"`
}

func (r *Resource) Super() *Resource {
	return r
}

type ResourceLoader interface {
	Super() *Resource
	Load(*ApiContext) *HttpError
}

type ResourceCreator interface {
	Super() *Resource
	Create(*ApiContext) *HttpError
}

type ResourceUpdater interface {
	Super() *Resource
	Update(*ApiContext, ResourceUpdater) *HttpError
}

type ResourceDeleter interface {
	Super() *Resource
	Delete(*ApiContext) *HttpError
}

type GETHandler struct {
	AccessTags Tags
	ShowTags   bool
}

func NewGETHandler(a Tags) *GETHandler {
	return &GETHandler{a, false}
}

func (handler *GETHandler) Handle(w http.ResponseWriter, r *http.Request, context *ApiContext, res ResourceLoader) {
	if !context.AuthenticateClient(w, r) {
		Log(context, WARNING, "Missing access tags")
		return
	}

	if !context.VerifyAccessTags(w, handler.AccessTags) {
		Log(context, WARNING, "Mismatched access tags for API signature")
		return
	}

	if err := res.Load(context); err != nil {
		RenderErrorResponse(w, context, err)
		return
	}

	if !context.VerifyAccessTags(w, res.Super().AccessTags) {
		return
	}

	if !handler.ShowTags {
		res.Super().AccessTags = nil
	}

	RenderJsonResponse(w, context, 200, res)
}

type POSTHandler struct {
	AccessTags Tags
}

func NewPOSTHandler(a Tags) *POSTHandler {
	return &POSTHandler{a}
}

func (handler *POSTHandler) Handle(w http.ResponseWriter, r *http.Request, context *ApiContext, res ResourceCreator) {
	var parent Resource

	if !context.AuthenticateClient(w, r) {
		Log(context, WARNING, "Missing access tags")
		return
	}

	if !context.VerifyAccessTags(w, handler.AccessTags) {
		Log(context, WARNING, "Mismatched access tags for API signature")
		return
	}

	level := len(context.Params)

	if level != 3 && level != 1 {
		panic("POST Handler expect a post to /a/b/c/ or /a/")
	}

	if level == 3 {
		if !LoadResource(context, context.Params[0], Base64Id(context.Params[1]), &parent) {
			RenderErrorResponse(w, context, NewHttpError(http.StatusNotFound, "Scope does not exists"))
			return
		}
		if !context.VerifyAccessTags(w, parent.AccessTags) {
			return
		}
	}

	if !ParseResource(r.Body, res) {
		RenderErrorResponse(w, context, NewHttpError(http.StatusBadRequest, "Failed to parse resource"))
		return
	}
	res.Super().Id = NewBase64Id()

	// For level 3 urls, we inherit tags from parent unless tags are specified in the request
	// for level 1 urls, we set defaults if necessary
	if level == 3 {
		res.Super().Parent = append([]Base64Id{parent.Id}, parent.Parent...)
		if res.Super().AccessTags == nil {
			res.Super().AccessTags = parent.AccessTags
		}
	} else {
		if res.Super().AccessTags == nil && context.Params[0] == "metrics" {
			res.Super().AccessTags = UserRoleTag
		}
	}

	if err := res.Create(context); err != nil {
		RenderErrorResponse(w, context, err)
		return
	}

	res.Super().AccessTags = nil

	RenderJsonResponse(w, context, 201, res)
}

type PUTHandler struct {
	AccessTags Tags
	ShowTags   bool
}

func NewPUTHandler(a Tags) *PUTHandler {
	return &PUTHandler{a, false}
}

func (handler *PUTHandler) Handle(w http.ResponseWriter, r *http.Request, context *ApiContext, resource ResourceUpdater, update ResourceUpdater) {
	if !context.AuthenticateClient(w, r) {
		Log(context, WARNING, "Missing access tags")
		return
	}

	if !context.VerifyAccessTags(w, handler.AccessTags) {
		Log(context, WARNING, "Mismatched access tags for API signature")
		return
	}

	if !LoadResource(context, context.Params[0], Base64Id(context.Params[1]), resource) {
		RenderErrorResponse(w, context, NewHttpErrorf(http.StatusNotFound, "%s was not found", r.RequestURI))
		return
	}

	if !context.VerifyAccessTags(w, resource.Super().AccessTags) {
		Log(context, WARNING, "Mismatched access tags for resource")
		return
	}

	if !ParseResource(r.Body, update) {
		RenderErrorResponse(w, context, NewHttpError(http.StatusBadRequest, "Failed to parse resource"))
		return
	}

	if err := resource.Update(context, update); err != nil {
		RenderErrorResponse(w, context, err)
		return
	}

	if !handler.ShowTags {
		resource.Super().AccessTags = nil
	}
	RenderJsonResponse(w, context, 200, resource)
}

type DELETEHandler struct {
	AccessTags Tags
}

func NewDELETEHandler(a Tags) *DELETEHandler {
	return &DELETEHandler{a}
}

func (handler *DELETEHandler) Handle(w http.ResponseWriter, r *http.Request, context *ApiContext, resource ResourceDeleter) {
	if !context.AuthenticateClient(w, r) {
		Log(context, WARNING, "Missing access tags")
		return
	}

	if !context.VerifyAccessTags(w, handler.AccessTags) {
		Log(context, WARNING, "Mismatched access tags for API signature")
		return
	}

	if !LoadResource(context, context.Params[0], Base64Id(context.Params[1]), resource) {
		RenderErrorResponse(w, context, NewHttpErrorf(http.StatusNotFound, "%s was not found", r.RequestURI))
		return
	}

	if !context.VerifyAccessTags(w, resource.Super().AccessTags) {
		Log(context, WARNING, "Mismatched access tags for resource")
		return
	}

	if err := resource.Delete(context); err != nil {
		RenderErrorResponse(w, context, err)
		return
	}

	RenderJsonResponse(w, context, 204, nil)
}

////////////////////////////////////////////////////////////////////////////

type NamedResource struct {
	Resource   `bson:",inline"`
	Name       string `json:"name" bson:"name"`
	Annotation string `json:"annotation" bson:"annotation"`
}
