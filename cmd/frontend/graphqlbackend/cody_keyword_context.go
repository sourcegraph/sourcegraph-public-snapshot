package graphqlbackend

import (
	"context"
	"fmt"
	"sort"

	"github.com/bbalet/stopwords"
	"math"
	"regexp"
	"strings"

	stemmer "github.com/reiver/go-porterstemmer"
)

var keywordSearchPatternType = "regexp"

const MAX_FILES_TO_RANK = 50000
const KEYWORD_CONTEXT_SEARCH_TIMEOUT = "0.5s"

type KeywordSearchArgs struct {
	Query string
	Repo  *string
	Limit *int32
}

func (r *schemaResolver) CodyKeywordContext(ctx context.Context, args *KeywordSearchArgs) ([]*FileMatchResolver, error) {
	limit := int32(20)
	if args.Limit != nil {
		limit = *args.Limit
	}

	repo := ".*"
	if args.Repo != nil {
		repo = *args.Repo
	}

	terms := queryTerms{}
	for _, part := range regexp.MustCompile(`\s+`).Split(args.Query, -1) {
		term := termFromStr(part)
		if len(term.Terms()) > 0 {
			terms = append(terms, term)
		}
	}
	query := fmt.Sprintf(`context:global select:file type:file count:200 timeout:%s repo:%s (?:%s)`, KEYWORD_CONTEXT_SEARCH_TIMEOUT, repo, strings.Join(terms.All(), "|"))

	searchResolver, err := NewBatchSearchImplementer(ctx, r.logger, r.db, r.enterpriseSearchJobs, &SearchArgs{Version: "V3", Query: query, PatternType: &keywordSearchPatternType})
	if err != nil {
		return []*FileMatchResolver{}, err
	}

	resultsResolver, err := searchResolver.Results(ctx)
	if err != nil {
		return []*FileMatchResolver{}, err
	}

	allResults := resultsResolver.Results()
	results := allResults[0:int(math.Min(float64(MAX_FILES_TO_RANK), float64(len(allResults))))]

	documents := make(resultDocuments, 0, len(results))

	for _, result := range results {
		fileMatchResolver, _ := result.ToFileMatch()
		if fileMatchResolver == nil {
			continue
		}

		file := fileMatchResolver.File()
		path := file.Path()
		content, err := file.Content(ctx, &GitTreeContentPageArgs{})
		if err != nil {
			continue
		}

		documents = append(documents, &resultDocument{
			matchResolver: fileMatchResolver,
			file:          file,
			path:          path,
			content:       content + " " + strings.Join(termsFromPath(path).All(), " "),
		})
	}

	stats := findStats(&terms, documents[0:int(math.Min(float64(MAX_FILES_TO_RANK), float64(len(documents))))])

	idfDict := idf(stats.TermTotalFiles, stats.TotalFiles)
	queryStems := terms.Stems()

	for _, document := range documents {
		score := idfLogScore(queryStems, idfDict, stats.FileTermsCount[document.path])

		if stats.FileBytesCount[document.path] > 10000 {
			score *= 0.1
		}

		document.score = score
	}

	sort.Sort(sort.Reverse(documents))

	matches := make([]*FileMatchResolver, 0, 0)
	for _, d := range documents[0:int(math.Min(float64(limit), float64(len(documents))))] {
		matches = append(matches, d.matchResolver)
	}

	return matches, nil
}

type resultDocument struct {
	matchResolver *FileMatchResolver
	file          *GitTreeEntryResolver
	path          string
	content       string
	score         float64
}

type resultDocuments []*resultDocument

func (r resultDocuments) Len() int           { return len(r) }
func (r resultDocuments) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r resultDocuments) Less(i, j int) bool { return r[i].score < r[j].score }

type documentStats struct {
	TotalFiles     int
	TermTotalFiles map[string]int
	FileTermsCount map[string]map[string]int
	FileBytesCount map[string]int
}

func findStats(terms *queryTerms, documents resultDocuments) *documentStats {
	stats := documentStats{}

	stats.TotalFiles = len(documents)
	stats.TermTotalFiles = map[string]int{}
	stats.FileTermsCount = map[string]map[string]int{}
	stats.FileBytesCount = map[string]int{}

	for _, document := range documents {
		for _, term := range *terms {
			r := regexp.MustCompile(fmt.Sprintf("(?i)(?:%s)", strings.Join(term.Terms(), "|")))
			matchesCount := len(r.FindAllString(document.content, -1))

			if matchesCount > 0 {
				stats.TermTotalFiles[term.stem] += 1
			}

			fileTerms := stats.FileTermsCount[document.path]
			if fileTerms == nil {
				fileTerms = map[string]int{}
			}

			fileTerms[term.stem] = matchesCount
			stats.FileTermsCount[document.path] = fileTerms
		}

		stats.FileBytesCount[document.path] = len(document.content)
	}

	return &stats
}

type queryTerm struct {
	base        string
	stem        string
	escaped     string
	escapedStem string
}

func (t *queryTerm) Terms() []string {
	ts := []string{}

	if t.base != "" {
		ts = append(ts, t.base)

		if t.base != t.stem && t.stem != "" {
			ts = append(ts, t.stem)
		}

		if t.base != t.escaped && t.escaped != "" {
			ts = append(ts, t.escaped)

			if t.escaped != t.escapedStem && t.escapedStem != "" {
				ts = append(ts, t.escapedStem)
			}
		}
	}

	return ts
}

type queryTerms []*queryTerm

func (ts *queryTerms) All() []string {
	terms := []string{}

	for _, t := range *ts {
		terms = append(terms, t.Terms()...)
	}

	return terms
}

func (ts *queryTerms) Stems() []string {
	stems := []string{}

	for _, t := range *ts {
		stem := t.stem
		if stem != "" {
			stems = append(stems, t.stem)
		}
	}

	return stems
}

func termFromStr(part string) *queryTerm {
	base := strings.ToLower(strings.TrimSpace(stopwords.CleanString(part, "en", true)))
	stem := stemmer.StemString(base)
	escaped := strings.ToLower(strings.TrimSpace(regexp.MustCompile(`[$()*+./?[\\\]^{|}-]`).ReplaceAllString(part, `\$0`)))
	escapedStem := stemmer.StemString(escaped)

	return &queryTerm{
		base:        base,
		stem:        stem,
		escaped:     escaped,
		escapedStem: escapedStem,
	}
}

func termsFromPath(path string) *queryTerms {
	terms := queryTerms{}
	for _, part := range regexp.MustCompile(`[/.]`).Split(path, -1) {
		terms = append(terms, termFromStr(part))
	}
	return &terms
}

func idf(termTotalFiles map[string]int, totalFiles int) map[string]float64 {
	logTotal := math.Log(float64(totalFiles))
	idf := map[string]float64{}

	for stem, count := range termTotalFiles {
		idf[stem] = logTotal - math.Log(float64(count))
	}

	return idf
}

func idfLogScore(stems []string, idfDict map[string]float64, termCounts map[string]int) float64 {
	score := float64(0)
	for _, stem := range stems {
		count := termCounts[stem]
		logScore := float64(0)
		if count != 0 {
			logScore = math.Log10(float64(count)) + 1
		}

		idfScore := idfDict[stem]
		if idfScore == 0 {
			idfScore = 1
		}

		idfLogScore := idfScore * logScore
		score += idfLogScore
	}
	return score
}
