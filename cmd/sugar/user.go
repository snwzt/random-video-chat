package sugar

import (
	"log"
	"os"
	"snwzt/random-video-chat/internal/handlers"
	"snwzt/random-video-chat/internal/server"
	"snwzt/random-video-chat/pkg/common"
	"snwzt/random-video-chat/pkg/service"
	"sync"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
)

type UserCmd struct {
	cmd *cobra.Command
}

func newUserCmd() *UserCmd {
	root := &UserCmd{}
	cmd := &cobra.Command{
		Use:           "user-service",
		Short:         "Run user service",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := godotenv.Load("../config/.env")
			if err != nil {
				log.Fatal("Error loading .env file")
			}

			var wg sync.WaitGroup

			redis, err := service.NewRedisStore(os.Getenv("REDIS_URI"))
			if err != nil {
				return err
			}

			instance := echo.New()
			userHttpHandle := &handlers.UserHTTPHandle{
				Redis: redis,
			}
			instance.Renderer = common.NewTemplate("../web/*.html")
			s := server.NewUserHTTPServer(":5000", instance, userHttpHandle)

			wg.Add(1)
			go func() {
				defer wg.Done()

				s.Run()
			}()

			h := &handlers.UserQueueHandle{
				Redis: redis,
			}
			userQueue := server.NewUserQueue(h)

			wg.Add(1)
			go func() {
				defer wg.Done()

				userQueue.Run()
			}()

			wg.Wait()

			return nil
		},
	}

	root.cmd = cmd
	return root
}
