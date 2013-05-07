package main

import (
	"container/list"
	"fmt"
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

func equals(args ...interface{}) (interface{}, string) {
	if args[0] == args[1] {
		return 1, ""
	}

	return 0, ""
}

func isEmpty(args ...interface{}) (interface{}, string) {
	lst, ok := args[0].(*list.List)

	if !ok {
		return nil, "Invalid type. Can only check if a list is empty."
	}

	if lst.Len() == 0 {
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
	lst, ok := args[0].(*list.List)

	if !ok {
		return nil, "Invalid type. Can only take the car of a list."
	}

	if lst.Len() == 0 {
		return nil, "Cannot take the car of an empty list."
	}

	return lst.Front().Value, ""
}

func comeFromBehind(args ...interface{}) (interface{}, string) {
	lst, ok := args[0].(*list.List)

	if !ok {
		return nil, "Invalid type. Can only take the cdr of a list."
	}

	if lst.Len() == 0 {
		return nil, "Cannot take the cdr of an empty list."
	}

	// TODO: Implement our own S-expressions because Go lists suck
	newList := list.New()
	for e := lst.Front().Next(); e != nil; e = e.Next() {
		newList.PushBack(e.Value)
	}

	return newList, ""
}

func cons(args ...interface{}) (interface{}, string) {
	lst, ok := args[1].(*list.List)

	if !ok {
		return nil, "Cannot cons to a non-list."
	}

	newList := list.New()
	newList.PushBack(args[0])
	for e := lst.Front(); e != nil; e = e.Next() {
		newList.PushBack(e.Value)
	}

	return newList, ""

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
