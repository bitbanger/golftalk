package main

import "errors"

type SexpPair struct {
	val interface{}
	next interface{}
}

func (pair *SexpPair) Len() (length int, err error) {
	if pair == nil {
		return
	}
	if pair.next == nil {
		//don't care what type next is if it's nil
		return 1, nil
	}
	if next, ok := pair.next.(*SexpPair); ok {
		length, err = next.Len()
		length++
		return
	}
	return 1, errors.New("pair does not represent a list")
}

func toList(items... interface{}) (head *SexpPair) {
	head = nil
	for i := len(items) - 1; i >= 0; i-- {
		head = &SexpPair{items[i], head}
	}
	return
}
