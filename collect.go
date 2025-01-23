package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kijimaD/inhale/stdlib"
)

// https://github.com/kd-collective/tools/blob/1261a24ceb1867ea7439eda244e53e7ace4ad777/internal/imports/fix.go#L152

type (
	PackageName = string
	Symbol      = string

	References = map[PackageName]map[Symbol]int
)

type visitFn func(node ast.Node) ast.Visitor

func (fn visitFn) Visit(node ast.Node) ast.Visitor {
	return fn(node)
}

func Run(w io.Writer, path string) {
	refs, err := walkDir(path)
	if err != nil {
		log.Fatal(err)
	}

	bs, err := json.Marshal(refs)
	if err != nil {
		log.Fatal(err)
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, bs, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(w, prettyJSON.String())
}

func walkDir(path string) (References, error) {
	refs := References{}
	visitFile := func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !isGoFile(f) {
			return err
		}

		fo, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fo.Close()
		var buf bytes.Buffer
		io.Copy(&buf, fo)

		fset := token.NewFileSet()
		fa, err := parser.ParseFile(fset, path, buf.String(), parser.AllErrors)
		if err != nil {
			return err
		}
		refs = collectReferences(fa, refs)

		return nil
	}
	err := filepath.Walk(path, visitFile)
	if err != nil {
		return nil, err
	}
	refs = filterStdLib(refs)

	return refs, nil
}

func isGoFile(f os.FileInfo) bool {
	name := f.Name()

	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func collectReferences(f *ast.File, refs References) References {
	var visitor visitFn
	visitor = func(node ast.Node) ast.Visitor {
		if node == nil {
			return visitor
		}
		switch v := node.(type) {
		case *ast.SelectorExpr:
			xident, ok := v.X.(*ast.Ident)
			if !ok {
				break
			}
			if xident.Obj != nil {
				// If the parser can resolve it, it's not a package ref.
				break
			}
			if !ast.IsExported(v.Sel.Name) {
				// Whatever this is, it's not exported from a package.
				break
			}
			pkgName := xident.Name
			r := refs[pkgName]
			if r == nil {
				r = make(map[string]int)
				refs[pkgName] = r
			}
			r[v.Sel.Name] += 1
		}
		return visitor
	}
	ast.Walk(visitor, f)

	return refs
}

// 標準ライブラリだけ残す
func filterStdLib(refs References) References {
	stdrefs := References{}
	for k, stds := range stdlib.PackageSymbols {
		stdbase := path.Base(k)
		if ref, ok := refs[stdbase]; ok {
			newpkgRef := map[string]int{}
			for method, count := range ref {
				for _, std := range stds {
					if method == std.Name {
						newpkgRef[method] = count
					}
				}
			}
			if len(newpkgRef) > 0 {
				stdrefs[k] = newpkgRef
			}
		}
	}

	return stdrefs
}

type ImportInfo struct {
	ImportPath string // import path, e.g. "crypto/rand".
	AliasName  string // import name, e.g. "crand", or "" if none.
}

func collectImports(f *ast.File) []ImportInfo {
	var imports []ImportInfo
	for _, imp := range f.Imports {
		var name string
		if imp.Name != nil {
			name = imp.Name.Name
		}
		if imp.Path.Value == `"C"` || name == "_" || name == "." {
			continue
		}
		path := strings.Trim(imp.Path.Value, `"`)
		imports = append(imports, ImportInfo{
			ImportPath: path,
			AliasName:  name,
		})
	}

	return imports
}
