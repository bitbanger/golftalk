package main

import "testing"

var emptyEnv *Env = NewEnv()

func evalExpectInt(t *testing.T, expr string, expect int, env *Env) {
	x, err := eval(expr, env)
	if err != "" {
		t.Error(expr, "gives error:", err)
		return
	}
	if x == nil {
		t.Error(expr, "gives nil want an int")
		return
	}
	i, typeOk := x.(int)
	if !typeOk {
		t.Errorf("%s gives %v, want an int\n", expr, x)
		return
	}
	if i != expect {
		t.Errorf("%s gives %d, want %d\n", expr, i, expect)
		return
	}
}

func evalExpectError(t *testing.T, expr string, expect string, env *Env) {
	x, err := eval(expr, env)
	if err == "" {
		t.Errorf("%s gives %v, want error: %s\n", expr, x, expect)
		return
	}
	if err != expect {
		t.Errorf("%s gives error: %s, want error: %s\n", expr, err, expect)
		return
	}
}

func TestAddition(t *testing.T) {
	addEnv := NewEnv()
	initGlobalEnv(addEnv)

	evalExpectInt(t,"(+ -5 12)", 7, addEnv)
	evalExpectInt(t, "(+ 7 100 99)", 206, addEnv)
	evalExpectInt(t, "(+ (+ 1 2) (+ 3 4))", 10, addEnv)
	evalExpectInt(t, "(+ 1)", 1, addEnv)
	evalExpectInt(t, "(+)", 0, addEnv)
	evalExpectError(t, "(+ hi there)", "Invalid types to add. Must all be int.", addEnv)
}

func TestSubtraction(t *testing.T) {
	env := NewEnv()
	initGlobalEnv(env)

	evalExpectInt(t,"(- 23 11)", 12, env)
	evalExpectInt(t, "(- 55 90 22)", -57, env)
	evalExpectInt(t, "(- (- 1 2) (- 3 4))", 0, env)
	evalExpectInt(t,"(- 5 )", -5, env)
	evalExpectError(t,"(-)", "Need at least 1 int to subtract.", env)
	evalExpectError(t, "(- go away)", "Invalid types to subtract. Must be int and int.", env)
}
