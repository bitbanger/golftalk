package main

import (
	"errors"
	"fmt"
)

type SexpPair struct {
	val     Expression
	next    Expression
	literal bool
}

var EmptyList *SexpPair = nil

//*SexpPair Should implement Expression
var _ Expression = EmptyList

func (pair *SexpPair) Len() (length int, err error) {
	if pair == EmptyList {
		return
	}
	if next, ok := pair.next.(*SexpPair); ok {
		length, err = next.Len()
		length++
		return
	}
	return 1, errors.New("pair does not represent a list")
}

func toList(items ...Expression) (head *SexpPair) {
	head = EmptyList
	for i := len(items) - 1; i >= 0; i-- {
		// TODO: Figure out if this should create literal lists or not
		head = &SexpPair{items[i], head, true}
	}
	return
}

// Get is a simple utility function to Get the nth item from a linked list.
func Get(lst *SexpPair, n int) Expression {
	obj := lst

	for i := 0; i < n; i++ {
		obj, _ = obj.next.(*SexpPair)
	}

	return obj.val
}

// ToSlice converts a linked list into a slice.
func ToSlice(lst *SexpPair) (result []Expression) {
	//FIXME: should probably be able to return an error if this errors
	count, _ := lst.Len()
	result = make([]Expression, 0, count)

	ok := true
	for e := lst; e != EmptyList && ok; e, ok = e.next.(*SexpPair) {
		result = append(result, e.val)
	}
	return
}

// SetIsLiteral allows a recursive toggle of a list's literal-ness.
// If a list is literal, it will never be executed as code.
func SetIsLiteral(lst *SexpPair, l bool) {
	if lst == EmptyList {
		return
	}

	lst.literal = l

	if nextLst, ok := lst.next.(*SexpPair); ok && nextLst != EmptyList {
		SetIsLiteral(nextLst, l)
	}
}

var coreFuncs = map[Symbol]CoreFunc{
	"call/cc": coreCallCC,

	"yknow":  coreDefine,
	"define": coreDefine,

	"insofaras": coreIf,
	"if":        coreIf,

	"bring-me-back-something-good": coreLambda,
	"lambda":                       coreLambda,

	"this-guy": coreQuote,
	"quote":    coreQuote,

	"crunch-crunch-crunch": coreApply,
	"apply":                coreApply,

	"let": coreLet,

	"cond": coreCond,

	"begin": coreBegin,

	"exit": haveANiceDay,
}

func (lst *SexpPair) Eval(stack *Stack, env *Env) (result Expression, nextEnv *Env, err string) {
	// Is the sexp literal?
	// If so, just return it.
	if lst == EmptyList || lst.literal {
		return lst, env, ""
	}

	// Validate argument list
	args, argsOk := lst.next.(*SexpPair)
	if !argsOk {
		return nil, nil, "Function has invalid argument list."
	}

	sym, _ := lst.val.(Symbol)

	// evaluate the first element of the list to get a function to apply to the rest of the list
	// TODO: Argument number checking
	evalFunc, funcErr := Eval(lst.val, env)
	if funcErr != "" {
		return nil, nil, funcErr
	}

	// Check if it should be interpreted as a procedure
	proc, wasProc := evalFunc.(Procedure)

	if wasProc {
		return Call(proc, args, env, stack)
	} else {
		return nil, nil, fmt.Sprintf("Function '%s' to execute was not a valid function.", lst.val)
	}

	panic(errors.New("list failed to evaluate correctly"))
}

func (l *SexpPair) String() (ret string) {
	ret = "("
	for ok := true; ok && l != EmptyList; l, ok = l.next.(*SexpPair) {
		ret = ret + SexpToString(l.val)
		if next, nextOk := l.next.(*SexpPair); nextOk && next != EmptyList {
			ret = ret + " "
		}
	}
	return ret + ")"
}

func (l *SexpPair) IsLiteral() bool {
	if l == EmptyList {
		return true
	}
	return l.literal
}
