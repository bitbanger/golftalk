package main

import (
	"fmt"
)

type StackFrame struct {
	Running    Procedure
	Args       *SexpPair
	CurrentEnv *Env
	Step       int
	StepInput  Expression
}

func (f *StackFrame) Run(stack *Stack, input Expression) (result Expression, nextEnv *Env, err string) {
	if f.Step == -1 && f.Running == nil {
		proc, ok := f.StepInput.(Procedure)
		if !ok {
			return nil, nil, fmt.Sprintf("Function '%s' to execute was not a valid function.", SexpToString(f.StepInput))
		}
		f.Running = proc
		f.Step = 0
	}
	return f.Running.Run(f, stack)
}

type Stack []StackFrame

func (s *Stack) Push(args *SexpPair, env *Env) {
	// Running will be set the next time this stack frame is run, to whatever
	// is fed to this special step as input (starts at step -1
	*s = Stack(append(*s, StackFrame{nil, args, env, -1, nil}))
}

func (s *Stack) Pop() {
	*s = (*s)[:len(*s)-1]
}

func (s *Stack) Empty() bool {
	return len(*s) == 0
}

func (s *Stack) RunTop(input Expression) (result Expression, nextEnv *Env, err string) {
	return (&((*s)[len(*s)-1])).Run(s, input)
}
