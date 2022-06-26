package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/tmr232/gengen"
	"go/ast"
	"go/format"
	"go/types"
	"golang.org/x/tools/go/packages"
	"io/ioutil"
	"log"
	"reflect"
	"strings"
)

type TypeInfo struct {
	PkgPath string
	Name    string
}

func (t TypeInfo) String() string {
	return fmt.Sprintf("%s.%s", t.PkgPath, t.Name)
}

var GeneratorType TypeInfo
var YieldType TypeInfo

func init() {
	generatorType := reflect.TypeOf(new(gengen.Generator[struct{}])).Elem()
	name, _, _ := strings.Cut(generatorType.Name(), "[")
	GeneratorType = TypeInfo{
		PkgPath: generatorType.PkgPath(),
		Name:    name,
	}

	YieldType = TypeInfo{
		PkgPath: GeneratorType.PkgPath,
		Name:    "Yield",
	}
}

type generatorDecls struct {
	// pkg is the package we're generating generators for
	pkg *packages.Package
	// decls are the declarations of all functions that return a generator.
	decls []*ast.FuncDecl
}

type IsGenVisitor struct {
	hasYield *bool
	pkg      *packages.Package
}

// Visit checks if gengen.Yield is used inside the given AST.
// It does that by checking if the code uses gengen.Yield in any way.
// There are currently no checks as to how gengen.Yield is used.
func (v *IsGenVisitor) Visit(n ast.Node) ast.Visitor {
	if *v.hasYield {
		// Already found a yield - no need to keep looking!
		return nil
	}
	if ident, isIdent := n.(*ast.Ident); isIdent {
		objectDefinition, exists := v.pkg.TypesInfo.Uses[ident]
		if exists && objectDefinition.Pkg() != nil {
			*v.hasYield = objectDefinition.Pkg().Path() == YieldType.PkgPath && objectDefinition.Name() == YieldType.Name
		}
		return nil
	}
	return v
}

// IsGenerator checks if a given ast.FuncDecl is a generator definition.
func IsGenerator(pkg *packages.Package, fdecl *ast.FuncDecl) (result bool) {
	results := fdecl.Type.Results
	if results == nil || len(results.List) != 1 {
		return false
	}

	// Ensure the return type is a gengen.Generator
	namedType, isNamed := pkg.TypesInfo.Types[results.List[0].Type].Type.(*types.Named)
	if !isNamed || namedType.Obj().Pkg().Path() != GeneratorType.PkgPath || namedType.Obj().Name() != GeneratorType.Name {

		return false
	}

	// Check for usage of gengen.Yield. If it does not exist - the function
	// may just be returning a generator.
	visitor := &IsGenVisitor{&result, pkg}
	ast.Walk(visitor, fdecl)
	return result
}

func formatSource(src []byte) []byte {
	formattedSrc, err := format.Source(src)
	if err != nil {
		// Should never happen, but can arise when developing this code.
		// The user can compile the output to see the error.
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return src
	}
	return formattedSrc
}

type ImportLine struct {
	Name *string
	Path string
}

func (imp ImportLine) String() string {
	if imp.Name != nil {
		return fmt.Sprintf("%s %s", *imp.Name, imp.Path)
	} else {
		return imp.Path
	}
}

type Imports []ImportLine

func (imports Imports) String() string {
	var out bytes.Buffer
	out.WriteString("import (\n")
	for _, imp := range imports {
		fmt.Fprintf(&out, "\t%s\n", imp)
	}
	out.WriteString(")\n")
	return out.String()
}

func isGeneratorSourceFile(file *ast.File) bool {
	if len(file.Comments) == 0 || len(file.Comments[0].List) == 0 {
		return false
	}
	return file.Comments[0].List[0].Text == "//go:build gengen"
}

func getFileImports(file *ast.File) Imports {
	imports := Imports{}
	for _, importSpec := range file.Imports {
		importLine := ImportLine{}
		if importSpec.Name != nil {
			importLine.Name = &importSpec.Name.Name
		}
		importLine.Path = importSpec.Path.Value
		imports = append(imports, importLine)
	}
	return imports
}

func main() {
	dir := "."
	tags := []string{"gengen"}

	cfg := &packages.Config{
		Mode:       packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedSyntax | packages.NeedName | packages.NeedImports,
		Context:    nil,
		Logf:       nil,
		Dir:        dir,
		Env:        nil,
		BuildFlags: []string{fmt.Sprintf("-tags=%s", strings.Join(tags, " "))},
		Fset:       nil,
		ParseFile:  nil,
		Tests:      true,
		Overlay:    nil,
	}

	pkgs, err := packages.Load(cfg)
	if err != nil {
		log.Fatal(err)
	}

	wiz := NewWizard()
	if wiz == nil {
		log.Fatal("Failed to initialize wizard.")
	}

	for _, pkg := range pkgs {
		// TODO: Get rid of the imports, as this is not where we generate them!
		for _, file := range pkg.Syntax {
			// Only copy & modify files that are generator source files.
			if !isGeneratorSourceFile(file) {
				continue
			}

			pkgWiz := wiz.WithPackage(pkg, getFileImports(file))

			out := bytes.Buffer{}

			ast.Inspect(file, func(node ast.Node) bool {
				switch node := node.(type) {
				case *ast.File:
					return true
				case *ast.Ident:
					// The only top-level ast.Ident node is the package name.
					// There is no node for the package definition, so it's just a naked ast.Ident.
					res, err := wiz.Render("package", struct{ PackageName string }{node.Name})
					if err != nil {
						log.Fatal("Failed to render package header.")
					}
					out.Write(res)
					return false
				case *ast.FuncDecl:
					if IsGenerator(pkg, node) {
						out.Write(pkgWiz.WithFunction(node).convertFunction())
						out.WriteString("\n")
					} else {
						format.Node(&out, pkg.Fset, node)
						out.WriteString("\n")
					}
					return false
				default:
					format.Node(&out, pkg.Fset, node)
					out.WriteString("\n")
					return false
				}
			})

			src := formatSource(out.Bytes())

			filepath := strings.TrimSuffix(pkg.Fset.Position(file.Pos()).Filename, ".go") + "_gengen.go"
			err = ioutil.WriteFile(filepath, src, 0644)
			if err != nil {
				log.Fatalf("writing output: %s", err)
			}
		}
	}
}
