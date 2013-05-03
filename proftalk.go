package main

import (
	"fmt"
	"strings"
	"regexp"
	"container/list"
	"strconv"
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
	} else {
		fmt.Printf("Environment failure!\nCouldn't find '%s'\n", val)
	}

	return nil
}

func NewEnv() *Env {
	env := &Env{}
	env.Dict = make(map[string]interface{})
	return env
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

func main() {
	s := "(1 two (4.12))"
	sexp, _ := parseSexp(splitByRegex(tokenize(s), "\\s+")).(*list.List)

	fmt.Println(sexpToString(sexp))
}

