package jsmm

import (
	"testing"
)

func Expect(t *testing.T, val MachineValue, expr string) {
	if val != nil {
		result := val.ToString()
		if result != expr {
			t.Error("Expected " + expr +", but got " + result)
		}
	} else {
		t.Error("Got nil result")
	}
}

func CompileAndRun(t *testing.T, expr string, extra interface{}) MachineValue {
	machine, err := Compile(expr)
	if err != nil {
		t.Error("Compile failed", err.Error())
		return nil
	}
	if extra != nil {
		err = ImportGlobal(machine, "test", extra)
		if err != nil {
			t.Error("Failed to import gloval 'test' variable")
			return nil
		}
	}
	v, exception := machine.Execute()
	if exception != nil {
		t.Error("Execute failed", exception.Error())
		return nil
	}
	if v == nil {
		t.Error("Return value is nil")
		return nil
	}
	return v
}

func TestSimple(t *testing.T) {
	v := CompileAndRun(t, "0 + 1 * 2 * (3 - 4) / -5 + 6", nil)

	Expect(t, v, "6.4")
}

func TestSimple2(t *testing.T) {
	v := CompileAndRun(t, "test > 2 && test < 4", 3)

	Expect(t, v, "true")
}

func TestSimple3(t *testing.T) {
	v := CompileAndRun(t, "[1,2,3][1]+[4,5,6][2]", nil)

	Expect(t, v, "8")
}

func TestSimple4(t *testing.T) {
	v := CompileAndRun(t, `{a: "b", "c": [1,8]}.c[0]+[1,2,3].length`, nil)

	Expect(t, v, "4")
}

func TestSimple5(t *testing.T) {
	v := CompileAndRun(t, `toString(1<2)+"ly"`, nil)

	Expect(t, v, "truely")
}

func TestSimple6(t *testing.T) {
	v := CompileAndRun(t, `[1,2,3].min()+[7,6,5,4].max()`, nil)

	Expect(t, v, "8")
}

func TestVar(t *testing.T) {
	val := [4]float64{1, 2, 3, 4}

	v := CompileAndRun(t, "test[0]+test[1]+test[2]", val)

	Expect(t, v, "6")
}

func TestVar2(t *testing.T) {
	val := struct{ A [4]float64 }{[4]float64{1, 2, 3, 4}}

	v := CompileAndRun(t, "test.A[0]+test.A[1]+test.A[2]", val)

	Expect(t, v, "6")
}

func TestTime1(t *testing.T) {
	v := CompileAndRun(t, `timeUTC("1970-01-01T00:00:01Z")==1 && timeUTC("2014-09-02T12:17:00Z")==1409660220 && timeUTC("now")>1443428707`, nil)

	Expect(t, v, "true")
}

func TestTime2(t *testing.T) {
	v := CompileAndRun(t, `timeUTC("1969-12-31T00:00:00Z")`, nil)

	Expect(t, v, "-86400")
}

func TestSelect1(t *testing.T) {
    v := CompileAndRun(t, `toString(select("country",[{"country":"UK", "color":"blue"},{color: "green"},{country:"FR","color":"yellow"},{country:"UK"}]))`, nil)

	Expect(t, v, `UK,,FR,UK`)
}

func TestMatchRegexp1(t *testing.T) {
    v := CompileAndRun(t, `matchRegexp("a(x+|y+)","zaxxxxon")`, nil)

    Expect(t, v, "true")
}

func TestMatchRegexp2(t *testing.T) {
    v := CompileAndRun(t, `matchRegexp("a(x+|y+)","zapyoon")`, nil)

    Expect(t, v, "false")
}

func TestMatchRegexp3(t *testing.T) {
    v := CompileAndRun(t, `matchRegexp("a(x+|y+)","")`, nil)

    Expect(t, v, "false")
}

func TestMatchRegexp4(t *testing.T) {
    v := CompileAndRun(t, `matchRegexp("#[-_a-zA-Z]+",["there is", "a #tag here", "but not here"])`, nil)

    Expect(t, v, "false")
}

func TestMatchRegexp5(t *testing.T) {
    v := CompileAndRun(t, `matchRegexp("#[-_a-zA-Z]+",["there #is", "a #tag here", "and #here"])`, nil)

    Expect(t, v, "true")
}

func TestToString(t *testing.T) {
    v := CompileAndRun(t, `toString([toString(matchRegexp), toString(true), toString(3.1415), toString(null)])`, nil)

    Expect(t, v, "function matchRegex(){ [Native code] },true,3.1415,")
}
 
