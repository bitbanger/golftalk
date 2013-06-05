package main

import (
	"fmt"
	"os"
)

type CoreFunc func(lst *SexpPair, env *Env) (result interface{}, nextEnv *Env, err string)

func coreDefine(lst *SexpPair, env *Env) (result interface{}, nextEnv *Env, err string) {
	args, _ := lst.next.(*SexpPair)

	sym, wasSym := Get(args, 0).(Symbol)
	symExp := Get(args, 1)

	if !wasSym {
		return nil, nil, "Symbol given to define wasn't a symbol."
	}

	evalExp, evalErr := Eval(symExp, env)

	if evalErr != "" {
		return nil, nil, evalErr
	}

	env.Dict[string(sym)] = evalExp

	return nil, nil, ""
}

func coreIf(lst *SexpPair, env *Env) (result interface{}, nextEnv *Env, err string) {
	args, _ := lst.next.(*SexpPair)

	test := Get(args, 0)
	conseq := Get(args, 1)
	alt := Get(args, 2)

	evalTest, testErr := Eval(test, env)

	if testErr != "" {
		return nil, nil, testErr
	}

	res, wasBool := evalTest.(bool)
	if !wasBool {
		return nil, nil, "Test given to conditional did not evaluate to a bool."
	}

	if res {
		return conseq, env, ""
	}

	return alt, env, ""
}

func coreLambda(lst *SexpPair, env *Env) (result interface{}, nextEnv *Env, err string) {
	args, _ := lst.next.(*SexpPair)

	symbols, symbolsOk := args.val.(*SexpPair)
	numSymbols, error := symbols.Len()
	if !symbolsOk || error != nil {
		return nil, nil, "Symbol list to bind within lambda wasn't a list."
	}

	exp := Get(lst, 2)

	lambVars := make([]Symbol, numSymbols)
	for i := range lambVars {
		lambVar, _ := Get(symbols, i).(Symbol)
		lambVars[i] = lambVar
	}

	return Proc{lambVars, exp, env}, env, ""
}

func coreQuote(lst *SexpPair, env *Env) (result interface{}, nextEnv *Env, err string) {
	args, _ := lst.next.(*SexpPair)

	if args == EmptyList {
		return nil, nil, "Need something to quote."
	}
	if args.next != EmptyList {
		return nil, nil, "Too many arguments to quote."
	}

	if argLst, ok := args.val.(*SexpPair); ok {
		SetIsLiteral(argLst, true)
		return argLst, env, ""
	}

	return args.val, env, ""
}

func coreApply(lst *SexpPair, env *Env) (result interface{}, nextEnv *Env, err string) {
	args, _ := lst.next.(*SexpPair)

	evalFunc, _ := Eval(Get(args, 0), env)
	proc, wasFunc := evalFunc.(func(args ...interface{}) (interface{}, string))
	evalList, _ := Eval(Get(args, 1), env)
	args, wasList := evalList.(*SexpPair)

	if !wasFunc {
		return nil, nil, "Function given to apply doesn't evaluate as a function."
	}

	if !wasList {
		return nil, nil, "List given to apply doesn't evaluate as a list."
	}

	argArr := ToSlice(args)
	result, err = proc(argArr...)
	return result, env, err
}

func coreLet(lst *SexpPair, env *Env) (result interface{}, nextEnv *Env, err string) {
	args, _ := lst.next.(*SexpPair)

	if length, _ := args.Len(); length != 2 {
		return nil, nil, "Let statements take two arguments: a list of bindings and an S-expression to evaluate."
	}

	// Check that our arguments are okay
	bindings, bindsOk := Get(args, 0).(*SexpPair)
	if !bindsOk {
		return nil, nil, "First argument to a let statement must be a list of bindings."
	} else if bindings.literal {
		return nil, nil, "List of bindings cannot be literal."
	}
	expression := Get(args, 1)

	// Set up parallel slices for binding
	var symbols []string
	var values []interface{}

	// Initialize the let environment first to allow lookups within itself
	letEnv := NewEnv()
	letEnv.Outer = env

	// Loop through all bindings and add them to the let environment
	ok := true
	bindNum := 0
	for sexp := bindings; sexp != EmptyList && ok; sexp, ok = sexp.next.(*SexpPair) {
		bindNum++
		binding, bindOk := sexp.val.(*SexpPair)

		// Check validity of binding (must be a symbol-value pair)
		if !bindOk {
			return nil, nil, fmt.Sprintf("Binding #%d is not an S-expression.", bindNum)
		} else if bindLength, _ := binding.Len(); bindLength != 2 {
			return nil, nil, fmt.Sprintf("Binding #%d does not have two elements.", bindNum)
		} else if binding.literal {
			return nil, nil, fmt.Sprintf("Binding #%d was literal; no binding may be literal.", bindNum)
		}

		// Check validity of symbol (must be a non-literal, non-empty string)
		// Duplicate definitions also may not exist, but we check this when we add the symbols to the environment dictionary to allow it in constant time
		symbol, symOk := binding.val.(Symbol)
		if !symOk || symbol == "" {
			return nil, nil, fmt.Sprintf("Binding #%d has a non-string, empty string, or string literal symbol.", bindNum)
		}
		symbols = append(symbols, string(symbol))

		// Evaluate the binding value before it's bound
		// Allow evaluation error to propagate outward, as usual
		// NOTE: Is it possible that, during sequential evaluation, the environment could be changed and cause the same text to evaluate as two different things?
		next, _ := binding.next.(*SexpPair)
		value, evalErr := Eval(next.val, letEnv)
		if evalErr != "" {
			return nil, nil, evalErr
		}
		values = append(values, value)
	}

	// Bind everything within a local environment
	// Detect duplicate symbol definitions here
	for i, _ := range symbols {
		if _, ok := letEnv.Dict[symbols[i]]; ok {
			return nil, nil, fmt.Sprintf("Binding #%d attempted to re-bind already bound symbol '%s'.", i+1, symbols[i])
		}
		letEnv.Dict[symbols[i]] = values[i]
	}
	// Return the expression in the let environment
	return expression, letEnv, ""
}

func coreCond(lst *SexpPair, env *Env) (result interface{}, nextEnv *Env, err string) {
	args, _ := lst.next.(*SexpPair)

	if length, _ := args.Len(); length == 0 {
		return nil, nil, "Must give at least one clause to cond."
	}

	ok := true
	clauseNum := 0
	for sexp := args; sexp != EmptyList && ok; sexp, ok = sexp.next.(*SexpPair) {
		clauseNum++

		// Check the validity of the clause.
		clause, clauseOk := sexp.val.(*SexpPair)
		if !clauseOk {
			return nil, nil, fmt.Sprintf("Clause #%d was not a list.", clauseNum)
		} else if length, _ := clause.Len(); length != 2 {
			return nil, nil, fmt.Sprintf("Clause #%d was a list with more than two elements.", clauseNum)
		} else if clause.literal {
			return nil, nil, fmt.Sprintf("Clause #%d was a literal list. Clauses may not be literal lists.", clauseNum)
		}

		// Evaluate the clause's test and check its validity.
		eval1, err1 := Eval(clause.val, env)
		if err1 != "" {
			return nil, nil, err1
		}
		testResult, resultOk := eval1.(bool)
		if !resultOk {
			return nil, nil, fmt.Sprintf("Clause #%d's test expression did not evaluate to a bool.", clauseNum)
		}

		// If the test passed, evaluate and return the result.
		if testResult {
			//FIXME: what if this isn't true?
			next := clause.next.(*SexpPair)
			return next.val, env, ""
		}
	}

	return nil, nil, "At least one test given to cond must pass."
}

func coreBegin(lst *SexpPair, env *Env) (result interface{}, nextEnv *Env, err string) {
	args, _ := lst.next.(*SexpPair)

	var res interface{} = nil
	err = ""
	numArgs, _ := args.Len()

	ourEnv := NewEnv()
	ourEnv.Outer = env

	for i := 1; i <= numArgs; i++ {
		res, err = Eval(Get(lst, i), ourEnv)
	}

	//FIXME last Eval could be done without recurring
	return res, env, err
}

func haveANiceDay(lst *SexpPair, env *Env) (result interface{}, nextEnv *Env, err string) {
	fmt.Println("\nhave a nice day ;)")
	os.Exit(0)

	return nil, nil, "Unreachable code."
}
