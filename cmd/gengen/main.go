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

func usesYield(pkg *packages.Package, node ast.Node) bool {
	usesYield := false
	ast.Inspect(node, func(node ast.Node) bool {
		if usesYield {
			return false
		}
		if ident, isIdent := node.(*ast.Ident); isIdent {
			objectDefinition, exists := pkg.TypesInfo.Uses[ident]
			if exists && objectDefinition.Pkg() != nil {
				usesYield = objectDefinition.Pkg().Path() == YieldType.PkgPath && objectDefinition.Name() == YieldType.Name
			}
			return false
		}
		return true
	})
	return usesYield
}

// IsGenerator checks if a given ast.FuncDecl is a generator definition.
func IsGenerator(pkg *packages.Package, fdecl *ast.FuncDecl) (result bool) {
	results := fdecl.Type.Results
	if results == nil || len(results.List) != 1 {
		return false
	}

	// Ensure the return type is a gengen.Generator
	namedType, isNamed := pkg.TypesInfo.Types[results.List[0].Type].Type.(*types.Named)
	if !isNamed || namedType.Obj().Pkg() == nil || namedType.Obj().Pkg().Path() != GeneratorType.PkgPath || namedType.Obj().Name() != GeneratorType.Name {

		return false
	}

	// Check for usage of gengen.Yield. If it does not exist - the function
	// may just be returning a generator.
	return usesYield(pkg, fdecl)
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

func renderGeneratorFile(wiz *Wizard, pkg *packages.Package, file *ast.File) []byte {
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
	return src
}

func loadPackages(dir string, tags ...string) ([]*packages.Package, error) {
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
	return pkgs, err
}

func main() {
	dir := "."
	buildTag := "gengen"

	pkgs, err := loadPackages(dir, buildTag)

	wiz := NewWizard()
	if wiz == nil {
		log.Fatal("Failed to initialize wizard.")
	}

	log.Println("Generating Generators!")

	visited := make(map[*ast.File]bool)

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			if visited[file] {
				continue
			}
			visited[file] = true

			if !isGeneratorSourceFile(file) {
				// Only copy & modify files that are generator source files.
				continue
			}

			sourcePath := pkg.Fset.Position(file.Pos()).Filename
			genPath := strings.TrimSuffix(sourcePath, ".go") + "_gengen.go"

			log.Printf("\t%s -> %s\n", sourcePath, genPath)

			src := renderGeneratorFile(wiz, pkg, file)

			err = ioutil.WriteFile(genPath, src, 0644)
			if err != nil {
				log.Fatalf("writing output: %s", err)
			}
		}
	}
}
