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
	"github.com/cloudsecurityalliance/ctpd/server/jsmm"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"reflect"
)

type ResultRow map[string]interface{}

type Result struct {
	Value       []ResultRow   `json:"value" bson:"value"`
	UpdateTime  ctp.Timestamp `json:"updateTime" bson:"updateTime"`
	AuthorityId *string       `json:"authorityId" bson:"authorityId"`
	Signature   *string       `json:"signature" bson:"signature"`
}

type Objective struct {
	Condition string      `json:"condition" bson:"condition"`
	Status    ctp.BoolErr `json:"status"    bson:"status"`
}

type Measurement struct {
	ctp.NamedResource `bson:",inline"`
	Metric            ctp.Link             `json:"metric"          bson:"metric"`
	Result            *Result              `json:"result"          bson:"result"`
	Objective         *Objective           `json:"objective"       bson:"objective"`
	CreateTrigger     *ctp.Link            `json:"createTrigger"   bson:"createTrigger,omitempty"`
	UserActivated     bool                 `json:"userActivated"   bson:"userActivated"`
	State             ctp.MeasurementState `json:"state"           bson:"state"`
}

func (measurement *Measurement) BuildLinks(context *ctp.ApiContext) {
	measurement.Self = ctp.NewLink(context.CtpBase, "@/measurements/$", measurement.Id)
	measurement.Scope = ctp.NewLink(context.CtpBase, "@/attributes/$", measurement.Parent[0])
}

func (measurement *Measurement) Load(context *ctp.ApiContext) *ctp.HttpError {
	if !ctp.LoadResource(context, "measurements", ctp.Base64Id(context.Params[1]), measurement) {
		return ctp.NewHttpError(http.StatusNotFound, "Not Found")
	}
	measurement.Metric = ctp.ExpandLink(context.CtpBase, measurement.Metric)
	measurement.BuildLinks(context)
	return nil
}

func (measurement *Measurement) Create(context *ctp.ApiContext) *ctp.HttpError {
	measurement.BuildLinks(context)
	measurement.Metric = ctp.ShortenLink(context.CtpBase, measurement.Metric)
	if !ctp.IsShortLink(measurement.Metric) {
		return ctp.NewBadRequestError("Invalid metric URL")
	}

	if !measurement.State.IsValid() {
		return ctp.NewBadRequestError("Invalid or missing state value")
	}

	if measurement.Result != nil {
		if err := measurementCheckResult(context, measurement); err != nil {
			return err
		}
	} else {
		if err := measurementCheckMetric(context, measurement); err != nil {
			return err
		}
	}

	if measurement.Objective != nil {
		if err := measurementObjectiveEvaluate(context, measurement); err != nil {
			return err
		}
	}

	if !ctp.CreateResource(context, "measurements", measurement) {
		return ctp.NewInternalServerError("Could not save measurement object")
	}
	return nil
}

func (measurement *Measurement) Update(context *ctp.ApiContext, update ctp.ResourceUpdater) *ctp.HttpError {
	measurement.BuildLinks(context)
	up, ok := update.(*Measurement)
	if !ok {
		return ctp.NewInternalServerError("Updated object is not a measurement") // should never happen
	}

	switch context.QueryParam {
	case "userActivated":
		switch up.State {
		case "activated":
			if measurement.State == "deactivated" {
				measurement.State = "pending" // FIXME: add backoffice logic for notification of state change?
			}
		case "deactivated":
			measurement.State = "deactivated"
			measurement.Result = nil
		default:
			return ctp.NewBadRequestError("state can only be 'activated' or 'deactivated'")
		}
	case "objective":
		measurement.Objective = up.Objective
		if err := measurementObjectiveEvaluate(context, measurement); err != nil {
			return err
		}
	case "result":
		if measurement.State == "deactivated" {
			return ctp.NewHttpError(http.StatusConflict, "Measurement is not in activated state.")
		}
		if measurement.State == "pending" {
			measurement.State = "activated"
		}

		if up.Result == nil {
			return ctp.NewBadRequestError("No result provided in request")
		}

		measurement.Result = up.Result

		if measurement.Result.UpdateTime.IsZero() {
			measurement.Result.UpdateTime = ctp.Now()
		}

		if err := measurementCheckMetric(context, measurement); err != nil {
			return err
		}

		if measurement.Objective != nil {
			if err := measurementObjectiveEvaluate(context, measurement); err != nil {
				return err
			}
		}

		measurementTriggersEvaluate(context, measurement)

	default:
		return ctp.NewBadRequestError("invalid query string") // should never happen, because already filtered in serve.go
	}

	if !ctp.UpdateResource(context, "measurements", measurement.Id, measurement) {
		return ctp.NewInternalServerError("Could not update measurement object")
	}
	return nil
}

func (measurement *Measurement) Delete(context *ctp.ApiContext) *ctp.HttpError {
	if !measurementDelete(context, measurement.Id) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not delete measurement")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////

func measurementCheckMetric(context *ctp.ApiContext, item *Measurement) *ctp.HttpError {
	var metric Metric

	if item.Metric == "" {
		return ctp.NewBadRequestError("Missing metric attribute in measurement.")
	}

	metricParams, ok := ctp.ParseLink(context.CtpBase, "@/metrics/$", item.Metric)
	if !ok {
		return ctp.NewBadRequestError("Metric URL is incorrect in measurement")
	}

	if !ctp.LoadResource(context, "metrics", ctp.Base64Id(metricParams[0]), &metric) {
		return ctp.NewBadRequestErrorf("Metric %s does not exist", ctp.ExpandLink(context.CtpBase, item.Metric))
	}
	return nil
}

func measurementCheckResult(context *ctp.ApiContext, item *Measurement) *ctp.HttpError {
	var metric Metric

	if item.Metric == "" {
		return ctp.NewBadRequestError("Missing metric attribute in measurement.")
	}

	if item.Result == nil {
		return ctp.NewBadRequestError("Missing result attribute in measurement.")
	}

	metricParams, ok := ctp.ParseLink(context.CtpBase, "@/metrics/$", item.Metric)
	if !ok {
		return ctp.NewBadRequestError("Metric URL is incorrect in measurement")
	}

	if !ctp.LoadResource(context, "metrics", ctp.Base64Id(metricParams[0]), &metric) {
		return ctp.NewBadRequestErrorf("Metric %s does not exist", ctp.ExpandLink(context.CtpBase, item.Metric))
	}

	for _, row := range item.Result.Value {
		if len(row) != len(metric.ResultFormat) {
			return ctp.NewBadRequestErrorf("Metric expects %d columns, but result value provides %d", len(metric.ResultFormat), len(row))
		}
		for name, cell := range row {
			metricDetailFound := false
			for _, metricDetail := range metric.ResultFormat {
				if metricDetail.Name == name {
					metricDetailFound = true
					kind := reflect.ValueOf(cell).Kind()
					switch metricDetail.Type {
					case "number":
						if kind != reflect.Float64 {
							return ctp.NewBadRequestError("Metric expects a number, but result is of different type")
						}
					case "boolean":
						if kind != reflect.Bool {
							return ctp.NewBadRequestError("Metric expects a boolean, but result is of different type")
						}
					case "string":
						if kind != reflect.String {
							return ctp.NewBadRequestError("Metric expects a string, but result is of different type")
						}
					default:
						return ctp.NewBadRequestError("Metric type information is incorrect")
					}
					break // from for
				}
			}
			if !metricDetailFound {
				return ctp.NewBadRequestErrorf("Metric does not describe '%s', which appears in result", name)
			}
		}
	}
	return nil
}

func importMeasurementResultInJSMM(machine *jsmm.Machine, result *Result) error {
	if result == nil {
		return nil
	}
	if err := jsmm.ImportGlobal(machine, "value", result.Value); err != nil {
		return err
	}
	if err := jsmm.ImportGlobal(machine, "updateTime", result.UpdateTime.String()); err != nil {
		return err
	}
	if err := jsmm.ImportGlobal(machine, "authorityId", result.AuthorityId); err != nil {
		return err
	}
	if err := jsmm.ImportGlobal(machine, "signature", result.Signature); err != nil {
		return err
	}
	return nil
}

func measurementObjectiveEvaluate(context *ctp.ApiContext, item *Measurement) *ctp.HttpError {
	item.Objective.Status = ctp.Terror

	ctp.Log(context, ctp.DEBUG, "Evaluating objective: %s\n", item.Objective.Condition)

	machine, err := jsmm.Compile(item.Objective.Condition)
	if err != nil {
		return ctp.NewBadRequestErrorf("Error in objective compilation - %s", err.Error())
	}

	if item.Result == nil {
		item.Objective.Status = ctp.Ttrue
		return nil
	}

	if context.DebugVM {
		machine.DebugMode(true)
	}

	if err := importMeasurementResultInJSMM(machine, item.Result); err != nil {
		return ctp.NewBadRequestErrorf("Error in objective evaluation while importing result - %s", err.Error)
	}

	v, exception := machine.Execute()
	if exception != nil {
		return ctp.NewBadRequestErrorf("Error in objective evaluation - %s", exception.Error())
	}
	if v != nil {
		item.Objective.Status = ctp.ToBoolErr(v.ToBoolean())
	}
	return nil
}

func measurementTriggersEvaluate(context *ctp.ApiContext, measurement *Measurement) {
	var trigger Trigger
	now := ctp.Now()

	mlink := ctp.ShortenLink(context.CtpBase, measurement.Self)
	query := context.Session.DB("ctp").C("triggers").Find(bson.M{"measurement": mlink})

	n, err := query.Count()
	if err == nil {
		ctp.Log(context, ctp.DEBUG, "Evaluating %d triggers related to measurement %s", n, measurement.Id)
	} else {
		ctp.Log(context, ctp.ERROR, "Failed to evaludate triggers related to measurement %s, %s", measurement.Id, err.Error())
		return
	}

	iter := query.Iter()
	for iter.Next(&trigger) {
		var err error
		var err_upd error
		var ok bool

		trigger.BuildLinks(context)

		ctp.Log(context, ctp.DEBUG, "Evaludating trigger %s, currently with status '%s'", trigger.Id, trigger.Status.String())

		switch trigger.Status {
		case ctp.Tfalse:
			ok, err = triggerCheckCondition(context, &trigger, measurement)
		case ctp.Ttrue:
			if uint(ctp.SecondsSince(trigger.StatusUpdateTime)) <= trigger.GuardTime {
				continue
			}
			ok, err = triggerCheckCondition(context, &trigger, measurement)
		case ctp.Terror:
			continue
		}

		switch {
		case err != nil:
			ctp.Log(context, ctp.ERROR, "Error in trigger %s for measurement %s", trigger.Id, measurement.Id)
			err_log := CreateErrorLogEntry(context, &trigger, err.Error())
			if err_log != nil {
				ctp.Log(context, ctp.ERROR, "Failed to create error log: %s", err_log.Error())
			}
			err_upd = context.Session.DB("ctp").C("triggers").Update(bson.M{"_id": trigger.Id}, bson.M{"$set": bson.M{"status": ctp.Terror, "statusUpdateTime": now.String()}})
		case ok:
			ctp.Log(context, ctp.DEBUG, "trigger %s is TRUE", trigger.Id)
			err_log := CreateNormalLogEntry(context, &trigger, measurement.Result, trigger.Tags)
			if err_log != nil {
				ctp.Log(context, ctp.ERROR, "Failed to create normal log: %s", err_log.Error())
			}
			err_upd = context.Session.DB("ctp").C("triggers").Update(bson.M{"_id": trigger.Id}, bson.M{"$set": bson.M{"status": ctp.Ttrue, "statusUpdateTime": now.String()}})
		default:
			ctp.Log(context, ctp.DEBUG, "Trigger %s is FALSE", trigger.Id)
			err_upd = context.Session.DB("ctp").C("triggers").Update(bson.M{"_id": trigger.Id}, bson.M{"$set": bson.M{"status": ctp.Tfalse, "statusUpdateTime": now.String()}})
		}

		if err_upd != nil {
			ctp.Log(context, ctp.ERROR, "Failed to update trigger 'status' and 'statusDateTime': %s", err_upd.Error())
		}
	}
}

////////////////////////////////////////////////////////////////////////////

func HandleGETMeasurement(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var measurement Measurement

	handler := ctp.NewGETHandler(ctp.UserRoleTag)

	handler.Handle(w, r, context, &measurement)
}

func HandlePOSTMeasurement(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var measurement Measurement

	handler := ctp.NewPOSTHandler(ctp.AdminRoleTag)

	handler.Handle(w, r, context, &measurement)
}

func HandlePUTMeasurement(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var measurement Measurement
	var update Measurement
	var access ctp.Tags

	switch context.QueryParam {
	case "userActivated":
		access = ctp.UserRoleTag
	case "result":
		access = ctp.AgentRoleTag
	case "objective":
		access = ctp.AdminRoleTag
	}
	handler := ctp.NewPUTHandler(access)

	handler.Handle(w, r, context, &measurement, &update)
}

func HandleDELETEMeasurement(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var measurement Measurement

	handler := ctp.NewDELETEHandler(ctp.AdminRoleTag)

	handler.Handle(w, r, context, &measurement)
}
