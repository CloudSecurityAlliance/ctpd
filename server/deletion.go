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
    "gopkg.in/mgo.v2/bson"
)

type deletecb func(*ctp.ApiContext, ctp.Base64Id) bool

func measurementDelete(context *ctp.ApiContext, id ctp.Base64Id) bool {
    return ctp.DeleteResource(context, "measurements", id)
}

func attributeDelete(context *ctp.ApiContext, id ctp.Base64Id) bool {
    if !IterateChildrenDelete(context, "measurements", "parent", id, measurementDelete) {
        return false
    }
    return ctp.DeleteResource(context, "attributes", id)
}

func assetDelete(context *ctp.ApiContext, id ctp.Base64Id) bool {
    if !IterateChildrenDelete(context, "attributes", "parent", id, attributeDelete) {
        return false
    }
    return ctp.DeleteResource(context, "assets", id)
}

func serviceViewDelete(context *ctp.ApiContext, id ctp.Base64Id) bool {
    if !IterateChildrenDelete(context, "assets", "parent", id, assetDelete) {
        return false
    }
    return ctp.DeleteResource(context, "serviceViews", id)
}

func IterateChildrenDelete(context *ctp.ApiContext, category string, selectorkey string, selectorvalue ctp.Base64Id, fn deletecb) bool {
    var item ctp.Resource

    query := context.Session.DB("ctp").C(category).Find(bson.M{selectorkey: string(selectorvalue)})
    iter := query.Iter()
    for iter.Next(&item) {
        if !fn(context, item.Id) {
            iter.Close()
            return false
        }
    }
    err := iter.Close()
    if err!=nil {
        return false
    }
    return true
}


