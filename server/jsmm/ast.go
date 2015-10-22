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
)

type Expression interface {
	String() string
	Compile(m *Machine)
}

//
// Literal number
//
type literalNumberExpr struct {
	value string
}

func (e *literalNumberExpr) Compile(m *Machine) {
	m.AddIParam(i_load_const, m.AddConst(NewNumberString(e.value)))
}
func (l *literalNumberExpr) String() string {
	return l.value
}

//
// Literal string
//
type literalStringExpr struct {
	value string
}

func (e *literalStringExpr) Compile(m *Machine) {
	m.AddIParam(i_load_const, m.AddConst(NewString(e.value)))
}
func (e *literalStringExpr) String() string {
	return fmt.Sprintf("\"%s\"", e.value)
}

//
// Literal boolean
//
type literalBooleanExpr struct {
	value string
}

func (e *literalBooleanExpr) Compile(m *Machine) {
	m.AddIParam(i_load_const, m.AddConst(NewBooleanString(e.value)))
}
func (e *literalBooleanExpr) String() string {
	return e.value
}

//
// Literal null
//
type literalNullExpr struct  {}

func (n *literalNullExpr) Compile(m *Machine) {
	m.AddIParam(i_load_const, m.AddConst(NewNull()))
}
func (n *literalNullExpr) String() string {
	return "null"
}

//
// List def expression
//

type ExpressionList []Expression

func NewExpressionList() ExpressionList {
	return make([]Expression, 0, 4)
}
func (e ExpressionList) Length() int {
    return len(e)
}
func (e ExpressionList) Append(i Expression) ExpressionList {
	return append(e, i)
}
func (e ExpressionList) Get(i int) Expression {
    return e[i]
}

func (e ExpressionList) String() string {
	r := ""
	for _, item := range e {
		r += " " + item.String()
	}
	return r
}

//
// Function call
//

type functionCallExpr struct {
	base      Expression
	fname     Expression
	arguments ExpressionList
}

func (e *functionCallExpr) Compile(m *Machine) {
    for i:=e.arguments.Length()-1; i>=0; i-- {
        e.arguments.Get(i).Compile(m)
    }
	e.base.Compile(m)
	e.fname.Compile(m)
	m.AddIParam(i_call, e.arguments.Length()+1)
}

func (e *functionCallExpr) String() string {
	return "(fcall " + e.base.String() + "." + e.fname.String() + " " + e.arguments.String() + ")"
}

//
// Attribute selection
//

type attributeSelectionExpr struct {
	base      Expression
	selection Expression
}

func (e *attributeSelectionExpr) Compile(m *Machine) {
	e.base.Compile(m)
	e.selection.Compile(m)
	m.AddI(i_get_index)
}

func (e *attributeSelectionExpr) String() string {
	return "(. " + e.base.String() + " " + e.selection.String() + ")"
}

//
// Attribute base
//

type getGlobalObjectExpr struct{}

func (e *getGlobalObjectExpr) Compile(m *Machine) {
	m.AddI(i_get_global)
}

func (e *getGlobalObjectExpr) String() string {
	return "(load_global)"
}

//
// Array def expression
//

type arrayDefExpr struct {
	value ExpressionList
}

func (e *arrayDefExpr) Compile(m *Machine) {
	m.AddI(i_newarray)
    for i:=0; i<e.value.Length(); i++ {
        e.value.Get(i).Compile(m)
    }
}
func (e *arrayDefExpr) String() string {
	return "(newaray " + e.value.String() + ")"
}

//
// Object def expression
//

type objectDefExpr struct {
	value ExpressionList
}

func (e *objectDefExpr) Compile(m *Machine) {
	m.AddI(i_newobject)
    for i:=0; i<e.value.Length(); i++ {
        e.value.Get(i).Compile(m)
    }
}
func (e *objectDefExpr) String() string {
	return "(newobject " + e.value.String() + ")"
}


/*
type identifierExpr struct {
    value string
}
func (e *identifierExpr)String() string {
    return e.value
}

type sliceExpr struct {
    source  Expression
    start   Expression
    stop    Expression
}
func (e *sliceExpr)String() string {
    return fmt.Sprintf("([:] %s %s %s)",e.source.String(),e.start.String(),e.stop.String())
}

type callExpr struct {
    funcName string
    params []Expression
}
func (e *callExpr)String() string {
    s := "(" + e.funcName
    for _,v := range e.params {
        s+=" " + v.String()
    }
    s+=")"
    return s
}
func newCallExpr(fname string) *callExpr {
    return &callExpr{fname,make([]Expression,0,2)}
}
func (e *callExpr)addParam(p Expression) *callExpr {
    e.params = append(e.params,p)
    return e
}
*/

//
// Binary expression (+,-,*,/,&,<,>,<=,>=,==,!=,...)
//

type binExpr struct {
	op int
	x  Expression
	y  Expression
}

func (e *binExpr) String() string {
	return fmt.Sprintf("(%s %s %s)", Ops[e.op].name, e.x.String(), e.y.String())
}
func (e *binExpr) Compile(m *Machine) {
	e.x.Compile(m)
	e.y.Compile(m)
	m.AddI(byte(e.op))
}

//
// Unary op expression
//

type unaryExpr struct {
	op int
	x  Expression
}

func (e *unaryExpr) String() string {
	return fmt.Sprintf("(%s %s)", Ops[e.op].name, e.x.String())
}
func (e *unaryExpr) Compile(m *Machine) {
	e.x.Compile(m)
	m.AddI(byte(e.op))
}

/*
type symbolExpr struct {
    identifier string
    base bool
}
func (e *symbolExpr)String() string {
    return e.identifier
}
func (e *symbolExpr)Compile(m *Machine) {
    m.AddI(loadcOp(m.AddConst(NewString(e.identifier))))
    if e.base {
        m.AddI(getBaseOp(1))
    } else {
        m.AddI(getBaseOp(0))
    }
}
*/
