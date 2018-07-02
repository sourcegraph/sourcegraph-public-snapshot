package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// extensionConfigurationSubjectConnection implements the GraphQL type
// ExtensionConfigurationSubjectConnection based on an underlying list of configuration subjects that is
// computed statically.
type extensionConfigurationSubjectConnection struct {
	connectionArgs
	subjects []api.ConfigurationSubject

	extension *registryExtensionMultiResolver

	// cache results because they are used by multiple fields
	once    sync.Once
	results []*extensionConfigurationSubjectEdge
	err     error
}

func (r *extensionConfigurationSubjectConnection) compute(ctx context.Context) ([]*extensionConfigurationSubjectEdge, error) {
	do := func() ([]*extensionConfigurationSubjectEdge, error) {
		subjects := r.subjects
		if r.connectionArgs.First != nil && len(subjects) > int(*r.connectionArgs.First) {
			subjects = subjects[:int(*r.connectionArgs.First)]
		}

		edges := make([]*extensionConfigurationSubjectEdge, len(subjects))
		for i, subject := range subjects {
			subjectResolver, err := configurationSubjectByDBID(ctx, subject)
			if err != nil {
				return nil, err
			}
			edges[i] = &extensionConfigurationSubjectEdge{
				node:      &extensionConfigurationSubject{subjectResolver},
				extension: r.extension,
			}
		}
		return edges, nil
	}
	r.once.Do(func() { r.results, r.err = do() })
	return r.results, r.err
}

func (r *extensionConfigurationSubjectConnection) Edges(ctx context.Context) ([]*extensionConfigurationSubjectEdge, error) {
	return r.compute(ctx)
}

func (r *extensionConfigurationSubjectConnection) Nodes(ctx context.Context) ([]*extensionConfigurationSubject, error) {
	edges, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	nodes := make([]*extensionConfigurationSubject, len(edges))
	for i, edge := range edges {
		nodes[i] = edge.node
	}
	return nodes, nil
}

func (r *extensionConfigurationSubjectConnection) TotalCount() int32 {
	return int32(len(r.subjects))
}

type extensionConfigurationSubjectEdge struct {
	node      *extensionConfigurationSubject
	extension *registryExtensionMultiResolver
}

func (r *extensionConfigurationSubjectEdge) Node() *extensionConfigurationSubject { return r.node }
func (r *extensionConfigurationSubjectEdge) Extension() *registryExtensionMultiResolver {
	return r.extension
}

func (r *extensionConfigurationSubjectEdge) IsEnabled(ctx context.Context) (bool, error) {
	settings, err := r.node.subject.LatestSettings(ctx)
	if err != nil || settings == nil {
		return false, err
	}

	v := readRegistryExtensionEnablement(r.extension.ExtensionID(), settings.Contents())
	return (v != nil && *v), nil
}

func (r *extensionConfigurationSubjectEdge) URL() (string, error) { return r.node.SettingsURL() }
