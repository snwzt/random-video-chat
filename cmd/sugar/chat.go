package sugar

import (
	"os"
	"snwzt/rvc/internal/handlers"
	"snwzt/rvc/services/chat"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type ChatCmd struct {
	cmd *cobra.Command
}

func newChatCmd(redis *redis.Client, zerolog *zerolog.Logger) *ChatCmd {
	root := &ChatCmd{}
	cmd := &cobra.Command{
		Use:           "chat-service",
		Short:         "Run chat service",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			instance := echo.New()
			chatServerHandle := &handlers.ChatServerHandle{
				Redis:  redis,
				Logger: zerolog,
			}
			s := chat.NewChatServer(":"+os.Getenv("CHAT_SERVICE_PORT"), instance, chatServerHandle, zerolog)

			s.Run()

			return nil
		},
	}

	root.cmd = cmd
	return root
}
