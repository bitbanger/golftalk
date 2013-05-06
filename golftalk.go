package main

import (
	"fmt"
	"strings"
	"regexp"
	"container/list"
	"strconv"
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

// get is a simple utility function to get the nth item from a linked list.
func get(lst *list.List, n int) interface{} {
	obj := lst.Front()

	for i := 0; i < n; i++ {
		obj = obj.Next()
	}

	return obj.Value
}

// toSlice converts a linked list into a slice.
func toSlice(lst *list.List) []interface{} {
	slice := make([]interface{}, lst.Len())
	i := 0
	for e := lst.Front(); e != nil; e = e.Next() {
		slice[i] = e.Value
		i++
	}
	return slice
}

// tokenize returns a string with all parentheses separated by whitespace, and all leading and trailing whitespace removed.
func tokenize(s string) string {
	return strings.Trim(strings.Replace(strings.Replace(s, "(", " ( ", -1), ")", " ) ", -1), " ")
}

// splitByRegex takes a string to split and a regular expression, and returns a linked list of all substrings separated by strings matching the provided regex.
func splitByRegex(str, regex string) *list.List {
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

// atomize infers the data type of a raw string and returns the string converted to this type.
// If it fails to safely convert the string, it simply returns it as a string again.
func atomize(str string) interface{} {
	// First, try to atomize it as an integer
	if i, err := strconv.ParseInt(str, 10, 32); err == nil {
		return int(i)
	}

	// That didn't work? Maybe it's a float
	// if f, err := strconv.ParseFloat(str, 32); err == nil {
	// 	return f
	// }

	// Fuck it; it's a string
	return str
}

// isBalanced determines if a string has balanced parentheses.
// This should be used to clean data in the REPL stage before it's parsed.
func isBalanced(line string) bool {
	val := 0
	for _, c := range line {
		if c == 40 {
			val++
		} else if c == 41 {
			val--
		}
		
		if val < 0 {
			return false
		}
	}
	
	return val == 0
}

// parseSexp takes a linked list of tokens, including parentheses, and uses the parentheses to "un-flatten" the list.
// The result is a recursively nested set of lists representing the structure of the S-expression.
// Note: the list is expected to have balanced parentheses.
func parseSexp(tokens *list.List) interface{} {
	token, _ := tokens.Remove(tokens.Front()).(string)

	if token == "(" {
		sexp := list.New()
		for true {
			firstTok, _ := tokens.Front().Value.(string)
			if firstTok == ")" {
				break
			}
			sexp.PushBack(parseSexp(tokens))
		}
		tokens.Remove(tokens.Front())
		return sexp
	} else {
		return atomize(token)
	}

	return nil
}

// sexpToString takes a parsed S-expression and returns a string representation, suitable for printing.
func sexpToString(sexp interface{}) string {
	if i, ok := sexp.(int); ok {
		return fmt.Sprintf("%d", i)
	}

	// if f, ok := sexp.(float64); ok {
	// 	return fmt.Sprintf("%f", f)
	// }

	if s, ok := sexp.(string); ok {
		return s
	}

	if l, ok := sexp.(*list.List); ok {
		ret := "("
		for e := l.Front(); e != nil; e = e.Next() {
			ret = ret + sexpToString(e.Value)
			if e.Next() != nil {
				ret = ret + " "
			}
		}
		return ret + ")"
	}

	return ""
}

// eval takes an S-expression and an environment, and returns the most simplified equivalent S-expression.
// Possible ways to simplify an S-expression include returning a literal value if the input was simply that literal value, looking up a symbol in the given environment (and its implied scope chain), and interpreting the S-expression as a function invocation.
// In the lattermost of evaluation strategies, the function may be provided as a literal or as a symbol referring to a function in the given scope chain; in other words, the first argument has eval recursively applied to it and must yield a function.
// If an error occurs at any point in the evaluation, eval returns an error string, and the returned value should be disregarded.
func eval(val interface{}, env *Env) (interface{}, string) {
	// Make sure the value is an S-expression
	sexp := val
	valStr, wasStr := val.(string)
	if wasStr {
		sexp = parseSexp(splitByRegex(tokenize(valStr), "\\s+"))
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
			return eval(lookupEnv.Dict[symbol], env)
		} else {
			return nil, fmt.Sprintf("'%s' not found in scope chain.", symbol)
		}
	}

	// Is the sexp just a list?
	// If so, let's apply the first symbol as a function to the rest of it!
	if lst, ok := sexp.(*list.List); ok {
		// The "car" of the list will be a symbol representing a function
		car, _ := lst.Front().Value.(string)

		switch car {
			case "insofaras":
				test := get(lst, 1)
				conseq := get(lst, 2)
				alt := get(lst, 3)
				
				
				evalTest, testErr := eval(test, env)
				
				if testErr != "" {
					return nil, testErr
				}
				
				result, wasInt := evalTest.(int)
				
				if !wasInt {
					return nil, "Test given to conditional evaluated as a non-integer."
				} else if result > 0 {
					return eval(conseq, env)
				} else {
					return eval(alt, env)
				}
			case "you-folks":
				literal := list.New()

				for e := lst.Front().Next(); e != nil; e = e.Next() {
					val, valErr := eval(e.Value, env)
					
					if valErr != "" {
						return nil, valErr
					}
					
					literal.PushBack(val)
				}

				return literal, ""
			case "yknow":
				sym, wasStr := get(lst, 1).(string)
				symExp := get(lst, 2)
				
				if !wasStr {
					return nil, "Symbol given to define wasn't a string."
				}
				
				// TODO: Is there an elegant way to do this?
				evalErr := ""
				env.Dict[sym], evalErr = eval(symExp, env)
				
				return nil, evalErr
			case "apply":
				evalFunc, _ := eval(get(lst, 1), env)
				proc, wasFunc := evalFunc.(func(args ...interface{}) (interface{}, string))
				evalList, _ := eval(get(lst, 2), env)
				args, wasList := evalList.(*list.List)
				
				if !wasFunc {
					return nil, "Function given to apply doesn't evaluate as a function."
				}
				
				if !wasList {
					return nil, "List given to apply doesn't evaluate as a list."
				}
				
				argArr := toSlice(args)
				return proc(argArr...)
			case "bring-me-back-something-good":
				vars, wasList := get(lst, 1).(*list.List)
				exp := get(lst, 2)
				
				if !wasList {
					return nil, "Symbol list to bind within lambda wasn't a list."
				}

				return func(args ...interface{}) (interface{}, string) {
					lambVars := make([]string, vars.Len())
					for i := range lambVars {
						// Outer scope handles possible non-string bindables
						lambVar, _ := get(vars, i).(string)
						lambVars[i] = lambVar
					}

					newEnv := MakeEnv(lambVars, args, env)

					return eval(exp, newEnv)
				}, ""
			case "exit":
				os.Exit(0)
			default:
				evalFunc, funcErr := eval(get(lst, 0), env)
				
				if funcErr != "" {
					return nil, funcErr
				}
				
				args := make([]interface{}, lst.Len() - 1)
				for i := range args {
					// TODO: Do we really need to evaluate here?
					// Lazy evaluation seems to be the way to go, but then wouldn't we have to evaluate arguments in a more limited scope?
					args[i], _ = eval(get(lst, i + 1), env)
				}
				
				proc, wasFunc := evalFunc.(func(args ...interface{}) (interface{}, string))
				if wasFunc {
					return proc(args...)
				} else {
					return nil, "Function to execute was not a valid function."
				}

		}
	}

	// No other choices left; the sexp must be a literal.
	// Let's just return it!
	return sexp, ""
}

// initGlobalEnv initializes the hierarchichal "root" environment with a few built-in functions.
func initGlobalEnv(globalEnv *Env) {
	globalEnv.Dict["+"] = func(args ...interface{}) (interface{}, string) {
		accumulator := 0
		for _, val := range args {
			i, ok := val.(int)
			if !ok {
				return nil, "Invalid types to add. Must all be int."
			}
			accumulator += i
		}
		return accumulator, ""
	}
	
	globalEnv.Dict["-"] = func(args ...interface{}) (interface{}, string) {
		switch len(args) {
			case 0:
				return nil, "Need at least 1 int to subtract."
			case 1:
				val, ok := args[0].(int)
				if !ok {
					return nil, "Invalid types to subtract. Must all be int."
				}
				return 0 - val, ""
		}

		accumulator := 0
		for idx, val := range args {
			i, ok := val.(int)
			if !ok {
				return nil, "Invalid types to subtract. Must all be int."
			}
			if idx == 0 {
				accumulator += i
			} else {
				accumulator -= i
			}
		}
		return accumulator, ""
	}
	
	globalEnv.Dict["*"] = func(args ...interface{}) (interface{}, string) {
		a, aok := args[0].(int)
		b, bok := args[1].(int)
		
		if !aok || !bok {
			return nil, "Invalid types to multiply. Must be int and int."
		}

		return a * b, ""
	}
	
	globalEnv.Dict["/"] = func(args ...interface{}) (interface{}, string) {
		a, aok := args[0].(int)
		b, bok := args[1].(int)
		
		if !aok || !bok {
			return nil, "Invalid types to divide. Must be int and int."
		}
		
		if b == 0 {
			return nil, "Division by zero is currently unsupported."
		}

		return a / b, ""
	}
	
	globalEnv.Dict["or"] = func(args ...interface{}) (interface{}, string) {
		a, aok := args[0].(int)
		b, bok := args[1].(int)
		
		if !aok || !bok {
			return nil, fmt.Sprintf("Invalid types to compare. Must be int and int. Got %d and %d", a, b)
		}
		
		if a > 0 || b > 0 {
			return 1, ""
		}
		
		return 0, ""
	}
	
	globalEnv.Dict["and"] = func(args ...interface{}) (interface{}, string) {
		a, aok := args[0].(int)
		b, bok := args[1].(int)
		
		if !aok || !bok {
			return nil, "Invalid types to compare. Must be int and int."
		}
		
		if a > 0 && b > 0 {
			return 1, ""
		}
		
		return 0, ""
	}
	
	globalEnv.Dict["not"] = func(args ...interface{}) (interface{}, string) {
		a, aok := args[0].(int)
		
		if !aok {
			return nil, "Invalid type to invert. Must be int."
		}
		
		if a > 0 {
			return 0, ""
		}
		
		return 1, ""
	}
	
	globalEnv.Dict["eq?"] = func(args ...interface{}) (interface{}, string) {
		if args[0] == args[1] {
			return 1, ""
		}
		
		return 0, ""
	}
	
	globalEnv.Dict["<"] = func(args ...interface{}) (interface{}, string) {
		a, aok := args[0].(int)
		b, bok := args[1].(int)
		
		if !aok || !bok {
			return nil, "Invalid types to compare. Must be int and int."
		}

		if a < b {
			return 1, ""
		}
		
		return 0, ""
	}
	
	// Neat!
	globalEnv.Dict[">"] = "(bring-me-back-something-good (a b) (< b a))"
	globalEnv.Dict["<="] = "(bring-me-back-something-good (a b) (or (< a b) (eq? a b)))"
	globalEnv.Dict[">="] = "(bring-me-back-something-good (a b) (or (> a b) (eq? a b)))"
	// Dat spaceship operator
	globalEnv.Dict["<==>"] = "(bring-me-back-something-good (a b) (insofaras (< a b) -1 (insofaras (> a b) 1 0"
	
	globalEnv.Dict["fib"] = "(bring-me-back-something-good (n) (insofaras (< n 2) n (+ (fib (- n 1)) (fib (- n 2)))))"
	globalEnv.Dict["fact"] = "(bring-me-back-something-good (n) (insofaras (eq? n 0) 1 (* n (fact (- n 1)))))"
}

func main() {
	globalEnv := NewEnv()

	initGlobalEnv(globalEnv)

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
		
		if !isBalanced(line) {
			fmt.Println("No.\n\tUnbalanced parentheses.")
			continue
		}
		
		if line != "" && line != "\n" {
			result, evalErr := eval(line, globalEnv)
			
			if evalErr != "" {
				fmt.Printf("No.\n\t%s\n", evalErr)
				continue
			}
			
			if result != nil {
				fmt.Println(sexpToString(result))
			}
		}
	}
}
