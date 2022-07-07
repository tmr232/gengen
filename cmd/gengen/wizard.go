package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
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

// Return a mapping between package paths and imports, and a set of all the import names
func createImportNameMapping(imports Imports) (map[string]string, map[string]bool) {
	mapping := make(map[string]string)
	importNames := make(map[string]bool)
	for _, importLine := range imports {
		packagePath := strings.Trim(importLine.Path, "\"")
		if importLine.Name != nil {
			name := *importLine.Name
			mapping[packagePath] = name
			importNames[name] = true
		} else {
			parts := strings.Split(packagePath, "/")
			name := parts[len(parts)-1]
			mapping[packagePath] = name
			importNames[name] = true
		}
	}
	return mapping, importNames
}

func (wiz *Wizard) WithPackage(pkg *packages.Package, imports Imports) *PkgWizard {
	importMapping, importNames := createImportNameMapping(imports)
	return &PkgWizard{
		Wizard:      *wiz,
		pkg:         pkg,
		imports:     importMapping,
		importNames: importNames,
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
	// A mapping between package paths to import names
	imports map[string]string

	// A set of all the imported names
	importNames map[string]bool
}

func (wiz *PkgWizard) WithFunction(fdecl *ast.FuncDecl) *FuncWizard {
	return &FuncWizard{
		PkgWizard:   *wiz,
		fdecl:       fdecl,
		definitions: make(map[types.Object]string),
		variables:   make(map[types.Object]string),
		names:       make(map[string]bool),
	}
}

type Block struct {
	seenReturn bool
}

type FuncWizard struct {
	PkgWizard
	fdecl       *ast.FuncDecl
	maxState    int
	definitions map[types.Object]string
	variables   map[types.Object]string
	names       map[string]bool
	jumpId      int
	adapterId   int
	extraState  []string
	loopStack   []LoopFrame
	blockStack  []Block
}

func (wiz *FuncWizard) EnterBlock() *FuncWizard {
	wiz.blockStack = append(wiz.blockStack, Block{})
	return wiz
}

func (wiz *FuncWizard) LeaveBlock() {
	wiz.blockStack = wiz.blockStack[:len(wiz.blockStack)-1]
}

func (wiz *FuncWizard) MarkReturn() {
	wiz.blockStack[len(wiz.blockStack)-1].seenReturn = true
}
func (wiz *FuncWizard) AfterReturn() bool {
	return wiz.blockStack[len(wiz.blockStack)-1].seenReturn
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

func (wiz *FuncWizard) AddFunctionArgument(obj types.Object) {
	wiz.variables[obj] = obj.Name()
	wiz.names[obj.Name()] = true
}

func (wiz *FuncWizard) DefineVariable(obj types.Object) (name string) {
	if obj.Name() == "_" {
		return "_"
	}
	namer := Namer{name: obj.Name()}
	for wiz.names[namer.Name()] {
		namer.Next()
	}
	wiz.names[namer.Name()] = true
	wiz.definitions[obj] = namer.Name()
	wiz.variables[obj] = namer.Name()
	return namer.Name()
}

func (wiz *FuncWizard) GetVariable(obj types.Object) (name string) {
	if obj.Name() == "_" {
		return "_"
	}
	name, exists := wiz.variables[obj]
	if exists {
		return name
	}
	isPackageName := wiz.importNames[obj.Name()]
	if isPackageName {
		return obj.Name()
	}

	panic(fmt.Sprintf("No variable for %s", obj.Name()))
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
	// A function is a block too :)
	defer wiz.EnterBlock().LeaveBlock()
	var out bytes.Buffer
	format.Node(&out, wiz.pkg.Fset, wiz.fdecl.Type)
	signature := out.String()

	// We only allow a single result
	if len(wiz.fdecl.Type.Results.List) != 1 {
		log.Fatalf("Expected a single result, got %d", len(wiz.fdecl.Type.Results.List))
	}

	generatorType := wiz.pkg.TypesInfo.TypeOf(wiz.fdecl.Type.Results.List[0].Type)
	namedType, isNamedType := generatorType.(*types.Named)
	if !isNamedType {
		panic("Generators only support named types.")
	}
	generatorItemType := namedType.TypeArgs().At(0)
	returnType := wiz.getTypeName(generatorItemType)

	//// We go through all the defs in the function,
	//// and define the relevant variables.
	//scope := wiz.pkg.TypesInfo.Scopes[wiz.fdecl.Type]
	//for _, name := range scope.Names() {
	//	fmt.Println(name, scope.Lookup(name).Type().String())
	//}
	////fmt.Println("SCOPED NAME", scope.Lookup("a").Type().)

	// Add all function arguments to the function
	// Otherwise - we won't have names for them!
	for _, param := range wiz.fdecl.Type.Params.List {
		for _, name := range param.Names {
			def := wiz.pkg.TypesInfo.Defs[name]
			wiz.AddFunctionArgument(def)
		}
	}

	var body strings.Builder
	for _, node := range wiz.fdecl.Body.List {
		body.WriteString(wiz.convertAst(node))
		body.WriteString("\n")
	}

	variables := make(map[string]string)

	for obj, name := range wiz.definitions {

		variables[name] = wiz.getTypeName(obj.Type())
	}

	src, err := wiz.Render("function", struct {
		Name         string
		Signature    string
		ReturnType   string
		Body         string
		State        map[string]string
		StateIndices []int
		ExtraState   []string
	}{
		Name:         wiz.fdecl.Name.Name,
		Signature:    signature,
		ReturnType:   returnType,
		Body:         body.String(),
		State:        variables,
		StateIndices: wiz.StateIndices(),
		ExtraState:   wiz.extraState,
	})
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(string(src))

	return src
}

func (wiz *FuncWizard) getTypeName(typ types.Type) string {
	if namedType, isNamedType := typ.(*types.Named); isNamedType {
		if namedType.Obj().Pkg() == nil {
			return namedType.Obj().Name()
		}
		if namedType.Obj().Pkg() == wiz.pkg.Types {
			// The type was defined in the same package, so we don't need to name the package
			// it was imported from.
			return namedType.Obj().Name()
		}
		packageName, exists := wiz.imports[namedType.Obj().Pkg().Path()]
		if exists {
			return fmt.Sprintf("%s.%s", packageName, namedType.Obj().Name())
		}
	}
	return typ.String()
}

func (wiz *FuncWizard) GenericAstVisitor() string { return "" }
func (wiz *FuncWizard) VisitReturnStmt(node *ast.ReturnStmt) string {
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
	wiz.MarkReturn()
	return string(returnStatement)
}
func (wiz *FuncWizard) VisitCallExpr(node *ast.CallExpr) string {
	if fun, isSelectorExpr := node.Fun.(*ast.SelectorExpr); isSelectorExpr {
		object := wiz.pkg.TypesInfo.Uses[fun.Sel]
		funcObject, isFunc := object.(*types.Func)
		if isFunc && funcObject.FullName() == YieldType.String() {
			// If we're after a return statement, we ignore this yield.
			if wiz.AfterReturn() {
				return ""
			}
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
	}
	args := make([]string, len(node.Args))
	for i := range args {
		args[i] = wiz.convertAst(node.Args[i])
	}
	return wiz.convertAst(node.Fun) + "(" + strings.Join(args, ", ") + ")"

}
func (wiz *FuncWizard) VisitExprStmt(node *ast.ExprStmt) string {
	return wiz.convertAst(node.X)
}
func (wiz *FuncWizard) VisitIdent(node *ast.Ident) string {
	/*
		Check defs & uses
		If this is a def - define the var, get the possibly new name
		If a use - get the name based on the uses object
	*/
	if node.Name == "nil" {
		// nil behaves a bit odd, so we handle it here.
		return node.String()
	}
	definition, exists := wiz.pkg.TypesInfo.Defs[node]
	if exists {
		return wiz.DefineVariable(definition)
	}
	usage, exists := wiz.pkg.TypesInfo.Uses[node]
	if _, isBuiltin := usage.(*types.Builtin); isBuiltin {
		return node.String()
	}
	if exists && usage.Pkg().Path() == wiz.pkg.PkgPath {
		return wiz.GetVariable(usage)
	}
	return node.String()
}
func (wiz *FuncWizard) VisitAssignStmt(node *ast.AssignStmt) string {
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
}
func (wiz *FuncWizard) VisitBasicLit(node *ast.BasicLit) string {
	var lit bytes.Buffer
	err := format.Node(&lit, wiz.pkg.Fset, node)
	if err != nil {
		log.Fatal(err)
	}
	return lit.String()
}

func (wiz *FuncWizard) VisitForStmt(node *ast.ForStmt) string {
	// Easiest case is a forever
	defer wiz.EnterLoop().ExitLoop()
	if node.Init == nil && node.Cond == nil && node.Post == nil {
		body := wiz.convertAst(node.Body)
		loop, err := wiz.Render("forever", struct {
			Loop LoopFrame
			Body string
		}{Loop: *wiz.GetLoopFrame(), Body: body})
		if err != nil {
			log.Fatal(err)
		}
		return string(loop)
	} else {
		// Regular C-style loop!
		init := wiz.convertAst(node.Init)
		cond := wiz.convertAst(node.Cond)
		post := wiz.convertAst(node.Post)
		body := wiz.convertAst(node.Body)
		loop, err := wiz.Render("for", struct {
			Init string
			Cond string
			Post string
			Body string
			Loop LoopFrame
		}{
			Init: init,
			Cond: cond,
			Post: post,
			Body: body,
			Loop: *wiz.GetLoopFrame(),
		})
		if err != nil {
			log.Fatal(err)
		}
		return string(loop)
	}
}
func (wiz *FuncWizard) VisitBlockStmt(node *ast.BlockStmt) string {
	defer wiz.EnterBlock().LeaveBlock()
	block := make([]string, len(node.List))
	for i, stmt := range node.List {
		block[i] = wiz.convertAst(stmt)
	}
	return strings.Join(block, "\n")
}
func (wiz *FuncWizard) VisitBinaryExpr(node *ast.BinaryExpr) string {
	x := wiz.convertAst(node.X)
	y := wiz.convertAst(node.Y)
	tok := node.Op.String()
	return fmt.Sprintf("%s %s %s", x, tok, y)
}
func (wiz *FuncWizard) VisitIncDecStmt(node *ast.IncDecStmt) string {
	x := wiz.convertAst(node.X)
	return fmt.Sprintf("%s%s", x, node.Tok)
}
func (wiz *FuncWizard) VisitDeclStmt(node *ast.DeclStmt) string {
	switch decl := node.Decl.(type) {
	case *ast.GenDecl:
		for _, spec := range decl.Specs {
			switch spec := spec.(type) {
			case *ast.ImportSpec:
				log.Fatal("There shouldn't be an import here!")
			case *ast.TypeSpec:
				log.Fatal("Neither should we have nested type defs")
			case *ast.ValueSpec:
				var assignments []string
				for i, name := range spec.Names {
					// First, we need to define the matching variable
					realName := wiz.DefineVariable(wiz.pkg.TypesInfo.Defs[name])

					// Then, if a value exists, we create an assignment
					if spec.Values != nil {
						realValue := wiz.convertAst(spec.Values[i])
						assignments = append(assignments, fmt.Sprintf("%s = %s", realName, realValue))
					}
				}
				return strings.Join(assignments, "\n")
			}
		}
	case *ast.FuncDecl:
		log.Fatal("Nested functions are currently unsupported.")
	}
	return ""
}
func (wiz *FuncWizard) VisitIfStmt(node *ast.IfStmt) string {
	loopId := wiz.GetIfId()
	init := wiz.convertAst(node.Init)
	body := wiz.convertAst(node.Body)
	else_ := wiz.convertAst(node.Else)
	cond := wiz.convertAst(node.Cond)
	if_, err := wiz.Render("if", struct {
		Init string
		Cond string
		Else string
		Body string
		If   int
	}{
		Init: init,
		Cond: cond,
		Else: else_,
		Body: body,
		If:   loopId,
	})
	if err != nil {
		log.Fatal(err)
	}
	return string(if_)
}
func (wiz *FuncWizard) VisitRangeStmt(node *ast.RangeStmt) string {
	rangeType := wiz.pkg.TypesInfo.TypeOf(node.X)
	defer wiz.EnterLoop().ExitLoop()
	switch rangeType := rangeType.(type) {
	case *types.Map:
		x := wiz.convertAst(node.X)
		keyType := rangeType.Key()
		valueType := rangeType.Elem()
		mapAdapterId := wiz.GetAdapterId()
		adapterName := fmt.Sprintf("__mapAdapter%d", mapAdapterId)
		mapAdapterDefinition := fmt.Sprintf("var %s gengen.Generator2[%s, %s]", adapterName, keyType, wiz.getTypeName(valueType))
		wiz.AddStateLine(mapAdapterDefinition)
		key := "_"
		value := "_"
		if node.Key != nil {
			key = wiz.convertAst(node.Key)
		}
		if node.Value != nil {
			value = wiz.convertAst(node.Value)
		}
		body := wiz.convertAst(node.Body)
		forLoop, err := wiz.Render("for-range-map", struct {
			Adapter   string
			Key       string
			KeyType   string
			Value     string
			ValueType string
			Map       string
			Loop      LoopFrame
			Body      string
		}{
			Adapter:   adapterName,
			Key:       key,
			KeyType:   keyType.String(),
			Value:     value,
			ValueType: wiz.getTypeName(valueType),
			Map:       x,
			Loop:      *wiz.GetLoopFrame(),
			Body:      body,
		})
		if err != nil {
			log.Fatal(err)
		}
		return string(forLoop)
	case *types.Slice, *types.Array:
		x := wiz.convertAst(node.X)
		valueType := rangeType.(interface{ Elem() types.Type }).Elem()
		mapAdapterId := wiz.GetAdapterId()
		adapterName := fmt.Sprintf("__sliceAdapter%d", mapAdapterId)
		mapAdapterDefinition := fmt.Sprintf("var %s gengen.Generator2[int, %s]", adapterName, wiz.getTypeName(valueType))
		wiz.AddStateLine(mapAdapterDefinition)
		key := "_"
		value := "_"
		if node.Key != nil {
			key = wiz.convertAst(node.Key)
		}
		if node.Value != nil {
			value = wiz.convertAst(node.Value)
		}
		body := wiz.convertAst(node.Body)
		name := wiz.getTypeName(valueType)
		forLoop, err := wiz.Render("for-range-slice", struct {
			Adapter   string
			Key       string
			KeyType   string
			Value     string
			ValueType string
			Slice     string
			Loop      LoopFrame
			Body      string
		}{
			Adapter:   adapterName,
			Key:       key,
			KeyType:   "int",
			Value:     value,
			ValueType: name,
			Slice:     x,
			Loop:      *wiz.GetLoopFrame(),
			Body:      body,
		})
		if err != nil {
			log.Fatal(err)
		}
		return string(forLoop)
	default:
		return wiz.Unsupported(node)
	}

}
func (wiz *FuncWizard) VisitUnaryExpr(node *ast.UnaryExpr) string {

	return node.Op.String() + wiz.convertAst(node.X)
}
func (wiz *FuncWizard) VisitSelectorExpr(node *ast.SelectorExpr) string {
	// The selector is always an ast.Ident, so we just take the name.
	// We cannot use the VisitIdent method as the selector does not represent a variable,
	// only an identifier.
	expr := wiz.convertAst(node.X) + "." + node.Sel.Name
	return expr
}
func (wiz *FuncWizard) VisitBranchStmt(node *ast.BranchStmt) string {
	if node.Label != nil {
		return wiz.Unsupported(node)
	}
	switch node.Tok {
	case token.BREAK:
		wiz.UseBreak()
		return fmt.Sprintf("goto __After%d", wiz.GetLoopId())
	case token.CONTINUE:
		wiz.UseContinue()
		return fmt.Sprintf("goto __Continue%d", wiz.GetLoopId())
	}
	return wiz.Unsupported(node)
}

func (wiz *FuncWizard) VisitCompositeLit(node *ast.CompositeLit) string {
	elts := make([]string, len(node.Elts))
	for i, elt := range node.Elts {
		elts[i] = wiz.convertAst(elt)
	}

	return wiz.convertAst(node.Type) + "{" + strings.Join(elts, ", ") + "}"

}
func (wiz *FuncWizard) VisitArrayType(node *ast.ArrayType) string {
	var arrType bytes.Buffer
	err := format.Node(&arrType, wiz.pkg.Fset, node)
	if err != nil {
		log.Fatal(err)
	}
	return arrType.String()
}
func (wiz *FuncWizard) convertAst(node ast.Node) string {
	if node == nil {
		// This saves some work with conditionally-nil nodes.
		// E.g. ast.IfStmt.Init
		return ""
	}
	visitor := GenericVisit[string](wiz, node)
	if visitor != nil {
		return visitor()
	} else {
		return wiz.Unsupported(node)
	}
}

func (wiz *FuncWizard) GetLoopFrame() *LoopFrame {
	return &wiz.loopStack[len(wiz.loopStack)-1]
}
func (wiz *FuncWizard) GetLoopId() int {
	return wiz.GetLoopFrame().Id
}

func (wiz *FuncWizard) UseContinue() {
	wiz.GetLoopFrame().HasContinue = true
}
func (wiz *FuncWizard) UseBreak() {
	wiz.GetLoopFrame().HasBreak = true
}

type LoopFrame struct {
	Id          int
	HasContinue bool
	HasBreak    bool
}

func (wiz *FuncWizard) EnterLoop() *FuncWizard {
	wiz.jumpId++
	loopId := wiz.jumpId
	wiz.loopStack = append(wiz.loopStack, LoopFrame{Id: loopId})
	return wiz
}

func (wiz *FuncWizard) ExitLoop() {
	wiz.loopStack = wiz.loopStack[0 : len(wiz.loopStack)-1]
}

func (wiz *FuncWizard) GetIfId() int {
	wiz.jumpId++
	return wiz.jumpId
}

func (wiz *FuncWizard) Unsupported(node ast.Node) string {
	var syntax bytes.Buffer
	ast.Fprint(&syntax, wiz.pkg.Fset, node, nil)
	var code bytes.Buffer
	format.Node(&code, wiz.pkg.Fset, node)
	return fmt.Sprintf("/*\n%s\n%s\n*/", code.String(), syntax.String())
}

func (wiz *FuncWizard) GetAdapterId() int {
	wiz.adapterId++
	return wiz.adapterId
}

func (wiz *FuncWizard) AddStateLine(definition string) {
	wiz.extraState = append(wiz.extraState, definition)
}
