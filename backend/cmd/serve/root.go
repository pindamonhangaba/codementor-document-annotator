package serve

import (
	"github.com/pindamonhangaba/monoboi/backend/cmd/daemon"

	"github.com/spf13/cobra"
)

func RegisterCommandRecursive(parent *cobra.Command) {
	parent.AddCommand(daemon.ServeAll())
}
