package main

import (
	"fmt"
	"go/ast"
	"log"
	"reflect"
)

type AstVisitor interface {
	AstVisitor()
}

func Visit(v AstVisitor, node ast.Node) {
	nodeType := reflect.TypeOf(node)
	visitorValue := reflect.ValueOf(v)
	methodValue := visitorValue.MethodByName("Visit" + nodeType.Elem().Name())
	if methodValue == *new(reflect.Value) {
		return
	}

	argType := methodValue.Type().In(0)
	if argType != nodeType {
		log.Fatalf("Function should accept %s but accepts %s instead.", argType, nodeType)
	}

	nodeValue := reflect.ValueOf(node)
	methodValue.Call([]reflect.Value{nodeValue})
}

type GenericAstVisitor[T any] interface {
	GenericAstVisitor() T
}

func GenericVisit[T any](v GenericAstVisitor[T], node ast.Node) func() T {
	nodeType := reflect.TypeOf(node)
	visitorValue := reflect.ValueOf(v)
	methodValue := visitorValue.MethodByName("Visit" + nodeType.Elem().Name())
	if methodValue == *new(reflect.Value) {
		return nil
	}

	argType := methodValue.Type().In(0)
	if argType != nodeType {
		log.Fatalf("Function should accept %s but accepts %s instead.", argType, nodeType)
	}

	nodeValue := reflect.ValueOf(node)
	return func() T { return methodValue.Call([]reflect.Value{nodeValue})[0].Interface().(T) }
}

type Magic struct{}

func (m *Magic) Visit(node ast.Node) {
	Visit(m, node)
}
func (m *Magic) AstVisitor()                          {}
func (m *Magic) VisitReturnStmt(node *ast.ReturnStmt) { fmt.Println("Yay") }
