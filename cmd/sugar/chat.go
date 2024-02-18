package sugar

import (
	"log"
	"os"
	"snwzt/random-video-chat/internal/handlers"
	"snwzt/random-video-chat/internal/server"
	"snwzt/random-video-chat/pkg/service"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

type ChatCmd struct {
	cmd *cobra.Command
}

func newChatCmd() *ChatCmd {
	root := &ChatCmd{}
	cmd := &cobra.Command{
		Use:           "chat-service",
		Short:         "Run chat service",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := godotenv.Load("../config/.env")
			if err != nil {
				log.Fatal("Error loading .env file")
			}

			redis, err := service.NewRedisStore(os.Getenv("REDIS_URI"))
			if err != nil {
				return err
			}

			instance := echo.New()
			chatHttpHandle := &handlers.ChatHTTPHandle{
				Redis: redis,
			}
			s := server.NewChatHTTPServer(":5001", instance, chatHttpHandle)

			s.Run()

			return nil
		},
	}

	root.cmd = cmd
	return root
}
