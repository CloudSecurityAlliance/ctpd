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
	"crypto/rand"
	"encoding/base64"
	"github.com/cloudsecurityalliance/ctpd/server/ctp"
	"net/http"
)

type Account ctp.Account

func (account *Account) BuildLinks(context *ctp.ApiContext) {
	account.Self = ctp.NewLink(context, "@/accounts/$", account.Id)
}

func (account *Account) Load(context *ctp.ApiContext) *ctp.HttpError {
	if !ctp.LoadResource(context, "accounts", ctp.Base64Id(context.Params[1]), account) {
        return ctp.NewHttpErrorf(http.StatusNotFound, "Account %s not found",context.Params[1])
	}
	account.BuildLinks(context)
	return nil
}

func (account *Account) Create(context *ctp.ApiContext) *ctp.HttpError {
	var key [24]byte

	account.BuildLinks(context)

	if account.Token == "" {
		_, err := rand.Read(key[:])
		if err != nil {
			return ctp.NewInternalServerError("Error generating key")
		}
		account.Token = base64.StdEncoding.EncodeToString(key[:])
	}

	if len(account.AccountTags.WithPrefix("account:")) == 0 {
		account.AccountTags.Append(ctp.NewTags("account:" + string(account.Id)))
	}

	if len(account.AccountTags.WithPrefix("role:")) == 0 {
		account.AccountTags.Append(ctp.UserRoleTag)
	}

	if !ctp.CreateResource(context, "accounts", account) {
		return ctp.NewHttpError(http.StatusInternalServerError, "Could not save account")
	}
	return nil
}

func (account *Account) Delete(context *ctp.ApiContext) *ctp.HttpError {
	if !ctp.DeleteResource(context, "accounts", account.Id) {
		return ctp.NewInternalServerError("Account deletion failed")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////

func HandleGETAccount(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var account Account

	handler := ctp.NewGETHandler(ctp.AdminRoleTag)

	handler.Handle(w, r, context, &account)
}

func HandlePOSTAccount(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var account Account

	handler := ctp.NewPOSTHandler(ctp.AdminRoleTag)

	handler.Handle(w, r, context, &account)
}

func HandleDELETEAccount(w http.ResponseWriter, r *http.Request, context *ctp.ApiContext) {
	var account Account

	handler := ctp.NewDELETEHandler(ctp.AdminRoleTag)

	handler.Handle(w, r, context, &account)
}
