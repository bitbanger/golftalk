package main

import (
	"fmt"
	"strings"
	"regexp"
	"container/list"
	"bufio"
	"os"
	"io"
)

// Env represents an "environment": a scope's mapping of symbol strings to values.
// Env also provides the ability to search up a scope chain for a value.
type Env struct {
	Dict map[string]interface{}
	Outer *Env
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
func MakeEnv(keys []string, vals []interface{}, outer *Env) *Env {
	env := &Env{}
	env.Dict = make(map[string]interface{})

	for i, key := range keys {
		env.Dict[key] = vals[i]
	}

	env.Outer = outer

	return env
}

type ContextExpression struct {
	val interface{}
	context *Env
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
	if i, ok := sexp.(int); ok {
		return fmt.Sprintf("%d", i)
	}

	// if f, ok := sexp.(float64); ok {
	// 	return fmt.Sprintf("%f", f)
	// }

	if s, ok := sexp.(string); ok {
		return s
	}

	if l, ok := sexp.(*SexpPair); ok {
		ret := "("
		for ; ok && l != EmptyList; l, ok = l.next.(*SexpPair) {
			ret = ret + SexpToString(l.val)
			if next, nextOk := l.next.(*SexpPair); nextOk && next != EmptyList {
				ret = ret + " "
			}
		}
		return ret + ")"
	}

	return ""
}

// Eval takes an S-expression and an environment, and returns the most simplified equivalent S-expression.
// Possible ways to simplify an S-expression include returning a literal value if the input was simply that literal value, looking up a symbol in the given environment (and its implied scope chain), and interpreting the S-expression as a function invocation.
// In the lattermost of evaluation strategies, the function may be provided as a literal or as a symbol referring to a function in the given scope chain; in other words, the first argument has Eval recursively applied to it and must yield a function.
// If an error occurs at any point in the evaluation, Eval returns an error string, and the returned value should be disregarded.
func Eval(val interface{}, env *Env) (interface{}, string) {
	// Make sure the value is an S-expression
	sexp := val
	valStr, wasStr := val.(string)
	if wasStr {
		var err error
		sexp, err = ParseLine(valStr)
		if err != nil {
			return nil, err.Error()
		}
	}
	
	// Is the sexp just a symbol?
	// If so, let's look it up and evaluate it!
	if symbol, ok := sexp.(string); ok {
		// Unless it starts with a quote...
		if strings.HasPrefix(symbol, "'") {
			return symbol[1:], ""
		}
		
		lookupEnv := env.Find(symbol)
		if lookupEnv != nil {
			return Eval(lookupEnv.Dict[symbol], env)
		} else {
			return nil, fmt.Sprintf("'%s' not found in scope chain.", symbol)
		}
	}

	// Is the sexp the empty list?
	if lst, ok := sexp.(*SexpPair); ok && lst == EmptyList {
		return lst, ""
	}

	// Is the sexp a *ContextExpression?
	// If so, evaluate it in that context
	if exp, ok := sexp.(*ContextExpression); ok {
		return Eval(exp.val, exp.context)
	}

	// Is the sexp just a list?
	// If so, let's apply the first symbol as a function to the rest of it!
	if lst, ok := sexp.(*SexpPair); ok {
		args, argsOk := lst.next.(*SexpPair)
		if !argsOk {
			return nil, "Function has invalid argument list."
		}

		switch lst.val {
			case "insofaras":
				test := Get(lst, 1)
				conseq := Get(lst, 2)
				alt := Get(lst, 3)
				
				
				evalTest, testErr := Eval(test, env)
				
				if testErr != "" {
					return nil, testErr
				}
				
				result, wasInt := evalTest.(int)
				
				if !wasInt {
					return nil, "Test given to conditional evaluated as a non-integer."
				} else if result > 0 {
					return Eval(conseq, env)
				} else {
					return Eval(alt, env)
				}
			case "you-folks":
				return args, ""
			case "this-guy":
				if args == EmptyList {
					return nil, "Need something to quote."
				}
				if args.next != EmptyList {
					return nil, "Too many arguments to quote."
				}
				return args.val, ""
			// TODO: Fix this in the case of quoted lists :(
			case "yknow":
				sym, wasStr := Get(lst, 1).(string)
				symExp := Get(lst, 2)
				
				if !wasStr {
					return nil, "Symbol given to define wasn't a string."
				}
				
				evalExp, evalErr := Eval(symExp, env)
				
				if evalErr != "" {
					return nil, evalErr
				}
				
				env.Dict[sym] = evalExp
				
				return nil, ""
			case "crunch-crunch-crunch":
				evalFunc, _ := Eval(Get(lst, 1), env)
				proc, wasFunc := evalFunc.(func(args ...interface{}) (interface{}, string))
				evalList, _ := Eval(Get(lst, 2), env)
				args, wasList := evalList.(*SexpPair)
				
				if !wasFunc {
					return nil, "Function given to apply doesn't evaluate as a function."
				}
				
				if !wasList {
					return nil, "List given to apply doesn't evaluate as a list."
				}
				
				argArr := ToSlice(args)
				return proc(argArr...)
			case "bring-me-back-something-good":
				symbols, symbolsOk := args.val.(*SexpPair)
				numSymbols, err := symbols.Len()
				if !symbolsOk || err != nil {
					return nil, "Symbol list to bind within lambda wasn't a list."
				}

				exp := Get(lst, 2)

				return func(env *Env, args ...interface{}) (interface{}, string) {
					lambVars := make([]string, numSymbols)
					for i := range lambVars {
						// Outer scope handles possible non-string bindables
						lambVar, _ := Get(symbols, i).(string)
						lambVars[i] = lambVar
					}

					newEnv := MakeEnv(lambVars, args, env)

					return Eval(exp, newEnv)
				}, ""
			case "exit":
				os.Exit(0)
			// TODO: Argument number checking
			default:
				evalFunc, funcErr := Eval(lst.val, env)
				if funcErr != "" {
					return nil, funcErr
				}
				proc, wasFunc := evalFunc.(func(env *Env, args ...interface{}) (interface{}, string))
				if !wasFunc {
					fmt.Println(lst.val)
					return nil, "Function to execute was not a valid function."
				}

				var argSlice []interface{}
				for arg, ok := args, true; arg != EmptyList; arg, ok = arg.next.(*SexpPair) {
					if !ok {
						return nil, "Argument list was not a list."
					}
					
					evalArg, evalErr := Eval(arg.val, env)
					
					// Errors propagate upward
					if evalErr != "" {
						return nil, evalErr
					}
					
					argSlice = append(argSlice, evalArg)
				}

				return proc(env, argSlice...)
		}
	}

	// No other choices left; the sexp must be a literal.
	// Let's just return it!
	return sexp, ""
}

// InitGlobalEnv initializes the hierarchichal "root" environment with a few built-in functions.
func InitGlobalEnv(globalEnv *Env) {
	globalEnv.Dict["+"] = add
	globalEnv.Dict["-"] = subtract
	globalEnv.Dict["*"] = multiply
	globalEnv.Dict["/"] = divide
	globalEnv.Dict["%"] = mod

	globalEnv.Dict["or"] = or
	globalEnv.Dict["and"] = and
	globalEnv.Dict["not"] = not

	globalEnv.Dict["eq?"] = equals
	globalEnv.Dict["empty?"] = isEmpty

	globalEnv.Dict["car"] = car
	globalEnv.Dict["come-from-behind"] = comeFromBehind
	globalEnv.Dict["cons"] = cons

	globalEnv.Dict["<"] = lessThan
	globalEnv.Dict[">"] = greaterThan
	globalEnv.Dict["<="] = lessThanOrEqual
	globalEnv.Dict[">="] = greaterThanOrEqual
	// Dat spaceship operator
	globalEnv.Dict["<==>"] = spaceship

	globalEnv.Dict["len"] = length
	globalEnv.Dict["fib"] = fib
	globalEnv.Dict["fact"] = fact
	
	globalEnv.Dict["map"] = mapOnto
	
	globalEnv.Dict["pow"] = pow
	globalEnv.Dict["powmod"] = powmod
	
	globalEnv.Dict["slice-left"] = sliceLeft
	globalEnv.Dict["slice-right"] = sliceRight
	globalEnv.Dict["split"] = split
	globalEnv.Dict["merge"] = merge
	globalEnv.Dict["merge-sort"] = mergeSort
}

func main() {
	globalEnv := NewEnv()

	InitGlobalEnv(globalEnv)

	in := bufio.NewReader(os.Stdin)

	for true {
		fmt.Print("golftalk~$ ")
		line, err := in.ReadString('\n')
		
		if err != nil {
			if err == io.EOF {
				fmt.Println()
				break
			} else {
				panic(err)
			}
		}

		if line != "" && line != "\n" {
			result, evalErr := Eval(line, globalEnv)
			
			if evalErr != "" {
				fmt.Printf("No.\n\t%s\n", evalErr)
				continue
			}
			
			if result != nil {
				fmt.Println(SexpToString(result))
			}
		}
	}
}
