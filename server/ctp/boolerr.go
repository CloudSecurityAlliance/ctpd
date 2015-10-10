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
    "errors"
)

type BoolErr int

const (
    Tfalse BoolErr = iota
    Ttrue
    Terror
)

func ToBoolErr(b bool) BoolErr {
    if b {
        return Ttrue
    }
    return Tfalse
}

func (t BoolErr) String() string {
    switch t {
    case Terror:
        return "error"
    case Tfalse:
        return "false"
    case Ttrue:
        return "true"
    }
    return "<Invalid>"
}

func (t *BoolErr) UnmarshalJSON(js []byte) error {
    switch string(js) {
    case `"error"`:
        *t = Terror
    case `"true"`:
        *t = Ttrue
    case `"false"`:
        *t = Tfalse
    default:
        return errors.New("Value must be either 'true', 'false' or 'error'")
    }
    return nil
}

func (t BoolErr) MarshalJSON() ([]byte, error) {
    switch t {
    case Terror:
        return []byte(`"error"`), nil
    case Tfalse:
        return []byte(`"false"`), nil
    case Ttrue:
        return []byte(`"true"`), nil
    }
    return nil, errors.New("Invalid BoolErr value")
}

