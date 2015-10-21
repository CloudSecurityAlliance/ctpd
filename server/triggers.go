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
	"fmt"
	"net/http"
	"github.com/cloudsecurityalliance/ctpd/server/ctp"
	"github.com/cloudsecurityalliance/ctpd/server/jsmm"
)

type Trigger struct {
	ctp.NamedResource `bson:",inline"`
	Measurement       ctp.Link      `json:"measurement" bson:"measurement"`
	Condition         string        `json:"condition" bson:"condition"`
	Notification      string        `json:"notification" bson:"notification"`
	GuardTime         uint          `json:"guardTime" bson:"guardTime"`
	Tags              []string      `json:"tags" bson:"tags"`
	Status            ctp.BoolErr   `json:"status" bson:"status"`
	StatusUpdateTime  ctp.Timestamp `json:"statusUpdateTime" bson:"statusUpdateTime"`
}

func (trigger *Trigger) BuildLinks(context *ctp.ApiContext) {
	trigger.Self = ctp.NewLink(context, "@/triggers/$", trigger.Id)
	trigger.Scope = ctp.NewLink(context, "@/serviceView/$", trigger.Parent[0])
}

func (trigger *Trigger) Load(context *ctp.ApiContext) *ctp.HttpError {
	if !ctp.LoadResource(context, "triggers", ctp.Base64Id(context.Params[1]), trigger) {
		return ctp.NewHttpError(http.StatusNotFound, "Not Found")
	}
	trigger.Measurement = ctp.ExpandLink(context, trigger.Measurement)
	trigger.BuildLinks(context)
	return nil
}

func (trigger *Trigger) Create(context *ctp.ApiContext) *ctp.HttpError {
	trigger.BuildLinks(context)
	trigger.Measurement = ctp.ShortenLink(context, trigger.Measurement)
	if !ctp.IsShortLink(trigger.Measurement) {
		return ctp.NewBadRequestError("Invalid measurement URL")
	}

	ok, err := triggerCheckCondition(context, trigger, nil)
	if err != nil {
		return ctp.NewBadRequestErrorf("%s", err.Error())
	}

	if ok {
		triggerLogAndNotify(context, trigger, nil)
	}

	if !ctp.CreateResource(context, "triggers", trigger) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not save object")
	}
	return nil
}

func (trigger *Trigger) Delete(context *ctp.ApiContext) *ctp.HttpError {
	if !ctp.DeleteResource(context, "triggers", trigger.Id) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not delete trigger")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////

func triggerLogAndNotify(context *ctp.ApiContext, trigger *Trigger, err error) {
	// TODO: log and notify
}

func triggerCheckCondition(context *ctp.ApiContext, trigger *Trigger, measurement *Measurement) (bool, error) {

    if measurement==nil {
        measurementParams, ok := ctp.ParseLink(context, "@/measurements/$", trigger.Measurement)
        if !ok {
            return false, fmt.Errorf("Measurement URL is incorrect")
        }

        measurement = new(Measurement)
        if !ctp.LoadResource(context, "measurements", ctp.Base64Id(measurementParams[0]), measurement) {
            return false, fmt.Errorf("Measurement %s does not exist", ctp.ExpandLink(context, trigger.Measurement))
        }
    }

	machine, err := jsmm.Compile(trigger.Condition)
	if err != nil {
		return false, ctp.NewBadRequestErrorf("Error in condition specification - %s", err.Error())
	}

    if context.DebugVM {
        machine.DebugMode(true)
    }

	if measurement.State != "activated" {
		return false, nil
	}
	if measurement.Result == nil {
		ctp.Log(context, ctp.ERROR, "In /measurements/%s, the state is activated but the value is null.", measurement.Id)
		return false, nil
	}

    if err := importMeasurementResultInJSMM(machine, measurement.Result); err != nil {
		return false, err
	}

	v, exception := machine.Execute()
	if exception != nil {
		return false, fmt.Errorf("Failed to evaluate condition: %s", exception.Error())
	}
	return v.ToBoolean(), nil
}

func HandleGETTrigger(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var trigger Trigger

	handler := ctp.NewGETHandler(ctp.UserRoleTag)

	handler.Handle(w, r, context, &trigger)
}

func HandlePOSTTrigger(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var trigger Trigger

	handler := ctp.NewPOSTHandler(ctp.AdminRoleTag)

	handler.Handle(w, r, context, &trigger)
}

func HandleDELETETrigger(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var trigger Trigger

	handler := ctp.NewDELETEHandler(ctp.AdminRoleTag)

	handler.Handle(w, r, context, &trigger)
}
