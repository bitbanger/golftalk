package main

type SexpPair struct {
	val interface{}
	next interface{}
}

func toList(items... interface{}) (head *SexpPair) {
	head = nil
	for i := len(items) - 1; i >= 0; i-- {
		head = &SexpPair{items[i], head}
	}
	return
}
