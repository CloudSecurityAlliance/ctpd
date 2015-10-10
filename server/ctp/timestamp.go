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
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Timestamp int64
type Duration int64

func Now() Timestamp {
	return Timestamp(time.Now().UTC().Unix())
}

func (t Timestamp) Time() time.Time {
	return time.Unix(int64(t), 0).UTC()
}

func ParseTimestamp(str string) (Timestamp, error) {
	var Y, M, D, h, m, s int

	n, err := fmt.Sscanf(str, "%d-%d-%dT%d:%d:%dZ", &Y, &M, &D, &h, &m, &s)
	if n == 3 || n == 5 || n == 6 {
		t := time.Date(Y, time.Month(M), D, h, m, s, 0, time.UTC)
		return Timestamp(t.UTC().Unix()), nil
	}
	return Timestamp(0), err
}

func (t Timestamp) String() string {
	utime := t.Time()
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ", utime.Year(), int(utime.Month()), utime.Day(), utime.Hour(), utime.Minute(), utime.Second())
}

func SecondsSince(t Timestamp) int64 {
	return int64(Now() - t)
}

func (t Timestamp) IsZero() bool {
	return t == 0
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	tm, err := ParseTimestamp(string(data))
	*t = tm
	return err
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	if t.Time().Year() < 0 || t.Time().Year() > 9999 {
		return nil, fmt.Errorf("Timestamp.MarshalJSON: year outside of range [0,9999]")
	}
	return []byte(`"` + t.String() + `"`), nil
}

func (t Timestamp) GetBSON() (interface{}, error) {
	return t.String(), nil
}

func (t *Timestamp) SetBSON(raw bson.Raw) error {
	var s string
	err := raw.Unmarshal(&s)
	if err != nil {
		return err
	}
	*t, err = ParseTimestamp(s)
	return err
}
