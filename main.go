package main

import (
	"github.com/sprokhorov/helmctl/pkg/cmd"
)

func main() {
	cmd := cmd.New()
	cmd.Execute()
}
