package query

import (
	"fmt"
	"regexp/syntax"

	"github.com/RoaringBitmap/roaring"
	"github.com/grafana/regexp"

	proto "github.com/sourcegraph/zoekt/grpc/protos/zoekt/webserver/v1"
)

func QToProto(q Q) *proto.Q {
	switch v := q.(type) {
	case RawConfig:
		return &proto.Q{Query: &proto.Q_RawConfig{RawConfig: v.ToProto()}}
	case *Regexp:
		return &proto.Q{Query: &proto.Q_Regexp{Regexp: v.ToProto()}}
	case *Symbol:
		return &proto.Q{Query: &proto.Q_Symbol{Symbol: v.ToProto()}}
	case *Language:
		return &proto.Q{Query: &proto.Q_Language{Language: v.ToProto()}}
	case *Const:
		return &proto.Q{Query: &proto.Q_Const{Const: v.Value}}
	case *Repo:
		return &proto.Q{Query: &proto.Q_Repo{Repo: v.ToProto()}}
	case *RepoRegexp:
		return &proto.Q{Query: &proto.Q_RepoRegexp{RepoRegexp: v.ToProto()}}
	case *BranchesRepos:
		return &proto.Q{Query: &proto.Q_BranchesRepos{BranchesRepos: v.ToProto()}}
	case *RepoIDs:
		return &proto.Q{Query: &proto.Q_RepoIds{RepoIds: v.ToProto()}}
	case *RepoSet:
		return &proto.Q{Query: &proto.Q_RepoSet{RepoSet: v.ToProto()}}
	case *FileNameSet:
		return &proto.Q{Query: &proto.Q_FileNameSet{FileNameSet: v.ToProto()}}
	case *Type:
		return &proto.Q{Query: &proto.Q_Type{Type: v.ToProto()}}
	case *Substring:
		return &proto.Q{Query: &proto.Q_Substring{Substring: v.ToProto()}}
	case *And:
		return &proto.Q{Query: &proto.Q_And{And: v.ToProto()}}
	case *Or:
		return &proto.Q{Query: &proto.Q_Or{Or: v.ToProto()}}
	case *Not:
		return &proto.Q{Query: &proto.Q_Not{Not: v.ToProto()}}
	case *Branch:
		return &proto.Q{Query: &proto.Q_Branch{Branch: v.ToProto()}}
	case *Boost:
		return &proto.Q{Query: &proto.Q_Boost{Boost: v.ToProto()}}
	default:
		// The following nodes do not have a proto representation:
		// - caseQ: only used internally, not by the RPC layer
		panic(fmt.Sprintf("unknown query node %T", v))
	}
}

func QFromProto(p *proto.Q) (Q, error) {
	switch v := p.Query.(type) {
	case *proto.Q_RawConfig:
		return RawConfigFromProto(v.RawConfig), nil
	case *proto.Q_Regexp:
		return RegexpFromProto(v.Regexp)
	case *proto.Q_Symbol:
		return SymbolFromProto(v.Symbol)
	case *proto.Q_Language:
		return LanguageFromProto(v.Language), nil
	case *proto.Q_Const:
		return &Const{Value: v.Const}, nil
	case *proto.Q_Repo:
		return RepoFromProto(v.Repo)
	case *proto.Q_RepoRegexp:
		return RepoRegexpFromProto(v.RepoRegexp)
	case *proto.Q_BranchesRepos:
		return BranchesReposFromProto(v.BranchesRepos)
	case *proto.Q_RepoIds:
		return RepoIDsFromProto(v.RepoIds)
	case *proto.Q_RepoSet:
		return RepoSetFromProto(v.RepoSet), nil
	case *proto.Q_FileNameSet:
		return FileNameSetFromProto(v.FileNameSet), nil
	case *proto.Q_Type:
		return TypeFromProto(v.Type)
	case *proto.Q_Substring:
		return SubstringFromProto(v.Substring), nil
	case *proto.Q_And:
		return AndFromProto(v.And)
	case *proto.Q_Or:
		return OrFromProto(v.Or)
	case *proto.Q_Not:
		return NotFromProto(v.Not)
	case *proto.Q_Branch:
		return BranchFromProto(v.Branch), nil
	case *proto.Q_Boost:
		return BoostFromProto(v.Boost)
	default:
		panic(fmt.Sprintf("unknown query node %T", p.Query))
	}
}

func RegexpFromProto(p *proto.Regexp) (*Regexp, error) {
	parsed, err := syntax.Parse(p.GetRegexp(), regexpFlags)
	if err != nil {
		return nil, err
	}
	return &Regexp{
		Regexp:        parsed,
		FileName:      p.GetFileName(),
		Content:       p.GetContent(),
		CaseSensitive: p.GetCaseSensitive(),
	}, nil
}

func (r *Regexp) ToProto() *proto.Regexp {
	return &proto.Regexp{
		Regexp:        r.Regexp.String(),
		FileName:      r.FileName,
		Content:       r.Content,
		CaseSensitive: r.CaseSensitive,
	}
}

func SymbolFromProto(p *proto.Symbol) (*Symbol, error) {
	expr, err := QFromProto(p.GetExpr())
	if err != nil {
		return nil, err
	}

	return &Symbol{
		Expr: expr,
	}, nil
}

func (s *Symbol) ToProto() *proto.Symbol {
	return &proto.Symbol{
		Expr: QToProto(s.Expr),
	}
}

func LanguageFromProto(p *proto.Language) *Language {
	return &Language{
		Language: p.GetLanguage(),
	}
}

func (l *Language) ToProto() *proto.Language {
	return &proto.Language{Language: l.Language}
}

func RepoFromProto(p *proto.Repo) (*Repo, error) {
	r, err := regexp.Compile(p.GetRegexp())
	if err != nil {
		return nil, err
	}
	return &Repo{
		Regexp: r,
	}, nil
}

func (q *Repo) ToProto() *proto.Repo {
	return &proto.Repo{
		Regexp: q.Regexp.String(),
	}
}

func RepoRegexpFromProto(p *proto.RepoRegexp) (*RepoRegexp, error) {
	r, err := regexp.Compile(p.GetRegexp())
	if err != nil {
		return nil, err
	}
	return &RepoRegexp{
		Regexp: r,
	}, nil
}

func (q *RepoRegexp) ToProto() *proto.RepoRegexp {
	return &proto.RepoRegexp{
		Regexp: q.Regexp.String(),
	}
}

func BranchesReposFromProto(p *proto.BranchesRepos) (*BranchesRepos, error) {
	brs := make([]BranchRepos, len(p.GetList()))
	for i, br := range p.GetList() {
		branchRepos, err := BranchReposFromProto(br)
		if err != nil {
			return nil, err
		}
		brs[i] = branchRepos
	}
	return &BranchesRepos{
		List: brs,
	}, nil
}

func (br *BranchesRepos) ToProto() *proto.BranchesRepos {
	list := make([]*proto.BranchRepos, len(br.List))
	for i, branchRepo := range br.List {
		list[i] = branchRepo.ToProto()
	}

	return &proto.BranchesRepos{
		List: list,
	}
}

func RepoIDsFromProto(p *proto.RepoIds) (*RepoIDs, error) {
	bm := roaring.NewBitmap()
	err := bm.UnmarshalBinary(p.GetRepos())
	if err != nil {
		return nil, err
	}

	return &RepoIDs{
		Repos: bm,
	}, nil
}

func (q *RepoIDs) ToProto() *proto.RepoIds {
	b, err := q.Repos.ToBytes()
	if err != nil {
		panic("unexpected error marshalling bitmap: " + err.Error())
	}
	return &proto.RepoIds{
		Repos: b,
	}
}

func BranchReposFromProto(p *proto.BranchRepos) (BranchRepos, error) {
	bm := roaring.NewBitmap()
	err := bm.UnmarshalBinary(p.GetRepos())
	if err != nil {
		return BranchRepos{}, err
	}
	return BranchRepos{
		Branch: p.GetBranch(),
		Repos:  bm,
	}, nil
}

func (br *BranchRepos) ToProto() *proto.BranchRepos {
	b, err := br.Repos.ToBytes()
	if err != nil {
		panic("unexpected error marshalling bitmap: " + err.Error())
	}

	return &proto.BranchRepos{
		Branch: br.Branch,
		Repos:  b,
	}
}

func RepoSetFromProto(p *proto.RepoSet) *RepoSet {
	return &RepoSet{
		Set: p.GetSet(),
	}
}

func (q *RepoSet) ToProto() *proto.RepoSet {
	return &proto.RepoSet{
		Set: q.Set,
	}
}

func FileNameSetFromProto(p *proto.FileNameSet) *FileNameSet {
	m := make(map[string]struct{}, len(p.GetSet()))
	for _, name := range p.GetSet() {
		m[name] = struct{}{}
	}
	return &FileNameSet{
		Set: m,
	}
}

func (q *FileNameSet) ToProto() *proto.FileNameSet {
	s := make([]string, 0, len(q.Set))
	for name := range q.Set {
		s = append(s, name)
	}
	return &proto.FileNameSet{
		Set: s,
	}
}

func TypeFromProto(p *proto.Type) (*Type, error) {
	child, err := QFromProto(p.GetChild())
	if err != nil {
		return nil, err
	}

	var kind uint8
	switch p.GetType() {
	case proto.Type_KIND_FILE_MATCH:
		kind = TypeFileMatch
	case proto.Type_KIND_FILE_NAME:
		kind = TypeFileName
	case proto.Type_KIND_REPO:
		kind = TypeRepo
	}

	return &Type{
		Child: child,
		// TODO: make proper enum types
		Type: kind,
	}, nil
}

func (q *Type) ToProto() *proto.Type {
	var kind proto.Type_Kind
	switch q.Type {
	case TypeFileMatch:
		kind = proto.Type_KIND_FILE_MATCH
	case TypeFileName:
		kind = proto.Type_KIND_FILE_NAME
	case TypeRepo:
		kind = proto.Type_KIND_REPO
	}

	return &proto.Type{
		Child: QToProto(q.Child),
		Type:  kind,
	}
}

func SubstringFromProto(p *proto.Substring) *Substring {
	return &Substring{
		Pattern:       p.GetPattern(),
		CaseSensitive: p.GetCaseSensitive(),
		FileName:      p.GetFileName(),
		Content:       p.GetContent(),
	}
}

func (q *Substring) ToProto() *proto.Substring {
	return &proto.Substring{
		Pattern:       q.Pattern,
		CaseSensitive: q.CaseSensitive,
		FileName:      q.FileName,
		Content:       q.Content,
	}
}

func OrFromProto(p *proto.Or) (*Or, error) {
	children := make([]Q, len(p.GetChildren()))
	for i, child := range p.GetChildren() {
		c, err := QFromProto(child)
		if err != nil {
			return nil, err
		}
		children[i] = c
	}
	return &Or{
		Children: children,
	}, nil
}

func (q *Or) ToProto() *proto.Or {
	children := make([]*proto.Q, len(q.Children))
	for i, child := range q.Children {
		children[i] = QToProto(child)
	}
	return &proto.Or{
		Children: children,
	}
}

func BoostFromProto(p *proto.Boost) (*Boost, error) {
	child, err := QFromProto(p.GetChild())
	if err != nil {
		return nil, err
	}
	return &Boost{
		Child: child,
		Boost: p.GetBoost(),
	}, nil
}

func (q *Boost) ToProto() *proto.Boost {
	return &proto.Boost{
		Child: QToProto(q.Child),
		Boost: q.Boost,
	}
}

func NotFromProto(p *proto.Not) (*Not, error) {
	child, err := QFromProto(p.GetChild())
	if err != nil {
		return nil, err
	}
	return &Not{
		Child: child,
	}, nil
}

func (q *Not) ToProto() *proto.Not {
	return &proto.Not{
		Child: QToProto(q.Child),
	}
}

func AndFromProto(p *proto.And) (*And, error) {
	children := make([]Q, len(p.GetChildren()))
	for i, child := range p.GetChildren() {
		c, err := QFromProto(child)
		if err != nil {
			return nil, err
		}
		children[i] = c
	}
	return &And{
		Children: children,
	}, nil
}

func (q *And) ToProto() *proto.And {
	children := make([]*proto.Q, len(q.Children))
	for i, child := range q.Children {
		children[i] = QToProto(child)
	}
	return &proto.And{
		Children: children,
	}
}

func BranchFromProto(p *proto.Branch) *Branch {
	return &Branch{
		Pattern: p.GetPattern(),
		Exact:   p.GetExact(),
	}
}

func (q *Branch) ToProto() *proto.Branch {
	return &proto.Branch{
		Pattern: q.Pattern,
		Exact:   q.Exact,
	}
}

func RawConfigFromProto(p *proto.RawConfig) (res RawConfig) {
	for _, protoFlag := range p.Flags {
		switch protoFlag {
		case proto.RawConfig_FLAG_ONLY_PUBLIC:
			res |= RcOnlyPublic
		case proto.RawConfig_FLAG_ONLY_PRIVATE:
			res |= RcOnlyPrivate
		case proto.RawConfig_FLAG_ONLY_FORKS:
			res |= RcOnlyForks
		case proto.RawConfig_FLAG_NO_FORKS:
			res |= RcNoForks
		case proto.RawConfig_FLAG_ONLY_ARCHIVED:
			res |= RcOnlyArchived
		case proto.RawConfig_FLAG_NO_ARCHIVED:
			res |= RcNoArchived
		}
	}
	return res
}

func (r RawConfig) ToProto() *proto.RawConfig {
	var flags []proto.RawConfig_Flag
	for _, flag := range flagNames {
		if r&flag.Mask != 0 {
			switch flag.Mask {
			case RcOnlyPublic:
				flags = append(flags, proto.RawConfig_FLAG_ONLY_PUBLIC)
			case RcOnlyPrivate:
				flags = append(flags, proto.RawConfig_FLAG_ONLY_PRIVATE)
			case RcOnlyForks:
				flags = append(flags, proto.RawConfig_FLAG_ONLY_FORKS)
			case RcNoForks:
				flags = append(flags, proto.RawConfig_FLAG_NO_FORKS)
			case RcOnlyArchived:
				flags = append(flags, proto.RawConfig_FLAG_ONLY_ARCHIVED)
			case RcNoArchived:
				flags = append(flags, proto.RawConfig_FLAG_NO_ARCHIVED)
			}
		}
	}
	return &proto.RawConfig{Flags: flags}
}
