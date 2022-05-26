package main

import (
	"fmt"
	"go/ast"
	"log"
	"reflect"
)

type VisitorDefinition interface {
	astVisitor()
}

type AstVisitor struct {
	v VisitorDefinition
}

func Visit(v VisitorDefinition, node ast.Node) {
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

func (v *AstVisitor) Visit(node ast.Node) {
	Visit(v.v, node)
}

type Magic struct{}

func (m *Magic) astVisitor()                          {}
func (m *Magic) VisitReturnStmt(node *ast.ReturnStmt) { fmt.Println("Yay") }
