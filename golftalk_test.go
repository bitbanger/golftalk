package main

import "testing"

var emptyEnv *Env = NewEnv()

func evalExpectInt64(t *testing.T, expr string, expect int64, env *Env) {
	x, err := eval(expr, env)
	if err != "" {
		t.Error(expr, "gives error:", err)
		return
	}
	if x == nil {
		t.Error(expr, "gives nil want an int64")
		return
	}
	i, typeOk := x.(int64)
	if !typeOk {
		t.Errorf("%s gives %v, want an int64\n", expr, x)
		return
	}
	if i != expect {
		t.Errorf("%s gives %d, want %d\n", expr, i, expect)
		return
	}
}

func TestAddition(t *testing.T) {
	addEnv := NewEnv()
	initGlobalEnv(addEnv)

	evalExpectInt64(t,"(+ 5 12)", 17, addEnv)
	evalExpectInt64(t, "(+ 7 100 99)", 206, addEnv)
}
