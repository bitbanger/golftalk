package main

type StackFrame struct {
	Running    Procedure
	Args       *SexpPair
	CurrentEnv *Env
	Step       int
	StepInput  Expression
}

func (f *StackFrame) Run(stack *Stack, input Expression) (result Expression, nextEnv *Env, err string) {
	return f.Running.Run(f, stack)
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

func (s *Stack) RunTop(input Expression) (result Expression, nextEnv *Env, err string) {
	return (&((*s)[len(*s)-1])).Run(s, input)
}
