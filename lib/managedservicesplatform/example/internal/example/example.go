package example

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/background"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/service"
)

type Config struct {
	Variable int
}

func (c *Config) Load(env *service.Env) {
	c.Variable = env.GetInt("VARIABLE", "13", "variable value")
}

type Service struct{}

var _ service.Service[Config] = Service{}

func (s Service) Name() string    { return "example" }
func (s Service) Version() string { return "dev" }
func (s Service) Initialize(
	ctx context.Context,
	logger log.Logger,
	contract service.Contract,
	config Config,
) (background.CombinedRoutine, error) {
	logger.Info("starting service")

	return background.CombinedRoutine{
		&httpRoutine{
			log: logger,
			Server: &http.Server{
				Addr: fmt.Sprintf(":%d", contract.Port),
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(fmt.Sprintf("Variable: %d", config.Variable)))
				}),
			},
		},
	}, nil
}

type httpRoutine struct {
	log log.Logger
	*http.Server
}

func (s *httpRoutine) Start() {
	if err := s.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.log.Error("error stopping server", log.Error(err))
	}
}

func (s *httpRoutine) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Server.Shutdown(ctx); err != nil {
		s.log.Error("error shutting down server", log.Error(err))
	} else {
		s.log.Info("server stopped")
	}
}
