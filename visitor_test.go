package main

import (
	"go/ast"
	"testing"
)

func TestHackyVisitor(t *testing.T) {
	magic := Magic{}
	visitor := AstVisitor{&magic}
	visitor.Visit(&ast.ReturnStmt{0, []ast.Expr{}})
	visitor.Visit(&ast.Ident{0, "a", nil})
}
