package main

import (
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("引数エラー: ディレクトリ指定が必要")
	}
	path := os.Args[1]

	Run(os.Stdout, path)
}
