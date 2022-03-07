package main

import (
	"os"

	"github.com/aerospike/aerostation/cmd/aeroctl/cmd"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	cmd.NewAeroctlCommand(os.Stdin, os.Stdout, os.Stderr)
	cmd.Execute()
}
