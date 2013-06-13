package main

type StackFrame struct {
	Running    Procedure
	Args       *SexpPair
	CurrentEnv *Env
	Step       int
	StepInput  Expression
}

type Stack []StackFrame

func (s *Stack) Push(proc Procedure, args *SexpPair, env *Env) {
	*s = Stack(append(*s, StackFrame{proc, args, env, 0, nil}))
}

func (s *Stack) Pop() {
	*s = (*s)[:len(*s)-1]
}

func (s *Stack) Empty() bool {
	return len(*s) == 0
}
