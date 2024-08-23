// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zoekt // import "github.com/sourcegraph/zoekt"

import (
	"math/rand"
	"reflect"

	proto "github.com/sourcegraph/zoekt/grpc/protos/zoekt/webserver/v1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func FileMatchFromProto(p *proto.FileMatch) FileMatch {
	lineMatches := make([]LineMatch, len(p.GetLineMatches()))
	for i, lineMatch := range p.GetLineMatches() {
		lineMatches[i] = LineMatchFromProto(lineMatch)
	}

	chunkMatches := make([]ChunkMatch, len(p.GetChunkMatches()))
	for i, chunkMatch := range p.GetChunkMatches() {
		chunkMatches[i] = ChunkMatchFromProto(chunkMatch)
	}

	return FileMatch{
		Score:              p.GetScore(),
		Debug:              p.GetDebug(),
		FileName:           string(p.GetFileName()), // Note: 🚨Warning, this filename may be a non-UTF8 string.
		Repository:         p.GetRepository(),
		Branches:           p.GetBranches(),
		LineMatches:        lineMatches,
		ChunkMatches:       chunkMatches,
		RepositoryID:       p.GetRepositoryId(),
		RepositoryPriority: p.GetRepositoryPriority(),
		Content:            p.GetContent(),
		Checksum:           p.GetChecksum(),
		Language:           p.GetLanguage(),
		SubRepositoryName:  p.GetSubRepositoryName(),
		SubRepositoryPath:  p.GetSubRepositoryPath(),
		Version:            p.GetVersion(),
	}
}

func (m *FileMatch) ToProto() *proto.FileMatch {
	lineMatches := make([]*proto.LineMatch, len(m.LineMatches))
	for i, lm := range m.LineMatches {
		lineMatches[i] = lm.ToProto()
	}

	chunkMatches := make([]*proto.ChunkMatch, len(m.ChunkMatches))
	for i, cm := range m.ChunkMatches {
		chunkMatches[i] = cm.ToProto()
	}

	return &proto.FileMatch{
		Score:              m.Score,
		Debug:              m.Debug,
		FileName:           []byte(m.FileName),
		Repository:         m.Repository,
		Branches:           m.Branches,
		LineMatches:        lineMatches,
		ChunkMatches:       chunkMatches,
		RepositoryId:       m.RepositoryID,
		RepositoryPriority: m.RepositoryPriority,
		Content:            m.Content,
		Checksum:           m.Checksum,
		Language:           m.Language,
		SubRepositoryName:  m.SubRepositoryName,
		SubRepositoryPath:  m.SubRepositoryPath,
		Version:            m.Version,
	}
}

func ChunkMatchFromProto(p *proto.ChunkMatch) ChunkMatch {
	ranges := make([]Range, len(p.GetRanges()))
	for i, r := range p.GetRanges() {
		ranges[i] = RangeFromProto(r)
	}

	symbols := make([]*Symbol, len(p.GetSymbolInfo()))
	for i, r := range p.GetSymbolInfo() {
		symbols[i] = SymbolFromProto(r)
	}

	return ChunkMatch{
		Content:      p.GetContent(),
		ContentStart: LocationFromProto(p.GetContentStart()),
		FileName:     p.GetFileName(),
		Ranges:       ranges,
		SymbolInfo:   symbols,
		Score:        p.GetScore(),
		DebugScore:   p.GetDebugScore(),
	}
}

func (cm *ChunkMatch) ToProto() *proto.ChunkMatch {
	ranges := make([]*proto.Range, len(cm.Ranges))
	for i, r := range cm.Ranges {
		ranges[i] = r.ToProto()
	}

	symbolInfo := make([]*proto.SymbolInfo, len(cm.SymbolInfo))
	for i, si := range cm.SymbolInfo {
		symbolInfo[i] = si.ToProto()
	}

	return &proto.ChunkMatch{
		Content:      cm.Content,
		ContentStart: cm.ContentStart.ToProto(),
		FileName:     cm.FileName,
		Ranges:       ranges,
		SymbolInfo:   symbolInfo,
		Score:        cm.Score,
		DebugScore:   cm.DebugScore,
	}
}

func RangeFromProto(p *proto.Range) Range {
	return Range{
		Start: LocationFromProto(p.GetStart()),
		End:   LocationFromProto(p.GetEnd()),
	}
}

func (r *Range) ToProto() *proto.Range {
	return &proto.Range{
		Start: r.Start.ToProto(),
		End:   r.End.ToProto(),
	}
}

func LocationFromProto(p *proto.Location) Location {
	return Location{
		ByteOffset: p.GetByteOffset(),
		LineNumber: p.GetLineNumber(),
		Column:     p.GetColumn(),
	}
}

func (l *Location) ToProto() *proto.Location {
	return &proto.Location{
		ByteOffset: l.ByteOffset,
		LineNumber: l.LineNumber,
		Column:     l.Column,
	}
}

func LineMatchFromProto(p *proto.LineMatch) LineMatch {
	lineFragments := make([]LineFragmentMatch, len(p.GetLineFragments()))
	for i, lineFragment := range p.GetLineFragments() {
		lineFragments[i] = LineFragmentMatchFromProto(lineFragment)
	}

	return LineMatch{
		Line:          p.GetLine(),
		LineStart:     int(p.GetLineStart()),
		LineEnd:       int(p.GetLineEnd()),
		LineNumber:    int(p.GetLineNumber()),
		Before:        p.GetBefore(),
		After:         p.GetAfter(),
		FileName:      p.GetFileName(),
		Score:         p.GetScore(),
		DebugScore:    p.GetDebugScore(),
		LineFragments: lineFragments,
	}
}

func (lm *LineMatch) ToProto() *proto.LineMatch {
	fragments := make([]*proto.LineFragmentMatch, len(lm.LineFragments))
	for i, fragment := range lm.LineFragments {
		fragments[i] = fragment.ToProto()
	}

	return &proto.LineMatch{
		Line:          lm.Line,
		LineStart:     int64(lm.LineStart),
		LineEnd:       int64(lm.LineEnd),
		LineNumber:    int64(lm.LineNumber),
		Before:        lm.Before,
		After:         lm.After,
		FileName:      lm.FileName,
		Score:         lm.Score,
		DebugScore:    lm.DebugScore,
		LineFragments: fragments,
	}
}

func SymbolFromProto(p *proto.SymbolInfo) *Symbol {
	if p == nil {
		return nil
	}

	return &Symbol{
		Sym:        p.GetSym(),
		Kind:       p.GetKind(),
		Parent:     p.GetParent(),
		ParentKind: p.GetParentKind(),
	}
}

func (s *Symbol) ToProto() *proto.SymbolInfo {
	if s == nil {
		return nil
	}

	return &proto.SymbolInfo{
		Sym:        s.Sym,
		Kind:       s.Kind,
		Parent:     s.Parent,
		ParentKind: s.ParentKind,
	}
}

func LineFragmentMatchFromProto(p *proto.LineFragmentMatch) LineFragmentMatch {
	return LineFragmentMatch{
		LineOffset:  int(p.GetLineOffset()),
		Offset:      p.GetOffset(),
		MatchLength: int(p.GetMatchLength()),
		SymbolInfo:  SymbolFromProto(p.GetSymbolInfo()),
	}
}

func (lfm *LineFragmentMatch) ToProto() *proto.LineFragmentMatch {
	return &proto.LineFragmentMatch{
		LineOffset:  int64(lfm.LineOffset),
		Offset:      lfm.Offset,
		MatchLength: int64(lfm.MatchLength),
		SymbolInfo:  lfm.SymbolInfo.ToProto(),
	}
}

func FlushReasonFromProto(p proto.FlushReason) FlushReason {
	switch p {
	case proto.FlushReason_FLUSH_REASON_TIMER_EXPIRED:
		return FlushReasonTimerExpired
	case proto.FlushReason_FLUSH_REASON_FINAL_FLUSH:
		return FlushReasonFinalFlush
	case proto.FlushReason_FLUSH_REASON_MAX_SIZE:
		return FlushReasonMaxSize
	default:
		return FlushReason(0)
	}
}

func (fr FlushReason) ToProto() proto.FlushReason {
	switch fr {
	case FlushReasonTimerExpired:
		return proto.FlushReason_FLUSH_REASON_TIMER_EXPIRED
	case FlushReasonFinalFlush:
		return proto.FlushReason_FLUSH_REASON_FINAL_FLUSH
	case FlushReasonMaxSize:
		return proto.FlushReason_FLUSH_REASON_MAX_SIZE
	default:
		return proto.FlushReason_FLUSH_REASON_UNKNOWN_UNSPECIFIED
	}
}

// Generate valid reasons for quickchecks
func (fr FlushReason) Generate(rand *rand.Rand, size int) reflect.Value {
	switch rand.Int() % 4 {
	case 1:
		return reflect.ValueOf(FlushReasonMaxSize)
	case 2:
		return reflect.ValueOf(FlushReasonFinalFlush)
	case 3:
		return reflect.ValueOf(FlushReasonTimerExpired)
	default:
		return reflect.ValueOf(FlushReason(0))
	}
}

func StatsFromProto(p *proto.Stats) Stats {
	return Stats{
		ContentBytesLoaded:    p.GetContentBytesLoaded(),
		IndexBytesLoaded:      p.GetIndexBytesLoaded(),
		Crashes:               int(p.GetCrashes()),
		Duration:              p.GetDuration().AsDuration(),
		FileCount:             int(p.GetFileCount()),
		ShardFilesConsidered:  int(p.GetShardFilesConsidered()),
		FilesConsidered:       int(p.GetFilesConsidered()),
		FilesLoaded:           int(p.GetFilesLoaded()),
		FilesSkipped:          int(p.GetFilesSkipped()),
		ShardsScanned:         int(p.GetShardsScanned()),
		ShardsSkipped:         int(p.GetShardsSkipped()),
		ShardsSkippedFilter:   int(p.GetShardsSkippedFilter()),
		MatchCount:            int(p.GetMatchCount()),
		NgramMatches:          int(p.GetNgramMatches()),
		NgramLookups:          int(p.GetNgramLookups()),
		Wait:                  p.GetWait().AsDuration(),
		MatchTreeConstruction: p.GetMatchTreeConstruction().AsDuration(),
		MatchTreeSearch:       p.GetMatchTreeSearch().AsDuration(),
		RegexpsConsidered:     int(p.GetRegexpsConsidered()),
		FlushReason:           FlushReasonFromProto(p.GetFlushReason()),
	}
}

func (s *Stats) ToProto() *proto.Stats {
	return &proto.Stats{
		ContentBytesLoaded:    s.ContentBytesLoaded,
		IndexBytesLoaded:      s.IndexBytesLoaded,
		Crashes:               int64(s.Crashes),
		Duration:              durationpb.New(s.Duration),
		FileCount:             int64(s.FileCount),
		ShardFilesConsidered:  int64(s.ShardFilesConsidered),
		FilesConsidered:       int64(s.FilesConsidered),
		FilesLoaded:           int64(s.FilesLoaded),
		FilesSkipped:          int64(s.FilesSkipped),
		ShardsScanned:         int64(s.ShardsScanned),
		ShardsSkipped:         int64(s.ShardsSkipped),
		ShardsSkippedFilter:   int64(s.ShardsSkippedFilter),
		MatchCount:            int64(s.MatchCount),
		NgramMatches:          int64(s.NgramMatches),
		NgramLookups:          int64(s.NgramLookups),
		Wait:                  durationpb.New(s.Wait),
		MatchTreeConstruction: durationpb.New(s.MatchTreeConstruction),
		MatchTreeSearch:       durationpb.New(s.MatchTreeSearch),
		RegexpsConsidered:     int64(s.RegexpsConsidered),
		FlushReason:           s.FlushReason.ToProto(),
	}
}

func ProgressFromProto(p *proto.Progress) Progress {
	return Progress{
		Priority:           p.GetPriority(),
		MaxPendingPriority: p.GetMaxPendingPriority(),
	}
}

func (p *Progress) ToProto() *proto.Progress {
	return &proto.Progress{
		Priority:           p.Priority,
		MaxPendingPriority: p.MaxPendingPriority,
	}
}

func SearchResultFromStreamProto(p *proto.StreamSearchResponse, repoURLs, lineFragments map[string]string) *SearchResult {
	if p == nil {
		return nil
	}

	return SearchResultFromProto(p.GetResponseChunk(), repoURLs, lineFragments)
}

func SearchResultFromProto(p *proto.SearchResponse, repoURLs, lineFragments map[string]string) *SearchResult {
	if p == nil {
		return nil
	}

	files := make([]FileMatch, len(p.GetFiles()))
	for i, file := range p.GetFiles() {
		files[i] = FileMatchFromProto(file)
	}

	return &SearchResult{
		Stats:    StatsFromProto(p.GetStats()),
		Progress: ProgressFromProto(p.GetProgress()),

		Files: files,

		RepoURLs:      repoURLs,
		LineFragments: lineFragments,
	}
}

func (sr *SearchResult) ToProto() *proto.SearchResponse {
	if sr == nil {
		return nil
	}

	files := make([]*proto.FileMatch, len(sr.Files))
	for i, file := range sr.Files {
		files[i] = file.ToProto()
	}

	return &proto.SearchResponse{
		Stats:    sr.Stats.ToProto(),
		Progress: sr.Progress.ToProto(),

		Files: files,
	}
}

func (sr *SearchResult) ToStreamProto() *proto.StreamSearchResponse {
	if sr == nil {
		return nil
	}

	return &proto.StreamSearchResponse{ResponseChunk: sr.ToProto()}
}

func RepositoryBranchFromProto(p *proto.RepositoryBranch) RepositoryBranch {
	return RepositoryBranch{
		Name:    p.GetName(),
		Version: p.GetVersion(),
	}
}

func (r *RepositoryBranch) ToProto() *proto.RepositoryBranch {
	return &proto.RepositoryBranch{
		Name:    r.Name,
		Version: r.Version,
	}
}

func RepositoryFromProto(p *proto.Repository) Repository {
	branches := make([]RepositoryBranch, len(p.GetBranches()))
	for i, branch := range p.GetBranches() {
		branches[i] = RepositoryBranchFromProto(branch)
	}

	subRepoMap := make(map[string]*Repository, len(p.GetSubRepoMap()))
	for name, repo := range p.GetSubRepoMap() {
		r := RepositoryFromProto(repo)
		subRepoMap[name] = &r
	}

	fileTombstones := make(map[string]struct{}, len(p.GetFileTombstones()))
	for _, file := range p.GetFileTombstones() {
		fileTombstones[file] = struct{}{}
	}

	return Repository{
		ID:                   p.GetId(),
		Name:                 p.GetName(),
		URL:                  p.GetUrl(),
		Source:               p.GetSource(),
		Branches:             branches,
		SubRepoMap:           subRepoMap,
		CommitURLTemplate:    p.GetCommitUrlTemplate(),
		FileURLTemplate:      p.GetFileUrlTemplate(),
		LineFragmentTemplate: p.GetLineFragmentTemplate(),
		priority:             p.GetPriority(),
		RawConfig:            p.GetRawConfig(),
		Rank:                 uint16(p.GetRank()),
		IndexOptions:         p.GetIndexOptions(),
		HasSymbols:           p.GetHasSymbols(),
		Tombstone:            p.GetTombstone(),
		LatestCommitDate:     p.GetLatestCommitDate().AsTime(),
		FileTombstones:       fileTombstones,
	}
}

func (r *Repository) ToProto() *proto.Repository {
	if r == nil {
		return nil
	}

	branches := make([]*proto.RepositoryBranch, len(r.Branches))
	for i, branch := range r.Branches {
		branches[i] = branch.ToProto()
	}

	subRepoMap := make(map[string]*proto.Repository, len(r.SubRepoMap))
	for name, repo := range r.SubRepoMap {
		subRepoMap[name] = repo.ToProto()
	}

	fileTombstones := make([]string, 0, len(r.FileTombstones))
	for file := range r.FileTombstones {
		fileTombstones = append(fileTombstones, file)
	}

	return &proto.Repository{
		Id:                   r.ID,
		Name:                 r.Name,
		Url:                  r.URL,
		Source:               r.Source,
		Branches:             branches,
		SubRepoMap:           subRepoMap,
		CommitUrlTemplate:    r.CommitURLTemplate,
		FileUrlTemplate:      r.FileURLTemplate,
		LineFragmentTemplate: r.LineFragmentTemplate,
		Priority:             r.priority,
		RawConfig:            r.RawConfig,
		Rank:                 uint32(r.Rank),
		IndexOptions:         r.IndexOptions,
		HasSymbols:           r.HasSymbols,
		Tombstone:            r.Tombstone,
		LatestCommitDate:     timestamppb.New(r.LatestCommitDate),
		FileTombstones:       fileTombstones,
	}
}

func IndexMetadataFromProto(p *proto.IndexMetadata) IndexMetadata {
	languageMap := make(map[string]uint16, len(p.GetLanguageMap()))
	for language, id := range p.GetLanguageMap() {
		languageMap[language] = uint16(id)
	}

	return IndexMetadata{
		IndexFormatVersion:    int(p.GetIndexFormatVersion()),
		IndexFeatureVersion:   int(p.GetIndexFeatureVersion()),
		IndexMinReaderVersion: int(p.GetIndexMinReaderVersion()),
		IndexTime:             p.GetIndexTime().AsTime(),
		PlainASCII:            p.GetPlainAscii(),
		LanguageMap:           languageMap,
		ZoektVersion:          p.GetZoektVersion(),
		ID:                    p.GetId(),
	}
}

func (m *IndexMetadata) ToProto() *proto.IndexMetadata {
	if m == nil {
		return nil
	}

	languageMap := make(map[string]uint32, len(m.LanguageMap))
	for language, id := range m.LanguageMap {
		languageMap[language] = uint32(id)
	}

	return &proto.IndexMetadata{
		IndexFormatVersion:    int64(m.IndexFormatVersion),
		IndexFeatureVersion:   int64(m.IndexFeatureVersion),
		IndexMinReaderVersion: int64(m.IndexMinReaderVersion),
		IndexTime:             timestamppb.New(m.IndexTime),
		PlainAscii:            m.PlainASCII,
		LanguageMap:           languageMap,
		ZoektVersion:          m.ZoektVersion,
		Id:                    m.ID,
	}
}

func RepoStatsFromProto(p *proto.RepoStats) RepoStats {
	return RepoStats{
		Repos:                      int(p.GetRepos()),
		Shards:                     int(p.GetShards()),
		Documents:                  int(p.GetDocuments()),
		IndexBytes:                 p.GetIndexBytes(),
		ContentBytes:               p.GetContentBytes(),
		NewLinesCount:              p.GetNewLinesCount(),
		DefaultBranchNewLinesCount: p.GetDefaultBranchNewLinesCount(),
		OtherBranchesNewLinesCount: p.GetOtherBranchesNewLinesCount(),
	}
}

func (s *RepoStats) ToProto() *proto.RepoStats {
	return &proto.RepoStats{
		Repos:                      int64(s.Repos),
		Shards:                     int64(s.Shards),
		Documents:                  int64(s.Documents),
		IndexBytes:                 s.IndexBytes,
		ContentBytes:               s.ContentBytes,
		NewLinesCount:              s.NewLinesCount,
		DefaultBranchNewLinesCount: s.DefaultBranchNewLinesCount,
		OtherBranchesNewLinesCount: s.OtherBranchesNewLinesCount,
	}
}

func RepoListEntryFromProto(p *proto.RepoListEntry) *RepoListEntry {
	if p == nil {
		return nil
	}

	return &RepoListEntry{
		Repository:    RepositoryFromProto(p.GetRepository()),
		IndexMetadata: IndexMetadataFromProto(p.GetIndexMetadata()),
		Stats:         RepoStatsFromProto(p.GetStats()),
	}
}

func (r *RepoListEntry) ToProto() *proto.RepoListEntry {
	if r == nil {
		return nil
	}

	return &proto.RepoListEntry{
		Repository:    r.Repository.ToProto(),
		IndexMetadata: r.IndexMetadata.ToProto(),
		Stats:         r.Stats.ToProto(),
	}
}

func MinimalRepoListEntryFromProto(p *proto.MinimalRepoListEntry) MinimalRepoListEntry {
	branches := make([]RepositoryBranch, len(p.GetBranches()))
	for i, branch := range p.GetBranches() {
		branches[i] = RepositoryBranchFromProto(branch)
	}

	return MinimalRepoListEntry{
		HasSymbols:    p.GetHasSymbols(),
		Branches:      branches,
		IndexTimeUnix: p.GetIndexTimeUnix(),
	}
}

func (m *MinimalRepoListEntry) ToProto() *proto.MinimalRepoListEntry {
	branches := make([]*proto.RepositoryBranch, len(m.Branches))
	for i, branch := range m.Branches {
		branches[i] = branch.ToProto()
	}
	return &proto.MinimalRepoListEntry{
		HasSymbols:    m.HasSymbols,
		Branches:      branches,
		IndexTimeUnix: m.IndexTimeUnix,
	}
}

func RepoListFromProto(p *proto.ListResponse) *RepoList {
	repos := make([]*RepoListEntry, len(p.GetRepos()))
	for i, repo := range p.GetRepos() {
		repos[i] = RepoListEntryFromProto(repo)
	}

	reposMap := make(map[uint32]MinimalRepoListEntry, len(p.GetReposMap()))
	for id, mle := range p.GetReposMap() {
		reposMap[id] = MinimalRepoListEntryFromProto(mle)
	}

	return &RepoList{
		Repos:    repos,
		ReposMap: reposMap,
		Crashes:  int(p.GetCrashes()),
		Stats:    RepoStatsFromProto(p.GetStats()),
	}
}

func (r *RepoList) ToProto() *proto.ListResponse {
	repos := make([]*proto.RepoListEntry, len(r.Repos))
	for i, repo := range r.Repos {
		repos[i] = repo.ToProto()
	}

	reposMap := make(map[uint32]*proto.MinimalRepoListEntry, len(r.ReposMap))
	for id, repo := range r.ReposMap {
		reposMap[id] = repo.ToProto()
	}

	return &proto.ListResponse{
		Repos:    repos,
		ReposMap: reposMap,
		Crashes:  int64(r.Crashes),
		Stats:    r.Stats.ToProto(),
	}
}

func (l *ListOptions) ToProto() *proto.ListOptions {
	if l == nil {
		return nil
	}
	var field proto.ListOptions_RepoListField
	switch l.Field {
	case RepoListFieldRepos:
		field = proto.ListOptions_REPO_LIST_FIELD_REPOS
	case RepoListFieldReposMap:
		field = proto.ListOptions_REPO_LIST_FIELD_REPOS_MAP
	}

	return &proto.ListOptions{
		Field: field,
	}
}

func ListOptionsFromProto(p *proto.ListOptions) *ListOptions {
	if p == nil {
		return nil
	}
	var field RepoListField
	switch p.GetField() {
	case proto.ListOptions_REPO_LIST_FIELD_REPOS:
		field = RepoListFieldRepos
	case proto.ListOptions_REPO_LIST_FIELD_REPOS_MAP:
		field = RepoListFieldReposMap
	}
	return &ListOptions{
		Field: field,
	}
}

func SearchOptionsFromProto(p *proto.SearchOptions) *SearchOptions {
	if p == nil {
		return nil
	}

	return &SearchOptions{
		EstimateDocCount:       p.GetEstimateDocCount(),
		Whole:                  p.GetWhole(),
		ShardMaxMatchCount:     int(p.GetShardMaxMatchCount()),
		TotalMaxMatchCount:     int(p.GetTotalMaxMatchCount()),
		ShardRepoMaxMatchCount: int(p.GetShardRepoMaxMatchCount()),
		MaxWallTime:            p.GetMaxWallTime().AsDuration(),
		FlushWallTime:          p.GetFlushWallTime().AsDuration(),
		MaxDocDisplayCount:     int(p.GetMaxDocDisplayCount()),
		MaxMatchDisplayCount:   int(p.GetMaxMatchDisplayCount()),
		NumContextLines:        int(p.GetNumContextLines()),
		ChunkMatches:           p.GetChunkMatches(),
		UseDocumentRanks:       p.GetUseDocumentRanks(),
		DocumentRanksWeight:    p.GetDocumentRanksWeight(),
		Trace:                  p.GetTrace(),
		DebugScore:             p.GetDebugScore(),
		UseBM25Scoring:         p.GetUseBm25Scoring(),
	}
}

func (s *SearchOptions) ToProto() *proto.SearchOptions {
	if s == nil {
		return nil
	}

	return &proto.SearchOptions{
		EstimateDocCount:       s.EstimateDocCount,
		Whole:                  s.Whole,
		ShardMaxMatchCount:     int64(s.ShardMaxMatchCount),
		TotalMaxMatchCount:     int64(s.TotalMaxMatchCount),
		ShardRepoMaxMatchCount: int64(s.ShardRepoMaxMatchCount),
		MaxWallTime:            durationpb.New(s.MaxWallTime),
		FlushWallTime:          durationpb.New(s.FlushWallTime),
		MaxDocDisplayCount:     int64(s.MaxDocDisplayCount),
		MaxMatchDisplayCount:   int64(s.MaxMatchDisplayCount),
		NumContextLines:        int64(s.NumContextLines),
		ChunkMatches:           s.ChunkMatches,
		UseDocumentRanks:       s.UseDocumentRanks,
		DocumentRanksWeight:    s.DocumentRanksWeight,
		Trace:                  s.Trace,
		DebugScore:             s.DebugScore,
		UseBm25Scoring:         s.UseBM25Scoring,
	}
}
