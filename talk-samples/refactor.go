package main

import "fmt"

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

	//for {
	//	if it.bookIndex < len(it.library.Rooms[it.roomIndex].Shelves[it.shelfIndex].Books) {
	//		return true
	//	}
	//	it.bookIndex = 0
	//
	//	it.shelfIndex++
	//	if it.shelfIndex < len(it.library.Rooms[it.roomIndex].Shelves) {
	//		continue
	//	}
	//	it.shelfIndex = 0
	//
	//	it.roomIndex++
	//	if it.roomIndex < len(it.library.Rooms) {
	//		continue
	//	}
	//	return false
	//}

}

func (it *LibraryBookIterator) Value() Book {
	return it.library.Rooms[it.roomIndex].Shelves[it.shelfIndex].Books[it.bookIndex]
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

	GeneratorPrintAllBooks(library)
}
