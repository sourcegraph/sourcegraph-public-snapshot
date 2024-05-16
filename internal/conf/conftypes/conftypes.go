package conftypes

import (
	"reflect"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	proto "github.com/sourcegraph/sourcegraph/internal/api/internalapi/v1"
)

// ServiceConnections represents configuration about how the deployment
// internally connects to services. These are settings that need to be
// propagated from the frontend to other services, so that the frontend
// can be the source of truth for all configuration.
type ServiceConnections struct {
	// GitServers is the addresses of gitserver instances that should be
	// talked to.
	GitServers []string `json:"gitServers"`

	// PostgresDSN is the PostgreSQL DB data source name.
	// eg: "postgres://sg@pgsql/sourcegraph?sslmode=false"
	PostgresDSN string `json:"postgresDSN"`

	// CodeIntelPostgresDSN is the PostgreSQL DB data source name for the
	// code intel database.
	// eg: "postgres://sg@pgsql/sourcegraph_codeintel?sslmode=false"
	CodeIntelPostgresDSN string `json:"codeIntelPostgresDSN"`

	// CodeInsightsDSN is the PostgreSQL DB data source name for the
	// code insights database.
	// eg: "postgres://sg@pgsql/sourcegraph_codeintel?sslmode=false"
	CodeInsightsDSN string `json:"codeInsightsPostgresDSN"`

	// Searchers is the addresses of searcher instances that should be talked to.
	Searchers []string `json:"searchers"`
	// Symbols is the addresses of symbol instances that should be talked to.
	Symbols []string `json:"symbols"`
	// Embeddings is the addresses of embeddings instances that should be talked to.
	Embeddings []string `json:"embeddings"`
	// Qdrant is the address of the Qdrant instance (or empty if disabled)
	Qdrant string `json:"qdrant"`
	// Zoekts is the addresses of Zoekt instances to talk to.
	Zoekts []string `json:"zoekts"`
	// ZoektListTTL is the TTL of the internal cache that Zoekt clients use to
	// cache the list of indexed repository. After TTL is over, new list will
	// get requested from Zoekt shards.
	ZoektListTTL time.Duration `json:"zoektListTTL"`
}

func (sc *ServiceConnections) ToProto() *proto.ServiceConnections {
	return &proto.ServiceConnections{
		GitServers:           sc.GitServers,
		PostgresDsn:          sc.PostgresDSN,
		CodeIntelPostgresDsn: sc.CodeIntelPostgresDSN,
		CodeInsightsDsn:      sc.CodeInsightsDSN,
		Searchers:            sc.Searchers,
		Symbols:              sc.Symbols,
		Embeddings:           sc.Embeddings,
		Qdrant:               sc.Qdrant,
		Zoekts:               sc.Zoekts,
		ZoektListTtl:         durationpb.New(sc.ZoektListTTL),
	}
}

func (sc *ServiceConnections) FromProto(in *proto.ServiceConnections) {
	*sc = ServiceConnections{
		GitServers:           in.GetGitServers(),
		PostgresDSN:          in.GetPostgresDsn(),
		CodeIntelPostgresDSN: in.GetCodeIntelPostgresDsn(),
		CodeInsightsDSN:      in.GetCodeInsightsDsn(),
		Searchers:            in.GetSearchers(),
		Symbols:              in.GetSymbols(),
		Embeddings:           in.GetEmbeddings(),
		Qdrant:               in.GetQdrant(),
		Zoekts:               in.GetZoekts(),
		ZoektListTTL:         in.GetZoektListTtl().AsDuration(),
	}
}

// RawUnified is the unparsed variant of conf.Unified.
type RawUnified struct {
	ID                 int32
	Site               string
	ServiceConnections ServiceConnections
}

func (r *RawUnified) ToProto() *proto.RawUnified {
	return &proto.RawUnified{
		Id:                 r.ID,
		Site:               r.Site,
		ServiceConnections: r.ServiceConnections.ToProto(),
	}
}

func (r *RawUnified) FromProto(in *proto.RawUnified) {
	*r = RawUnified{
		ID:   in.GetId(),
		Site: in.GetSite(),
	}
	r.ServiceConnections.FromProto(in.GetServiceConnections())
}

// Equal tells if the two configurations are equal or not.
func (r RawUnified) Equal(other RawUnified) bool {
	return r.Site == other.Site && reflect.DeepEqual(r.ServiceConnections, other.ServiceConnections)
}

// Bedrock ModelIDs can either be just a <model_id> string as specified in https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html. Or when using provisioned capacity a string like `<model_id>/<provisioned_capacity>` where the <provisioned_capacity> refers to the arn in https://docs.aws.amazon.com/bedrock/latest/APIReference/API_CreateProvisionedModelThroughput.html#API_CreateProvisionedModelThroughput_RequestSyntax
type BedrockModelIdParts struct {
	Model               string
	ProvisionedCapacity string
}

func ParseBedrockModelId(modelId string) BedrockModelIdParts {
	parts := strings.SplitN(modelId, "/", 2)

	parsed := BedrockModelIdParts{
		Model: parts[0],
	}
	if len(parts) == 2 {
		parsed.ProvisionedCapacity = parts[1]
	}
	return parsed
}

func CasefoldBedrockModelId(modelId string) string {
	// Bedrock models are case sensitive if they contain a ARN
	// make sure to only lowercase the non ARN part
	parsed := ParseBedrockModelId(modelId)
	parsed.Model = strings.ToLower(parsed.Model)

	if parsed.ProvisionedCapacity != "" {
		return strings.Join([]string{parsed.Model, parsed.ProvisionedCapacity}, "/")
	}
	return parsed.Model
}
