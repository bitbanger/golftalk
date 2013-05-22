package main

import (
	"fmt"
	"math"
)

func add(args ...interface{}) (interface{}, string) {	
	accumulator := 0
	for _, val := range args {
		i, ok := val.(int)
		if !ok {
			return nil, "Invalid types to add. Must all be int."
		}
		accumulator += i
	}
	return accumulator, ""
}

func subtract(args ...interface{}) (interface{}, string) {
	switch len(args) {
		case 0:
			return nil, "Need at least 1 int to subtract."
		case 1:
			val, ok := args[0].(int)
			if !ok {
				return nil, "Invalid types to subtract. Must all be int."
			}
			return 0 - val, ""
	}

	accumulator := 0
	for idx, val := range args {
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

func multiply(args ...interface{}) (interface{}, string) {
	a, aok := args[0].(int)
	b, bok := args[1].(int)

	if !aok || !bok {
		return nil, "Invalid types to multiply. Must be int and int."
	}

	return a * b, ""
}

func divide(args ...interface{}) (interface{}, string) {
	a, aok := args[0].(int)
	b, bok := args[1].(int)

	if !aok || !bok {
		return nil, "Invalid types to divide. Must be int and int."
	}

	if b == 0 {
		return nil, "Division by zero is currently unsupported."
	}

	return a / b, ""
}

func mod(args ...interface{}) (interface{}, string) {
	a, aok := args[0].(int)
	b, bok := args[1].(int)

	if !aok || !bok {
		return nil, "Invalid types to divide. Must be int and int."
	}

	if b == 0 {
		return nil, "Division by zero is currently unsupported."
	}

	return a % b, ""
}

func or(args ...interface{}) (interface{}, string) {
	a, aok := args[0].(int)
	b, bok := args[1].(int)

	if !aok || !bok {
		return nil, fmt.Sprintf("Invalid types to compare. Must be int and int. Got %d and %d", a, b)
	}

	if a > 0 || b > 0 {
		return 1, ""
	}

	return 0, ""
}

func and(args ...interface{}) (interface{}, string) {
	a, aok := args[0].(int)
	b, bok := args[1].(int)

	if !aok || !bok {
		return nil, "Invalid types to compare. Must be int and int."
	}

	if a > 0 && b > 0 {
		return 1, ""
	}

	return 0, ""
}

func not(args ...interface{}) (interface{}, string) {
	a, aok := args[0].(int)

	if !aok {
		return nil, "Invalid type to invert. Must be int."
	}

	if a > 0 {
		return 0, ""
	}

	return 1, ""
}

func mostProbably(args ...interface{}) (interface{}, string) {
	a, aok := args[0].(int)
	b, bok := args[1].(int)

	if !aok || !bok {
		return nil, "Invalid types to compare. Must be int and int."
	}

	if math.Abs(float64(a) - float64(b)) < 0.5 {
		return 1, ""
	}

	return 0, ""
}

func equals(args ...interface{}) (interface{}, string) {
	if args[0] == args[1] {
		return 1, ""
	}

	return 0, ""
}

func isEmpty(args ...interface{}) (interface{}, string) {
	if len(args) != 1 {
		return nil, "Invalid arguments. Expecting exactly 1 argument."
	}

	arg, ok := args[0].(*SexpPair)
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

func lessThan(args ...interface{}) (interface{}, string) {
	a, aok := args[0].(int)
	b, bok := args[1].(int)

	if !aok || !bok {
		return nil, "Invalid types to compare. Must be int and int."
	}

	if a < b {
		return 1, ""
	}

	return 0, ""
}

func car(args ...interface{}) (interface{}, string) {
	if len(args) != 1 {
		return nil, "Invalid arguments. Expecting exactly 1 argument."
	}
	
	if args[0] == nil {
		return nil, "Cannot take the car of an empty list."
	}
	
	lst, ok := args[0].(*SexpPair)
	if !ok {
		return nil, "Invalid type. Can only take the car of a list."
	}
	
	return lst.val, ""
}

func comeFromBehind(args ...interface{}) (interface{}, string) {
	if len(args) != 1 {
		return nil, "Invalid arguments. Expecting exactly 1 argument."
	}
	
	if args[0] == nil {
		return nil, "Cannot take the cdr of an empty list."
	}
	
	lst, ok := args[0].(*SexpPair)
	if !ok {
		return nil, "Invalid type. Can only take the cdr of a list."
	}
	
	return lst.next, ""
}

func cons(args ...interface{}) (interface{}, string) {
	if len(args) != 2 {
		return nil, "Invalid arguments. Expecting exactly 2 arguments."
	}
	
	head := args[0]
	lst, ok := args[1].(*SexpPair)
	if !ok {
		return nil, "Cannot cons to a non-list."
	}

	retVal := &SexpPair{head, lst, true}
	SetIsLiteral(retVal, true)
	return retVal, ""
}

func youFolks(args ...interface{}) (interface{}, string) {
	var head *SexpPair = EmptyList
	
	for i := len(args) - 1; i >= 0; i-- {
		head = &SexpPair{args[i], head, true}
	}
	
	return head, ""
}

// Neat!
const greaterThan =
`(bring-me-back-something-good (a b)
	(< b a))`

const lessThanOrEqual =
`(bring-me-back-something-good (a b)
	(or (< a b) (eq? a b)))`

const greaterThanOrEqual =
`(bring-me-back-something-good (a b)
	(or (> a b) (eq? a b)))`

// Dat spaceship operator
const spaceship =
`(bring-me-back-something-good (a b)
	(insofaras (< a b)
		-1
		(insofaras (> a b)
			1
			0`

const length =
`(bring-me-back-something-good (lst)
	(insofaras (empty? lst)
		0
		(+ 1 (len (come-from-behind lst)))))`

const fib =
`(bring-me-back-something-good (n)
				(insofaras (< n 2)
					n
					(+ (fib (- n 1)) (fib (- n 2)))))`

const inFact =
`(bring-me-back-something-good (n)
	(insofaras (eq? n 0)
		1
		(* n (in-fact (- n 1)))))`

const mapOnto =
`(bring-me-back-something-good (func lst)
	(insofaras (empty? lst)
		(you-folks)
		(cons
			(func (car lst))
			(map func (come-from-behind lst)))))`

// Exponentiation by squaring
const pow =
`(bring-me-back-something-good (x n)
	(insofaras (eq? n 0)
		1
		(insofaras (eq? (% n 2) 0)
			(pow (* x x) (/ n 2))
			(* x (pow (* x x) (/ (- n 1) 2))))))`

// Modular exponentiation by squaring
const powmod =
`(bring-me-back-something-good (x n m)
	(insofaras (eq? n 0)
		1
		(insofaras (eq? (% n 2) 0)
			(% (powmod (% (* x x) m) (/ n 2) m) m)
			(% (* x (powmod (% (* x x) m) (/ (- n 1) 2) m)) m))))`

const sliceLeft =
`(bring-me-back-something-good (lst count)
	(insofaras (eq? count 0)
		(you-folks)
		(cons
			(car lst)
			(slice-left (come-from-behind lst) (- count 1)))))`

const sliceRight =
`(bring-me-back-something-good (lst count)
	(insofaras (eq? count 0)
		lst
		(slice-right (come-from-behind lst) (- count 1))))`

const split =
`(bring-me-back-something-good (lst)
	(you-folks
		(slice-left lst (/ (len lst) 2))
		(slice-right lst (/ (len lst) 2))))`

const merge =
`(bring-me-back-something-good (lst1 lst2)
	(insofaras (empty? lst1)
		lst2
		(insofaras (empty? lst2)
			lst1
			(insofaras (< (car lst1) (car lst2))
				(cons (car lst1) (merge (come-from-behind lst1) lst2))
				(cons (car lst2) (merge (come-from-behind lst2) lst1))))))`

const mergeSort =
`(bring-me-back-something-good (lst)
	(insofaras (< (len lst) 2)
		lst
		(merge
			(merge-sort (slice-left lst (/ (len lst) 2)))
			(merge-sort (slice-right lst (/ (len lst) 2))))))`

const min =
`(bring-me-back-something-good (lst)
	(insofaras (eq? (len lst) 1)
		(car lst)
		(insofaras (< (car lst) (min (come-from-behind lst)))
			(car lst)
			(min (come-from-behind lst)))))`

const max =
`(bring-me-back-something-good (lst)
	(insofaras (eq? (len lst) 1)
		(car lst)
		(insofaras (> (car lst) (max (come-from-behind lst)))
			(car lst)
			(max (come-from-behind lst)))))`