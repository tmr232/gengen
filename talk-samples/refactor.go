//go:build gengen

package main

import (
	"fmt"
	"github.com/tmr232/gengen"
)

//go:generate go run github.com/tmr232/gengen/cmd/gengen

type Book struct {
	Name      string
	Author    string
	Published int
}

func (b Book) String() string {
	return fmt.Sprintf("\"%s\" / %s (%d)", b.Name, b.Author, b.Published)
}

type Shelf struct {
	Books []Book
}

type Room struct {
	Shelves []Shelf
}

type Library struct {
	Rooms []Room
}

func PrintAllBooks(library Library) {
	for _, room := range library.Rooms {
		for _, shelf := range room.Shelves {
			for _, book := range shelf.Books {
				fmt.Println(book)
			}
		}
	}
}

type StopIterationError struct{}

func (s StopIterationError) Error() string {
	return "StopIterationError"
}

var StopIteration = StopIterationError{}

func ForEachBook(library Library, callback func(Book) error) error {
	for _, room := range library.Rooms {
		for _, shelf := range room.Shelves {
			for _, book := range shelf.Books {
				err := callback(book)
				if err == StopIteration {
					return nil
				}
				return err
			}
		}
	}
	return nil
}

//func FindBook(library Library, predicate func(Book) bool) (Book, found bool) {
//	var result Book
//	var found bool
//
//	ForEachBook(library, func(book Book) error {
//		if predicate(book) {
//			result = book
//			found = true
//			return StopIteration
//		}
//		return nil
//	})
//
//	return result, found
//}

//func GenFindBook(library Library, predicate func(Book) bool) (Book, found bool) {
//	it := IterBooks(library)
//	for it.Next() {
//		book := it.Value()
//		if predicate(book) {
//			return book, true
//		}
//	}
//	return Book{}, false
//}

type LibraryBookIterator struct {
	bookIndex  int
	shelfIndex int
	roomIndex  int
	library    Library
}

func NewBookIterator(library Library) *LibraryBookIterator {
	return &LibraryBookIterator{
		bookIndex:  -1,
		shelfIndex: 0,
		roomIndex:  0,
		library:    library,
	}
}

func (it *LibraryBookIterator) Next() bool {
	it.bookIndex++
	for it.bookIndex >= len(it.library.Rooms[it.roomIndex].Shelves[it.shelfIndex].Books) {
		it.bookIndex = 0
		it.shelfIndex++
		for it.shelfIndex >= len(it.library.Rooms[it.roomIndex].Shelves) {
			it.shelfIndex = 0
			it.roomIndex++
			if it.roomIndex >= len(it.library.Rooms) {
				return false
			}
		}
	}

	return true
}

func (it *LibraryBookIterator) Value() Book {
	return it.library.Rooms[it.roomIndex].Shelves[it.shelfIndex].Books[it.bookIndex]
}

type ClosureIterator[T any] struct {
	Advance func(withValue func(value T) bool, withError func(err error) bool, exhausted func() bool) bool
	value   T
	err     error
}

func (it *ClosureIterator[T]) Value() T {
	return it.value
}

func (it *ClosureIterator[T]) Err() error {
	return it.err
}

func (it *ClosureIterator[T]) Next() bool {
	withValue := func(value T) bool {
		it.value = value
		return true
	}
	withError := func(err error) bool {
		it.err = err
		return false
	}
	exhausted := func() bool { return false }
	return it.Advance(withValue, withError, exhausted)
}

func TalkBookIterator(library Library) ClosureIterator[Book] {
	bookIndex := -1
	shelfIndex := 0
	roomIndex := 0
	return ClosureIterator[Book]{
		Advance: func(withValue func(value Book) bool, withError func(err error) bool, exhausted func() bool) bool {
			bookIndex++
			for bookIndex >= len(library.Rooms[roomIndex].Shelves[shelfIndex].Books) {
				bookIndex = 0
				shelfIndex++
				for shelfIndex >= len(library.Rooms[roomIndex].Shelves) {
					shelfIndex = 0
					roomIndex++
					if roomIndex >= len(library.Rooms) {
						return exhausted()
					}
				}
			}

			return withValue(library.Rooms[roomIndex].Shelves[shelfIndex].Books[bookIndex])
		},
	}
}

func TalkIteratorPrintAllBooks(library Library) {
	it := TalkBookIterator(library)
	for it.Next() {
		fmt.Println(it.Value())
	}
}

func IteratorPrintAllBooks(library Library) {
	it := NewBookIterator(library)
	for it.Next() {
		fmt.Println(it.Value())
	}
}

func GeneratorPrintAllBooks(library Library) {
	it := IterBooks(library)
	for it.Next() {
		fmt.Println(it.Value())
	}
}

func IterBooks(library Library) gengen.Generator[Book] {
	for _, room := range library.Rooms {
		for _, shelf := range room.Shelves {
			for _, book := range shelf.Books {
				gengen.Yield(book)
			}
		}
	}
	return nil
}

func main() {
	library := Library{
		[]Room{
			Room{
				[]Shelf{
					Shelf{
						[]Book{
							Book{"The Fellowship of the Ring", "J.R.R Tolkien", 1954},
							Book{"The Hobbit", "J.R.R Tolkien", 1937},
						},
					},
					Shelf{},
				},
			},
		},
	}

	TalkIteratorPrintAllBooks(library)
}

/*
Filtering is a good example, as you might want to use it as a middle-layer
and only get what you need when you need it.
This means that you _have_ to have an iterator.
*/
