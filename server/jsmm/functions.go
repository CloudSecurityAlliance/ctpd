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

package jsmm

import (
	"regexp"
	"strings"
	"time"
)

func ToString(m *Machine, fn *Function, paramCount int) (int, *MachineException) {
	// stack -1: global object
	// stack -2: object to convert
	m.Push(NewString(m.Get(-2).ToString()))
	return 1, nil
}

func ToBoolean(m *Machine, fn *Function, paramCount int) (int, *MachineException) {
	// stack -1: global object
	// stack -2: object to convert
	m.Push(NewBoolean(m.Get(-2).ToBoolean()))
	return 1, nil
}

func ToNumber(m *Machine, fn *Function, paramCount int) (int, *MachineException) {
	// stack -1: global object
	// stack -2: object to convert
	m.Push(NewNumber(m.Get(-2).ToNumber()))
	return 1, nil
}

func ToArray(m *Machine, fn *Function, paramCount int) (int, *MachineException) {
	// stack -1: global object
	// stack -2: object to convert
	objref := m.Get(-2)

	switch objref.Type() {
	case TypeArray:
		m.Push(objref)
	case TypeString:
		reader := strings.NewReader(objref.ToString())
		a := NewArray()
		index := uint32(0)
		for {
			ch, _, err := reader.ReadRune()
			if err != nil {
				break
			}
			a.SetUInt32Property(index, NewString(string(ch)))
			index++
		}
		m.Push(a)
	case TypeNumber, TypeBoolean, TypeNull:
		m.Push(NewArray(objref))
	default:
		m.Push(NewNull())
	}
	return 1, nil
}

func TimeUTC(m *Machine, fn *Function, paramCount int) (int, *MachineException) {
	// stack -1: global object
	// stack -2: object to convert
	objref := m.Get(-2)

	if objref.Type() != TypeString {
		//m.Push(NewNumber(-1))
		return 0, NewMachineException("Missing time expression in call to timeUTC()")
	}
	ts := objref.ToString()
	if ts == "now" {
		t := time.Now().UTC().Unix()
		m.Push(NewNumber(float64(t)))
		return 1, nil
	}
	tp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		//m.Push(NewNumber(-1))
		return 0, NewMachineException("Time format error in timeUTC(), " + err.Error())
	}
	t := tp.Unix()
	m.Push(NewNumber(float64(t)))
	return 1, nil
}

func MatchRegexp(m *Machine, fn *Function, paramCount int) (int, *MachineException) {
	// stack -1: global object
	// stack -2: regexp string
	// stack -3: string or array of strings
	reref := m.Get(-2)
	objref := m.Get(-3)

	if reref.Type() != TypeString {
		m.Push(NewNull())
		return 1, nil
	}

	re, err := regexp.CompilePOSIX(reref.ToString())
	if err != nil {
		return 0, NewMachineException("matchRegex failed, " + err.Error())
	}

	switch objref.Type() {
	case TypeString:
		m.Push(NewBoolean(re.MatchString(objref.ToString())))
	case TypeArray:
		var i uint32

		array := objref.(*Array)
		retval := false

		for i = 0; i < array.length; i++ {
			val, err := array.GetUInt32Property(i)
			if err == nil && val.Type() == TypeString {
				retval = re.MatchString(val.ToString())
				if !retval {
				    break
                }
			}
		}
		m.Push(NewBoolean(retval))
	default:
		return 0, NewMachineException("matchRegex expects a string or an array as parameters")
	}
	return 1, nil
}

func Select(m *Machine, fn *Function, paramCount int) (int, *MachineException) {
	// stack -1: global object
	// stack -2: select key
	// stack -3: array of objects
	keyref := m.Get(-2)
	objref := m.Get(-3)

	if keyref.Type() != TypeString || objref.Type() != TypeArray {
		return 0, NewMachineException("select expects a string key and an array as parameters, got (%s,%s) instead", TypeOf(keyref), TypeOf(objref))
	}

	var i uint32
	result := NewArray()
	array := objref.(*Array)
	key := keyref.ToString()

	for i = 0; i < array.length; i++ {
		val, err := array.GetUInt32Property(i)
		if err == nil {
			val2, err2 := val.GetProperty(key)
			if err2 == nil {
				result.Push(val2)
			}
		}
	}
	m.Push(result)
	return 1, nil
}

func ArrayMin(m *Machine, fn *Function, paramCount int) (int, *MachineException) {
	// stack -1: table object
	arrayref := m.Get(-1)

	if array, ok := arrayref.(*Array); ok {
		var i uint32
		var ref MachineValue

		ref = NullConst

		for i = 0; i < array.length; i++ {

			val, err := array.GetUInt32Property(i)
			if err == nil && val.Type() != TypeNull {
				if ref == NullConst {
					ref = val
				} else {
					if lessThan(val, ref) {
						ref = val
					}
				}
			}
		}
		m.Push(ref)
	} else {
		return 0, NewMachineException("Array method called on non-array object")
	}
	return 1, nil
}

func ArrayMax(m *Machine, fn *Function, paramCount int) (int, *MachineException) {
	// stack -1: table object
	arrayref := m.Get(-1)

	if array, ok := arrayref.(*Array); ok {
		var i uint32
		var ref MachineValue

		ref = NullConst

		for i = 0; i < array.length; i++ {
			val, err := array.GetUInt32Property(i)
			if err == nil && val.Type() != TypeNull {
				if ref == NullConst {
					ref = val
				} else {
					if !lessThan(val, ref) {
						ref = val
					}
				}
			}
		}
		m.Push(ref)
	} else {
		return 0, NewMachineException("Array method called on non-array object")
	}
	return 1, nil
}

func ToJSON(m *Machine, fn *Function, paramCount int) (int, *MachineException) {
	m.Push(NewString(m.Get(-2).ToJSON()))
	return 1, nil
}
