package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// extensionConfigurationSubjectConnection implements the GraphQL type
// ExtensionConfigurationSubjectConnection based on an underlying list of configuration subjects that is
// computed statically.
type extensionConfigurationSubjectConnection struct {
	connectionArgs
	subjects []api.ConfigurationSubject
}

func (r *extensionConfigurationSubjectConnection) Nodes(ctx context.Context) ([]*extensionConfigurationSubject, error) {
	subjects := r.subjects
	if r.connectionArgs.First != nil && len(subjects) > int(*r.connectionArgs.First) {
		subjects = subjects[:int(*r.connectionArgs.First)]
	}

	resolvers := make([]*extensionConfigurationSubject, len(subjects))
	for i, subject := range subjects {
		subjectResolver, err := configurationSubjectByDBID(ctx, subject)
		if err != nil {
			return nil, err
		}
		resolvers[i] = &extensionConfigurationSubject{subjectResolver}
	}
	return resolvers, nil
}

func (r *extensionConfigurationSubjectConnection) TotalCount() int32 {
	return int32(len(r.subjects))
}
