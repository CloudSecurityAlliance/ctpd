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
	"math"
)

const (
	i_load_const = iota
	i_get_global
	i_get_index
	i_set_index
	i_array_append
	i_add
	i_sub
	i_mul
	i_div
	i_mod
	i_equ
	i_neq
	i_lt
	i_gt
	i_lte
	i_gte
	i_and
	i_or
	i_not
	i_neg
	i_call
	i_newarray
	i_newobject
	/**/
	I_COUNT
)

type Op struct {
	name   string
	length int
	exec   func(m *Machine) *MachineException
}

var Ops = [I_COUNT]Op{
	{"load_const", 4,
		func(m *Machine) *MachineException {
			m.Push(m.GetConst(m.GetIParam()))
			return nil
		}},
	{"get_global", 1,
		func(m *Machine) *MachineException {
			//key := m.Get(-1).ToString()
			//m.Pop(1)
			//r := m.context.GetProperty(key)
			//if r.Type()==TypeNull {
			//    return NewMachineException("ReferenceError: %s is not defined.",key)
			//}
			m.Push(m.context)
			return nil
		}},
	{"get_index", 1,
		func(m *Machine) *MachineException {
			objectref := m.Get(-2)
			key := m.Get(-1).ToString()
			m.Pop(2)
            r, err := objectref.GetProperty(key)
            if err!=nil {
                return err
            }
            m.Push(r)
            return nil
		}},
	{"set_index", 1,
		func(m *Machine) *MachineException {
			objectref := m.Get(-3)
			key := m.Get(-2).ToString()
			val := m.Get(-1)
			m.Pop(2)
		    return objectref.SetProperty(key, val)
		}},
	{"array_append", 1,
		func(m *Machine) *MachineException {
			arrayref := m.Get(-2)
			val := m.Get(-1)
			m.Pop(1)
			if array, ok := arrayref.(*Array); ok {
				array.Append(val)
				return nil
			}
			return NewMachineException("array_append called on something not a table")
		}},
	{"add", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-2)
			b := m.Get(-1)
			m.Pop(2)

			if a.Type() == TypeString || b.Type() == TypeString {
				m.Push(NewString(a.ToString() + b.ToString()))
			} else {
				m.Push(NewNumber(a.ToNumber() + b.ToNumber()))
			}
			return nil
		}},
	{"sub", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-2).ToNumber()
			b := m.Get(-1).ToNumber()
			m.Pop(2)
			m.Push(NewNumber(a - b))
			return nil
		}},
	{"mul", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-2).ToNumber()
			b := m.Get(-1).ToNumber()
			m.Pop(2)
			m.Push(NewNumber(a * b))
			return nil
		}},
	{"div", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-2).ToNumber()
			b := m.Get(-1).ToNumber()
			m.Pop(2)
			if b == 0 {
				if a == 0 {
					m.Push(NewNumber(math.NaN()))
				} else {
					m.Push(NewNumber(math.Inf(int(a))))
				}
			} else {
				m.Push(NewNumber(a / b))
			}
			return nil
		}},
	{"mod", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-2).ToNumber()
			b := m.Get(-1).ToNumber()
			m.Pop(2)
			m.Push(NewNumber(math.Mod(a, b)))
			return nil
		}},
	{"equ", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-2)
			b := m.Get(-1)
			m.Pop(2)
			m.Push(NewBoolean(equal(a, b)))
			return nil
		}},
	{"neq", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-2)
			b := m.Get(-1)
			m.Pop(2)
			m.Push(NewBoolean(!equal(a, b)))
			return nil
		}},
	{"lt", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-2)
			b := m.Get(-1)
			m.Pop(2)
			m.Push(NewBoolean(lessThan(a, b)))
			return nil
		}},
	{"gt", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-2)
			b := m.Get(-1)
			m.Pop(2)
			m.Push(NewBoolean(!equal(a, b) && !lessThan(a, b)))
			return nil
		}},
	{"lte", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-2)
			b := m.Get(-1)
			m.Pop(2)
			m.Push(NewBoolean(lessThan(a, b) || equal(a, b)))
			return nil
		}},
	{"gte", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-2)
			b := m.Get(-1)
			m.Pop(2)
			m.Push(NewBoolean(!lessThan(a, b)))
			return nil
		}},
	{"and", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-2)
			b := m.Get(-1)
			m.Pop(2)
			if a.ToBoolean() == false {
				m.Push(a)
			} else {
				m.Push(b)
			}
			return nil
		}},
	{"or", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-2)
			b := m.Get(-1)
			m.Pop(2)
			if a.ToBoolean() == true {
				m.Push(a)
			} else {
				m.Push(b)
			}
			return nil
		}},
	{"not", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-1).ToBoolean()
			m.Pop(1)
			m.Push(NewBoolean(!a))
			return nil
		}},
	{"neg", 1,
		func(m *Machine) *MachineException {
			a := m.Get(-1).ToNumber()
			m.Pop(1)
			m.Push(NewNumber(-a))
			return nil
		}},
	{"call", 4,
		func(m *Machine) *MachineException {
			objectref := m.Get(-2)
			key := m.Get(-1).ToString()
			paramcount := m.GetIParam()
			m.Pop(1)

            functionref, err := objectref.GetProperty(key)
            if err!=nil {
                return err
            }
            if functionref.Type() != TypeFunction {
				if function, ok := functionref.(*Function); ok {
					result_len, merror := function.Call(m, paramcount)
					if merror == nil {
						if result_len == 1 {
							result := m.Get(-1)
							m.Pop(paramcount + 1)
							m.Push(result)
						} else {
							m.Pop(paramcount)
						}
					}
					return merror
				}
			}
			return NewMachineException("TypeError: '%s' is not a function", key)
		}},
	{"newarray", 1,
		func(m *Machine) *MachineException {
			//paramcount := m.GetIParam()
			array := NewArray()
			//for i:=0; i<paramcount; i++ {
			//    array.SetProperty(strconv.Itoa(i),m.Get(-i-1))
			//}
			//m.Pop(paramcount)
			m.Push(array)
			return nil
		}},
	{"newobject", 1,
		func(m *Machine) *MachineException {
            object := NewObject()
            m.Push(object)
            return nil
		}},
}

func equalAsNumbers(a, b MachineValue) bool {
	an := a.ToNumber()
	bn := b.ToNumber()
	if math.IsNaN(an) || math.IsNaN(bn) {
		return false
	}
	if an == bn {
		return true
	}
	return false
}
func equal(a, b MachineValue) bool {
	if a.Type() == b.Type() {
		switch a.Type() {
		case TypeNull:
			return true
		case TypeNumber:
			return equalAsNumbers(a, b)
		case TypeString:
			return a.ToString() == b.ToString()
		case TypeBoolean:
			return a.ToBoolean() == b.ToBoolean()
		default:
			return a == b
		}
	}
	// Note: Simplified
	return equalAsNumbers(a, b)
}
func lessThanAsNumbers(a, b MachineValue) bool {
	an := a.ToNumber()
	bn := b.ToNumber()
	if math.IsNaN(an) || math.IsNaN(bn) {
		return false
	}
	if an < bn {
		return true
	}
	return false
}
func lessThan(a, b MachineValue) bool {
	if a.Type() == b.Type() {
		switch a.Type() {
		case TypeNumber:
			return lessThanAsNumbers(a, b)
		case TypeString:
			return a.ToString() < b.ToString()
		case TypeBoolean:
			return (a.ToBoolean() == false) && (b.ToBoolean() == true)
		default:
			return false
		}
	}
	// Note: Simplified
	return lessThanAsNumbers(a, b)
}

/*
func (op loadOp) exec(m *Machine) *MachineException {
	m.Push(m.Get(int(op)))
	m.IncPc()
	return nil
}
func (op loadOp) String() string {
	return fmt.Sprintf("load %d", op)
}
*/
/*
type newobjectOp int
func (op newOp)exec(m* Machine) *MachineException {
    a := NewObject(ArrayPrototype)
    for i:=0;i<int(op);i++ {
        x:=m.Get(-int(op)+i)
        PutIndex(a,uint32(i),x)
    }
    m.Pop(int(op))
    m.Push(a)
    m.IncPc()
    return nil
}
func (op narrayOp)String() string {
    return fmt.Sprintf("newarray %d",op)
}
*/
