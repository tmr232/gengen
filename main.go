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

func getGeneratorDefinitions(dir string, tags []string) {
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

	for _, pkg := range pkgs {
		fmt.Println("Package name: ", pkg.ID)
		for _, f := range pkg.Syntax {
			fmt.Println("Syntax name: ", f.Name.Name)

			for _, decl := range f.Decls {
				switch decl := decl.(type) {
				case *ast.FuncDecl:
					fmt.Println("Function: ", decl.Name)
					results := decl.Type.Results
					if results == nil {
						continue
					}
					for _, result := range results.List {
						fmt.Println("Result type: ", pkg.TypesInfo.Types[result.Type].Type.String())
					}
				}
			}
		}

		//for id, obj := range pkg.TypesInfo.Defs {
		//	fmt.Printf("%s: %q defines %v\n",
		//		pkg.Fset.Position(id.Pos()), id.Name, obj)
		//}
		//for id, obj := range pkg.TypesInfo.Uses {
		//	fmt.Printf("%s: %q uses %v\n",
		//		pkg.Fset.Position(id.Pos()), id.Name, obj)
		//}
	}
}

func main() {
	getGeneratorDefinitions(".", []string{"gengen"})
}
