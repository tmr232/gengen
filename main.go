package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/ast"
	"go/format"
	"golang.org/x/tools/go/packages"
	"io/ioutil"
	"log"
	"path"
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
		Mode:       packages.NeedTypes | packages.NeedTypesInfo | packages.NeedFiles | packages.NeedSyntax | packages.NeedName,
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
					if pkg.TypesInfo.Types[results.List[0].Type].Type.String() != "github.com/tmr232/gengen/gengen.Generator[int]" {
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

type Wizard struct {
	template *template.Template
}

func NewWizard() *Wizard {
	t, err := template.New("core").Parse(coreTemplate)
	if err != nil {
		log.Fatal(err)
	}
	return &Wizard{template: t}
}

func (wiz *Wizard) Render(name string, data any) ([]byte, error) {
	var out bytes.Buffer
	err := wiz.template.ExecuteTemplate(&out, name, data)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
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

func main() {
	dir := "./sample"
	buildTag := "gengen"

	generatorDefs := getGeneratorDefinitions(dir, []string{buildTag})
	for _, genDecls := range generatorDefs {
		fmt.Println(genDecls.pkg.ID, genDecls.pkg.Name)
		for _, fdef := range genDecls.decls {
			fmt.Println("    ", fdef.Name.Name)
		}
	}

	wiz := NewWizard()
	if wiz == nil {
		log.Fatal("Failed to initialize wizard.")
	}

	for _, genDef := range generatorDefs {
		// For starters - let's just create a file!
		src, err := wiz.Render("package", struct{ PackageName string }{genDef.pkg.Name})
		src = formatSource(src)
		if err != nil {
			log.Fatal(err)
		}
		filepath := path.Join(dir, genDef.pkg.Name+"_gengen.go")
		err = ioutil.WriteFile(filepath, src, 0644)
		if err != nil {
			log.Fatalf("writing output: %s", err)
		}
	}
}
