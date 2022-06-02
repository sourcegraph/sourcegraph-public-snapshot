package jobutil

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
)

// NewFeelingLuckyJob generates an opportunistic search query by applying various rules on
// the input string.
func NewFeelingLuckyJob(inputs *run.SearchInputs, plan query.Plan) []job.Job {
	children := make([]job.Job, 0, len(plan))
	for _, b := range plan {
		firstJob, err := NewBasicJob(inputs, b)
		if err != nil {
			panic("basic invalid job, very rare corner case!") // FIXME
		}
		generatedJobs := []job.Job{}
		for _, newBasic := range BuildBasic(b) { // Basic queries generated are sequenced.
			child, err := NewBasicJob(inputs, newBasic)
			if err != nil {
				panic("generated an invalid basic query D:")
			}
			generatedJobs = append(generatedJobs, child)
		}
		children = append(children, NewSequentialJob(firstJob, NewOrJob(generatedJobs...)))
		// ^ issue is dedupe, check: node represents a function
		// ^ deduplication? also check repo:github.com/sourcegraph/sourcegraph "patternType"
		// children = append(children, NewOrJob(sequentialJobs...))
	}
	return children
}

func BuildBasic(b query.Basic) []query.Basic {
	bs := []query.Basic{}
	if g := UnquotedPatterns(b); g != nil {
		bs = append(bs, *g)
	}
	/* // races UnquotePatterns! (Or guarantees if we see everything, but these are different matches)
	if g := UnorderedPatterns(b); g != nil {
		bs = append(bs, *g)
	}
	*/
	// compose unquoted => unordered separately
	/* // is this racing the above, and then we get 500 results?
	if g := UnquotedPatterns(b); g != nil {
		if h := UnorderedPatterns(*g); h != nil {
			bs = append(bs, *h)
		}
	}
	*/
	/*
		if g := PatternsAsRepoPaths(b); g != nil {
			if h := UnquotedPatterns(*g); h != nil {
				if i := UnorderedPatterns(*h); i != nil {
					bs = append(bs, *i)
				}
				bs = append(bs, *h)
			}
			bs = append(bs, *g)
		}
	*/
	return bs
}

// UnquotePatterns unquotes quoted patterns in queries (removing
// quotes, and honoring escape sequences).
func UnquotedPatterns(b query.Basic) *query.Basic {
	// Go all the way back to the original string, and parse it as regex to discover quoted patterns.
	rawParseTree, err := query.Parse(query.StringHuman(b.ToParseTree()), query.SearchTypeRegex)
	if err != nil {
		return nil
	}

	changed := false

	newParseTree := query.MapPattern(rawParseTree, func(value string, negated bool, annotation query.Annotation) query.Node {
		if annotation.Labels.IsSet(query.Quoted) {
			fmt.Printf("see quoted: %s\n", value)
			/*
				v, err := strconv.Unquote(value) // Lazy--shouldn't actually unquote `` but whatever!!
				if err != nil {
					fmt.Printf("error")
					return query.Pattern{
						Value:      value,
						Negated:    negated,
						Annotation: annotation,
					}
				}
			*/
			changed = true
			annotation.Labels.Unset(query.Quoted)
			annotation.Labels.Set(query.Literal)
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}
		return query.Pattern{
			Value:      value,
			Negated:    negated,
			Annotation: annotation,
		}
	})

	if !changed {
		return nil
	}

	newNodes, err := query.Sequence(query.For(query.SearchTypeLiteralDefault))(newParseTree)
	if err != nil {
		return nil
	}

	newBasic, err := query.ToBasicQuery(newNodes)
	if err != nil {
		panic(err.Error())
	}

	return &newBasic
}

// UnorderedPatterns generates a query that interprets all recognized patterns
// as unordered terms (`and`-ed terms). Brittle assumption: only for queries in
// default/literal mode, where all terms are space-separated and spaces are
// unescapable, implying we can obtain patterns with a split on space.
func UnorderedPatterns(b query.Basic) *query.Basic {
	var andPatterns []query.Node
	query.VisitPattern([]query.Node{b.Pattern}, func(value string, negated bool, annotation query.Annotation) {
		if negated {
			// append negated terms as-is.
			andPatterns = append(andPatterns, query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			})
			return
		}
		for _, p := range strings.Split(value, " ") {
			andPatterns = append(andPatterns, query.Pattern{
				Value:      p,
				Negated:    negated,
				Annotation: annotation, // danger: preserves invalid range
			})
		}
	})
	return &query.Basic{
		Parameters: b.Parameters,
		Pattern:    query.Operator{Kind: query.And, Operands: andPatterns, Annotation: query.Annotation{}},
	}
}

func PatternsAsRepoPaths(b query.Basic) *query.Basic {
	rawParseTree, err := query.Parse(query.StringHuman(b.ToParseTree()), query.SearchTypeLiteralDefault) // We're going all the way back baby.
	if err != nil {
		return nil
	}

	repoParams := []query.Node{}

	changed := false // important, removing this leads to dupes (why?)

	// gotta collect terms to promote as repo, becauase doing it in-place
	// creates patterns inside concat nodes, which just makes life more
	// difficult. Trying to Map concat nodes is also a pain. Just delete
	// patterns, and add the repo filters back.
	newParseTree := query.MapPattern(rawParseTree, func(value string, negated bool, annotation query.Annotation) query.Node {
		r := regexp.MustCompile(`(https://)?github.com/.*`)
		v := r.FindString(value)
		if v == "" {
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}
		v = strings.TrimPrefix(v, "https://")
		v = strings.TrimPrefix(v, "github.com/")
		v = strings.TrimSuffix(v, "/")

		if v != value {
			changed = true
		}
		repoParams = append(repoParams, query.Parameter{
			Field:   query.FieldRepo,
			Value:   v,
			Negated: negated,
		})
		return nil
	})

	if !changed {
		return nil
	}

	// Gotta reduce this, or we won't be able to partition to basic
	// query--the basic partitioning is not super smart.
	newParseTree = query.NewOperator(append(newParseTree, repoParams...), query.And)
	newNodes, err := query.Sequence(query.For(query.SearchTypeLiteralDefault))(newParseTree)
	if err != nil {
		return nil
	}

	newBasic, err := query.ToBasicQuery(newNodes)
	if err != nil {
		panic(err.Error())
	}

	return &newBasic
}

func ParametersToNodes(parameters []query.Parameter) []query.Node {
	var nodes []query.Node
	for _, n := range parameters {
		nodes = append(nodes, query.Node(n))
	}
	return nodes
}

func NodesToParameters(nodes []query.Node) []query.Parameter {
	b, _ := query.ToBasicQuery(nodes)
	return b.Parameters
}
