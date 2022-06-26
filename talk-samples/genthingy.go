//go:build gengen

package main

import (
	"github.com/tmr232/gengen"
	"github.com/tmr232/gengen/talk-samples/extrapackage"
)

func KindaUseless() {}

//go:generate go run github.com/tmr232/gengen/cmd/gengen
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

func Nothing() gengen.Generator[extrapackage.Thing] {
	return nil
	gengen.Yield(extrapackage.Thing{})
}
