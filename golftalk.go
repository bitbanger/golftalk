package main

import (
	"bufio"
	"container/list"
	"fmt"
	"io"
	"os"
	"regexp"
)

// Should we use the actual Scheme names?
var USE_SCHEME_NAMES bool = true

// Env represents an "environment": a scope's mapping of symbol strings to values.
// Env also provides the ability to search up a scope chain for a value.
type Env struct {
	Dict  map[string]interface{}
	Outer *Env
}

type Proc struct {
	Name    string
	Vars    []Symbol
	Exp     interface{}
	EvalEnv *Env
}

// Find returns the closest parent scope with an extant mapping between a given symbol and any value.
func (e Env) Find(val string) *Env {
	if e.Dict[val] != nil {
		return &e
	} else if e.Outer != nil {
		return e.Outer.Find(val)
	}

	return nil
}

// NewEnv returns an initialized environment.
func NewEnv() *Env {
	env := &Env{}
	env.Dict = make(map[string]interface{})
	return env
}

// MakeEnv returns an environment initialized with two parallel symbol-value slices and a parent environment pointer.
func MakeEnv(keys []Symbol, vals []interface{}, outer *Env) *Env {
	env := &Env{}
	env.Dict = make(map[string]interface{})

	for i, key := range keys {
		env.Dict[string(key)] = vals[i]
	}

	env.Outer = outer

	return env
}

// SplitByRegex takes a string to split and a regular expression, and returns a linked list of all substrings separated by strings matching the provided regex.
func SplitByRegex(str, regex string) *list.List {
	re := regexp.MustCompile(regex)
	matches := re.FindAllStringIndex(str, -1)

	result := list.New()

	start := 0
	for _, match := range matches {
		result.PushBack(str[start:match[0]])
		start = match[1]

	}

	result.PushBack(str[start:len(str)])

	return result
}

// SexpToString takes a parsed S-expression and returns a string representation, suitable for printing.
func SexpToString(sexp interface{}) string {
	switch sexp := sexp.(type) {
	case int:
		return fmt.Sprintf("%d", sexp)

	case float64:
		return fmt.Sprintf("%g", sexp)

	case bool:
		if sexp {
			return "#t"
		}
		return "#f"

	case string:
		return sexp

	case Proc:
		if sexp.Name != "" {
			return fmt.Sprintf("#<procedure:%s>", sexp.Name)
		}

		return "#<procedure>"

	// TODO: Make core functions and non-lambda builtins display names as well, somehow.

	case *SexpPair:
		return sexp.String()
	case Symbol:
		return sexp.String()
	}
	return ""
}

// Eval takes an S-expression and an environment, and returns the most simplified equivalent S-expression.
// Possible ways to simplify an S-expression include returning a literal value if the input was simply that literal value, looking up a symbol in the given environment (and its implied scope chain), and interpreting the S-expression as a function invocation.
// In the lattermost of evaluation strategies, the function may be provided as a literal or as a symbol referring to a function in the given scope chain; in other words, the first argument has Eval recursively applied to it and must yield a function.
// If an error occurs at any point in the evaluation, Eval returns an error string, and the returned value should be disregarded.
func Eval(inVal interface{}, inEnv *Env) (interface{}, string) {
	val := inVal
	env := inEnv

	for {
		sexp := val

		// Is the sexp a literal list?
		if lst, ok := sexp.(*SexpPair); ok && (lst == EmptyList || lst.literal) {
			return lst, ""
		}

		//Is the sexp an evaluable expression?
		if expr, ok := sexp.(Expression); ok {
			result, nextEnv, err := expr.Eval(env)
			if err != "" {
				return result, err
			}
			val, env = result, nextEnv
			continue
		}

		// No other choices left; the sexp must be a literal.
		// Let's just return it!
		return sexp, ""
	}

	return nil, "Eval is seriously broken."
}

// InitGlobalEnv initializes the hierarchichal "root" environment with a few built-in functions and constants.
func InitGlobalEnv(globalEnv *Env) {
	globalEnv.Dict["pi"] = 3.141592653589793
	globalEnv.Dict["euler"] = 2.718281828459045

	globalEnv.Dict["+"] = add
	globalEnv.Dict["-"] = subtract
	globalEnv.Dict["*"] = multiply
	globalEnv.Dict["/"] = divide
	globalEnv.Dict["%"] = mod
	globalEnv.Dict["sqrt"] = sqrt

	globalEnv.Dict["or"] = or
	globalEnv.Dict["and"] = and
	globalEnv.Dict["not"] = not

	globalEnv.Dict["eq?"] = equals
	globalEnv.Dict["most-probably?"] = mostProbably
	globalEnv.Dict["empty?"] = isEmpty

	globalEnv.Dict["one-less-car"] = car
	if USE_SCHEME_NAMES {
		globalEnv.Dict["car"] = car
	}
	globalEnv.Dict["come-from-behind"] = comeFromBehind
	if USE_SCHEME_NAMES {
		globalEnv.Dict["cdr"] = comeFromBehind
	}
	globalEnv.Dict["cons"] = cons
	globalEnv.Dict["you-folks"] = youFolks
	if USE_SCHEME_NAMES {
		globalEnv.Dict["list"] = youFolks
	}

	globalEnv.Dict["<"] = lessThan

	for key, val := range globalEnv.Dict {
		globalEnv.Dict[key], _ = Eval(val, globalEnv)
	}

	globalEnv.Dict["readln"] = readLine

	libraryExprs, _ := ParseLine(libraryCode)
	for _, expr := range libraryExprs {
		Eval(expr, globalEnv)
	}
	if USE_SCHEME_NAMES {
		globalEnv.Dict["fact"] = globalEnv.Dict["in-fact"]
	}
}

func main() {
	globalEnv := NewEnv()

	InitGlobalEnv(globalEnv)

	in := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("golftalk~$ ")
		line, err := in.ReadString('\n')

		if err != nil {
			if err == io.EOF {
				fmt.Println("\n\nhave a nice day ;)")
				break
			} else {
				panic(err)
			}
		}

		if line != "" && line != "\n" {
			sexps, parseErr := ParseLine(line)
			if parseErr != nil {
				fmt.Printf("No.\n\t%s\n", parseErr.Error())
				continue
			}

			for _, sexp := range sexps {
				result, evalErr := Eval(sexp, globalEnv)

				if evalErr != "" {
					fmt.Printf("No.\n\t%s\n", evalErr)
					continue
				}

				if result != nil {
					if sexpResult, ok := result.(*SexpPair); ok && (sexpResult == EmptyList || sexpResult.literal) {
						fmt.Print("'")
					}
					fmt.Println(SexpToString(result))
				}
			}
		}
	}
}
