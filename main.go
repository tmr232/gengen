package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"log"
	"strings"
	"text/template"
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

type generatorDecls struct {
	pkg   *packages.Package
	decls []*ast.FuncDecl
}

func getGeneratorDefinitions(dir string, tags []string) []generatorDecls {
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

	var generatorFunctionDefs []generatorDecls
	for _, pkg := range pkgs {
		var decls []*ast.FuncDecl
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
					decls = append(decls, decl)
				}
			}
		}
		if len(decls) > 0 {
			generatorFunctionDefs = append(generatorFunctionDefs, generatorDecls{
				pkg:   pkg,
				decls: decls,
			})
		}
	}

	return generatorFunctionDefs
}

//go:embed gengen.tmpl
var coreTemplate string

func funWithTemplates() {
	t, err := template.New("package").Parse(coreTemplate)
	if err != nil {
		log.Fatal(err)
	}
	var out bytes.Buffer
	err = t.ExecuteTemplate(&out, "package", map[string]string{"packageName": "test"})
	if err != nil {
		fmt.Println(out.String())
	}
	err = t.ExecuteTemplate(&out, "sub", nil)
	if err != nil {
		fmt.Println(out.String())
	}

}

func main() {
	generatorDefs := getGeneratorDefinitions(".", []string{"gengen"})
	for _, genDecls := range generatorDefs {
		fmt.Println(genDecls.pkg.ID, genDecls.pkg.Name)
		for _, fdef := range genDecls.decls {
			fmt.Println("    ", fdef.Name.Name)
		}
	}

	funWithTemplates()
}
