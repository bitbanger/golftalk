package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
)

func add(args ...interface{}) (interface{}, string) {
	useFloat := false
	accumulator := 0
	fAccumulator := 0.0
	for _, val := range args {
		i, wasInt := val.(int)
		f, wasFloat := val.(float64)

		if wasFloat {
			// Floats can only be added to the float accumulator
			fAccumulator += f
			useFloat = true
		} else if wasInt {
			// Ints can be added to floats and ints
			accumulator += i
			fAccumulator += float64(i)
		} else {
			return nil, "Invalid types to add. Must all be int or float."
		}
	}

	if useFloat {
		return fAccumulator, ""
	}

	return accumulator, ""
}

func subtract(args ...interface{}) (interface{}, string) {
	switch len(args) {
	case 0:
		return nil, "Need at least 1 value to subtract."
	case 1:
		i, wasInt := args[0].(int)
		f, wasFloat := args[0].(float64)

		if wasInt {
			return 0 - i, ""
		} else if wasFloat {
			return 0 - f, ""
		} else if !wasInt && !wasFloat {
			return nil, "Invalid types to subtract. Must all be int or float."
		}

		return nil, "Something went very wrong in the subtract function."
	}

	useFloat := false
	accumulator := 0
	fAccumulator := 0.0
	for idx, val := range args {
		i, wasInt := val.(int)
		f, wasFloat := val.(float64)

		if !wasInt && !wasFloat {
			return nil, "Invalid types to subtract. Must all be int or float."
		}

		if idx == 0 {
			if wasInt {
				accumulator += i
				fAccumulator += float64(i)
			} else if wasFloat {
				fAccumulator += f
				useFloat = true
			}
		} else {
			if wasInt {
				accumulator -= i
				fAccumulator -= float64(i)
			} else if wasFloat {
				fAccumulator -= f
				useFloat = true
			}
		}
	}

	if useFloat {
		return fAccumulator, ""
	}

	return accumulator, ""
}

func multiply(args ...interface{}) (interface{}, string) {
	useFloat := false
	accumulator := 1
	fAccumulator := 1.0
	for _, val := range args {
		i, wasInt := val.(int)
		f, wasFloat := val.(float64)

		if wasFloat {
			fAccumulator *= f
			useFloat = true
		} else if wasInt {
			accumulator *= i
			fAccumulator *= float64(i)
		} else {
			return nil, "Invalid types to multiply. Must all be int or float."
		}
	}

	if useFloat {
		return fAccumulator, ""
	}

	return accumulator, ""
}

func divide(args ...interface{}) (interface{}, string) {
	useFloat := false
	accumulator := 0
	fAccumulator := 0.0
	for idx, val := range args {
		i, wasInt := val.(int)
		f, wasFloat := val.(float64)

		// Initialize accumulators with the first value
		if idx == 0 {
			if wasInt {
				accumulator = i
				fAccumulator = float64(i)
			} else {
				fAccumulator = f
				useFloat = true
			}

			continue
		}

		if (wasInt && i == 0) || (wasFloat && f == 0.0) {
			return nil, "Division by zero is currently unsupported."
		}

		if wasFloat {
			fAccumulator /= f
			useFloat = true
		} else if wasInt {
			accumulator /= i
			fAccumulator /= float64(i)
		} else {
			return nil, "Invalid types to divide. Must all be int or float."
		}
	}

	// If we used floats but it was equivalent to an integer, return an integer
	if float64(accumulator) == fAccumulator {
		return accumulator, ""
	}

	if useFloat {
		return fAccumulator, ""
	}

	return accumulator, ""
}

func mod(args ...interface{}) (interface{}, string) {
	a, aok := args[0].(int)
	b, bok := args[1].(int)

	if !aok || !bok {
		return nil, "Invalid types to divide. Must be int and int."
	}

	if b == 0 {
		return nil, "Division by zero is currently unsupported."
	}

	return a % b, ""
}

func sqrt(args ...interface{}) (interface{}, string) {
	i, wasInt := args[0].(int)
	f, wasFloat := args[0].(float64)

	result := 0.0

	if wasInt {
		result = math.Sqrt(float64(i))
	} else if wasFloat {
		result = math.Sqrt(f)
	} else {
		return nil, "Invalid type for square root. Must be int or float."
	}

	// Return an int iff we got a perfect square
	if math.Floor(result) == result {
		return int(result), ""
	}

	return result, ""
}

func or(args ...interface{}) (interface{}, string) {
	a, aok := args[0].(bool)
	b, bok := args[1].(bool)

	if !aok || !bok {
		return nil, fmt.Sprintf("Invalid types to compare. Must be bool and bool.")
	}

	return a || b, ""
}

func and(args ...interface{}) (interface{}, string) {
	a, aok := args[0].(bool)
	b, bok := args[1].(bool)

	if !aok || !bok {
		return nil, "Invalid types to compare. Must be bool and bool."
	}

	return a && b, ""
}

func not(args ...interface{}) (interface{}, string) {
	a, aok := args[0].(bool)

	if !aok {
		return nil, "Invalid type to invert. Must be bool."
	}

	return !a, ""
}

func mostProbably(args ...interface{}) (interface{}, string) {
	i1, wasInt1 := args[0].(int)
	f1, wasFloat1 := args[0].(float64)

	i2, wasInt2 := args[1].(int)
	f2, wasFloat2 := args[1].(float64)

	if (!wasInt1 && !wasFloat1) || (!wasInt2 && !wasFloat2) {
		return nil, "Invalid types to compare. Each must be int or float."
	}

	if wasInt1 && wasInt2 {
		if i1 == i2 {
			return true, ""
		}
	} else if wasInt1 && wasFloat2 {
		if math.Abs(float64(i1)-f2) < 0.5 {
			return true, ""
		}
	} else if wasFloat1 && wasInt2 {
		if math.Abs(f1-float64(i2)) < 0.5 {
			return true, ""
		}
	} else if wasFloat1 && wasFloat2 {
		if math.Abs(f1-f2) < 0.5 {
			return true, ""
		}
	}

	return false, ""
}

func readLine(args ...interface{}) (interface{}, string) {
	in := bufio.NewReader(os.Stdin)
	line, err := in.ReadString('\n')

	if err != nil {
		return nil, err.Error()
	}

	return line, ""
}

func equals(args ...interface{}) (interface{}, string) {
	return args[0] == args[1], ""
}

func isEmpty(args ...interface{}) (interface{}, string) {
	if len(args) != 1 {
		return nil, "Invalid arguments. Expecting exactly 1 argument."
	}

	arg, ok := args[0].(*SexpPair)
	if !ok {
		return nil, "Invalid type. Can only check if a list is empty."
	}

	return arg == EmptyList, ""
}

func lessThan(args ...interface{}) (interface{}, string) {
	i1, wasInt1 := args[0].(int)
	f1, wasFloat1 := args[0].(float64)

	i2, wasInt2 := args[1].(int)
	f2, wasFloat2 := args[1].(float64)

	if (!wasInt1 && !wasFloat1) || (!wasInt2 && !wasFloat2) {
		return nil, "Invalid types to compare. Each must be int or float."
	}

	if wasInt1 {
		f1 = float64(i1)
	}
	if wasInt2 {
		f2 = float64(i2)
	}

	return f1 < f2, ""
}

func car(args ...interface{}) (interface{}, string) {
	if len(args) != 1 {
		return nil, "Invalid arguments. Expecting exactly 1 argument."
	}

	if args[0] == nil {
		return nil, "Cannot take the car of an empty list."
	}

	lst, ok := args[0].(*SexpPair)
	if !ok {
		return nil, "Invalid type. Can only take the car of a list."
	}

	return lst.val, ""
}

func comeFromBehind(args ...interface{}) (interface{}, string) {
	if len(args) != 1 {
		return nil, "Invalid arguments. Expecting exactly 1 argument."
	}

	if args[0] == nil {
		return nil, "Cannot take the cdr of an empty list."
	}

	lst, ok := args[0].(*SexpPair)
	if !ok {
		return nil, "Invalid type. Can only take the cdr of a list."
	}

	return lst.next, ""
}

func cons(args ...interface{}) (interface{}, string) {
	if len(args) != 2 {
		return nil, "Invalid arguments. Expecting exactly 2 arguments."
	}

	head := args[0]
	lst, ok := args[1].(*SexpPair)
	if !ok {
		return nil, "Cannot cons to a non-list."
	}

	retVal := &SexpPair{head, lst, true}
	SetIsLiteral(retVal, true)
	return retVal, ""
}

func youFolks(args ...interface{}) (interface{}, string) {
	var head *SexpPair = EmptyList

	for i := len(args) - 1; i >= 0; i-- {
		head = &SexpPair{args[i], head, true}
	}

	return head, ""
}

var goLibraryProcs map[string]goProcPtr = map[string]goProcPtr{
	"+":                add,
	"-":                subtract,
	"*":                multiply,
	"/":                divide,
	"%":                mod,
	"sqrt":             sqrt,
	"or":               or,
	"and":              and,
	"not":              not,
	"eq?":              equals,
	"most-probably?":   mostProbably,
	"empty?":           isEmpty,
	"one-less-car":     car,
	"come-from-behind": comeFromBehind,
	"cons":             cons,
	"you-folks":        youFolks,
	"<":                lessThan,
	"readln":           readLine,
}

var alternateNames map[string]string = map[string]string{
	"car":  "car",
	"cdr":  "comeFromBehind",
	"list": "youFolks",
	"fact": "in-fact",
}

//proftalk library code
const libraryCode = `
(yknow >
	(bring-me-back-something-good (a b)
		(< b a)
	)
)

(yknow <=
	(bring-me-back-something-good (a b)
		(or (< a b) (eq? a b))
	)
)

(yknow >=
	(bring-me-back-something-good (a b)
		(or (> a b) (eq? a b))
	)
)

(yknow <==>
	(bring-me-back-something-good (a b)
		(cond
			((> a b) 1)
			((eq? a b) 0)
			((< a b) -1)
		)
	)
)

(yknow len
	(bring-me-back-something-good (lst)
		(insofaras (empty? lst)
			0
			(+ 1 (len (come-from-behind lst)))
		)
	)
)

(yknow fib
	(bring-me-back-something-good (n)
		(insofaras (< n 2)
			n
			(+ (fib (- n 1)) (fib (- n 2)))
		)
	)
)

(yknow in-fact
	(bring-me-back-something-good (n)
		(insofaras (eq? n 0)
			1
			(* n (in-fact (- n 1)))
		)
	)
)

(yknow map
	(bring-me-back-something-good (func lst)
		(insofaras (empty? lst)
			(you-folks)
			(cons
				(func (one-less-car lst))
				(map func (come-from-behind lst))
			)
		)
	)
)

(yknow foldl
	(bring-me-back-something-good (func start lst)
		(cond
			((empty? lst) start)
			(#t (foldl func (func start (one-less-car lst)) (come-from-behind lst)))
		)
	)
)

(yknow pow
	(bring-me-back-something-good (x n)
		(cond
			((eq? n 0) 1)
			((eq? (% n 2) 0) (pow (* x x) (/ n 2)))
			(#t (* x (pow (* x x) (/ (- n 1) 2))))
		)
	)
)

(yknow powmod
	(bring-me-back-something-good (x n m)
		(cond
			((eq? n 0)
				1)
			((eq? (% n 2) 0)
				(% (powmod (% (* x x) m) (/ n 2) m) m))
			(#t
				(% (* x (powmod (% (* x x) m) (/ (- n 1) 2) m)) m))
		)
	)
)

(yknow slice-left
	(bring-me-back-something-good (lst count)
		(insofaras (eq? count 0)
			(you-folks)
			(cons
				(one-less-car lst)
				(slice-left (come-from-behind lst) (- count 1))
			)
		)
	)
)

(yknow slice-right
	(bring-me-back-something-good (lst count)
		(insofaras (eq? count 0)
			lst
			(slice-right (come-from-behind lst) (- count 1))
		)
	)
)

(yknow split
	(bring-me-back-something-good (lst)
		(you-folks
			(slice-left lst (/ (len lst) 2))
			(slice-right lst (/ (len lst) 2))
		)
	)
)

(yknow merge
	(bring-me-back-something-good (lst1 lst2)
		(cond
			((empty? lst1)
				lst2)
			((empty? lst2)
				lst1)
			((< (one-less-car lst1) (one-less-car lst2))
				(cons (one-less-car lst1) (merge (come-from-behind lst1) lst2)))
			(#t
				(cons (one-less-car lst2) (merge (come-from-behind lst2) lst1)))
		)
	)
)

(yknow merge-sort
	(bring-me-back-something-good (lst)
		(insofaras (< (len lst) 2)
			lst
			(let (
				(left-half (slice-left lst (/ (len lst) 2)))
				(right-half (slice-right lst (/ (len lst) 2))))
			(merge (merge-sort left-half) (merge-sort right-half)))
		)
	)
)

(yknow min
	(bring-me-back-something-good (lst)
		(cond
			((eq? (len lst) 1)
				(one-less-car lst))
			((< (one-less-car lst) (min (come-from-behind lst)))
				(one-less-car lst))
			(#t
				(min (come-from-behind lst)))
		)
	)
)

(yknow max
	(bring-me-back-something-good (lst)
		(cond
			((eq? (len lst) 1)
				(one-less-car lst))
			((> (one-less-car lst) (max (come-from-behind lst)))
				(one-less-car lst))
			(#t
				(max (come-from-behind lst)))
		)
	)
)

(yknow range
	(bring-me-back-something-good (a b)
		(cond
			((eq? a b) '())
			((> a b) (cons a (range (- a 1) b)))
			(#t (cons a (range (+ a 1) b)))
		)
	)
)

(yknow srange
	(bring-me-back-something-good (n)
		(range 0 n)
	)
)

(yknow rrange
	(bring-me-back-something-good (n)
		(range n 0)
	)
)

(yknow reverse
	(bring-me-back-something-good (lst)
		(let (
			(reverse-helper (bring-me-back-something-good (lst work)
				(cond
					((empty? lst) work)
					(#t (reverse-helper (come-from-behind lst) (cons (one-less-car lst) work)))))))

			(reverse-helper lst '())
		)
	)
)

(yknow repeat
	(bring-me-back-something-good (val num)
		(insofaras (<= num 0)
			'()
			(cons val (repeat val (- num 1)))
		)
	)
)

(yknow append
	(bring-me-back-something-good (lst1 lst2)
		(let (
			(append-elem (bring-me-back-something-good (lst elem)
				(cond
					((empty? lst)
						(you-folks elem))
					(#t
						(cons (one-less-car lst) (append-elem (come-from-behind lst) elem)))))))
		(cond
			((empty? lst2)
				lst1)
			(#t
				(append
					(append-elem lst1 (car lst2))
					(come-from-behind lst2)
				)
			)
		))
	)
)

(yknow sum
	(bring-me-back-something-good (lst)
		(insofaras (empty? lst)
			0
			(+ (one-less-car lst) (sum (come-from-behind lst)))
		)
	)
)

(yknow count
	(bring-me-back-something-good (elem lst)
		(insofaras (empty? lst)
			0
			(sum (map (bring-me-back-something-good (val) (eq? val elem)) lst))
		)
	)
)
`
