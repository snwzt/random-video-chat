package sugar

import (
	"log"
	"os"
	"snwzt/random-video-chat/internal/handlers"
	"snwzt/random-video-chat/internal/server"
	"snwzt/random-video-chat/pkg/service"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type ForwarderCmd struct {
	cmd *cobra.Command
}

func newForwarderCmd() *ForwarderCmd {
	root := &ForwarderCmd{}
	cmd := &cobra.Command{
		Use:           "forwarder-service",
		Short:         "Run forwarder service",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := godotenv.Load("../config/.env")
			if err != nil {
				log.Fatal("Error loading .env file")
			}

			cancelChan := make(chan string)
			defer close(cancelChan)

			redis, err := service.NewRedisStore(os.Getenv("REDIS_URI"))
			if err != nil {
				return err
			}

			log.Println(os.Getenv("REDIS_URI"))

			log.Println("Connected to Redis")

			h := &handlers.ForwarderHandle{
				Redis:           redis,
				CancelForwarder: cancelChan,
			}
			forwarder := server.NewForwarder(h)

			log.Println("Starting Forwarder")

			forwarder.Run()

			return nil
		},
	}

	root.cmd = cmd
	return root
}
