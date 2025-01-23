package main

import (
	"bytes"
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
          "github.com/kijimaD/example"
  )

  func main() {
          fmt.Println("hello")

          title := strings.Title("hello world")
          rep := strings.Repeat("a", 10)
          fmt.Println(title, rep)
          example.Hello("hi")
  }`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "dummy.go", src, parser.AllErrors)
	assert.NoError(t, err)

	refs := References{}
	collectReferences(f, refs)
	expected := map[string]map[string]int(map[string]map[string]int{
		"fmt":     map[string]int{"Println": 2},
		"strings": map[string]int{"Repeat": 1, "Title": 1},
		"example": map[string]int{"Hello": 1},
	},
	)
	assert.Equal(t, expected, refs)
}

func TestCollectImports(t *testing.T) {
	src := `package main

  import (
          "fmt"
          abc "fmt"
          "strings"
          "github.com/kijimaD/example"
  )

  func main() {
          fmt.Println("hello")

          title := strings.Title("hello world")
          rep := strings.Repeat("a", 10)
          fmt.Println(title, rep)
          abc.Println("hello")
  }`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "dummy.go", src, parser.AllErrors)
	assert.NoError(t, err)

	expect := []ImportInfo{
		{
			ImportPath: "fmt",
			AliasName:  "",
		},
		{
			ImportPath: "fmt",
			AliasName:  "abc",
		},
		{
			ImportPath: "strings",
			AliasName:  "",
		},
		{
			ImportPath: "github.com/kijimaD/example",
			AliasName:  "",
		},
	}
	assert.Equal(t, expect, collectImports(f))
}

func TestFilterStdlib(t *testing.T) {
	refs := map[string]map[string]int(map[string]map[string]int{
		"fmt":     map[string]int{"Println": 2},
		"strings": map[string]int{"Repeat": 1, "Title": 1, "NotExistMethod": 1},
		"os":      map[string]int{"NotExistMethod": 1},
		"example": map[string]int{"Hello": 1},
	})

	stdrefs := filterStdLib(refs)
	expect := map[string]map[string]int(map[string]map[string]int{
		"fmt":     map[string]int{"Println": 2},
		"strings": map[string]int{"Repeat": 1, "Title": 1},
	})
	assert.Equal(t, expect, stdrefs)
}

func TestWalkDir(t *testing.T) {
	refs, err := walkDir("./testdata")
	assert.NoError(t, err)

	// とりあえず標準ライブラリ以外は除外されている
	expect := map[string]map[string]int{
		"go/ast":     map[string]int{"File": 3, "FuncDecl": 1},
		"bufio":      map[string]int{"NewReader": 1},
		"bytes":      map[string]int{"Buffer": 3, "IndexByte": 1, "LastIndex": 1, "NewReader": 1, "ReplaceAll": 1},
		"context":    map[string]int{"Context": 1},
		"fmt":        map[string]int{"Fprint": 1},
		"go/format":  map[string]int{"Source": 1},
		"io":         map[string]int{"EOF": 1, "Reader": 1},
		"os":         map[string]int{"Exit": 1},
		"go/parser":  map[string]int{"AllErrors": 3, "Mode": 3, "ParseComments": 3, "ParseFile": 4, "SkipObjectResolution": 2},
		"go/printer": map[string]int{"Config": 1, "TabIndent": 1, "UseSpaces": 1},
		"regexp":     map[string]int{"MustCompile": 1},
		"strconv":    map[string]int{"Unquote": 1},
		"strings":    map[string]int{"Contains": 2, "HasPrefix": 5},
		"testing":    map[string]int{"M": 1},
		"go/token":   map[string]int{"FileSet": 2, "NewFileSet": 3},
	}
	assert.Equal(t, expect, refs)
}

func TestRun(t *testing.T) {
	buf := bytes.Buffer{}
	Run(&buf, "./testdata")

	expect := `{
	"bufio": {
		"NewReader": 1
	},
	"bytes": {
		"Buffer": 3,
		"IndexByte": 1,
		"LastIndex": 1,
		"NewReader": 1,
		"ReplaceAll": 1
	},
	"context": {
		"Context": 1
	},
	"fmt": {
		"Fprint": 1
	},
	"go/ast": {
		"File": 3,
		"FuncDecl": 1
	},
	"go/format": {
		"Source": 1
	},
	"go/parser": {
		"AllErrors": 3,
		"Mode": 3,
		"ParseComments": 3,
		"ParseFile": 4,
		"SkipObjectResolution": 2
	},
	"go/printer": {
		"Config": 1,
		"TabIndent": 1,
		"UseSpaces": 1
	},
	"go/token": {
		"FileSet": 2,
		"NewFileSet": 3
	},
	"io": {
		"EOF": 1,
		"Reader": 1
	},
	"os": {
		"Exit": 1
	},
	"regexp": {
		"MustCompile": 1
	},
	"strconv": {
		"Unquote": 1
	},
	"strings": {
		"Contains": 2,
		"HasPrefix": 5
	},
	"testing": {
		"M": 1
	}
}
`
	assert.Equal(t, expect, buf.String())
}
