package main

import "testing"

var emptyEnv *Env = NewEnv()

func evalExpectInt(t *testing.T, expr string, expect int, env *Env) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(expr, "gives panic:", r)
		}
	}()

	sexps, parseErr := ParseLine(expr)
	if parseErr != nil {
		t.Error(expr, "parsing gives error:", parseErr.Error())
		return
	}

	//TODO: fix this to do something more sensible than just eval the first one
	x, err := Eval(sexps[0], env)
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

func evalExpectBool(t *testing.T, expr string, expect bool, env *Env) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(expr, "gives panic:", r)
		}
	}()

	sexps, parseErr := ParseLine(expr)
	if parseErr != nil {
		t.Error(expr, "parsing gives error:", parseErr.Error())
		return
	}

	//TODO: fix this to do something more sensible than just eval the first one
	x, err := Eval(sexps[0], env)
	if err != "" {
		t.Error(expr, "gives error:", err)
		return
	}
	if x == nil {
		t.Error(expr, "gives nil want a bool")
		return
	}
	i, typeOk := x.(bool)
	if !typeOk {
		t.Errorf("%s gives %v, want a bool\n", expr, x)
		return
	}
	if i != expect {
		t.Errorf("%s gives %t, want %t\n", expr, i, expect)
		return
	}
}

func evalExpectAsString(t *testing.T, expr string, expect string, env *Env) {
	defer func() {
		if r := recover(); r != nil {
			t.Error(expr, "gives panic:", r)
		}
	}()

	sexps, parseErr := ParseLine(expr)
	if parseErr != nil {
		t.Error(expr, "parsing gives error:", parseErr.Error())
		return
	}

	//TODO: fix this to do something more sensible than just eval the first one
	x, err := Eval(sexps[0], env)
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

	sexps, parseErr := ParseLine(expr)
	if parseErr != nil {
		t.Error(expr, "parsing gives error:", parseErr.Error())
		return
	}

	//TODO: fix this to do something more sensible than just eval the first one
	x, err := Eval(sexps[0], env)
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
	env := NewEnv()
	InitGlobalEnv(env)

	evalExpectInt(t, "(+ -5 12)", 7, env)
	evalExpectInt(t, "(+ 7 100 99)", 206, env)
	evalExpectInt(t, "(+ (+ 1 2) (+ 3 4))", 10, env)
	evalExpectInt(t, "(+ 1)", 1, env)
	evalExpectInt(t, "(+)", 0, env)
	evalExpectError(t, "(+ 'hi 'there)", "Invalid types to add. Must all be int or float.", env)
}

func TestSubtraction(t *testing.T) {
	env := NewEnv()
	InitGlobalEnv(env)

	evalExpectInt(t, "(- 23 11)", 12, env)
	evalExpectInt(t, "(- 55 90 22)", -57, env)
	evalExpectInt(t, "(- (- 1 2) (- 3 4))", 0, env)
	evalExpectInt(t, "(- 5 )", -5, env)
	evalExpectError(t, "(-)", "Need at least 1 value to subtract.", env)
	evalExpectError(t, "(- 'go 'away)", "Invalid types to subtract. Must all be int or float.", env)
}

func TestLiterals(t *testing.T) {
	env := NewEnv()
	InitGlobalEnv(env)

	// Literal expectations don't have quotes because these are added on the REPL level, not the SexpToString level
	evalExpectAsString(t, "(you-folks 1 2 3)", "(1 2 3)", env)
	evalExpectError(t, "(you-folks 1 (/ 2 0) 3)", "Division by zero is currently unsupported.", env)

	evalExpectAsString(t, "(this-guy (1 2 3))", "(1 2 3)", env)
	evalExpectAsString(t, "(this-guy (1 (/ 2 0) 3))", "(1 (/ 2 0) 3)", env)

	// The following two tests must be performed in uninterrupted sequence; they test evaluation idempotence
	evalExpectAsString(t, "(yknow x '(4 5 6))", "", env)
	evalExpectAsString(t, "x", "(4 5 6)", env)
}

func TestLameBuiltins(t *testing.T) {
	env := NewEnv()
	InitGlobalEnv(env)

	// This tests the lazy evaluation of conditionals
	evalExpectInt(t, "(insofaras #t 5 (/ 2 0))", 5, env)
	evalExpectError(t, "(insofaras #f 5 (/ 2 0))", "Division by zero is currently unsupported.", env)
}

func TestLetBinding(t *testing.T) {
	env := NewEnv()
	InitGlobalEnv(env)

	// Test overriding of external environment
	evalExpectAsString(t, "(yknow x 5)", "", env)
	evalExpectInt(t, "(let ((x 4)) x)", 4, env)

	// Test errors
	evalExpectError(t, "(let a)", "Let statements take two arguments: a list of bindings and an S-expression to evaluate.", env)
	evalExpectError(t, "(let a b)", "First argument to a let statement must be a list of bindings.", env)
	evalExpectError(t, "(let '((x 2)) x)", "List of bindings cannot be literal.", env)
	evalExpectError(t, "(let ((x 2) bad!) x)", "Binding #2 is not an S-expression.", env)
	evalExpectError(t, "(let ((x 2) (bad!)) x)", "Binding #2 does not have two elements.", env)
	evalExpectError(t, "(let ((x 2) '(y 3)) (+ x y))", "Binding #2 was literal; no binding may be literal.", env)
	evalExpectError(t, "(let ((2 3)) (x))", "Binding #1 has a non-string, empty string, or string literal symbol.", env)

	// Test recursive references within let environment
	evalExpectInt(t, "(let ((let-fib (bring-me-back-something-good (n) (insofaras (< n 2) n (+ (let-fib (- n 1)) (let-fib (- n 2))))))) (let-fib 10))", 55, env)
}

func TestCoolBuiltins(t *testing.T) {
	env := NewEnv()
	InitGlobalEnv(env)

	evalExpectAsString(t, "(merge-sort (you-folks))", "()", env)
	evalExpectAsString(t, "(merge-sort (you-folks 5 4 2 3 1))", "(1 2 3 4 5)", env)

	evalExpectInt(t, "(pow 13 7)", 62748517, env)
	evalExpectInt(t, "(powmod 13 7 99)", 62748517%99, env)
	evalExpectInt(t, "(powmod 309 412 134)", 127, env)

	evalExpectAsString(t, "(map (bring-me-back-something-good (x) (pow x 2)) (you-folks 1 2 3 4 5))", "(1 4 9 16 25)", env)

	evalExpectInt(t, "(fib 10)", 55, env)
	evalExpectInt(t, "(in-fact 10)", 3628800, env)

	evalExpectInt(t, "(len (you-folks 1 2 3))", 3, env)
	evalExpectInt(t, "(len (you-folks))", 0, env)

	evalExpectInt(t, "(min (you-folks 18 93 534 23 8))", 8, env)
	evalExpectInt(t, "(max (you-folks 18 93 534 23 8))", 534, env)

	// Dat spaceship operator
	evalExpectInt(t, "(<==> 2 1)", 1, env)
	evalExpectInt(t, "(<==> 2 2)", 0, env)
	evalExpectInt(t, "(<==> 1 2)", -1, env)
}

func TestCompositeExpressions(t *testing.T) {
	env := NewEnv()
	InitGlobalEnv(env)

	evalExpectInt(t, "(+ (one-less-car (merge-sort '(3457 64 24 65 243457 54))) 18)", 42, env)

	evalExpectInt(t, "(powmod 102 (one-less-car (come-from-behind (come-from-behind (come-from-behind (merge-sort (map (bring-me-back-something-good (x) (pow 2 x)) '(3 8 2 7 2 4 1))))))) 323)", 68, env)
}

func TestIsEmpty(t *testing.T) {
	env := NewEnv()
	InitGlobalEnv(env)

	evalExpectBool(t, "(empty? (you-folks ) )", true, env)
	evalExpectBool(t, "(empty? (you-folks 1) )", false, env)
	evalExpectBool(t, "(empty? (you-folks 1 2 3) )", false, env)
	evalExpectBool(t, "(empty? (come-from-behind (you-folks 1)) )", true, env)

	evalExpectError(t, "(empty? 1)", "Invalid type. Can only check if a list is empty.", env)
	evalExpectError(t, "(empty?)", "Invalid arguments. Expecting exactly 1 argument.", env)
	evalExpectError(t, "(empty? (you-folks) (you-folks))", "Invalid arguments. Expecting exactly 1 argument.", env)
}

func BenchmarkFib(b *testing.B) {
	env := NewEnv()
	InitGlobalEnv(env)

	expr := &SexpPair{Symbol("fib"), &SexpPair{int(25), EmptyList, false}, false}

	b.ResetTimer()
	for t := 0; t < b.N; t++ {
		result, err := Eval(expr, env)
		if err != "" {
			b.Error("fib returned error:", err)
			continue
		}
		i, ok := result.(int)
		if !ok {
			b.Error("fib did not return an int!")
			continue
		}
		if i != 75025 {
			b.Error("fib returned wrong result!")
			continue
		}
	}
}
