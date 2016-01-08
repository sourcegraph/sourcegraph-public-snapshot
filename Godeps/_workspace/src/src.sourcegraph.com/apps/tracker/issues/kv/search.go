package kv

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzers/keyword_analyzer"
	"github.com/blevesearch/bleve/analysis/language/en"
	"golang.org/x/net/context"
	"src.sourcegraph.com/apps/tracker/issues"

	"src.sourcegraph.com/sourcegraph/conf/feature"
)

const indexFilename = "index.bleve"

var index bleve.Index

// issueDocument represents the data that will be stored for each issue in the
// search index.
type issueDocument struct {
	ID   string
	Repo string
	issues.Issue
	Comments []issues.Comment
}

func (s service) Search(ctx context.Context, opt issues.SearchOptions) (issues.SearchResponse, error) {
	repoIsInitalized, err := index.GetInternal([]byte(opt.Repo.URI))
	if err != nil {
		return issues.SearchResponse{}, err
	}
	if len(repoIsInitalized) < 1 {
		// TODO run this operation in parallel?
		s.indexAll(ctx, opt.Repo)
		index.SetInternal([]byte(opt.Repo.URI), []byte{1})
	}

	mq1 := bleve.NewMatchQuery(opt.Query).SetField("Issue.Title").SetBoost(3)
	mq2 := bleve.NewMatchQuery(opt.Query).SetField("Comments.Body")
	tq := bleve.NewTermQuery(opt.Repo.URI).SetField("Repo")
	q := bleve.NewBooleanQueryMinShould(
		[]bleve.Query{tq},
		[]bleve.Query{mq1, mq2},
		[]bleve.Query{},
		1,
	)
	sr := bleve.NewSearchRequest(q)
	sr.Highlight = bleve.NewHighlightWithStyle("html")
	sr.Highlight.AddField("Issue.Title")
	sr.Highlight.AddField("Comments.Body")
	sr.Fields = []string{"*"}
	sr.From = (opt.Page - 1) * opt.PerPage
	sr.Size = opt.PerPage

	res, err := index.Search(sr)
	if err != nil {
		return issues.SearchResponse{}, err
	}

	var results []issues.SearchResult
	for _, hit := range res.Hits {
		var comment string
		// HACK: We only want to display a comment as part of a search result its
		// body text matched the current query. We can determine this by checking
		// if the fragment has highlighted the matching portion of the comment with
		// an HTML "mark" tag. Ideally there should be an easier way to determine
		// if a specific field contains a match, but this will suffice.
		if strings.Contains(hit.Fragments["Comments.Body"][0], "<mark>") {
			comment = hit.Fragments["Comments.Body"][0]
		}

		// bleve stores time fields in time.RFC3339 format by default.
		createdAt, err := time.Parse(time.RFC3339, hit.Fields["Issue.Comment.CreatedAt"].(string))
		if err != nil {
			return issues.SearchResponse{}, err
		}

		// TODO: Unmarshal fields from the hit.Fields map in a more type-safe manner.
		results = append(results, issues.SearchResult{
			ID:      hit.Fields["ID"].(string),
			Title:   template.HTML(hit.Fragments["Issue.Title"][0]),
			Comment: template.HTML(comment),
			User: issues.User{
				Login:   hit.Fields["Issue.Comment.User.Login"].(string),
				HTMLURL: template.URL(hit.Fields["Issue.Comment.User.HTMLURL"].(string)),
			},
			CreatedAt: createdAt,
			State:     issues.State(hit.Fields["Issue.State"].(string)),
		})
	}

	return issues.SearchResponse{Results: results, Total: res.Total}, nil
}

// index adds or replaces a single issue in a repo's search index.
func (s service) index(ctx context.Context, rs issues.RepoSpec, id uint64) {
	if !feature.Features.TrackerSearch {
		return
	}

	issue, err := s.Get(ctx, rs, id)
	if err != nil {
		log.Println("IndexIssue:", err)
	}

	comments, err := s.ListComments(ctx, rs, issue.ID, nil)
	if err != nil {
		log.Println("IndexIssue:", err)
	}

	index.Index(globalID(rs.URI, issue.ID), issueDocument{
		ID:       strconv.FormatUint(issue.ID, 10),
		Repo:     rs.URI,
		Issue:    issue,
		Comments: comments,
	})
}

// indexAll generates an index for all issues for a given repo.
func (s service) indexAll(ctx context.Context, rs issues.RepoSpec) {
	batch := index.NewBatch()
	allIssues, err := s.List(ctx, issues.RepoSpec{URI: rs.URI}, issues.IssueListOptions{
		State: issues.AllStates,
	})
	if err != nil {
		log.Println(err)
	}

	for _, issue := range allIssues {
		comments, err := s.ListComments(ctx, issues.RepoSpec{URI: rs.URI}, issue.ID, nil)
		if err != nil {
			log.Println("indexAll:", err)
		}

		batch.Index(globalID(rs.URI, issue.ID), issueDocument{
			ID:       strconv.FormatUint(issue.ID, 10),
			Repo:     rs.URI,
			Issue:    issue,
			Comments: comments,
		})
	}

	if batch.Size() > 0 {
		err = index.Batch(batch)
		if err != nil {
			log.Println("indexAll:", err)
		}
	}
}

func getIndex() bleve.Index {
	path := filepath.Join(os.Getenv("SGPATH"), "tmp", "tracker")

	index, err := bleve.Open(filepath.Join(path, indexFilename))
	if err == bleve.ErrorIndexPathDoesNotExist {
		return newIndex()
	} else if err != nil {
		log.Fatal(err)
	}

	return index
}

func newIndex() bleve.Index {
	path := filepath.Join(os.Getenv("SGPATH"), "tmp", "tracker")
	os.MkdirAll(path, os.FileMode(0755))

	kwMapping := bleve.NewTextFieldMapping()
	kwMapping.Analyzer = keyword_analyzer.Name
	stdMapping := bleve.NewTextFieldMapping()
	stdMapping.Analyzer = en.AnalyzerName

	issueMapping := bleve.NewDocumentMapping()
	issueMapping.AddFieldMappingsAt("Title", stdMapping)

	commentsMapping := bleve.NewDocumentMapping()
	commentsMapping.AddFieldMappingsAt("Body", stdMapping)

	documentMapping := bleve.NewDocumentMapping()
	documentMapping.AddFieldMappingsAt("Repo", kwMapping)
	documentMapping.AddSubDocumentMapping("Issue", issueMapping)
	documentMapping.AddSubDocumentMapping("Comments", commentsMapping)

	mapping := bleve.NewIndexMapping()
	mapping.DefaultMapping = documentMapping

	index, err := bleve.New(filepath.Join(path, indexFilename), mapping)
	if err != nil {
		log.Fatal(err)
	}

	return index
}

func globalID(repo string, issueID uint64) string {
	return fmt.Sprintf("%s-%d", repo, issueID)
}
