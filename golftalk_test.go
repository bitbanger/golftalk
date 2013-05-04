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

	evalExpectInt64(t,"(+ -5 12)", 7, addEnv)
	evalExpectInt64(t, "(+ 7 100 99)", 206, addEnv)
	evalExpectInt64(t, "(+ (+ 1 2) (+ 3 4))", 10, addEnv)
	evalExpectInt64(t, "(+ 1)", 1, addEnv)
	evalExpectInt64(t, "(+)", 0, addEnv)
	evalExpectError(t, "(+ hi there)", "Invalid types to add. Must be int and int.", addEnv)
}

func TestSubtraction(t *testing.T) {
	env := NewEnv()
	initGlobalEnv(env)

	evalExpectInt64(t,"(- 23 11)", 12, env)
	evalExpectInt64(t, "(- 55 90 22)", -57, env)
	evalExpectInt64(t, "(- (- 1 2) (- 3 4))", 0, env)
	evalExpectInt64(t,"(- 5 )", -5, env)
	evalExpectError(t,"(-)", "Need at least 1 int to subtract.", env)
	evalExpectError(t, "(- go away)", "Invalid types to subtract. Must be int and int.", env)
}
