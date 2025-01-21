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
          title := strings.Title("hello world")
          rep := strings.Repeat("a", 10)
          fmt.Println(title, rep)
  }`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "blank.go", src, parser.AllErrors)
	assert.NoError(t, err)

	refs := collectReferences(f)
	expected := map[string]map[string]bool{
		"fmt": map[string]bool{
			"Println": true,
		},
		"strings": map[string]bool{
			"Repeat": true,
			"Title":  true,
		},
	}
	assert.Equal(t, expected, refs)
}

func TestWalkDir(t *testing.T) {
	assert.NoError(t, walkDir("/home/orange/Project/inhale"))
}
