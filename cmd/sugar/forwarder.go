package sugar

import (
	"snwzt/rvc/internal/handlers"
	"snwzt/rvc/services/forwarder"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type ForwarderCmd struct {
	cmd *cobra.Command
}

func newForwarderCmd(redis *redis.Client, zerolog *zerolog.Logger) *ForwarderCmd {
	root := &ForwarderCmd{}
	cmd := &cobra.Command{
		Use:           "forwarder-service",
		Short:         "Run forwarder service",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cancelChan := make(chan string)
			defer close(cancelChan)

			forwarderOperationsHandle := &handlers.ForwarderOperationsHandle{
				Redis:           redis,
				CancelForwarder: cancelChan,
				Logger:          zerolog,
			}
			forwarder := forwarder.NewForwarder(forwarderOperationsHandle)

			forwarder.Run()

			return nil
		},
	}

	root.cmd = cmd
	return root
}
