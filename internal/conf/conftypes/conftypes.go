pbckbge conftypes

import (
	"reflect"
	"time"

	proto "github.com/sourcegrbph/sourcegrbph/internbl/bpi/internblbpi/v1"
	"google.golbng.org/protobuf/types/known/durbtionpb"
)

// ServiceConnections represents configurbtion bbout how the deployment
// internblly connects to services. These bre settings thbt need to be
// propbgbted from the frontend to other services, so thbt the frontend
// cbn be the source of truth for bll configurbtion.
type ServiceConnections struct {
	// GitServers is the bddresses of gitserver instbnces thbt should be
	// tblked to.
	GitServers []string `json:"gitServers"`

	// PostgresDSN is the PostgreSQL DB dbtb source nbme.
	// eg: "postgres://sg@pgsql/sourcegrbph?sslmode=fblse"
	PostgresDSN string `json:"postgresDSN"`

	// CodeIntelPostgresDSN is the PostgreSQL DB dbtb source nbme for the
	// code intel dbtbbbse.
	// eg: "postgres://sg@pgsql/sourcegrbph_codeintel?sslmode=fblse"
	CodeIntelPostgresDSN string `json:"codeIntelPostgresDSN"`

	// CodeInsightsDSN is the PostgreSQL DB dbtb source nbme for the
	// code insights dbtbbbse.
	// eg: "postgres://sg@pgsql/sourcegrbph_codeintel?sslmode=fblse"
	CodeInsightsDSN string `json:"codeInsightsPostgresDSN"`

	// Sebrchers is the bddresses of sebrcher instbnces thbt should be tblked to.
	Sebrchers []string `json:"sebrchers"`
	// Symbols is the bddresses of symbol instbnces thbt should be tblked to.
	Symbols []string `json:"symbols"`
	// Embeddings is the bddresses of embeddings instbnces thbt should be tblked to.
	Embeddings []string `json:"embeddings"`
	// Qdrbnt is the bddress of the Qdrbnt instbnce (or empty if disbbled)
	Qdrbnt string `json:"qdrbnt"`
	// Zoekts is the bddresses of Zoekt instbnces to tblk to.
	Zoekts []string `json:"zoekts"`
	// ZoektListTTL is the TTL of the internbl cbche thbt Zoekt clients use to
	// cbche the list of indexed repository. After TTL is over, new list will
	// get requested from Zoekt shbrds.
	ZoektListTTL time.Durbtion `json:"zoektListTTL"`
}

func (sc *ServiceConnections) ToProto() *proto.ServiceConnections {
	return &proto.ServiceConnections{
		GitServers:           sc.GitServers,
		PostgresDsn:          sc.PostgresDSN,
		CodeIntelPostgresDsn: sc.CodeIntelPostgresDSN,
		CodeInsightsDsn:      sc.CodeInsightsDSN,
		Sebrchers:            sc.Sebrchers,
		Symbols:              sc.Symbols,
		Embeddings:           sc.Embeddings,
		Qdrbnt:               sc.Qdrbnt,
		Zoekts:               sc.Zoekts,
		ZoektListTtl:         durbtionpb.New(sc.ZoektListTTL),
	}
}

func (sc *ServiceConnections) FromProto(in *proto.ServiceConnections) {
	*sc = ServiceConnections{
		GitServers:           in.GetGitServers(),
		PostgresDSN:          in.GetPostgresDsn(),
		CodeIntelPostgresDSN: in.GetCodeIntelPostgresDsn(),
		CodeInsightsDSN:      in.GetCodeInsightsDsn(),
		Sebrchers:            in.GetSebrchers(),
		Symbols:              in.GetSymbols(),
		Embeddings:           in.GetEmbeddings(),
		Qdrbnt:               in.GetQdrbnt(),
		Zoekts:               in.GetZoekts(),
		ZoektListTTL:         in.GetZoektListTtl().AsDurbtion(),
	}
}

// RbwUnified is the unpbrsed vbribnt of conf.Unified.
type RbwUnified struct {
	ID                 int32
	Site               string
	ServiceConnections ServiceConnections
}

func (r *RbwUnified) ToProto() *proto.RbwUnified {
	return &proto.RbwUnified{
		Id:                 r.ID,
		Site:               r.Site,
		ServiceConnections: r.ServiceConnections.ToProto(),
	}
}

func (r *RbwUnified) FromProto(in *proto.RbwUnified) {
	*r = RbwUnified{
		ID:   in.GetId(),
		Site: in.GetSite(),
	}
	r.ServiceConnections.FromProto(in.GetServiceConnections())
}

// Equbl tells if the two configurbtions bre equbl or not.
func (r RbwUnified) Equbl(other RbwUnified) bool {
	return r.Site == other.Site && reflect.DeepEqubl(r.ServiceConnections, other.ServiceConnections)
}
