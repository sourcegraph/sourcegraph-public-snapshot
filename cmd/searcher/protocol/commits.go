package protocol

import (
	"encoding/json"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	proto "github.com/sourcegraph/sourcegraph/internal/searcher/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CommitMatch struct {
	Oid       api.CommitID
	Author    Signature      `json:",omitempty"`
	Committer Signature      `json:",omitempty"`
	Parents   []api.CommitID `json:",omitempty"`
	// Refs       []string       `json:",omitempty"`
	// SourceRefs []string       `json:",omitempty"`

	Message       result.MatchedString `json:",omitempty"`
	Diff          result.MatchedString `json:",omitempty"`
	ModifiedFiles []string             `json:",omitempty"`
}

func (cm *CommitMatch) ToProto() *proto.CommitMatch {
	parents := make([]string, 0, len(cm.Parents))
	for _, parent := range cm.Parents {
		parents = append(parents, string(parent))
	}
	return &proto.CommitMatch{
		Oid:       string(cm.Oid),
		Author:    SignatureToProto(&cm.Author),
		Committer: SignatureToProto(&cm.Committer),
		Parents:   parents,
		// Refs:          cm.Refs,
		// SourceRefs:    cm.SourceRefs,
		Message:       matchedStringToProto(cm.Message),
		Diff:          matchedStringToProto(cm.Diff),
		ModifiedFiles: cm.ModifiedFiles,
	}
}

func CommitMatchFromProto(p *proto.CommitMatch) CommitMatch {
	parents := make([]api.CommitID, 0, len(p.GetParents()))
	for _, parent := range p.GetParents() {
		parents = append(parents, api.CommitID(parent))
	}
	return CommitMatch{
		Oid:       api.CommitID(p.GetOid()),
		Author:    SignatureFromProto(p.GetAuthor()),
		Committer: SignatureFromProto(p.GetCommitter()),
		Parents:   parents,
		// Refs:          p.GetRefs(),
		// SourceRefs:    p.GetSourceRefs(),
		Message:       matchedStringFromProto(p.GetMessage()),
		Diff:          matchedStringFromProto(p.GetDiff()),
		ModifiedFiles: p.GetModifiedFiles(),
	}
}

func matchedStringFromProto(p *proto.CommitMatch_MatchedString) result.MatchedString {
	ranges := make([]result.Range, 0, len(p.GetRanges()))
	for _, rr := range p.GetRanges() {
		ranges = append(ranges, rangeFromProto(rr))
	}
	return result.MatchedString{
		Content:       p.GetContent(),
		MatchedRanges: ranges,
	}
}

func matchedStringToProto(ms result.MatchedString) *proto.CommitMatch_MatchedString {
	rrs := make([]*proto.CommitMatch_Range, 0, len(ms.MatchedRanges))
	for _, rr := range ms.MatchedRanges {
		rrs = append(rrs, rangeToProto(rr))
	}
	return &proto.CommitMatch_MatchedString{
		Content: ms.Content,
		Ranges:  rrs,
	}
}

func rangeToProto(r result.Range) *proto.CommitMatch_Range {
	return &proto.CommitMatch_Range{
		Start: locationToProto(r.Start),
		End:   locationToProto(r.End),
	}
}

func rangeFromProto(p *proto.CommitMatch_Range) result.Range {
	return result.Range{
		Start: locationFromProto(p.GetStart()),
		End:   locationFromProto(p.GetEnd()),
	}
}

func locationToProto(l result.Location) *proto.CommitMatch_Location {
	return &proto.CommitMatch_Location{
		Offset: uint32(l.Offset),
		Line:   uint32(l.Line),
		Column: uint32(l.Column),
	}
}

func locationFromProto(p *proto.CommitMatch_Location) result.Location {
	return result.Location{
		Offset: int(p.GetOffset()),
		Line:   int(p.GetLine()),
		Column: int(p.GetColumn()),
	}
}

type Signature struct {
	Name  string `json:",omitempty"`
	Email string `json:",omitempty"`
	Date  time.Time
}

func SignatureFromProto(p *proto.CommitMatch_Signature) Signature {
	return Signature{
		Name:  p.GetName(),
		Email: p.GetEmail(),
		Date:  p.GetDate().AsTime(),
	}
}

func SignatureToProto(s *Signature) *proto.CommitMatch_Signature {
	return &proto.CommitMatch_Signature{
		Name:  s.Name,
		Email: s.Email,
		Date:  timestamppb.New(s.Date),
	}
}

type CommitSearchRequest struct {
	Repo                 api.RepoName
	Revisions            []string
	Query                CommitSearchNode
	IncludeDiff          bool
	Limit                int
	IncludeModifiedFiles bool
}

func (r *CommitSearchRequest) ToProto() *proto.CommitSearchRequest {
	revs := make([]*proto.RevisionSpecifier, 0, len(r.Revisions))
	for _, rev := range r.Revisions {
		revs = append(revs, &proto.RevisionSpecifier{RevSpec: rev})
	}
	return &proto.CommitSearchRequest{
		Repo:                 string(r.Repo),
		Revisions:            revs,
		Query:                r.Query.ToProto(),
		IncludeDiff:          r.IncludeDiff,
		Limit:                int64(r.Limit),
		IncludeModifiedFiles: r.IncludeModifiedFiles,
	}
}

func SearchRequestFromProto(p *proto.CommitSearchRequest) (*CommitSearchRequest, error) {
	query, err := CommitSearchNodeFromProto(p.GetQuery())
	if err != nil {
		return nil, err
	}

	revisions := make([]string, 0, len(p.GetRevisions()))
	for _, rev := range p.GetRevisions() {
		revisions = append(revisions, rev.GetRevSpec())
	}

	return &CommitSearchRequest{
		Repo:                 api.RepoName(p.GetRepo()),
		Revisions:            revisions,
		Query:                query,
		IncludeDiff:          p.GetIncludeDiff(),
		Limit:                int(p.GetLimit()),
		IncludeModifiedFiles: p.GetIncludeModifiedFiles(),
	}, nil
}

type SearchEventMatches []CommitMatch

type SearchEventDone struct {
	LimitHit bool
	Error    string
}

func (s SearchEventDone) Err() error {
	if s.Error != "" {
		var e gitdomain.RepoNotExistError
		if err := json.Unmarshal([]byte(s.Error), &e); err == nil {
			return &e
		}
		return errors.New(s.Error)
	}
	return nil
}

func NewSearchEventDone(limitHit bool, err error) SearchEventDone {
	event := SearchEventDone{
		LimitHit: limitHit,
	}
	var notExistError *gitdomain.RepoNotExistError
	if errors.As(err, &notExistError) {
		b, _ := json.Marshal(notExistError)
		event.Error = string(b)
	} else if err != nil {
		event.Error = err.Error()
	}
	return event
}
