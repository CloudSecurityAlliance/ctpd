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
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"github.com/cloudsecurityalliance/ctpd/server/ctp"
	"strconv"
)

type CollectionItem struct {
	Link ctp.Link `json:"link"`
	Name string   `json:"name,omitempty"`
}

type Collection struct {
	ctp.Resource     `bson:",inline"`
	CollectionLength int              `json:"collectionLength"`
	ReturnedLength   int              `json:"returnedLength"`
	CollectionType   string           `json:"collectionType"`
	Items            []CollectionItem `json:"collection"`
}

func HandleGETCollection(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var item ctp.NamedResource
	var parent ctp.Resource
	var query *mgo.Query
	var collectionType string
	var skip, page, items int
	var err error

	collection := new(Collection)
	selector := make(bson.M)

	if name, ok := r.URL.Query()["name"]; ok {
		selector["name"] = name
	}

	page_query := r.URL.Query().Get("page")
	items_query := r.URL.Query().Get("items")
	if page_query != "" || items_query != "" {
		if page_query == "" || items_query == "" {
			ctp.RenderErrorResponse(w, context, ctp.NewHttpError(http.StatusBadRequest, "Must specify both 'page' and 'items' in query string."))
			return
		}
		if page, err = strconv.Atoi(page_query); err != nil || page < 0 {
			ctp.RenderErrorResponse(w, context, ctp.NewHttpError(http.StatusBadRequest, "page must be a positive number."))
			return
		}
		if items, err = strconv.Atoi(items_query); err != nil || items <= 0 {
			ctp.RenderErrorResponse(w, context, ctp.NewHttpError(http.StatusBadRequest, "items must be a non-zero positive number."))
		}
		skip = items * page
	}

	if !context.AuthenticateClient(w, r) {
		return
	}

	var mgoCollection *mgo.Collection

	if len(context.Params) == 1 {
		collectionType = context.Params[0]
		mgoCollection = context.Session.DB("ctp").C(collectionType)

		switch collectionType {
		case "serviceViews":
			if !context.VerifyAccessTags(w, ctp.UserRoleTag) {
				return
			}
			if !ctp.MatchTags(context.AccountTags, ctp.AdminRoleTag) {
				selector["accessTags"] = bson.M{"$in": context.AccountTags.WithPrefix("account:")}
			}
		case "metrics":
			if !context.VerifyAccessTags(w, ctp.UserRoleTag) {
				return
			}
		default:
			if !context.VerifyAccessTags(w, ctp.AdminRoleTag) {
				return
			}
		}
	} else {
		if !context.VerifyAccessTags(w, ctp.UserRoleTag) {
			return
		}

		if !ctp.LoadResource(context, context.Params[0], ctp.Base64Id(context.Params[1]), &parent) {
			ctp.RenderErrorResponse(w, context, ctp.NewNotFoundErrorf("Not found - /%s/%s does not exist", context.Params[0], context.Params[1]))
			return
		}

		if !context.VerifyAccessTags(w, parent.AccessTags) {
			return
		}
		collection.Scope = ctp.NewLink(context, "/$/$", context.Params[0], context.Params[1])

		collectionType = context.Params[2]
        if context.Params[2]=="indicators" {
            mgoCollection = context.Session.DB("ctp").C("measurements")
        } else {
            mgoCollection = context.Session.DB("ctp").C(collectionType)
        }
		selector["parent"] = context.Params[1]
	}

	query = mgoCollection.Find(selector)

	collection_length, err := query.Count()
	if err != nil {
		ctp.RenderErrorResponse(w, context, ctp.NewInternalServerError(err))
		return
	}

	query = query.Sort("$natural").Skip(skip).Limit(items)

	returned_length, err := query.Count()
	if err != nil {
		ctp.RenderErrorResponse(w, context, ctp.NewInternalServerError(err))
		return
	}

	collection.Self = ctp.Link(r.URL.RequestURI())
	collection.CollectionLength = collection_length
	collection.ReturnedLength = returned_length
	collection.CollectionType = collectionType
	collection.Items = make([]CollectionItem, 0, returned_length)

	iter := query.Iter()
	for iter.Next(&item) {
		collection.Items = append(collection.Items, CollectionItem{
			Link: ctp.NewLink(context, "@/$/$", collectionType, item.Id),
			Name: item.Name,
		})
	}

	if err := iter.Close(); err != nil {
		ctp.RenderErrorResponse(w, context, ctp.NewInternalServerError(err))
		return
	}

	ctp.RenderJsonResponse(w, context, 200, collection)
}
