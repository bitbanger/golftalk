package main

import (
	"errors"
	"fmt"
	"os"
)

type CoreFunc func(frame *StackFrame, stack *Stack) (result Expression, nextEnv *Env, done bool, err string)

func coreCallCC(frame *StackFrame, stack *Stack) (result Expression, nextEnv *Env, done bool, err string) {
	return nil, nil, true, "call/cc not implemented"
	//FIXME: actually do this
	/*
		args, _ := lst.next.(*SexpPair)

		arg := Get(args, 0)
		evalArg, evalErr := Eval(arg, env)

		if evalErr != "" {
			return nil, nil, evalErr
		}

		proc, wasProc := evalArg.(Procedure)

		if !wasProc {
			return nil, nil, "Argument to call/cc must be a procedure."
		}

		var throw goProcPtr = func(args ...Expression) (Expression, string) {
			fmt.Println("NON-LOCAL CONTROL FLOW FTW!")

			return nil, ""
		}

		throwProc := &GoProc{"throw", throw}
		throwPair := &SexpPair{throwProc, EmptyList, false}

		return proc.Apply(throwPair, env)
	*/
}

func coreDefine(frame *StackFrame, _ *Stack) (result Expression, nextEnv *Env, done bool, err string) {
	args, env := frame.Args, frame.CurrentEnv
	frame.Step++

	switch frame.Step {
	case 1:
		sym, wasSym := Get(args, 0).(Symbol)
		symExp := Get(args, 1)

		if !wasSym {
			return nil, nil, true, "Symbol given to define wasn't a symbol."
		}

		return symExp, env, false, ""
	case 2:
		evalExp := frame.StepInput

		// Sym should be ok from previous step
		sym := args.val.(Symbol)

		// If we're binding a function to this name, make sure the function literal knows what it's called.
		// This is just to conform with Racket's function display technique. It's not used in the actual execution of the function!
		if proc, wasProc := evalExp.(Procedure); wasProc {
			proc.GiveName(string(sym))
			env.Dict[sym] = proc
		} else {
			env.Dict[sym] = evalExp
		}
		return PTBlank, nil, true, ""
	}
	panic(errors.New(fmt.Sprintf("Invalid step %d in define", frame.Step)))
}

func coreIf(frame *StackFrame, _ *Stack) (result Expression, nextEnv *Env, done bool, err string) {
	args, env := frame.Args, frame.CurrentEnv
	frame.Step++

	switch frame.Step {
	case 1:
		test := Get(args, 0)
		return test, env, false, ""
	case 2:
		res, wasBool := frame.StepInput.(PTBool)
		if !wasBool {
			return nil, nil, true, "Test given to conditional did not evaluate to a bool."
		}

		if res {
			conseq := Get(args, 1)
			return conseq, env, true, ""
		}

		alt := Get(args, 2)
		return alt, env, true, ""
	}
	panic(errors.New(fmt.Sprintf("Invalid step %d in if", frame.Step)))
}

func coreLambda(frame *StackFrame, _ *Stack) (result Expression, nextEnv *Env, done bool, err string) {
	args, env := frame.Args, frame.CurrentEnv

	symbols, symbolsOk := args.val.(*SexpPair)
	numSymbols, error := symbols.Len()
	if !symbolsOk || error != nil {
		return nil, nil, true, "Symbol list to bind within lambda wasn't a list."
	}

	exp := Get(args, 2)

	lambVars := make([]Symbol, numSymbols)
	for i := range lambVars {
		lambVar, _ := Get(symbols, i).(Symbol)
		lambVars[i] = lambVar
	}

	return &Proc{"", lambVars, exp, env}, env, true, ""
}

func coreQuote(frame *StackFrame, _ *Stack) (result Expression, nextEnv *Env, done bool, err string) {
	args, env := frame.Args, frame.CurrentEnv

	if args == EmptyList {
		return nil, nil, true, "Need something to quote."
	}
	if args.next != EmptyList {
		return nil, nil, true, "Too many arguments to quote."
	}

	if argLst, ok := args.val.(*SexpPair); ok {
		SetIsLiteral(argLst, true)
		return argLst, env, true, ""
	}

	return args.val, env, true, ""
}

func coreApply(frame *StackFrame, _ *Stack) (result Expression, nextEnv *Env, done bool, err string) {
	args, env := frame.Args, frame.CurrentEnv
	frame.Step++

	switch frame.Step {
	case 1:
		return Get(args, 0), env, false, ""
	case 2:
		evalFunc := frame.StepInput
		proc, wasFunc := evalFunc.(Procedure)
		if !wasFunc {
			return nil, nil, true, "Function given to apply doesn't evaluate as a function."
		}
		frame.Args.val = proc

		return Get(args, 1), env, false, ""
	case 3:
		evalList := frame.StepInput
		newArgs, wasList := evalList.(*SexpPair)
		if !wasList {
			return nil, nil, true, "List given to apply doesn't evaluate as a list."
		}
		args.next = newArgs

		return args, env, true, ""
	}
	panic(errors.New(fmt.Sprintf("Invalid step %d in apply", frame.Step)))
}

func coreLet(frame *StackFrame, _ *Stack) (result Expression, nextEnv *Env, done bool, err string) {
	args, env := frame.Args, frame.CurrentEnv
	frame.Step++

	if frame.Step == 1 {
		if length, _ := args.Len(); length != 2 {
			return nil, nil, true, "Let statements take two arguments: a list of bindings and an S-expression to evaluate."
		}

		letEnv := NewEnv()
		letEnv.Outer = env
		env = letEnv
		frame.CurrentEnv = env

		// Check that our arguments are okay
		bindings, bindsOk := Get(args, 0).(*SexpPair)
		if !bindsOk {
			return nil, nil, true, "First argument to a let statement must be a list of bindings."
		} else if bindings.literal {
			return nil, nil, true, "List of bindings cannot be literal."
		}

		//Store expression to evaluate in the new environment temporarily
		letEnv.Dict["__let_expression__"] = Get(args, 1)

		//args now holds bindings only
		frame.Args = bindings
		args = frame.Args
	} else {
		// Bind previous variable

		// Binding should be ok from last step
		binding := args.val.(*SexpPair)

		// Symbol should be ok from last step
		symbol := binding.val.(Symbol)
		env.Dict[symbol] = frame.StepInput

		// All done, onto next arg
		var argsOk bool
		frame.Args, argsOk = args.next.(*SexpPair)
		args = frame.Args
		if !argsOk {
			return nil, nil, true, "let: binding list not a list"
		}
	}

	if args == EmptyList {
		// All done binding, grab expression
		expression := env.Dict["__let_expression__"]
		delete(env.Dict, "__let_expression__")

		// And Go
		return expression, env, true, ""
	}

	bindNum := frame.Step
	binding, bindOk := args.val.(*SexpPair)
	// Check validity of binding (must be a symbol-value pair)
	if !bindOk {
		return nil, nil, true, fmt.Sprintf("Binding #%d is not an S-expression.", bindNum)
	} else if bindLength, _ := binding.Len(); bindLength != 2 {
		return nil, nil, true, fmt.Sprintf("Binding #%d does not have two elements.", bindNum)
	} else if binding.literal {
		return nil, nil, true, fmt.Sprintf("Binding #%d was literal; no binding may be literal.", bindNum)
	}

	// Check validity of symbol (must be a non-literal, non-empty string)
	symbol, symOk := binding.val.(Symbol)
	if !symOk || symbol == "" {
		return nil, nil, true, fmt.Sprintf("Binding #%d has a non-string, empty string, or string literal symbol.", bindNum)
	}
	if symbol == "__let_expression__" {
		return nil, nil, true, "let: unable to bind internal symbol \"__let_expression__\"."
	}
	if _, ok := env.Dict[symbol]; ok {
		return nil, nil, true, fmt.Sprintf("Binding #%d attempted to re-bind already bound symbol '%s'.", bindNum, symbol)
	}

	// Evaluate the binding value before it's bound
	// NOTE: Is it possible that, during sequential evaluation, the environment could be changed and cause the same text to evaluate as two different things?
	next, _ := binding.next.(*SexpPair)
	return next.val, env, false, ""
}

func coreCond(frame *StackFrame, _ *Stack) (result Expression, nextEnv *Env, done bool, err string) {
	args, env := frame.Args, frame.CurrentEnv
	frame.Step++

	if args == EmptyList {
		return nil, nil, true, "Must give at least one clause to cond."
	}

	if frame.Step > 1 {
		// Get test result from last time
		testResult, resultOk := frame.StepInput.(PTBool)
		if !resultOk {
			return nil, nil, true, fmt.Sprintf("Clause #%d's test expression did not evaluate to a bool.", frame.Step-1)
		}
		// If the test passed, evaluate and return the result.
		if testResult {
			//clause valid from last time
			clause := args.val.(*SexpPair)
			//FIXME: what if this isn't true?
			next := clause.next.(*SexpPair)
			return next.val, env, true, ""
		}

		// Done with last step's condition, on to the next
		args, argsOk := args.next.(*SexpPair)
		frame.Args = args
		if !argsOk {
			return nil, nil, true, "cond: argument list not a list"
		}
		if args == EmptyList {
			return nil, nil, true, "At least one test given to cond must pass."
		}
	}

	clauseNum := frame.Step

	// Check the validity of the clause.
	clause, clauseOk := args.val.(*SexpPair)
	if !clauseOk {
		return nil, nil, true, fmt.Sprintf("Clause #%d was not a list.", clauseNum)
	} else if length, _ := clause.Len(); length != 2 {
		return nil, nil, true, fmt.Sprintf("Clause #%d was a list with more than two elements.", clauseNum)
	} else if clause.literal {
		return nil, nil, true, fmt.Sprintf("Clause #%d was a literal list. Clauses may not be literal lists.", clauseNum)
	}

	// Evaluate the clause's test
	return clause.val, env, false, ""
}

func coreBegin(frame *StackFrame, _ *Stack) (result Expression, nextEnv *Env, done bool, err string) {
	if frame.Step == 0 {
		frame.Step++
		ourEnv := NewEnv()
		ourEnv.Outer = frame.CurrentEnv
		frame.CurrentEnv = ourEnv
	}

	args, env := frame.Args, frame.CurrentEnv

	if args == EmptyList {
		//all done, return last result
		return frame.StepInput, env, true, ""
	}

	var argsOk bool
	frame.Args, argsOk = args.next.(*SexpPair)
	if !argsOk {
		return nil, nil, true, "begin: expression list not a list"
	}
	return args.val, env, false, ""
}

func haveANiceDay(frame *StackFrame, _ *Stack) (result Expression, nextEnv *Env, done bool, err string) {
	fmt.Println("\nhave a nice day ;)")
	os.Exit(0)

	return nil, nil, true, "Unreachable code."
}
