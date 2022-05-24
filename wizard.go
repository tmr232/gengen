package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/format"
	"go/types"
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
		PkgWizard:   *wiz,
		fdecl:       fdecl,
		definitions: make(map[types.Object]string),
		names:       make(map[string]bool),
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
	fdecl       *ast.FuncDecl
	maxState    int
	definitions map[types.Object]string
	names       map[string]bool
	loopId      int
}

type Namer struct {
	name string
	id   int
}

func (n *Namer) Next() {
	n.id++
}

func (n *Namer) Name() string {
	if n.id > 0 {
		return fmt.Sprintf("%s%d", n.name, n.id)
	}
	return n.name
}

func (wiz *FuncWizard) DefineVariable(obj types.Object) (name string) {
	namer := Namer{name: obj.Name()}
	for wiz.names[namer.Name()] {
		namer.Next()
	}
	wiz.names[namer.Name()] = true
	wiz.definitions[obj] = namer.Name()
	return namer.Name()
}

func (wiz *FuncWizard) GetVariable(obj types.Object) (name string) {
	return wiz.definitions[obj]
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

	//// We go through all the defs in the function,
	//// and define the relevant variables.
	//scope := wiz.pkg.TypesInfo.Scopes[wiz.fdecl.Type]
	//for _, name := range scope.Names() {
	//	fmt.Println(name, scope.Lookup(name).Type().String())
	//}
	////fmt.Println("SCOPED NAME", scope.Lookup("a").Type().)

	var body strings.Builder
	for _, node := range wiz.fdecl.Body.List {
		body.WriteString(wiz.convertAst(node))
		body.WriteString("\n")
	}

	variables := make(map[string]string)

	for obj, name := range wiz.definitions {
		variables[name] = obj.Type().String()
	}

	src, err := wiz.Render("function", struct {
		Name         string
		Signature    string
		ReturnType   string
		Body         string
		State        map[string]string
		StateIndices []int
	}{
		Name:         wiz.fdecl.Name.Name,
		Signature:    signature,
		ReturnType:   returnType,
		Body:         body.String(),
		State:        variables,
		StateIndices: wiz.StateIndices(),
	})
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(src))

	if wiz.fdecl.Name.Name == "Range" {
		var funcAst bytes.Buffer
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
		//case *ast.AssignStmt:
	case *ast.Ident:
		/*
			Check defs & uses
			If this is a def - define the var, get the possibly new name
			If a use - get the name based on the uses object
		*/
		definition, exists := wiz.pkg.TypesInfo.Defs[node]
		var name string
		if exists {
			name = wiz.DefineVariable(definition)
		} else {
			usage := wiz.pkg.TypesInfo.Uses[node]
			name = wiz.GetVariable(usage)
		}
		return name
	case *ast.AssignStmt:
		var lhs []string
		for _, expr := range node.Lhs {
			lhs = append(lhs, wiz.convertAst(expr))
		}

		var rhs []string
		for _, expr := range node.Rhs {
			rhs = append(rhs, wiz.convertAst(expr))
		}

		tok := node.Tok.String()
		if tok == ":=" {
			tok = "="
		}

		return strings.Join(lhs, ", ") + " " + tok + " " + strings.Join(rhs, ", ")
	case *ast.BasicLit:
		var lit bytes.Buffer
		err := format.Node(&lit, wiz.pkg.Fset, node)
		if err != nil {
			log.Fatal(err)
		}
		return lit.String()
	case *ast.ForStmt:
		// Easiest case is a forever
		if node.Init == nil && node.Cond == nil && node.Post == nil {
			loopId := wiz.GetLoopId()
			body := wiz.convertAst(node.Body)
			loop, err := wiz.Render("forever", struct {
				Loop int
				Body string
			}{Loop: loopId, Body: body})
			if err != nil {
				log.Fatal(err)
			}
			return string(loop)
		} else {
			// Regula C-style loop!
			loopId := wiz.GetLoopId()
			body := wiz.convertAst(node.Body)
			init := wiz.convertAst(node.Init)
			post := wiz.convertAst(node.Post)
			cond := wiz.convertAst(node.Cond)
			loop, err := wiz.Render("for", struct {
				Init string
				Cond string
				Post string
				Body string
				Loop int
			}{
				Init: init,
				Cond: cond,
				Post: post,
				Body: body,
				Loop: loopId,
			})
			if err != nil {
				log.Fatal(err)
			}
			return string(loop)
		}
	case *ast.BlockStmt:
		block := make([]string, len(node.List))
		for i, stmt := range node.List {
			block[i] = wiz.convertAst(stmt)
		}
		return strings.Join(block, "\n")
	case *ast.BinaryExpr:
		x := wiz.convertAst(node.X)
		y := wiz.convertAst(node.Y)
		tok := node.Op.String()
		return fmt.Sprintf("%s %s %s", x, tok, y)
	case *ast.IncDecStmt:
		x := wiz.convertAst(node.X)
		return fmt.Sprintf("%s%s", x, node.Tok)
	}
	return wiz.Unsupported(node)
}

func (wiz *FuncWizard) GetLoopId() int {
	wiz.loopId++
	return wiz.loopId
}

func (wiz *FuncWizard) Unsupported(node ast.Node) string {
	var syntax bytes.Buffer
	ast.Fprint(&syntax, wiz.pkg.Fset, node, nil)
	var code bytes.Buffer
	format.Node(&code, wiz.pkg.Fset, node)
	return fmt.Sprintf("/*\n%s\n%s\n*/", code.String(), syntax.String())
}
