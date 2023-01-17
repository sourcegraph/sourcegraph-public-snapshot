package actor

import "google.golang.org/grpc/metadata"

const metadataKeyActor = "actor"

type ActorPropagator struct{}

func (a ActorPropagator) ExtractContext() metadata.MD {
}
