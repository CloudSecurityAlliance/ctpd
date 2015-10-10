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

type MeasurementState string

func (s MeasurementState) IsValid() bool {
	switch string(s) {
	case "pending":
		return true
	case "activated":
		return true
	case "deactivated":
		return true
	}
	return false
}

func (s *MeasurementState) UnmarshalJSON(js []byte) error {
	if len(js) >= 2 {
		*s = MeasurementState(js[1 : len(js)-1])
		if s.IsValid() {
			return nil
		}
	}
	return errors.New("Invalid state property")
}

