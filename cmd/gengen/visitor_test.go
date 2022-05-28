package main

import (
	"go/ast"
	"testing"
)

func TestHackyVisitor(t *testing.T) {
	magic := Magic{}
	magic.Visit(&ast.ReturnStmt{0, []ast.Expr{}})
	magic.Visit(&ast.Ident{0, "a", nil})
}
