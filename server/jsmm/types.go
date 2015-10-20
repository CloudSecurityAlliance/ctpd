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
	"fmt"
	"math"
	"strconv"
)

type MachineType int

const (
	TypeError MachineType = iota
	TypeNull
	TypeBoolean
	TypeNumber
	TypeString
	TypeObject
	TypeArray
	TypeFunction
)

type JSFunction func(*Machine, *Function, int) (int, *MachineException)

/*
 *  Machinevalue describes an interface implemented by
 *  all types, whether primitive or objects.
 */


type MachineValue interface {
    Type() MachineType
	ToString() string
	ToNumber() float64
	ToBoolean() bool
	ToJSON() string
	GetProperty(prop string) (MachineValue, *MachineException)
	SetProperty(prop string, val MachineValue) *MachineException
}

/*
 * TypeOf returns a string representation of the type
 */

func TypeOf(m MachineValue) string {
	switch m.Type() {
	case TypeError:
		return "error"
	case TypeNull:
		return "null"
	case TypeBoolean:
		return "boolean"
	case TypeNumber:
		return "number"
	case TypeString:
		return "string"
	case TypeObject:
		return "object"
	case TypeArray:
		return "array"
	case TypeFunction:
		return "function"
	}
	return "unknown"
}

/*
 * default behaviour of GetProperty and SetProperty
 */

func failGetProperty(v MachineValue, prop string) (MachineValue, *MachineException) {
    return NewNull(), NewMachineException("BadType: cannot read property '%s' of %s", prop, TypeOf(v))
}

func failSetProperty(v MachineValue, prop string) *MachineException {
	return NewMachineException("BadType: cannot set property %s of %s", prop, TypeOf(v))
}

//
// The Number type
//
type Number struct {
	value float64
}

func NewNumber(f float64) MachineValue {
	return &Number{f}
}
func NewNumberString(v string) MachineValue {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		panic("syntax error in number " + v)
	}
	return NewNumber(f)
}
func (n *Number) Type() MachineType {
    return TypeNumber
}
func (n *Number) ToBoolean() bool {
	if n.value == 0 || math.IsNaN(n.value) {
		return false
	}
	return true
}
func (n *Number) ToString() string {
	return fmt.Sprintf("%g", n.value)
}
func (n *Number) ToNumber() float64 {
	return n.value
}
func (n *Number) ToJSON() string {
	return n.ToString()
}

func (n *Number) GetProperty(prop string) (MachineValue, *MachineException) {
    return failGetProperty(n,prop)
}
func (n *Number) SetProperty(prop string, val MachineValue) *MachineException {
    return failSetProperty(n,prop)
}


//
// The String type
//
type String struct {
	value string
}

func NewString(v string) MachineValue {
	return &String{v}
}
func (s *String) Type() MachineType {
    return TypeString
}
func (s *String) ToString() string {
	return s.value
}
func (s *String) ToNumber() float64 {
	num, err := strconv.ParseFloat(s.value, 64)
	if err != nil {
		return math.NaN()
	}
	return num
}
func (s *String) ToBoolean() bool {
	return s.value != ""
}

func (s *String) ToJSON() string {
	return `"` + s.ToString() + `"`
}
func (s *String) GetProperty(prop string) (MachineValue, *MachineException) {
    return failGetProperty(s,prop)
}
func (s *String) SetProperty(prop string, val MachineValue) *MachineException {
    return failSetProperty(s,prop)
}

//
// The Boolean type
//
type Boolean struct {
	value bool
}

func NewBoolean(b bool) MachineValue {
	return &Boolean{b}
}
func NewBooleanString(v string) MachineValue {
	if v == "true" {
		return NewBoolean(true)
	}
	if v != "false" {
		panic("Boolean string representation must be 'true' or 'false'")
	}
	return NewBoolean(false)
}
func (b *Boolean) Type() MachineType {
    return TypeBoolean
}
func (b *Boolean) ToString() string {
	if b.value {
		return "true"
	}
	return "false"
}
func (b *Boolean) ToNumber() float64 {
	if b.value {
		return 1.0
	}
	return 0.0
}
func (b *Boolean) ToBoolean() bool {
	return b.value
}
func (b *Boolean) ToJSON() string {
	return b.ToString()
}
func (b *Boolean) GetProperty(prop string) (MachineValue, *MachineException) {
    return failGetProperty(b,prop)
}
func (b *Boolean) SetProperty(prop string, val MachineValue) *MachineException {
    return failSetProperty(b,prop)
}

//
// The null type
//
type Null struct {}

var NullConst = &Null{}

func NewNull() MachineValue {
	return NullConst
}
func (u *Null) Type() MachineType {
    return TypeNull
}
func (u *Null) ToString() string {
	return ""
}
func (u *Null) ToNumber() float64 {
	return 0
}
func (u *Null) ToBoolean() bool {
	return false
}
func (u *Null) ToJSON() string {
	return u.ToString()
}
func (u *Null) GetProperty(prop string) (MachineValue, *MachineException) {
    return failGetProperty(u,prop)
}
func (u *Null) SetProperty(prop string, val MachineValue) *MachineException {
    return failSetProperty(u,prop)
}

//
// The object type
//

type Object struct {
    // not a SimpleType !
	class     string
	prototype MachineValue
	value     map[string]MachineValue
}

func CreateObjectWithPrototype(klass string, prototype MachineValue) Object {
	return Object{
		class:     klass,
		prototype: prototype,
		value:     make(map[string]MachineValue),
	}
}

var defaultObjectPrototype = CreateObjectWithPrototype("Object", nil)

func NewObject() *Object {
    o := CreateObjectWithPrototype("Object",&defaultObjectPrototype)
    return &o
}

func (o *Object) GetProperty(prop string) (MachineValue, *MachineException) {
    if val, ok := o.value[prop]; ok {
        return val, nil
    }
    if o.prototype!=nil {
        _, err := o.prototype.GetProperty(prop)
        if err!=nil {
            return failGetProperty(o,prop)
        }
    }
	return NullConst, nil
}

func (o *Object) SetProperty(prop string, val MachineValue) *MachineException {
    o.value[prop] = val
    return nil
}

func (o *Object) Type() MachineType {
	return TypeObject
}
func (o *Object) ToString() string {
	return "[Object " + o.class + "]"
}
func (o *Object) ToNumber() float64 {
	return math.NaN()
}
func (o *Object) ToBoolean() bool {
	return true
}
func (o *Object) ToJSON() string {
	s := "{"
	first := true
	for key, val := range o.value {
		if !first {
			s += ","
		} else {
			first = false
		}
		s += `"` + key + `":` + val.ToJSON()
	}
	s += "}"
	return s
}

//
// The array type
//

type Array struct {
	Object
	length uint32
}

func NewArray(init ...MachineValue) *Array {
	a := &Array{CreateObjectWithPrototype("Array",&defaultObjectPrototype), 0}
	a.Object.SetProperty("min", NewFunction("min",ArrayMin))
	a.Object.SetProperty("max", NewFunction("max",ArrayMax))
	for k, v := range init {
		a.SetProperty(strconv.Itoa(k), v)
	}
	return a
}

func (a *Array) GetProperty(prop string) (MachineValue, *MachineException) {
	if prop == "length" {
		return NewNumber(float64(a.length)), nil
	}
	return a.Object.GetProperty(prop)
}

func (a *Array) SetProperty(prop string, val MachineValue) *MachineException {
	if num, ok := IsUInt32(prop); ok {
		if num >= a.length {
			a.length = num + 1
		}
	}
	return a.Object.SetProperty(prop, val)
}

func (a *Array) ToString() string {
    var i uint32
	var s string

	for i = 0; i < a.length; i++ {
		if i > 0 {
			s += ","
		}
		val, err := a.GetUInt32Property(i)
        if err==nil {
            s += val.ToString()
        }
	}
	return s
}

func (a *Array) GetUInt32Property(prop uint32) (MachineValue, *MachineException) {
	return a.GetProperty(strconv.FormatUint(uint64(prop), 10))
}

func (a *Array) SetUInt32Property(prop uint32, val MachineValue) *MachineException {
	return a.SetProperty(strconv.FormatUint(uint64(prop), 10), val)
}

func (a *Array) Push(val MachineValue) {
	a.SetUInt32Property(a.length, val)
}

func (a *Array) Type() MachineType {
	return TypeArray
}
func (a *Array) ToJSON() string {
	var i uint32
	s := "["

	for i = 0; i < a.length; i++ {
		if i > 0 {
			s += ","
		}
		val, err := a.GetUInt32Property(i)
        if err!=nil {
            s += val.ToJSON()
        }
	}
	s += "]"
	return s
}

func (a *Array) Append(val MachineValue) {
	prop := strconv.FormatUint(uint64(a.length), 10)
	a.length++
	a.SetProperty(prop, val)
}

//
// The function type
//

type Function struct {
	Object
    name string
	call JSFunction
}

func NewFunction(name string, call JSFunction) *Function {
	return &Function{CreateObjectWithPrototype("Function",&defaultObjectPrototype), name, call}
}
func (f *Function) Type() MachineType {
	return TypeFunction
}
func (f *Function) ToString() string {
	return fmt.Sprintf("function %s(){ [Native code] }",f.name)
}
func (f *Function) ToNumber() float64 {
	return math.NaN()
}
func (f *Function) ToBoolean() bool {
	return true
}
func (f *Function) Call(m *Machine, paramcount int) (int, *MachineException) {
	return f.call(m, f, paramcount)
}
func (f *Function) ToJSON() string {
	return "null"
}

//
// Helper functions
//

func ToUInt32(val MachineValue) uint32 {
	n := val.ToNumber()
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return 0
	}
	return uint32(n)
}

func IsUInt32(s string) (uint32, bool) {
	r, e := strconv.ParseUint(s, 0, 32)
	return uint32(r), e == nil
}
