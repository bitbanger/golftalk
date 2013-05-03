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

// TODO: Is this really the best way to do this recursive type embedding thing?
type Environment interface {
	Find(string) *Env
}

type Env struct {
	Dict map[string]interface{}
	Outer Environment
}

func (e Env) Find(val string) *Env {
	if e.Dict[val] != nil {
		return &e
	} else if e.Outer != nil {
		return e.Outer.Find(val)
	}

	return nil
}

func NewEnv() *Env {
	env := &Env{}
	env.Dict = make(map[string]interface{})
	return env
}

func MakeEnv(keys []string, vals []interface{}, outer *Env) *Env {
	env := &Env{}
	env.Dict = make(map[string]interface{})

	for i, key := range keys {
		env.Dict[key] = vals[i]
	}

	env.Outer = outer

	return env
}

func get(lst *list.List, n int) interface{} {
	obj := lst.Front()

	for i := 0; i < n; i++ {
		obj = obj.Next()
	}

	return obj.Value
}

func toSlice(lst *list.List) []interface{} {
	slice := make([]interface{}, lst.Len())
	i := 0
	for e := lst.Front(); e != nil; e = e.Next() {
		slice[i] = e.Value
		i++
	}
	return slice
}

func tokenize(s string) string {
	return strings.Trim(strings.Replace(strings.Replace(s, "(", " ( ", -1), ")", " ) ", -1), " ")
}

func splitByRegex(str, regex string) *list.List {
	re := regexp.MustCompile(regex)
	matches := re.FindAllStringIndex(str, -1)

	// Worst case, this is one greater than the number of matches
	// result := make([]string, len(matches) + 1)

	result := list.New()

	start := 0
	for _, match := range matches {
		// result[i] = str[start:match[0]]
		result.PushBack(str[start:match[0]])
		start = match[1]

	}

	result.PushBack(str[start:len(str)])

	return result
}

// And here's where we abandon type safety...
func atomize(str string) interface{} {
	// First, try to atomize it as an integer
	i, err := strconv.ParseInt(str, 10, 32)
	if err == nil {
		return i
	}

	// That didn't work? Maybe it's a float
	f, err2 := strconv.ParseFloat(str, 32)
	if err2 == nil {
		return f
	}

	// Fuck it; it's a string
	return str
}

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
	} else if token == ")" {
		fmt.Println("Unexpected )")
		return nil
	} else {
		return atomize(token)
	}

	return nil
}

func sexpToString(sexp interface{}) string {
	i, good := sexp.(int64)
	if good {
		return fmt.Sprintf("%d", i)
	}

	f, good2 := sexp.(float64)
	if good2 {
		return fmt.Sprintf("%f", f)
	}

	s, good3 := sexp.(string)
	if good3 {
		return s
	}

	l, good4 := sexp.(*list.List)
	if good4 {
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

func eval(sexp interface{}, env *Env) interface{} {
	// Is the sexp just a symbol?
	// If so, let's look it up!
	symbol, good := sexp.(string)
	if good {
		lookupEnv := env.Find(symbol)
		if lookupEnv != nil {
			return lookupEnv.Dict[symbol]
		} else {
			fmt.Printf("No.\n\t'%s' is an unresolvable symbol.\n", symbol)
			return nil
		}
	}

	// Is the sexp just a list?
	// If so, let's apply the first symbol as a function to the rest of it!
	lst, good2 := sexp.(*list.List)
	if good2 {
		// The "car" of the list will be a symbol representing a function
		car, _ := lst.Front().Value.(string)

		switch car {
			case "insofaras":
				test := get(lst, 1)
				conseq := get(lst, 2)
				alt := get(lst, 3)

				result, wasInt := eval(test, env).(int64)

				if(wasInt && result > 0) {
					return eval(conseq, env)
				} else {
					return eval(alt, env)
				}
			case "you-folks":
				literal := list.New()

				for e := lst.Front().Next(); e != nil; e = e.Next() {
					literal.PushBack(e.Value)
				}

				return literal
			case "yknow":
				sym, _ := get(lst, 1).(string)
				symExp := get(lst, 2)

				env.Dict[sym] = eval(symExp, env)
			case "apply":
				proc, _ := eval(get(lst, 1), env).(func(args ...interface{}) interface{})
				args, _ := eval(get(lst, 2), env).(*list.List)
				argArr := toSlice(args)
				return proc(argArr...)
			case "bring-me-back-something-good":
				vars, _ := get(lst, 1).(*list.List)
				exp := get(lst, 2)

				return func(args ...interface{}) interface{} {
					lambVars := make([]string, vars.Len())
					for i := range lambVars {
						lambVar, _ := get(vars, i).(string)
						lambVars[i] = lambVar
					}

					newEnv := MakeEnv(lambVars, args, env)

					return eval(exp, newEnv)
				};
			case "exit":
				os.Exit(0)
			default:
				args := make([]interface{}, lst.Len() - 1)
				for i := range args {
					args[i] = eval(get(lst, i + 1), env)
				}

				proc, wasFunc := eval(get(lst, 0), env).(func(args ...interface{}) interface{})
				if wasFunc {
					return proc(args...)
				} else {
					fmt.Printf("No.\n\t'%s' is not a valid function.\n", get(lst, 0))
				}

		}
	}

	// No other choices left; the sexp must be a literal.
	// Let's just return it!
	return sexp
}

func main() {
	// s := "(yknow make-two (bring-me-back-something-good (a) 2))"
	// s2 := "(+ 3 7)"
	// sexp := parseSexp(splitByRegex(tokenize(s), "\\s+"))
	// sexp2 := parseSexp(splitByRegex(tokenize(s2), "\\s+"))

	globalEnv := NewEnv()

	globalEnv.Dict["+"] = func(args ...interface{}) interface{} {
		// TODO: Implment type safety here!
		a, _ := args[0].(int64)
		b, _ := args[1].(int64)

		return a + b
	};

	in := bufio.NewReader(os.Stdin)

	for true {
		fmt.Print("golftalk~$ ")
		line, err := in.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println() //end line with prompt on it
				break //We should end the REPL if EOF is reached
			} else {
				panic(err) //something went wrong...
			}
		}
		if line != "" && line != "\n" {
			result := eval(parseSexp(splitByRegex(tokenize(line), "\\s+")), globalEnv)
			if result != nil {
				fmt.Println(result)
			}
		}
	}

	// fmt.Println(eval(sexp, globalEnv))
	// fmt.Println(sexpToString(eval(sexp2, globalEnv)))
}

