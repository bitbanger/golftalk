package main

import (
	"fmt"
)

func add(env *Env, args ...interface{}) (interface{}, string) {
	evalArgs := make([]interface{}, len(args))
	for i, arg := range args {
		evalArg, err := Eval(arg, env)
		
		if err != "" {
			return nil, err
		}
		
		evalArgs[i] = evalArg
	}
	
	accumulator := 0
	for _, val := range evalArgs {
		i, ok := val.(int)
		if !ok {
			return nil, "Invalid types to add. Must all be int."
		}
		accumulator += i
	}
	return accumulator, ""
}

func subtract(env *Env, args ...interface{}) (interface{}, string) {
	evalArgs := make([]interface{}, len(args))
	for i, arg := range args {
		evalArg, err := Eval(arg, env)
		
		if err != "" {
			return nil, err
		}
		
		evalArgs[i] = evalArg
	}
	
	switch len(args) {
		case 0:
			return nil, "Need at least 1 int to subtract."
		case 1:
			val, ok := evalArgs[0].(int)
			if !ok {
				return nil, "Invalid types to subtract. Must all be int."
			}
			return 0 - val, ""
	}

	accumulator := 0
	for idx, val := range evalArgs {
		i, ok := val.(int)
		if !ok {
			return nil, "Invalid types to subtract. Must all be int."
		}
		if idx == 0 {
			accumulator += i
		} else {
			accumulator -= i
		}
	}
	return accumulator, ""
}

func multiply(env *Env, args ...interface{}) (interface{}, string) {
	a0, err0 := Eval(args[0], env)
	if err0 != "" {
		return nil, err0
	}
	
	a1, err1 := Eval(args[1], env)
	if err1 != "" {
		return nil, err1
	}
	
	a, aok := a0.(int)
	b, bok := a1.(int)

	if !aok || !bok {
		return nil, "Invalid types to multiply. Must be int and int."
	}

	return a * b, ""
}

func divide(env *Env, args ...interface{}) (interface{}, string) {
	a0, err0 := Eval(args[0], env)
	if err0 != "" {
		return nil, err0
	}
	
	a1, err1 := Eval(args[1], env)
	if err1 != "" {
		return nil, err1
	}
	
	a, aok := a0.(int)
	b, bok := a1.(int)

	if !aok || !bok {
		return nil, "Invalid types to divide. Must be int and int."
	}

	if b == 0 {
		return nil, "Division by zero is currently unsupported."
	}

	return a / b, ""
}

func mod(env *Env, args ...interface{}) (interface{}, string) {
	a0, err0 := Eval(args[0], env)
	if err0 != "" {
		return nil, err0
	}
	
	a1, err1 := Eval(args[1], env)
	if err1 != "" {
		return nil, err1
	}
	
	a, aok := a0.(int)
	b, bok := a1.(int)

	if !aok || !bok {
		return nil, "Invalid types to divide. Must be int and int."
	}

	if b == 0 {
		return nil, "Division by zero is currently unsupported."
	}

	return a % b, ""
}

func or(env *Env, args ...interface{}) (interface{}, string) {
	a0, err0 := Eval(args[0], env)
	if err0 != "" {
		return nil, err0
	}
	
	a1, err1 := Eval(args[1], env)
	if err1 != "" {
		return nil, err1
	}
	
	a, aok := a0.(int)
	b, bok := a1.(int)

	if !aok || !bok {
		return nil, fmt.Sprintf("Invalid types to compare. Must be int and int. Got %d and %d", a, b)
	}

	if a > 0 || b > 0 {
		return 1, ""
	}

	return 0, ""
}

func and(env *Env, args ...interface{}) (interface{}, string) {
	a0, err0 := Eval(args[0], env)
	if err0 != "" {
		return nil, err0
	}
	
	a1, err1 := Eval(args[1], env)
	if err1 != "" {
		return nil, err1
	}
	
	a, aok := a0.(int)
	b, bok := a1.(int)

	if !aok || !bok {
		return nil, "Invalid types to compare. Must be int and int."
	}

	if a > 0 && b > 0 {
		return 1, ""
	}

	return 0, ""
}

func not(env *Env, args ...interface{}) (interface{}, string) {
	a0, err0 := Eval(args[0], env)
	if err0 != "" {
		return nil, err0
	}
	
	a, aok := a0.(int)

	if !aok {
		return nil, "Invalid type to invert. Must be int."
	}

	if a > 0 {
		return 0, ""
	}

	return 1, ""
}

func equals(env *Env, args ...interface{}) (interface{}, string) {
	a0, err0 := Eval(args[0], env)
	if err0 != "" {
		return nil, err0
	}
	
	a1, err1 := Eval(args[1], env)
	if err1 != "" {
		return nil, err1
	}
	
	if a0 == a1 {
		return 1, ""
	}

	return 0, ""
}

func isEmpty(env *Env, args ...interface{}) (interface{}, string) {
	if len(args) != 1 {
		return nil, "Invalid arguments. Expecting exactly 1 argument."
	}
	
	a0, err0 := Eval(args[0], env)
	if err0 != "" {
		return nil, err0
	}

	arg, ok := a0.(*SexpPair)
	if !ok {
		return nil, "Invalid type. Can only check if a list is empty."
	}
	
	if arg == nil {
		return 1, ""
	}
	
	argLen, _ := arg.Len()
	
	if argLen == 1 && arg.val == "you-folks" {
		return 1, ""
	}

	return 0, ""
}

func lessThan(env *Env, args ...interface{}) (interface{}, string) {
	a0, err0 := Eval(args[0], env)
	if err0 != "" {
		return nil, err0
	}
	
	a1, err1 := Eval(args[1], env)
	if err1 != "" {
		return nil, err1
	}
	
	a, aok := a0.(int)
	b, bok := a1.(int)

	if !aok || !bok {
		return nil, "Invalid types to compare. Must be int and int."
	}

	if a < b {
		return 1, ""
	}

	return 0, ""
}

func car(env *Env, args ...interface{}) (interface{}, string) {
	if len(args) != 1 {
		return nil, "Invalid arguments. Expecting exactly 1 argument."
	}
	
	a0, err0 := Eval(args[0], env)
	if err0 != "" {
		return nil, err0
	}
	
	if a0 == nil {
		return nil, "Cannot take the car of an empty list."
	}
	
	lst, ok := a0.(*SexpPair)
	if !ok {
		return nil, "Invalid type. Can only take the car of a list."
	}
	
	return lst.val, ""
}

func comeFromBehind(env *Env, args ...interface{}) (interface{}, string) {
	if len(args) != 1 {
		return nil, "Invalid arguments. Expecting exactly 1 argument."
	}
	
	a0, err0 := Eval(args[0], env)
	if err0 != "" {
		return nil, err0
	}
	
	if a0 == nil {
		return nil, "Cannot take the cdr of an empty list."
	}
	
	lst, ok := a0.(*SexpPair)
	if !ok {
		return nil, "Invalid type. Can only take the cdr of a list."
	}
	
	return lst.next, ""
}

func cons(env *Env, args ...interface{}) (interface{}, string) {
	if len(args) != 2 {
		return nil, "Invalid arguments. Expecting exactly 2 arguments."
	}
	
	a0, err0 := Eval(args[0], env)
	if err0 != "" {
		return nil, err0
	}
	
	a1, err1 := Eval(args[1], env)
	if err1 != "" {
		return nil, err1
	}

	head := a0
	lst, ok := a1.(*SexpPair)
	if !ok {
		return nil, "Cannot cons to a non-list."
	}

	return &SexpPair{head, lst}, ""
}

// Neat!
const greaterThan = "(bring-me-back-something-good (a b) (< b a))"
const lessThanOrEqual = "(bring-me-back-something-good (a b) (or (< a b) (eq? a b)))"
const greaterThanOrEqual = "(bring-me-back-something-good (a b) (or (> a b) (eq? a b)))"

// Dat spaceship operator
const spaceship = "(bring-me-back-something-good (a b) (insofaras (< a b) -1 (insofaras (> a b) 1 0"

const length = "(bring-me-back-something-good (lst) (insofaras (empty? lst) 0 (+ 1 (len (come-from-behind lst)))))"

const fib = "(bring-me-back-something-good (n) (insofaras (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))"
const fact = "(bring-me-back-something-good (n) (insofaras (eq? n 0) 1 (* n (fact (- n 1)))))"

const mapOnto = "(bring-me-back-something-good (func lst) (insofaras (empty? lst) (you-folks) (cons (func (car lst)) (map func (come-from-behind lst)))))"

// Exponentiation by squaring
const pow = "(bring-me-back-something-good (x n) (insofaras (eq? n 0) 1 (insofaras (eq? (% n 2) 0) (pow (* x x) (/ n 2)) (* x (pow (* x x) (/ (- n 1) 2))))))"

// Modular exponentiation by squaring
const powmod = "(bring-me-back-something-good (x n m) (insofaras (eq? n 0) 1 (insofaras (eq? (% n 2) 0) (% (powmod (% (* x x) m) (/ n 2) m) m) (% (* x (powmod (% (* x x) m) (/ (- n 1) 2) m)) m))))"
