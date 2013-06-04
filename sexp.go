package main

import (
	"errors"
	"fmt"
	"os"
)

type SexpPair struct {
	val     interface{}
	next    interface{}
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

func toList(items ...interface{}) (head *SexpPair) {
	head = EmptyList
	for i := len(items) - 1; i >= 0; i-- {
		// TODO: Figure out if this should create literal lists or not
		head = &SexpPair{items[i], head, true}
	}
	return
}

// Get is a simple utility function to Get the nth item from a linked list.
func Get(lst *SexpPair, n int) interface{} {
	obj := lst

	for i := 0; i < n; i++ {
		obj, _ = obj.next.(*SexpPair)
	}

	return obj.val
}

// ToSlice converts a linked list into a slice.
func ToSlice(lst *SexpPair) (result []interface{}) {
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

func (lst *SexpPair) Eval(env *Env) (result interface{}, nextEnv *Env, err string) {
	// Is the sexp literal?
	// If so, just return it.
	if lst == EmptyList || lst.literal {
		return lst, env, ""
	}

	args, argsOk := lst.next.(*SexpPair)
	if !argsOk {
		return nil, nil, "Function has invalid argument list."
	}

	//If s is not a symbol, s wil be "", which will fall to default correctly
	switch s, _ := lst.val.(Symbol); s {
	case "let":
		if length, _ := args.Len(); length != 2 {
			return nil, nil, "Let statements take two arguments: a list of bindings and an S-expression to evaluate."
		}

		// Check that our arguments are okay
		bindings, bindsOk := Get(lst, 1).(*SexpPair)
		if !bindsOk {
			return nil, nil, "First argument to a let statement must be a list of bindings."
		} else if bindings.literal {
			return nil, nil, "List of bindings cannot be literal."
		}
		expression := Get(lst, 2)

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

	case "cond":
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
	case "if":
		if !USE_SCHEME_NAMES {
			break
		}
		fallthrough
	case "insofaras":
		test := Get(lst, 1)
		conseq := Get(lst, 2)
		alt := Get(lst, 3)

		evalTest, testErr := Eval(test, env)

		if testErr != "" {
			return nil, nil, testErr
		}

		result, wasBool := evalTest.(bool)
		if !wasBool {
			return nil, nil, "Test given to conditional did not evaluate to a bool."
		}

		if result {
			return conseq, env, ""
		} else {
			return alt, env, ""
		}
	case "quote":
		if !USE_SCHEME_NAMES {
			break
		}
		fallthrough
	case "this-guy":
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
	case "define":
		if !USE_SCHEME_NAMES {
			break
		}
		fallthrough
	case "yknow":
		sym, wasSym := Get(lst, 1).(Symbol)
		symExp := Get(lst, 2)

		if !wasSym {
			return nil, nil, "Symbol given to define wasn't a symbol."
		}

		evalExp, evalErr := Eval(symExp, env)

		if evalErr != "" {
			return nil, nil, evalErr
		}

		env.Dict[string(sym)] = evalExp

		return nil, nil, ""
	case "apply":
		if !USE_SCHEME_NAMES {
			break
		}
		fallthrough
	case "crunch-crunch-crunch":
		evalFunc, _ := Eval(Get(lst, 1), env)
		proc, wasFunc := evalFunc.(func(args ...interface{}) (interface{}, string))
		evalList, _ := Eval(Get(lst, 2), env)
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
	case "begin":
		var result interface{} = nil
		err := ""
		numArgs, _ := args.Len()

		ourEnv := NewEnv()
		ourEnv.Outer = env

		for i := 1; i <= numArgs; i++ {
			result, err = Eval(Get(lst, i), ourEnv)
		}

		//FIXME last Eval could be done without recurring
		return result, env, err
	case "lambda":
		if !USE_SCHEME_NAMES {
			break
		}
		fallthrough
	case "bring-me-back-something-good":
		symbols, symbolsOk := args.val.(*SexpPair)
		numSymbols, err := symbols.Len()
		if !symbolsOk || err != nil {
			return nil, nil, "Symbol list to bind within lambda wasn't a list."
		}

		exp := Get(lst, 2)

		lambVars := make([]Symbol, numSymbols)
		for i := range lambVars {
			lambVar, _ := Get(symbols, i).(Symbol)
			lambVars[i] = lambVar
		}

		return Proc{lambVars, exp, env}, env, ""
	case "exit":
		fmt.Println("\nhave a nice day ;)")
		os.Exit(0)
	// TODO: Argument number checking
	default:
		evalFunc, funcErr := Eval(lst.val, env)
		if funcErr != "" {
			return nil, nil, funcErr
		}

		var argSlice []interface{}
		for arg, ok := args, true; arg != EmptyList; arg, ok = arg.next.(*SexpPair) {
			if !ok {
				return nil, nil, "Argument list was not a list."
			}

			evalArg, evalErr := Eval(arg.val, env)

			// Errors propagate upward
			if evalErr != "" {
				return nil, nil, evalErr
			}

			argSlice = append(argSlice, evalArg)
		}

		fun, wasFunc := evalFunc.(func(args ...interface{}) (interface{}, string))
		proc, wasProc := evalFunc.(Proc)

		if wasProc {
			// Bind params to args in a new environment
			argSlice := ToSlice(args)
			var evalErr string
			for i, _ := range argSlice {
				argSlice[i], evalErr = Eval(argSlice[i], env)

				if evalErr != "" {
					return nil, nil, evalErr
				}
			}

			newEnv := MakeEnv(proc.Vars, argSlice, proc.EvalEnv)

			// Set the expression to be evaluated
			return proc.Exp, newEnv, ""
		} else if wasFunc {
			result, err = fun(argSlice...)
			return result, nextEnv, err
		} else {
			return nil, nil, fmt.Sprintf("Function '%s' to execute was not a valid function.", lst.val)
		}
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
