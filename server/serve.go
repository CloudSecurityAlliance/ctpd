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

var ctpUrlMap = map[string]ctp.HandlerFunc{
	// Official API calls
	"GET:/":                            HandleGETBaseURI,
	"GET:/serviceViews":                HandleGETCollection,
	"GET:/serviceViews/$":              HandleGETServiceView,
	"GET:/serviceViews/$/triggers":     HandleGETCollection,
	"GET:/serviceViews/$/dependencies": HandleGETCollection,
	"GET:/serviceViews/$/assets":       HandleGETCollection,
	"GET:/serviceViews/$/logs":         HandleGETCollection,
	"GET:/assets/$":                    HandleGETAsset,
	"GET:/assets/$/attributes":         HandleGETCollection,
	"GET:/attributes/$":                HandleGETAttribute,
	"GET:/attributes/$/measurements":   HandleGETCollection,
	"GET:/measurements/$":              HandleGETMeasurement,
	"GET:/metrics":                     HandleGETCollection,
	"GET:/metrics/$":                   HandleGETMetric,
	"GET:/triggers/$":                  HandleGETTrigger,
	"GET:/dependencies/$":              ctp.HandleNotImplemented,
	"GET:/logs/$":                      HandleGETLogEntry,
	"PUT:/measurements/$/?initiate":    HandlePUTMeasurement,
	"POST:/serviceViews/$/triggers":    HandlePOSTTrigger,
	"DELETE:/triggers/$":               ctp.HandleNotImplemented,

	// Unoficial backoffice API
	"GET:/serviceViews/$?tags":        HandleGETTags,
	"PUT:/serviceViews/$?tags":        HandlePUTTags,
	"GET:/assets/$?tags":              HandleGETTags,
	"PUT:/assets/$?tags":              HandlePUTTags,
	"GET:/attributes/$?tags":          HandleGETTags,
	"PUT:/attributes/$?tags":          HandlePUTTags,
	"GET:/measurements/$?tags":        HandleGETTags,
	"PUT:/measurements/$?tags":        HandlePUTTags,
	"GET:/metrics/$?tags":             HandleGETTags,
	"PUT:/metrics/$?tags":             HandlePUTTags,
	"GET:/accounts/$?tags":          HandleGETTags,
	"PUT:/accounts/$?tags":          HandlePUTTags,
	"GET:/triggers/$?tags":            HandleGETTags,
	"PUT:/triggers/$?tags":            HandlePUTTags,
	"GET:/dependencies/$?tags":        ctp.HandleNotImplemented,
	"PUT:/dependencies/$?tags":        ctp.HandleNotImplemented,
	"GET:/logs/$?tags":                HandleGETTags,
	"PUT:/logs/$?tags":                HandlePUTTags,
	"PUT:/measurements/$?result":      HandlePUTMeasurement,
	"PUT:/measurements/$?objective":   HandlePUTMeasurement,
	"POST:/serviceViews":              HandlePOSTServiceView,
	"POST:/serviceViews/$/assets":     HandlePOSTAsset,
	"POST:/assets/$/attributes":       HandlePOSTAttribute,
	"POST:/attributes/$/measurements": HandlePOSTMeasurement,
	"POST:/metrics":                   HandlePOSTMetric,
	"POST:/dependencies":              ctp.HandleNotImplemented,
	"DELETE:/serviceViews/$":          HandleDELETEServiceView,
	"DELETE:/assets/$":                HandleDELETEAsset,
	"DELETE:/attributes/$":            HandleDELETEAttribute,
	"DELETE:/measurements/$":          HandleDELETEMeasurement,
	"DELETE:/metrics/$":               HandleDELETEMetric,
	"DELETE:/dependencies/$":          ctp.HandleNotImplemented,
	"DELETE:/logs/$":                  ctp.HandleNotImplemented,
	"GET:/accounts/$":               HandleGETAccount,
	"POST:/accounts":                HandlePOSTAccount,
	"GET:/accounts":                 HandleGETCollection,
	"PUT:/accounts/$":               ctp.HandleNotImplemented,
	"DELETE:/accounts/$":            HandleDELETEAccount,
}

type CtpApiHandlerMux struct {
	Configuration ctp.Configuration
}

func NewCtpApiHandlerMux(conf ctp.Configuration) *CtpApiHandlerMux {
	return &CtpApiHandlerMux{conf}
}

var muxRunOnce bool = false

func (mux *CtpApiHandlerMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	context, err := ctp.NewApiContext(r, mux.Configuration)
	if err != nil {
		ctp.RenderErrorResponse(w, context, ctp.NewHttpError(500, "Context creation failure"))
		return
	}
	defer context.Close()

	if handlerfunc, ok := ctpUrlMap[context.Signature]; ok {
		ctp.Log(context, ctp.INFO, "Serving request '%s' with signature '%s'", r.RequestURI, context.Signature)
		handlerfunc(w, r, context)
	} else {
		ctp.Log(nil, ctp.WARNING, "Not found: %s\n", context.Signature)
		http.NotFound(w, r)
	}
}
