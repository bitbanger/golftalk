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

// Get is a simple utility function to Get the nth item from a linked list.
func Get(lst *SexpPair, n int) interface{} {
	obj := lst

	for i := 0; i < n; i++ {
		obj, _ = obj.next.(*SexpPair)
	}

	return obj.val
}

// ToSlice converts a linked list into a slice.
func ToSlice(lst *SexpPair) (result []interface{}) {
	ok := true
	for e := lst ; e != nil && ok; e, ok = GetSexp(e.next) {
		result = append(result, e.val)
	}
	return
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

// Atomize infers the data type of a raw string and returns the string converted to this type.
// If it fails to safely convert the string, it simply returns it as a string again.
func Atomize(str string) interface{} {
	// First, try to Atomize it as an integer
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

// IsBalanced determines if a string has balanced parentheses.
// This should be used to clean data in the REPL stage before it's parsed.
func IsBalanced(line string) bool {
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

// ParseSexp takes a linked list of tokens, including parentheses, and uses the parentheses to "un-flatten" the list.
// The result is a recursively nested set of lists representing the structure of the S-expression.
// Note: the list is expected to have balanced parentheses.
func ParseSexp(tokens *list.List) interface{} {
	token, _ := tokens.Remove(tokens.Front()).(string)

	if token == "(" {
		sexpHead := &SexpPair{"DUMMY", nil}
		sexp := sexpHead
		for true {
			firstTok, _ := tokens.Front().Value.(string)
			if firstTok == ")" {
				break
			}
			newPair := &SexpPair{ParseSexp(tokens), sexp.next}
			sexp.next = newPair
			sexp = newPair
		}
		tokens.Remove(tokens.Front())
		return sexpHead.next
	} else {
		return Atomize(token)
	}

	return nil
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

	if l, ok := GetSexp(sexp); ok {
		ret := "("
		for ; ok && l != nil; l, ok = l.next.(*SexpPair) {
			ret = ret + SexpToString(l.val)
			if _, nextOk := l.next.(*SexpPair); nextOk {
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
		spacedParenStr := strings.Trim(strings.Replace(strings.Replace(valStr, "(", " ( ", -1), ")", " ) ", -1), " ")
		sexp = ParseSexp(SplitByRegex(spacedParenStr, "\\s+"))
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
	if lst, ok := GetSexp(sexp); ok && lst == nil {
		return sexp, ""
	}

	// Is the sexp just a list?
	// If so, let's apply the first symbol as a function to the rest of it!
	if lst, ok := sexp.(*SexpPair); ok {
		// The "car" of the list will be a symbol representing a function
		car, _ := lst.val.(string)
		args, argsOk := GetSexp(lst.next)
		if !argsOk {
			return nil, "Function has invalid argument list."
		}

		switch car {
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
				return lst, ""
			case "yknow":
				sym, wasStr := Get(lst, 1).(string)
				symExp := Get(lst, 2)
				
				if !wasStr {
					return nil, "Symbol given to define wasn't a string."
				}
				
				env.Dict[sym] = symExp
				
				return nil, ""
			case "apply":
				evalFunc, _ := Eval(Get(lst, 1), env)
				proc, wasFunc := evalFunc.(func(args ...interface{}) (interface{}, string))
				evalList, _ := Eval(Get(lst, 2), env)
				args, wasList := GetSexp(evalList)
				
				if !wasFunc {
					return nil, "Function given to apply doesn't evaluate as a function."
				}
				
				if !wasList {
					return nil, "List given to apply doesn't evaluate as a list."
				}
				
				argArr := ToSlice(args)
				return proc(argArr...)
			case "bring-me-back-something-good":
				symbols, symbolsOk := GetSexp(args.val)
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
			default:
				evalFunc, funcErr := Eval(car, env)
				if funcErr != "" {
					return nil, funcErr
				}
				proc, wasFunc := evalFunc.(func(env *Env, args ...interface{}) (interface{}, string))
				if !wasFunc {
					return nil, "Function to execute was not a valid function."
				}
				
				return proc(env, ToSlice(args)...)
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
		
		if !IsBalanced(line) {
			fmt.Println("No.\n\tUnbalanced parentheses.")
			continue
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
