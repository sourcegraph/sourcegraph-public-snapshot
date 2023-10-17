package example

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sourcegraph/log"

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
func (s Service) Start(
	ctx context.Context,
	logger log.Logger,
	contract service.Contract,
	config Config,
) error {
	logger.Info("starting service")
	return http.ListenAndServe(
		fmt.Sprintf(":%d", contract.Port),
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(fmt.Sprintf("Variable: %d", config.Variable)))
		}))
}
