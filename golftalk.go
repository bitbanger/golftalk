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

// Should we use the actual Scheme names?
var USE_SCHEME_NAMES bool = true

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
	sexp := val
		
	// Is the sexp just a symbol?
	// If so, let's look it up and evaluate it!
	if symbol, ok := sexp.(string); ok {
		// Unless it starts with a quote...
		if strings.HasPrefix(symbol, "'") {
			return symbol, ""
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

	// Is the sexp an executable list?
	// If so, let's apply the first symbol as a function to the rest of it!
	if lst, ok := sexp.(*SexpPair); ok && !lst.literal {
		args, argsOk := lst.next.(*SexpPair)
		if !argsOk {
			return nil, "Function has invalid argument list."
		}

		switch lst.val {
			case "let":
				if length, _ := args.Len(); length != 2 {
					return nil, "Let statements take two arguments: a list of bindings and an S-expression to evaluate."
				}
				
				// Check that our arguments are okay
				bindings, bindsOk := Get(lst, 1).(*SexpPair)
				if !bindsOk {
					return nil, "First argument to a let statement must be a list of bindings."
				} else if bindings.literal {
					return nil, "List of bindings cannot be literal."
				}
				expression := Get(lst, 2)
				
				// Set up parallel slices for binding
				var symbols []string
				var values []interface{}
				
				// Initialize the let environment first to allow lookups within itself
				letEnv := NewEnv()
				letEnv.Outer = env
				
				// Loop through all bindings and add them to the let environment
				ok := true
				bindNum := 0
				for sexp := bindings; sexp != EmptyList && ok; sexp, ok = sexp.next.(*SexpPair) {
					bindNum++
					binding, bindOk := sexp.val.(*SexpPair)
					
					// Check validity of binding (must be a symbol-value pair)
					if !bindOk {
						return nil, fmt.Sprintf("Binding #%d is not an S-expression.", bindNum)
					} else if bindLength, _ := binding.Len(); bindLength != 2 {
						return nil, fmt.Sprintf("Binding #%d does not have two elements.", bindNum)
					} else if binding.literal {
						return nil, fmt.Sprintf("Binding #%d was literal; no binding may be literal.", bindNum)
					}
					
					// Check validity of symbol (must be a non-literal, non-empty string)
					// Duplicate definitions also may not exist, but we check this when we add the symbols to the environment dictionary to allow it in constant time
					symbol, symOk := binding.val.(string)
					if !symOk || symbol == "" || symbol[0] == '\'' {
						return nil, fmt.Sprintf("Binding #%d has a non-string, empty string, or string literal symbol.", bindNum)
					}
					symbols = append(symbols, symbol)
					
					// Evaluate the binding value before it's bound
					// Allow evaluation error to propagate outward, as usual
					// NOTE: Is it possible that, during sequential evaluation, the environment could be changed and cause the same text to evaluate as two different things?
					next, _ := binding.next.(*SexpPair)
					value, evalErr := Eval(next.val, letEnv)
					if evalErr != "" {
						return nil, evalErr
					}
					values = append(values, value)
				}
				
				// Bind everything within a local environment
				// Detect duplicate symbol definitions here
				for i, _ := range symbols {
					if _, ok := letEnv.Dict[symbols[i]]; ok {
						return nil, fmt.Sprintf("Binding #%d attempted to re-bind already bound symbol '%s'.", i + 1, symbols[i])
					}
					letEnv.Dict[symbols[i]] = values[i]
				}
				
				// Return the evaluation of the expression in the let environment
				return Eval(expression, letEnv)
				
			case "cond":
				if length, _ := args.Len(); length == 0 {
					return nil, "Must give at least one clause to cond."
				}
				
				ok := true
				clauseNum := 0
				for sexp := args; sexp != EmptyList && ok; sexp, ok = sexp.next.(*SexpPair) {
					clauseNum++
					
					// Check the validity of the clause.
					clause, clauseOk := sexp.val.(*SexpPair)
					if !clauseOk {
						return nil, fmt.Sprintf("Clause #%d was not a list.", clauseNum)
					} else if length, _ := clause.Len(); length != 2 {
						return nil, fmt.Sprintf("Clause #%d was a list with more than two elements.", clauseNum)
					} else if clause.literal {
						return nil, fmt.Sprintf("Clause #%d was a literal list. Clauses may not be literal lists.", clauseNum)
					}
					
					// Evaluate the clause's test and check its validity.
					eval1, err1 := Eval(clause.val, env)
					if err1 != "" {
						return nil, err1
					}
					testResult, resultOk := eval1.(int)
					if !resultOk {
						return nil, fmt.Sprintf("Clause #%d's test expression did not evaluate to an int.", clauseNum)
					}
					
					// If the test passed, evaluate and return the result.
					if testResult > 0 {
						next, _ := clause.next.(*SexpPair)
						eval2, err2 := Eval(next.val, env)
						if err2 != "" {
							return nil, err2
						}
						return eval2, ""
					}
				}
				
				return nil, "At least one test to cond must pass."
			case "if":
				if !USE_SCHEME_NAMES {
					break
				}
				fallthrough
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
			case "quote":
				if !USE_SCHEME_NAMES {
					break
				}
				fallthrough
			case "this-guy":
				if args == EmptyList {
					return nil, "Need something to quote."
				}
				if args.next != EmptyList {
					return nil, "Too many arguments to quote."
				}
				
				if argLst, ok := args.val.(*SexpPair); ok {
					SetIsLiteral(argLst, true)
					return argLst, ""
				}
				
				return args.val, ""
			case "define":
				if !USE_SCHEME_NAMES {
					break
				}
				fallthrough
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
			case "apply":
				if !USE_SCHEME_NAMES {
					break
				}
				fallthrough
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
			case "lambda":
				if !USE_SCHEME_NAMES {
					break
				}
				fallthrough
			case "bring-me-back-something-good":
				symbols, symbolsOk := args.val.(*SexpPair)
				numSymbols, err := symbols.Len()
				if !symbolsOk || err != nil {
					return nil, "Symbol list to bind within lambda wasn't a list."
				}

				exp := Get(lst, 2)

				return func(args ...interface{}) (interface{}, string) {
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
				fmt.Println("\nhave a nice day ;)")
				os.Exit(0)
			// TODO: Argument number checking
			default:
				evalFunc, funcErr := Eval(lst.val, env)
				if funcErr != "" {
					return nil, funcErr
				}
				proc, wasFunc := evalFunc.(func(args ...interface{}) (interface{}, string))
				if !wasFunc {
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

				return proc(argSlice...)
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
	
	globalEnv.Dict[">"], _ = ParseLine(greaterThan)
	globalEnv.Dict["<="], _ = ParseLine(lessThanOrEqual)
	globalEnv.Dict[">="], _ = ParseLine(greaterThanOrEqual)
	// Dat spaceship operator
	globalEnv.Dict["<==>"], _ = ParseLine(spaceship)

	globalEnv.Dict["len"], _ = ParseLine(length)
	globalEnv.Dict["fib"], _ = ParseLine(fib)
	globalEnv.Dict["in-fact"], _ = ParseLine(inFact)
	if USE_SCHEME_NAMES {
		globalEnv.Dict["fact"] = globalEnv.Dict["in-fact"]
	}
	
	globalEnv.Dict["map"], _ = ParseLine(mapOnto)
	globalEnv.Dict["foldl"], _ = ParseLine(foldLeft)
	
	globalEnv.Dict["pow"], _ = ParseLine(pow)
	globalEnv.Dict["powmod"], _ = ParseLine(powmod)
	
	globalEnv.Dict["slice-left"], _ = ParseLine(sliceLeft)
	globalEnv.Dict["slice-right"], _ = ParseLine(sliceRight)
	globalEnv.Dict["split"], _ = ParseLine(split)
	globalEnv.Dict["merge"], _ = ParseLine(merge)
	globalEnv.Dict["merge-sort"], _ = ParseLine(mergeSort)
	
	globalEnv.Dict["min"], _ = ParseLine(min)
	globalEnv.Dict["max"], _ = ParseLine(max)
	
	globalEnv.Dict["range"], _ = ParseLine(numRange)
	globalEnv.Dict["srange"], _ = ParseLine(sRange)
	globalEnv.Dict["rrange"], _ = ParseLine(rRange)
	
	globalEnv.Dict["reverse"], _ = ParseLine(reverse)
	
	globalEnv.Dict["readln"] = readLine
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
				fmt.Println("\n\nhave a nice day ;)")
				break
			} else {
				panic(err)
			}
		}

		if line != "" && line != "\n" {
			sexp, parseErr := ParseLine(line)
			if parseErr != nil {
				fmt.Printf("No.\n\t%s\n", parseErr.Error())
				continue
			}
			
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
