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

type Typer interface {
    Type() MachineType
}

type MachineValue interface {
	Typer

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

func TypeOf(m Typer) string {
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
 * SimpleType is used as base type to describe the default
 * behaviour of GetProperty and SetProperty
 */

type SimpleType MachineType

func (dummy *SimpleType) Type() MachineType {
    return MachineType(*dummy)
}

func (dummy *SimpleType) GetProperty(prop string) (MachineValue, *MachineException) {
    return NewNull(), NewMachineException("BadType: cannot read property %s of %s", prop, TypeOf(dummy))
}

func (dummy *SimpleType) SetProperty(prop string, val MachineValue) *MachineException {
	return NewMachineException("BadType: cannot set property %s of %s", prop, TypeOf(dummy))
}

//
// The Number type
//
type Number struct {
	SimpleType
	value float64
}

func NewNumber(f float64) MachineValue {
	return &Number{SimpleType(TypeNumber), f}
}
func NewNumberString(v string) MachineValue {
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		panic("syntax error in number " + v)
	}
	return NewNumber(f)
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

//
// The String type
//
type String struct {
	SimpleType
	value string
}

func NewString(v string) MachineValue {
	return &String{SimpleType(TypeString), v}
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

//
// The Boolean type
//
type Boolean struct {
	SimpleType
	value bool
}

func NewBoolean(b bool) MachineValue {
	return &Boolean{SimpleType(TypeBoolean), b}
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

//
// The null type
//
type Null struct {
	SimpleType
}

var NullConst = &Null{SimpleType(TypeNull)}

func NewNull() MachineValue {
	return NullConst
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

//
// The object type
//

type Object struct {
    // not a SimpleType !
	class     string
	prototype *Object
	value     map[string]MachineValue
}

func NewObjectWithPrototype(prototype *Object) *Object {
	klass := "Undefined"
	if prototype != nil {
		klass = prototype.class
	}
	return &Object{
		class:     klass,
		prototype: prototype,
		value:     make(map[string]MachineValue),
	}
}

var ObjectPrototype = NewObjectWithPrototype(nil)

func NewObject() *Object {
	return NewObjectWithPrototype(ObjectPrototype)
}

func (o *Object) GetProperty(prop string) (MachineValue, *MachineException) {
    if val, ok := o.value[prop]; ok {
        return val, nil
    }
    if o.prototype!=nil {
        return o.prototype.GetProperty(prop)
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
	a := &Array{*ObjectPrototype, 0}
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
	SimpleType
    name string
	call JSFunction
}

func NewFunction(name string, call JSFunction) *Function {
	return &Function{SimpleType(TypeFunction), name, call}
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
