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
func (wiz *Wizard) Render(name string, data any) ([]byte, error) {
	var out bytes.Buffer
	err := wiz.template.ExecuteTemplate(&out, name, data)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

type PkgWizard struct {
	Wizard
	pkg *packages.Package
}

func (wiz *PkgWizard) WithFunction(fdecl *ast.FuncDecl) *FuncWizard {
	return &FuncWizard{
		PkgWizard: *wiz,
		fdecl:     fdecl,
	}
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
		f := wiz.WithFunction(fdecl).convertFunction()
		functions = append(functions, string(f))
	}
	return functions
}

type FuncWizard struct {
	PkgWizard
	fdecl    *ast.FuncDecl
	maxState int
}

func (wiz *FuncWizard) StateIndices() []int {
	indices := make([]int, wiz.maxState+1)
	for i := range indices {
		indices[i] = i
	}
	return indices
}

func (wiz *FuncWizard) NextIndex() int {
	wiz.maxState += 1
	return wiz.maxState
}

func (wiz *FuncWizard) convertFunction() []byte {

	var out bytes.Buffer
	format.Node(&out, wiz.pkg.Fset, wiz.fdecl.Type)
	signature := out.String()

	// We only allow a single result
	if len(wiz.fdecl.Type.Results.List) != 1 {
		log.Fatalf("Expected a single result, got %d", len(wiz.fdecl.Type.Results.List))
	}

	_, returnType, _ := strings.Cut(wiz.pkg.TypesInfo.TypeOf(wiz.fdecl.Type.Results.List[0].Type).String(), "[")
	returnType = strings.TrimSuffix(returnType, "]")

	var body strings.Builder
	for _, node := range wiz.fdecl.Body.List {
		body.WriteString(wiz.convertAst(node))
		body.WriteString("\n")
	}

	src, err := wiz.Render("function", struct {
		Name         string
		Signature    string
		ReturnType   string
		Body         string
		State        string
		StateIndices []int
	}{
		Name:         wiz.fdecl.Name.Name,
		Signature:    signature,
		ReturnType:   returnType,
		Body:         body.String(),
		State:        "",
		StateIndices: wiz.StateIndices(),
	})
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(src))

	var funcAst bytes.Buffer
	if wiz.fdecl.Name.Name == "Yield" {
		ast.Fprint(&funcAst, wiz.pkg.Fset, wiz.fdecl.Body, nil)
		fmt.Println(funcAst.String())
	}

	return src
}

func (wiz *FuncWizard) convertAst(node ast.Node) string {
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

			yield, err := wiz.Render("yield", struct {
				YieldValue string
				Next       int
			}{
				YieldValue: yieldValue.String(),
				Next:       wiz.NextIndex(),
			})
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
