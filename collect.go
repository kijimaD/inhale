package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
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

	return refs, nil
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

func isGoFile(f os.FileInfo) bool {
	name := f.Name()

	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}
