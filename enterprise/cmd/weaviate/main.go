package weaviate

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"time"

	"github.com/sourcegraph/log"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/fault"
	"github.com/weaviate/weaviate/entities/models"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const port = "8181"

func start(ctx context.Context, observationCtx *observation.Context, cgf *Config) error {
	logger := observationCtx.Logger

	// Initialize tracing/metrics
	observationCtx = observation.NewContext(logger, observation.Honeycomb(&honey.Dataset{
		Name:       "weaviate",
		SampleRate: 20,
	}))

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	logger.Info("listening", log.String("addr", addr))

	cmd := exec.Command(cgf.Path)
	cmd.Env = append(cmd.Env, fmt.Sprintf("OPENAI_APIKEY=%s", cgf.OpenAIApiKey))

	err := cmd.Start()
	if err != nil {
		return err
	}

	// Create schema with retry
	for i := 0; i < 3; i++ {
		time.Sleep(10 * time.Second)
		logger.Info("creating schema", log.Int("attempt", i))
		err = createSchemaIfNotExists(ctx, logger)
		if err == nil {
			logger.Info("success", log.Int("attempt", i))
			break
		}
		logger.Error(fmt.Sprintf("error creating schema %s, type=%T", err, err))
		var wErr *fault.WeaviateClientError
		if errors.As(err, &wErr) {
			logger.Error("DerivedFromError", log.Error(wErr.DerivedFromError))
		}
	}

	return err
}

func createSchemaIfNotExists(ctx context.Context, logger log.Logger) error {
	var wErr *fault.WeaviateClientError

	err := createSchema(ctx, "Code")
	if errors.As(err, &wErr) {
		logger.Info("error creating \"code\" schema", log.Error(wErr.DerivedFromError))
		if wErr.StatusCode == 422 {
			err = nil
		}
	}
	if err != nil {
		return errors.Wrap(err, "error creating \"code\" schema")
	}

	err = createSchema(ctx, "Text")
	if errors.As(err, &wErr) {
		logger.Info("error creating \"text\" schema", log.Error(wErr.DerivedFromError))
		if wErr.StatusCode == 422 {
			err = nil
		}
	}
	if err != nil {
		return errors.Wrap(err, "error creating \"text\" schema")
	}

	return nil
}

func createSchema(ctx context.Context, typ string) error {
	client, err := weaviate.NewClient(weaviate.Config{Host: "localhost:8181", Scheme: "http"})
	if err != nil {
		return errors.Wrap(err, "error getting weaviate client")
	}

	class := &models.Class{
		Class: typ,
		Properties: []*models.Property{
			{Name: "filename", DataType: []string{"string"}},
			{Name: "content", DataType: []string{"string"}},
			{Name: "repository", DataType: []string{"string"}},
			{Name: "revision", DataType: []string{"string"}},
			{Name: "start_line", DataType: []string{"int"}},
			{Name: "end_line", DataType: []string{"int"}},
		},
		Vectorizer: "text2vec-openai",
	}
	return client.Schema().ClassCreator().WithClass(class).Do(ctx)
}
