package sample

import xyz "fmt"

type SomeGenError struct{}

func (s SomeGenError) Error() string {
	xyz.Println("abc")
	return "Ooh! Error!"
}
