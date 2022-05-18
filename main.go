package main

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"log"
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

func main() {
	pkg := parsePackage()
	fmt.Println(pkg.Types.Name())
	fmt.Println(pkg.TypesInfo.Defs)
	thing := pkg.Syntax[0].Decls[3].(*ast.FuncDecl).Body.List[1].(*ast.ReturnStmt).Results[0]
	fmt.Println(pkg.TypesInfo.TypeOf(thing))
	//cli.MakeCliApp(goat.App("Gengen", goat.Action(start))).Run(os.Args)
}
