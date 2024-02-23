package sugar

import (
	"os"
	"snwzt/rvc/internal/handlers"
	"snwzt/rvc/pkg/common"
	"snwzt/rvc/services/db"
	"snwzt/rvc/services/user"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type UserCmd struct {
	cmd *cobra.Command
}

func newUserCmd(redis *redis.Client, zerolog *zerolog.Logger) *UserCmd {
	root := &UserCmd{}
	cmd := &cobra.Command{
		Use:           "user-service",
		Short:         "Run user service",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			var wg sync.WaitGroup
			var err error

			instance := echo.New()

			instance.Renderer, err = common.NewTemplate("web/*.html")
			if err != nil {
				zerolog.Err(err).Msg("unable to load templates")
			}

			redis, err := db.NewRedisStore(os.Getenv("REDIS_URI"))
			if err != nil {
				zerolog.Err(err).Msg("unable to connect to redis")
			}
			userHttpHandle := &handlers.UserServerHandle{
				Redis: redis,
			}

			s := user.NewUserServer(":"+os.Getenv("USER_SERVICE_PORT"), instance, userHttpHandle, zerolog)

			wg.Add(1)
			go func() {
				defer wg.Done()

				s.Run()
			}()

			userOperationsHandle := &handlers.UserOperationsHandle{
				Redis:  redis,
				Logger: zerolog,
			}
			userOperations := user.NewUserOperations(userOperationsHandle)

			wg.Add(1)
			go func() {
				defer wg.Done()

				userOperations.Run()
			}()

			wg.Wait()

			return nil
		},
	}

	root.cmd = cmd
	return root
}
