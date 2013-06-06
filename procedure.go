package main

import (
	"fmt"
)

type Procedure interface {
	//Runs the procedure with the given arguments
	Apply(args *SexpPair, env *Env) (result interface{}, newEnv *Env, error string)

	//Sets the name of the Procedure if it doesn't have one already
	GiveName(name string)

	//returns a string representation of the function
	String() string
}

//TODO: decide if this should be called PTProc or something
type Proc struct {
	Name    string
	Vars    []Symbol
	Exp     interface{}
	EvalEnv *Env
}

var _ Procedure = &Proc{}

func (p *Proc) Apply(args *SexpPair, env *Env) (result interface{}, newEnv *Env, error string) {
	argSlice := ToSlice(args)
	var evalErr string
	for i, _ := range argSlice {
		argSlice[i], evalErr = Eval(argSlice[i], env)

		if evalErr != "" {
			return nil, nil, evalErr
		}
	}

	newEnv = MakeEnv(p.Vars, argSlice, p.EvalEnv)

	// Set the expression to be evaluated
	return p.Exp, newEnv, ""
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
