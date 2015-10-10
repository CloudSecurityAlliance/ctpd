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
	"fmt"
    "net/http"
)

type HttpError struct {
	code    int
	message string
}

func (e *HttpError) Error() string {
	return e.message
}

func (e *HttpError) StatusCode() int {
	return e.code
}

func NewHttpError(code int, msg interface{}) *HttpError {
    switch msg.(type) {
    case string:
        return &HttpError{code, msg.(string)}
    case error:
        return &HttpError{code, msg.(error).Error()}
    case fmt.Stringer:
        return &HttpError{code, msg.(fmt.Stringer).String()}
    }
    return &HttpError{code, fmt.Sprintf("%v",msg)}
}

func NewHttpErrorf(code int, format string, params ...interface{}) *HttpError {
    return &HttpError{code, fmt.Sprintf(format,params...)}
}

func NewInternalServerError(msg interface{}) *HttpError {
    return NewHttpError(http.StatusInternalServerError, msg)
}

func NewInternalServerErrorf(format string, params ...interface{}) *HttpError {
    return NewHttpErrorf(http.StatusInternalServerError, format, params...)
}

func NewBadRequestError(msg interface{}) *HttpError {
    return NewHttpError(http.StatusBadRequest, msg)
}

func NewBadRequestErrorf(format string, params ...interface{}) *HttpError {
    return NewHttpErrorf(http.StatusBadRequest, format, params...)
}

func NewNotFoundError(msg interface{}) *HttpError {
    return NewHttpError(http.StatusNotFound, msg)
}

func NewNotFoundErrorf(format string, params ...interface{}) *HttpError {
    return NewHttpErrorf(http.StatusNotFound, format, params...)
}

