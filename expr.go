package main

import (
	"fmt"
)

type Expression interface {
	Eval(env *Env) (result interface{}, nextEnv *Env, err string)
	String() string
	IsLiteral() bool
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

func (s Symbol) IsLiteral() bool {
	return false
}

type PTInt int

//PTInt should implement Expression
var _ Expression = PTInt(0)

func (i PTInt) Eval(env *Env) (result interface{}, nextEnv *Env, err string) {
	return i, env, ""
}

func (i PTInt) String() string {
	return fmt.Sprintf("%d", i)
}

func (_ PTInt) IsLiteral() bool {
	return true
}

type PTFloat float64

//PTFloat should implement Expression
var _ Expression = PTFloat(0.0)

func (f PTFloat) Eval(env *Env) (result interface{}, nextEnv *Env, err string) {
	return f, env, ""
}

func (f PTFloat) String() string {
	return fmt.Sprintf("%g", f)
}

func (_ PTFloat) IsLiteral() bool {
	return true
}

type PTBool bool

//PTBool should implement Expression
var _ Expression = PTBool(false)

func (b PTBool) Eval(env *Env) (result interface{}, nextEnv *Env, err string) {
	return b, env, ""
}

func (b PTBool) String() string {
	if b {
		return "#t"
	}
	return "#f"
}

func (_ PTBool) IsLiteral() bool {
	return true
}

type QuotedSymbol string

func (s QuotedSymbol) Eval(env *Env) (result interface{}, nextEnv *Env, err string) {
	return s, env, ""
}

func (s QuotedSymbol) String() string {
	return "'" + string(s)
}

func (_ QuotedSymbol) IsLiteral() bool {
	return true
}

//Used only to have some special functions be able to return nothing
type PTBlankType struct{}

var PTBlank Expression = PTBlankType{}

func (_ PTBlankType) Eval(env *Env) (result interface{}, nextEnv *Env, err string) {
	return PTBlank, env, ""
}

func (_ PTBlankType) String() string {
	return ""
}

func (_ PTBlankType) IsLiteral() bool {
	return true
}
