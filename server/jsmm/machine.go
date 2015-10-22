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
	"io"
	"log"
)

type MachineException struct {
	msg string
}

func NewMachineException(format string, args ...interface{}) *MachineException {
	return &MachineException{fmt.Sprintf(format, args...)}
}

func (e *MachineException) Error() string {
	return e.msg
}

type Machine struct {
	constants           []MachineValue
	constantStringTable map[string]int
	stack               []MachineValue
	code                []byte
	pc                  int
	context             Object
	debug_mode          bool
}

func NewMachine() *Machine {
	m := &Machine{
		constants:           make([]MachineValue, 0, 10),
		constantStringTable: make(map[string]int),
		stack:               make([]MachineValue, 0, 10),
		code:                make([]byte, 0, 10),
		pc:                  -1,
		context:             CreateObjectWithPrototype("GlobalObject", NewNull()),
		debug_mode:          false,
	}
	m.context.SetProperty("toString", NewFunction("toString", ToString))
	m.context.SetProperty("toBoolean", NewFunction("toBoolean", ToString))
	m.context.SetProperty("toNumber", NewFunction("toNumber", ToString))
	m.context.SetProperty("timeUTC", NewFunction("timeUTC", TimeUTC))
	m.context.SetProperty("matchRegexp", NewFunction("matchRegex", MatchRegexp))
	m.context.SetProperty("select", NewFunction("select", Select))
	return m
}

func Compile(expr string) (*Machine, error) {
	ast, err := Parse(expr)
	if err != nil {
		return nil, err
	}
	m := NewMachine()
	ast.Compile(m)
	return m, nil
}

func (m *Machine) DebugMode(debug bool) {
	m.debug_mode = debug
}

func (m *Machine) AddConst(c MachineValue) int {
	if c.Type() == TypeString {
		str := c.ToString()
		if pos, ok := m.constantStringTable[str]; ok {
			return pos
		}
		m.constantStringTable[str] = len(m.constants)
	}
	m.constants = append(m.constants, c)
	return len(m.constants) - 1
}

func (m *Machine) GetConst(pos int) MachineValue {
	return m.constants[pos]
}

func (m *Machine) Push(e MachineValue) int {
	if e == nil {
		panic("Pushing a nil on the stack")
	}
	m.stack = append(m.stack, e)
	return m.Top()
}

func (m *Machine) Pop(count int) int {
	m.stack = m.stack[:len(m.stack)-count]
	return m.Top()
}

func (m *Machine) Get(pos int) MachineValue {
	if pos < 0 {
		pos = len(m.stack) + pos
	}
	if pos < 0 {
		panic(fmt.Sprintf("Unknown index %d in machine stack (pc=%d)", pos, m.pc))
	}
	return m.stack[pos]
}

func (m *Machine) Top() int {
	return len(m.stack) - 1
}

func (m *Machine) AddI(i byte) *Machine {
	m.code = append(m.code, i)
	return m
}

func (m *Machine) AddIParam(i byte, p int) *Machine {
	m.code = append(m.code, i)
	m.code = append(m.code, byte((p>>16)&0xff))
	m.code = append(m.code, byte((p>>8)&0xff))
	m.code = append(m.code, byte(p&0xff))
	return m
}

func (m *Machine) GetIParamAt(index int) int {
	var compl uint32 = 0xFF000000
	var param uint32 = (uint32(m.code[index+1]) << 16) | (uint32(m.code[index+2]) << 8) | uint32(m.code[index+3])
	if (param & 0x800000) != 0 {
		param |= compl
	}
	return int(param)
}

func (m *Machine) GetIParam() int {
	return m.GetIParamAt(m.pc)
}

func (m *Machine) Call(pc int) *MachineException {
	save_pc := m.pc
	m.pc = pc
	for m.pc < len(m.code) {
		op := m.code[m.pc]
		if m.debug_mode {
			log.Printf("pc=%d, st=%d, opcode=%d, opname=%s\n", m.pc, m.Top(), op, Ops[op].name)
		}
		if err := Ops[op].exec(m); err != nil {
			return err
		}
		if m.debug_mode {
			if m.Top() >= 0 {
				top := m.Get(m.Top())
				log.Printf("st -> %s: %s", TypeOf(top), top.ToString())
			}
		}
		m.pc += Ops[op].length
	}
	if m.pc >= len(m.code) {
		m.pc = -1
	}
	m.pc = save_pc
	return nil
}

func (m *Machine) Execute() (MachineValue, *MachineException) {
	if err := m.Call(0); err != nil {
		return nil, err
	}
	if m.Top() >= 0 {
		return m.Get(m.Top()), nil
	}
	return nil, nil
}

func (m *Machine) GlobalObject() *Object {
	return &m.context
}

//
// Auxiliairy functions
//

func DumpStack(w io.Writer, m *Machine) {
	fmt.Fprintf(w, "stack:\n")
	for i := m.Top(); i >= 0; i-- {
		if m.Get(i).Type() == TypeString {
			fmt.Fprintf(w, "%4d: \"%s\"\n", i, m.Get(i).ToString())
		} else {
			fmt.Fprintf(w, "%4d: %s\n", i, m.Get(i).ToString())
		}
	}
}
func DumpCode(w io.Writer, m *Machine) {
	fmt.Fprintf(w, "constants:\n")
	for i := 0; i < len(m.constants); i++ {
		if m.GetConst(i).Type() == TypeString {
			fmt.Fprintf(w, "%4d: \"%s\"\n", i, m.GetConst(i).ToString())
		} else {
			fmt.Fprintf(w, "%4d: %s\n", i, m.GetConst(i).ToString())
		}
	}
	fmt.Println("code:")
	i := 0
	for i < len(m.code) {
		op := m.code[i]
		// FIXME: check code index
		if Ops[op].length == 1 {
			fmt.Printf("%4d: %02x             %s\n", i, op, Ops[op].name)
			i++
		} else {
			fmt.Printf("%4d: %02x %02x %02x %02x    %s %d\n", i, op, m.code[i+1], m.code[i+2], m.code[i+3], Ops[op].name, m.GetIParamAt(i))
			i += 4
		}
	}
}
