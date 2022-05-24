package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"github.com/tmr232/gengen/gengen"
	"go/ast"
	"go/format"
	"golang.org/x/tools/go/packages"
	"io/ioutil"
	"log"
	"path"
	"strings"
)

type generatorDecls struct {
	// pkg is the package we're generating generators for
	pkg *packages.Package
	// decls are the declarations of all functions that return a generator.
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

					// TODO: Use proper type information here.
					if !strings.HasPrefix(
						pkg.TypesInfo.Types[results.List[0].Type].Type.String(),
						"github.com/tmr232/gengen/gengen.Generator[",
					) {
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

func collectImports(pkg *packages.Package) Imports {
	var imports Imports
	for _, f := range pkg.Syntax {
		for _, imp := range f.Imports {
			if imp.Name != nil {
				imports = append(imports, ImportLine{Name: &imp.Name.Name, Path: imp.Path.Value})
			} else {
				imports = append(imports, ImportLine{Path: imp.Path.Value})
			}
		}
	}
	return imports
}

type ImportPathsVisitor struct {
	paths map[string]bool
	pkg   *packages.Package
}

func (v ImportPathsVisitor) Visit(n ast.Node) ast.Visitor {
	ident, isIdent := n.(*ast.Ident)
	if isIdent {
		object, exists := v.pkg.TypesInfo.Uses[ident]
		if exists && object.Pkg() != nil {
			v.paths[object.Pkg().Path()] = true
		}
		return nil
	}
	return v
}

func FindUsedImports(genDef generatorDecls) map[string]bool {
	paths := make(map[string]bool)
	visitor := ImportPathsVisitor{
		paths: paths,
		pkg:   genDef.pkg,
	}

	for _, node := range genDef.decls {
		ast.Walk(visitor, node)
	}

	return paths
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
		finalImports := CollectUsedImports(genDef)
		functions := wiz.WithPackage(genDef.pkg).convertFunctions(genDef)
		src, err := wiz.Render("package",
			struct {
				PackageName string
				Imports     Imports
				Functions   []string
			}{
				PackageName: genDef.pkg.Name,
				Imports:     finalImports,
				Functions:   functions,
			})
		src = formatSource(src)
		if err != nil {
			log.Fatal(err)
		}

		filepath := path.Join(dir, genDef.pkg.Name+"_gengen.go")
		err = ioutil.WriteFile(filepath, src, 0644)
		if err != nil {
			log.Fatalf("writing output: %s", err)
		}
		//fmt.Println(string(src))
	}

	sample := sampleGenerator()
	for sample.Next() {
		fmt.Println(sample.Value())
	}
	if sample.Error() != nil {
		fmt.Println("Oh no! Error!")
	}
}

func CollectUsedImports(genDef generatorDecls) Imports {
	// First, find all the imports
	imports := collectImports(genDef.pkg)
	// Then, all the imports used by the generators
	usedImports := FindUsedImports(genDef)
	// And only use the intersection!
	var finalImports Imports
	for _, importLine := range imports {
		if usedImports[strings.Trim(importLine.Path, "\"")] {
			finalImports = append(finalImports, importLine)
		}
	}

	fmt.Println("Used Imports ", finalImports)
	return finalImports
}

func sampleGenerator() gengen.Generator[int] {
	done := false
	return &gengen.GeneratorFunction[int]{Advance: func() (hasValue bool, value int, err error) {
		if done {
			return
		}
		done = true
		return true, 42, nil
	}}
}
