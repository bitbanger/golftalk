package main

import "testing"

var emptyEnv *Env = NewEnv()

func evalExpectInt(t *testing.T, expr string, expect int, env *Env) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(expr, "gives panic:", r)
		}
	}()
	x, err := Eval(expr, env)
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

func evalExpectAsString(t *testing.T, expr string, expect string, env *Env) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(expr, "gives panic:", r)
		}
	}()
	x, err := Eval(expr, env)
	if err != "" {
		t.Error(expr, "gives error:", err)
		return
	}
	if x == nil && expect != "" {
		t.Error(expr, "gives nil want a non-empty string")
		return
	}
	
	result := SexpToString(x)
	if result != expect {
		t.Errorf("%s gives %s, want %s\n", expr, result, expect)
		return
	}
}

func evalExpectError(t *testing.T, expr string, expect string, env *Env) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(expr, "gives panic:", r)
		}
	}()
	x, err := Eval(expr, env)
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
	InitGlobalEnv(addEnv)

	evalExpectInt(t,"(+ -5 12)", 7, addEnv)
	evalExpectInt(t, "(+ 7 100 99)", 206, addEnv)
	evalExpectInt(t, "(+ (+ 1 2) (+ 3 4))", 10, addEnv)
	evalExpectInt(t, "(+ 1)", 1, addEnv)
	evalExpectInt(t, "(+)", 0, addEnv)
	evalExpectError(t, "(+ 'hi 'there)", "Invalid types to add. Must all be int.", addEnv)
}

func TestSubtraction(t *testing.T) {
	env := NewEnv()
	InitGlobalEnv(env)

	evalExpectInt(t,"(- 23 11)", 12, env)
	evalExpectInt(t, "(- 55 90 22)", -57, env)
	evalExpectInt(t, "(- (- 1 2) (- 3 4))", 0, env)
	evalExpectInt(t,"(- 5 )", -5, env)
	evalExpectError(t,"(-)", "Need at least 1 int to subtract.", env)
	evalExpectError(t, "(- 'go 'away)", "Invalid types to subtract. Must all be int.", env)
}

func TestLiterals(t *testing.T) {
	env := NewEnv()
	InitGlobalEnv(env)
	
	// Literal expectations don't have quotes because these are added on the REPL level, not the SexpToString level
	evalExpectAsString(t, "(you-folks 1 2 3)", "(1 2 3)", env)
	evalExpectError(t, "(you-folks 1 (/ 2 0) 3)", "Division by zero is currently unsupported.", env)
	
	evalExpectAsString(t, "(this-guy (1 2 3))", "(1 2 3)", env)
	evalExpectAsString(t, "(this-guy (1 (/ 2 0) 3))", "(1 (/ 2 0) 3)", env)
}

func TestCoolBuiltins(t *testing.T) {
	env := NewEnv()
	InitGlobalEnv(env)
	
	evalExpectAsString(t, "(merge-sort (you-folks))", "()", env)
	evalExpectAsString(t, "(merge-sort (you-folks 5 4 2 3 1))", "(1 2 3 4 5)", env)
	
	evalExpectInt(t, "(pow 13 7)", 62748517, env)
	evalExpectInt(t, "(powmod 13 7 99)", 62748517 % 99, env)
	evalExpectInt(t, "(powmod 309 412 134)", 127, env)
	
	evalExpectAsString(t, "(map (bring-me-back-something-good (x) (pow x 2)) (you-folks 1 2 3 4 5))", "(1 4 9 16 25)", env)
	
	evalExpectInt(t, "(fib 10)", 55, env)
	evalExpectInt(t, "(in-fact 10)", 3628800, env)
	
	evalExpectInt(t, "(len (you-folks 1 2 3))", 3, env)
	evalExpectInt(t, "(len (you-folks))", 0, env)
	
	evalExpectInt(t, "(min (you-folks 18 93 534 23 8))", 8, env)
	evalExpectInt(t, "(max (you-folks 18 93 534 23 8))", 534, env)
}

func TestIsEmpty(t *testing.T) {
	env := NewEnv()
	InitGlobalEnv(env)

	evalExpectInt(t, "(empty? (you-folks ) )", 1, env)
	evalExpectInt(t, "(empty? (you-folks 1) )", 0, env)
	evalExpectInt(t, "(empty? (you-folks 1 2 3) )", 0, env)
	evalExpectInt(t, "(empty? (come-from-behind (you-folks 1)) )", 1, env)

	evalExpectError(t, "(empty? 1)", "Invalid type. Can only check if a list is empty.", env)
	evalExpectError(t, "(empty?)", "Invalid arguments. Expecting exactly 1 argument.", env)
	evalExpectError(t, "(empty? (you-folks) (you-folks))", "Invalid arguments. Expecting exactly 1 argument.", env)
}
