package main

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"log"
	"strings"
)

func parsePackage() *packages.Package {
	cfg := &packages.Config{
		Mode:    packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedSyntax,
		Context: nil,
		Logf:    nil,
		//Dir:        "C:\\Code\\Personal\\gengen",
		Dir:        "C:\\Code\\Personal\\go-explore\\itertools",
		Env:        nil,
		BuildFlags: nil,
		Fset:       nil,
		ParseFile:  nil,
		Tests:      true,
		Overlay:    nil,
	}

	pkgs, err := packages.Load(cfg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(pkgs)
	printFunInfo(pkgs)

	if len(pkgs) != 1 {
		log.Fatalf("error: %d packages found", len(pkgs))
	}
	return pkgs[0]
}

type Args struct {
	Path string `alias:"path"`
}

func start(args Args) error {
	fmt.Println(args.Path)
	return nil
}

func printFunInfo(pkgs []*packages.Package) {
	for _, pkg := range pkgs {
		fmt.Println("Package name: ", pkg.ID)
		for _, f := range pkg.Syntax {
			fmt.Println("Syntax name: ", f.Name.Name)

			for _, decl := range f.Decls {
				switch decl := decl.(type) {
				case *ast.FuncDecl:
					fmt.Println("Function: ", decl.Name)
				}
			}
		}
	}
}

func getGeneratorDefinitions(dir string, tags []string) []*ast.FuncDecl {
	cfg := &packages.Config{
		Mode:       packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedSyntax,
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

	var generatorFunctionDefs []*ast.FuncDecl
	// TODO: We'll need to separate functions into packages for generation.
	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			for _, decl := range f.Decls {
				switch decl := decl.(type) {
				case *ast.FuncDecl:
					results := decl.Type.Results
					if results == nil {
						continue
					}
					if len(results.List) != 1 {
						continue
					}
					if pkg.TypesInfo.Types[results.List[0].Type].Type.String() != "github.com/tmr232/gengen.Generator[int]" {
						continue
					}
					generatorFunctionDefs = append(generatorFunctionDefs, decl)
				}
			}
		}
	}

	return generatorFunctionDefs
}

func main() {
	generatorDefs := getGeneratorDefinitions(".", []string{"gengen"})
	for _, fdef := range generatorDefs {
		fmt.Println(fdef.Name.Name)
	}
}
