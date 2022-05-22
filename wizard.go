package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/format"
	"golang.org/x/tools/go/packages"
	"log"
	"strings"
	"text/template"
)

type Wizard struct {
	template *template.Template
}

//go:embed gengen.tmpl
var coreTemplate string

func NewWizard() *Wizard {
	funcMap := template.FuncMap{
		"join":       strings.Join,
		"trimPrefix": strings.TrimPrefix,
	}
	t, err := template.New("core").Funcs(funcMap).Parse(coreTemplate)
	if err != nil {
		log.Fatal(err)
	}

	return &Wizard{template: t}
}

func (wiz *Wizard) WithPackage(pkg *packages.Package) *PkgWizard {
	return &PkgWizard{
		Wizard: *wiz,
		pkg:    pkg,
	}
}

type PkgWizard struct {
	Wizard
	pkg *packages.Package
}

func (wiz *Wizard) Render(name string, data any) ([]byte, error) {
	var out bytes.Buffer
	err := wiz.template.ExecuteTemplate(&out, name, data)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

func (wiz *PkgWizard) convertFunctions(generatorDecl generatorDecls) []string {
	var functions []string
	pkg := generatorDecl.pkg
	fset := pkg.Fset
	for id, obj := range pkg.TypesInfo.Defs {
		fmt.Printf("%s: %q defines %v\n",
			fset.Position(id.Pos()), id.Name, obj)
	}
	for id, obj := range pkg.TypesInfo.Uses {
		fmt.Printf("%s: %q uses %v\n",
			fset.Position(id.Pos()), id.Name, obj)
	}
	for _, fdecl := range generatorDecl.decls {
		f := wiz.convertFunction(fdecl)
		functions = append(functions, string(f))
	}
	return functions
}

func (wiz *PkgWizard) convertFunction(fdecl *ast.FuncDecl) []byte {

	var out bytes.Buffer
	format.Node(&out, wiz.pkg.Fset, fdecl.Type)
	signature := out.String()

	// We only allow a single result
	if len(fdecl.Type.Results.List) != 1 {
		log.Fatalf("Expected a single result, got %d", len(fdecl.Type.Results.List))
	}

	_, returnType, _ := strings.Cut(wiz.pkg.TypesInfo.TypeOf(fdecl.Type.Results.List[0].Type).String(), "[")
	returnType = strings.TrimSuffix(returnType, "]")

	var body strings.Builder
	for _, node := range fdecl.Body.List {
		body.WriteString(wiz.convertAst(node))
		body.WriteString("\n")
	}

	src, err := wiz.Render("function", struct {
		Name       string
		Signature  string
		ReturnType string
		Body       string
		State      string
	}{
		Name:       fdecl.Name.Name,
		Signature:  signature,
		ReturnType: returnType,
		Body:       body.String(),
		State:      "",
	})
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(src))

	var funcAst bytes.Buffer
	if fdecl.Name.Name == "Yield" {
		ast.Fprint(&funcAst, wiz.pkg.Fset, fdecl.Body, nil)
		fmt.Println(funcAst.String())
	}

	return src
}

func (wiz *PkgWizard) convertAst(node ast.Node) string {
	switch node := node.(type) {
	case *ast.ReturnStmt:
		if len(node.Results) != 1 {
			log.Fatalf("Expected 1 result, got %d", len(node.Results))
		}
		var retval bytes.Buffer
		err := format.Node(&retval, wiz.pkg.Fset, node.Results[0])
		if err != nil {
			log.Fatal(err)
		}

		returnStatement, err := wiz.Render("return", struct{ ReturnValue string }{ReturnValue: retval.String()})
		if err != nil {
			log.Fatal(err)
		}
		return string(returnStatement)
	case *ast.CallExpr:
		object := wiz.pkg.TypesInfo.Uses[node.Fun.(*ast.SelectorExpr).Sel]
		if object.String() == "func github.com/tmr232/gengen/gengen.Yield(value any)" {
			// Yield only accepts one argument
			if len(node.Args) != 1 {
				log.Fatal("Yield accepts a single argument.")
			}
			var yieldValue bytes.Buffer
			format.Node(&yieldValue, wiz.pkg.Fset, node.Args[0])

			yield, err := wiz.Render("yield", struct{ YieldValue string }{YieldValue: yieldValue.String()})
			if err != nil {
				log.Fatal(err)
			}
			return string(yield)
		}
		return "//NO!"
	case *ast.ExprStmt:
		return wiz.convertAst(node.X)
	}
	return "// Unsupported!"
}
