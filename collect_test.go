package main

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectReferences(t *testing.T) {
	src := `package main

  import (
          "fmt"
          "strings"
  )

  func main() {
          fmt.Println("hello")

          title := strings.Title("hello world")
          rep := strings.Repeat("a", 10)
          fmt.Println(title, rep)
  }`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "dummy.go", src, parser.AllErrors)
	assert.NoError(t, err)

	refs := References{}
	collectReferences(f, refs)
	expected := map[string]map[string]int(map[string]map[string]int{
		"fmt":     map[string]int{"Println": 2},
		"strings": map[string]int{"Repeat": 1, "Title": 1}},
	)
	assert.Equal(t, expected, refs)
}

func TestWalkDir(t *testing.T) {
	refs, err := walkDir("./testdata")
	assert.NoError(t, err)

	// TODO: パッケージ名をインポート名に変える
	expect := map[string]map[string]int{
		"ast":     map[string]int{"File": 3, "FuncDecl": 1},
		"astutil": map[string]int{"Imports": 1},
		"bufio":   map[string]int{"NewReader": 1},
		"bytes":   map[string]int{"Buffer": 3, "IndexByte": 1, "LastIndex": 1, "NewReader": 1, "ReplaceAll": 1},
		"context": map[string]int{"Context": 1},
		"event":   map[string]int{"Start": 1},
		"fmt":     map[string]int{"Fprint": 1},
		"format":  map[string]int{"Source": 1},
		"io":      map[string]int{"EOF": 1, "Reader": 1},
		"os":      map[string]int{"Exit": 1},
		"parser":  map[string]int{"AllErrors": 3, "Mode": 3, "ParseComments": 3, "ParseFile": 4, "SkipObjectResolution": 2},
		"printer": map[string]int{"Config": 1, "TabIndent": 1, "UseSpaces": 1},
		"regexp":  map[string]int{"MustCompile": 1},
		"strconv": map[string]int{"Unquote": 1},
		"strings": map[string]int{"Contains": 2, "HasPrefix": 5},
		"testenv": map[string]int{"ExitIfSmallMachine": 1},
		"testing": map[string]int{"M": 1},
		"token":   map[string]int{"FileSet": 2, "NewFileSet": 3},
	}
	assert.Equal(t, expect, refs)
}
