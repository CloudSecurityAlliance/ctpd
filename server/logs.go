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

type LogEntry struct {
	ctp.Resource `bson:",inline"`
	Trigger      ctp.Link      `json:"trigger" bson:"trigger"`
	CreationTime ctp.Timestamp `json:"creationTime" bson:"creationTime"`
	Result       *Result       `json:"result,omitempty" bson:"result,omitempty"`
	Error        *string       `json:"error,omitempty" bson:"error,omitempty"`
	Tags         []string      `json:"tags" bson:"tags"`
}

func (log *LogEntry) BuildLinks(context *ctp.ApiContext) {
	log.Self = ctp.NewLink(context, "@/logs/$", log.Id)
}

func (log *LogEntry) Load(context *ctp.ApiContext) *ctp.HttpError {
	if !ctp.LoadResource(context, "logs", ctp.Base64Id(context.Params[1]), log) {
		return ctp.NewHttpError(http.StatusNotFound, "Not Found")
	}
	log.BuildLinks(context)
	return nil
}

func (log *LogEntry) Create(context *ctp.ApiContext) *ctp.HttpError {
	log.BuildLinks(context)
	//log.CreationTime = ctp.Now()
	if !ctp.CreateResource(context, "logs", log) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not save object")
	}
	return nil
}

/*
func (log *LogEntry) Delete(context *ctp.ApiContext) *ctp.HttpError {
	if !logDelete(context, log.Id) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not delete log")
	}
	return nil
}
*/

func CreateNormalLogEntry(context *ctp.ApiContext, trigger *Trigger, result *Result, tags []string) *ctp.HttpError {
	var log = new(LogEntry)
	log.Id = ctp.NewBase64Id()
	log.Parent = trigger.Parent
	log.AccessTags = trigger.AccessTags
    log.CreationTime = ctp.Now()
	log.Trigger = trigger.Self
	log.Result = result
	log.Tags = tags
	return log.Create(context)
}

func CreateErrorLogEntry(context *ctp.ApiContext, trigger *Trigger, errmsg string) *ctp.HttpError {
	var log = new(LogEntry)
	log.Id = ctp.NewBase64Id()
	log.Parent = trigger.Parent
	log.AccessTags = trigger.AccessTags
    log.CreationTime = ctp.Now()
	log.Trigger = trigger.Self
	log.Error = &errmsg
	log.Tags = []string{"error"}
	return log.Create(context)
}

////////////////////////////////////////////////////////////////////////////

func HandleGETLogEntry(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var log LogEntry

	handler := ctp.NewGETHandler(ctp.UserRoleTag)

	handler.Handle(w, r, context, &log)
}

/*
func HandlePOSTLogEntry(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var log LogEntry

	handler := ctp.NewPOSTHandler(ctp.AdminRoleTag)

	handler.Handle(w, r, context, &log)
}

func HandleDELETELogEntry(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var log LogEntry

	handler := ctp.NewDELETEHandler(ctp.AdminRoleTag)

	handler.Handle(w, r, context, &log)
}
*/
