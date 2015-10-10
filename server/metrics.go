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
	//"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

type measurementParameter struct {
	Name  string `json:"name"  bson:"name"`
	Type  string `json:"type"  bson:"type"`
	Value string `json:"value" bson:"value"`
}

type resultColumnFormat struct {
	Name string `json:"name" bson:"name"`
	Type string `json:"type" bson:"type"`
}

type Metric struct {
	ctp.NamedResource     `bson:",inline"`
	BaseMetric            string                 `json:"baseMetric"            bson:"baseMetric"`
	MeasurementParameters []measurementParameter `json:"measurementParameters" bson:"measurementParameters"`
	ResultFormat          []resultColumnFormat   `json:"resultFormat"          bson:"resultFormat"`
}

func (metric *Metric) BuildLinks(context *ctp.ApiContext) {
	metric.Self = ctp.NewLink(context, "@/metrics/$", metric.Id)
	metric.Scope = ctp.NewLink(context, "@/")
}

func (metric *Metric) Load(context *ctp.ApiContext) *ctp.HttpError {
	if !ctp.LoadResource(context, "metrics", ctp.Base64Id(context.Params[1]), metric) {
		return ctp.NewHttpError(http.StatusNotFound, "Not Found")
	}
	metric.BuildLinks(context)
	return nil
}

func (metric *Metric) Create(context *ctp.ApiContext) *ctp.HttpError {
	metric.BuildLinks(context)

	if metric.AccessTags == nil {
		metric.AccessTags = ctp.NewTags("access:user")
	}

	if !ctp.CreateResource(context, "metrics", metric) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not save object")
	}
	return nil
}

func (metric *Metric) Delete(context *ctp.ApiContext) *ctp.HttpError {
	metricUrl := ctp.NewLink(context, "@/metrics/$", metric.Id) // just to create a clean URL

	query := context.Session.DB("ctp").C("measurements").Find(bson.M{"metric": metricUrl})
	count, err := query.Count()
	if err != nil {
		return ctp.NewInternalServerError(err)
	}
	if count > 0 {
		return ctp.NewHttpError(http.StatusConflict, "Metric cannot be deleted because it is still in use by some measurements.")
	}

	if !ctp.DeleteResource(context, "metrics", metric.Id) {
		return ctp.NewInternalServerError("Metric deletion failed")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////

func HandleGETMetric(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var metric Metric

	handler := ctp.NewGETHandler(ctp.UserAccess)

	handler.Handle(w, r, context, &metric)
}

func HandlePOSTMetric(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var metric Metric

	handler := ctp.NewPOSTHandler(ctp.AdminAccess)

	handler.Handle(w, r, context, &metric)
}

func HandleDELETEMetric(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var metric Metric

	handler := ctp.NewDELETEHandler(ctp.AdminAccess)

	handler.Handle(w, r, context, &metric)
}
