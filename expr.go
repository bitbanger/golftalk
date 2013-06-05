package main

import (
	"fmt"
)

type Expression interface {
	Eval(env *Env) (result interface{}, nextEnv *Env, err string)
	String() string
}

type Symbol string

//Symbol should implement Expression
var _ Expression = Symbol("")

func (s Symbol) Eval(env *Env) (result interface{}, nextEnv *Env, err string) {
	lookupEnv := env.Find(string(s))
	if lookupEnv != nil {
		return lookupEnv.Dict[string(s)], env, ""
	} else {
		return nil, nil, fmt.Sprintf("'%s' not found in scope chain.", s)
	}
}

func (s Symbol) String() string {
	return string(s)
}
