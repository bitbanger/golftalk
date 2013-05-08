package main

import "errors"

type SexpPair struct {
	val interface{}
	next interface{}
}

var EmptyList *SexpPair = nil

func (pair *SexpPair) Len() (length int, err error) {
	if pair == EmptyList {
		return
	}
	if next, ok := pair.next.(*SexpPair); ok {
		length, err = next.Len()
		length++
		return
	}
	return 1, errors.New("pair does not represent a list")
}

func toList(items... interface{}) (head *SexpPair) {
	head = EmptyList
	for i := len(items) - 1; i >= 0; i-- {
		head = &SexpPair{items[i], head}
	}
	return
}
