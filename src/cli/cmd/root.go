package cmd

import (
  "github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
    Use:   "./sdk",
    Short: "The storj sdk is a batteries-included development kit for storj",
    Long:
`The storj sdk includes all of the pieces necessary to run a local storj stack
on your personal computer. Other than its dependency on docker-engine, it has
everything you need included.`,
}
