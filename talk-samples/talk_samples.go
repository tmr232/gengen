package main

import "fmt"

func

// Iterators

func IteratorInterface() {
	iterator := getIterator()
	for iterator.Next() {
		value := iterator.Value()
		fmt.Println(value)
	}
	if iterator.Err() != nil {
		panic(iterator.Err())
	}
}

// Implementing Iterators

// State-Block
// All vars declared here

type MyIterator[T any] struct{}
func NewIterator[T any]() *MyIterator[T] {
	return new(MyIterator[T])
}
func (it* MyIterator[T]) Value() T {
	return *new(T)
}
func (it* MyIterator[T]) Err() error {
	return nil
}
func (it* MyIterator[T]) Next() bool {
	return false
}

func init() {
	NewIterator[int]()
}