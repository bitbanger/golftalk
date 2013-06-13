package main

import (
	"fmt"
)

type Procedure interface {
	//Runs/resumes the procedure
	Run(frame *StackFrame, stack *Stack) (result Expression, newEnv *Env, err string)

	//Sets the name of the Procedure if it doesn't have one already
	GiveName(name string)

	Expression
}

//TODO: decide if this should be called PTProc or something
type Proc struct {
	Name    string
	Vars    []Symbol
	Exp     Expression
	EvalEnv *Env
}

var _ Procedure = &Proc{}

func (p *Proc) Run(frame *StackFrame, stack *Stack) (result Expression, newEnv *Env, err string) {
	args, env := frame.Args, frame.CurrentEnv
	frame.Step++
	curVar := frame.Step - 1 // Starts at 0
	if frame.Step == 1 {
		// Set up a new environment to do bindings in
		frame.CurrentEnv = NewEnv()
		frame.CurrentEnv.Outer = env
		env = frame.CurrentEnv
	} else {
		// Bind the last evaluation to the last var
		env.Dict[p.Vars[curVar-1]] = frame.StepInput
	}
	if curVar < len(p.Vars) {
		//Get next argument to bind
		bindingExpr := args.val
		var argsOk bool
		frame.Args, argsOk = args.next.(*SexpPair)
		if !argsOk {
			return nil, nil, "Invalid argument list"
		}

		// Evaluate the value of the binding in the outer Env!
		return bindingExpr, env.Outer, ""
	}

	if args != EmptyList {
		return nil, nil, "Too many arguments"
	}

	// Don't need our frame anymore
	stack.Pop()

	// Set the expression to be evaluated
	return p.Exp, env, ""
}

func (p *Proc) GiveName(name string) {
	if p.Name == "" {
		p.Name = name
	}
}

func (p *Proc) String() string {
	if p.Name != "" {
		return fmt.Sprintf("#<procedure:%s>", p.Name)
	}

	return "#<procedure>"
}

func (p *Proc) Eval(_ *Stack, env *Env) (result Expression, nextEnv *Env, err string) {
	return p, env, ""
}

func (_ *Proc) IsLiteral() bool {
	return true
}

type goProcPtr func(args ...Expression) (Expression, string)

type GoProc struct {
	Name    string
	funcPtr goProcPtr
}

func (g *GoProc) Run(frame *StackFrame, stack *Stack) (result Expression, nextEnv *Env, err string) {
	args, env := frame.Args, frame.CurrentEnv
	frame.Step++

	if frame.Step == 1 {
		//I need a new env for my local variables :(
		frame.CurrentEnv = NewEnv()
		frame.CurrentEnv.Outer = env
		env = frame.CurrentEnv
	} else {
		entry := &SexpPair{frame.StepInput, EmptyList, false}
		if tail, ok := env.Dict["__goproc_run_tail__"]; ok {
			tail.(*SexpPair).next = entry
		} else {
			// If first entry, save it as the head, too
			env.Dict["__goproc_run_head__"] = entry
		}
		env.Dict["__goproc_run_tail__"] = entry
	}

	if args != EmptyList {
		bindingExpr := args.val

		var argsOk bool
		frame.Args, argsOk = args.next.(*SexpPair)
		if !argsOk {
			return nil, nil, "Invalid argument list"
		}

		// Evaluate the value of the binding in the outer Env!
		return bindingExpr, env.Outer, ""
	}

	// All done, don't need our frame anymore
	stack.Pop()

	evaluatedArgs, _ := env.Dict["__goproc_run_head__"].(*SexpPair)
	argSlice := ToSlice(evaluatedArgs)
	result, err = g.funcPtr(argSlice...)
	return result, env.Outer, err
}

func (g *GoProc) GiveName(name string) {
	if g.Name == "" {
		g.Name = name
	}
}

func (g *GoProc) String() string {
	if g.Name != "" {
		return fmt.Sprintf("#<procedure:%s>", g.Name)
	}

	return "#<procedure>"
}

func (g *GoProc) Eval(_ *Stack, env *Env) (result Expression, nextEnv *Env, err string) {
	return g, env, ""
}

func (_ *GoProc) IsLiteral() bool {
	return true
}

func (f CoreFunc) Run(frame *StackFrame, stack *Stack) (result Expression, nextEnv *Env, err string) {
	var done bool
	result, nextEnv, done, err = f(frame, stack)
	if done {
		stack.Pop()
	}
	return
}

func (_ CoreFunc) GiveName(name string) {
	//dont care
	return
}

func (_ CoreFunc) String() string {
	return "#<core procedure>"
}

func (f CoreFunc) Eval(_ *Stack, env *Env) (result Expression, nextEnv *Env, err string) {
	return f, env, ""
}

func (_ CoreFunc) IsLiteral() bool {
	return true
}
