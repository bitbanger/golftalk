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

type Env struct {
	Dict map[string]interface{}
	Outer *Env
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

	result := list.New()

	start := 0
	for _, match := range matches {
		result.PushBack(str[start:match[0]])
		start = match[1]

	}

	result.PushBack(str[start:len(str)])

	return result
}

// And here's where we abandon type safety...
func atomize(str string) interface{} {
	// First, try to atomize it as an integer
	if i, err := strconv.ParseInt(str, 10, 32); err == nil {
		return i
	}

	// That didn't work? Maybe it's a float
	// if f, err := strconv.ParseFloat(str, 32); err == nil {
	// 	return f
	// }

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
	if i, ok := sexp.(int64); ok {
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

func eval(sexp interface{}, env *Env) interface{} {
	// Is the sexp just a symbol?
	// If so, let's look it up!
	if symbol, ok := sexp.(string); ok {
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
	if lst, ok := sexp.(*list.List); ok {
		// The "car" of the list will be a symbol representing a function
		car, _ := lst.Front().Value.(string)

		switch car {
			case "insofaras":
				test := get(lst, 1)
				conseq := get(lst, 2)
				alt := get(lst, 3)

				if result, wasInt := eval(test, env).(int64); wasInt && result > 0 {
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
				return nil
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
	globalEnv := NewEnv()

	globalEnv.Dict["+"] = func(args ...interface{}) interface{} {
		a, aok := args[0].(int64)
		b, bok := args[1].(int64)
		
		if !aok || !bok {
			fmt.Println("No.\n\tInvalid types.")
			return nil
		}

		return a + b
	};

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
			result := eval(parseSexp(splitByRegex(tokenize(line), "\\s+")), globalEnv)
			if result != nil {
				fmt.Println(result)
			}
		}
	}
}
