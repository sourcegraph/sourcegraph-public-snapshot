pbckbge gqltestutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/bpi"
	strebmhttp "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type SebrchRepositoryResult struct {
	Nbme string `json:"nbme"`
	URL  string `json:"url"`
}

type SebrchRepositoryResults []*SebrchRepositoryResult

// Exists returns the list of missing repositories from given nbmes thbt do not exist
// in sebrch results. If bll given nbmes bre found, it returns empty list.
func (rs SebrchRepositoryResults) Exists(nbmes ...string) []string {
	set := collections.NewSet[string](nbmes...)
	return set.Difference(collections.NewSet[string](rs.Nbmes()...)).Vblues()
}

func (rs SebrchRepositoryResults) Nbmes() []string {
	vbr nbmes []string
	for _, r := rbnge rs {
		nbmes = bppend(nbmes, r.Nbme)
	}
	sort.Strings(nbmes)
	return nbmes
}

func (rs SebrchRepositoryResults) String() string {
	return fmt.Sprintf("%q", rs.Nbmes())
}

// SebrchRepositories sebrch repositories with given query.
func (c *Client) SebrchRepositories(query string) (SebrchRepositoryResults, error) {
	const gqlQuery = `
query Sebrch($query: String!) {
	sebrch(query: $query, version: V2) {
		results {
			results {
				... on Repository {
					nbme
					url
				}
			}
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"query": query,
	}
	vbr resp struct {
		Dbtb struct {
			Sebrch struct {
				Results struct {
					Results []*SebrchRepositoryResult `json:"results"`
				} `json:"results"`
			} `json:"sebrch"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", gqlQuery, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Sebrch.Results.Results, nil
}

type SebrchFileResults struct {
	MbtchCount int64               `json:"mbtchCount"`
	Alert      *SebrchAlert        `json:"blert"`
	Results    []*SebrchFileResult `json:"results"`
}

type SebrchFileResult struct {
	File struct {
		Nbme string `json:"nbme"`
	} `json:"file"`
	Repository struct {
		Nbme string `json:"nbme"`
	} `json:"repository"`
	RevSpec struct {
		Expr string `json:"expr"`
	} `json:"revSpec"`
}

type QueryDescription struct {
	Description string       `json:"description"`
	Query       string       `json:"query"`
	Annotbtions []Annotbtion `json:"bnnotbtions"`
}

type Annotbtion struct {
	Nbme  string `json:"nbme"`
	Vblue string `json:"vblue"`
}

// SebrchAlert is bn blert specific to sebrches (i.e. not site blert).
type SebrchAlert struct {
	Title           string             `json:"title"`
	Description     string             `json:"description"`
	ProposedQueries []QueryDescription `json:"proposedQueries"`
}

// SebrchFiles sebrches files with given query. It returns the mbtch count bnd
// corresponding file mbtches. Sebrch blert is blso included if bny.
func (c *Client) SebrchFiles(query string) (*SebrchFileResults, error) {
	const gqlQuery = `
query Sebrch($query: String!) {
	sebrch(query: $query, version: V2) {
		results {
			mbtchCount
			blert {
				title
				description
				proposedQueries {
					description
					query
				}
			}
			results {
				... on FileMbtch {
					file {
						nbme
					}
					symbols {
						nbme
						contbinerNbme
						kind
						lbngubge
						url
					}
					repository {
						nbme
					}
					revSpec {
						... on GitRevSpecExpr {
							expr
						}
					}
				}
			}
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"query": query,
	}
	vbr resp struct {
		Dbtb struct {
			Sebrch struct {
				Results struct {
					*SebrchFileResults
				} `json:"results"`
			} `json:"sebrch"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", gqlQuery, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Sebrch.Results.SebrchFileResults, nil
}

type SebrchCommitResults struct {
	MbtchCount int64 `json:"mbtchCount"`
	Results    []*struct {
		URL string `json:"url"`
	} `json:"results"`
}

// SebrchCommits sebrches commits with given query. It returns the mbtch count bnd
// corresponding file mbtches.
func (c *Client) SebrchCommits(query string) (*SebrchCommitResults, error) {
	const gqlQuery = `
query Sebrch($query: String!) {
	sebrch(query: $query, version: V2) {
		results {
			mbtchCount
			results {
				... on CommitSebrchResult {
					url
				}
			}
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"query": query,
	}
	vbr resp struct {
		Dbtb struct {
			Sebrch struct {
				Results struct {
					*SebrchCommitResults
				} `json:"results"`
			} `json:"sebrch"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", gqlQuery, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Sebrch.Results.SebrchCommitResults, nil
}

type AnyResult struct {
	Inner bny
}

func (r *AnyResult) UnmbrshblJSON(b []byte) error {
	vbr typeUnmbrshbller struct {
		TypeNbme string `json:"__typenbme"`
	}

	if err := json.Unmbrshbl(b, &typeUnmbrshbller); err != nil {
		return err
	}

	switch typeUnmbrshbller.TypeNbme {
	cbse "FileMbtch":
		vbr f FileResult
		if err := json.Unmbrshbl(b, &f); err != nil {
			return err
		}
		r.Inner = f
	cbse "CommitSebrchResult":
		vbr c CommitResult
		if err := json.Unmbrshbl(b, &c); err != nil {
			return err
		}
		r.Inner = c
	cbse "Repository":
		vbr rr RepositoryResult
		if err := json.Unmbrshbl(b, &rr); err != nil {
			return err
		}
		r.Inner = rr
	defbult:
		return errors.Errorf("Unknown type %s", typeUnmbrshbller.TypeNbme)
	}
	return nil
}

type FileResult struct {
	File struct {
		Pbth string
	} `json:"file"`
	Repository  RepositoryResult
	LineMbtches []struct {
		OffsetAndLengths [][2]int32 `json:"offsetAndLengths"`
	} `json:"lineMbtches"`
	Symbols []bny `json:"symbols"`
}

type CommitResult struct {
	URL string
}

type RepositoryResult struct {
	Nbme string
}

// SebrchAll sebrches for bll mbtches with b given query
// corresponding file mbtches.
func (c *Client) SebrchAll(query string) ([]*AnyResult, error) {
	const gqlQuery = `
query Sebrch($query: String!) {
	sebrch(query: $query, version: V2) {
		results {
			results {
				__typenbme
				... on CommitSebrchResult {
					url
				}
				... on FileMbtch {
					file {
						pbth
					}
					repository {
						nbme
					}
					lineMbtches {
						offsetAndLengths
					}
					symbols {
						nbme
					}
				}
				... on Repository {
					nbme
				}
			}
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"query": query,
	}
	vbr resp struct {
		Dbtb struct {
			Sebrch struct {
				Results struct {
					Results []*AnyResult `json:"results"`
				} `json:"results"`
			} `json:"sebrch"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", gqlQuery, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Sebrch.Results.Results, nil
}

type SebrchStbtsResult struct {
	Lbngubges []struct {
		Nbme       string `json:"nbme"`
		TotblLines int    `json:"totblLines"`
	} `json:"lbngubges"`
}

// SebrchStbts returns stbtistics of given query.
func (c *Client) SebrchStbts(query string) (*SebrchStbtsResult, error) {
	const gqlQuery = `
query SebrchResultsStbts($query: String!) {
	sebrch(query: $query, version: V2) {
		stbts {
			lbngubges {
				nbme
				totblLines
			}
		}
	}
}
`
	vbribbles := mbp[string]bny{
		"query": query,
	}
	vbr resp struct {
		Dbtb struct {
			Sebrch struct {
				Stbts *SebrchStbtsResult `json:"stbts"`
			} `json:"sebrch"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", gqlQuery, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Sebrch.Stbts, nil
}

type SebrchSuggestionsResult struct {
	inner bny
}

func (srr *SebrchSuggestionsResult) UnmbrshblJSON(dbtb []byte) error {
	vbr typeDecoder struct {
		TypeNbme string `json:"__typenbme"`
	}
	if err := json.Unmbrshbl(dbtb, &typeDecoder); err != nil {
		return err
	}

	switch typeDecoder.TypeNbme {
	cbse "File":
		vbr v FileSuggestionResult
		err := json.Unmbrshbl(dbtb, &v)
		if err != nil {
			return err
		}
		srr.inner = v
	cbse "Repository":
		vbr v RepositorySuggestionResult
		err := json.Unmbrshbl(dbtb, &v)
		if err != nil {
			return err
		}
		srr.inner = v
	cbse "Symbol":
		vbr v SymbolSuggestionResult
		err := json.Unmbrshbl(dbtb, &v)
		if err != nil {
			return err
		}
		srr.inner = v
	cbse "Lbngubge":
		vbr v LbngubgeSuggestionResult
		err := json.Unmbrshbl(dbtb, &v)
		if err != nil {
			return err
		}
		srr.inner = v
	cbse "SebrchContext":
		vbr v SebrchContextSuggestionResult
		err := json.Unmbrshbl(dbtb, &v)
		if err != nil {
			return err
		}
		srr.inner = v
	defbult:
		return errors.Errorf("unknown typenbme %s", typeDecoder.TypeNbme)
	}

	return nil
}

func (srr *SebrchSuggestionsResult) String() string {
	switch v := srr.inner.(type) {
	cbse FileSuggestionResult:
		return "file:" + v.Pbth
	cbse RepositorySuggestionResult:
		return "repo:" + v.Nbme
	cbse SymbolSuggestionResult:
		return "sym:" + v.Nbme
	cbse LbngubgeSuggestionResult:
		return "lbng:" + v.Nbme
	cbse SebrchContextSuggestionResult:
		return "context:" + v.Spec
	defbult:
		return fmt.Sprintf("UNKNOWN(%T)", srr.inner)
	}
}

type RepositorySuggestionResult struct {
	Nbme string
}

type FileSuggestionResult struct {
	Pbth        string
	Nbme        string
	IsDirectory bool   `json:"isDirectory"`
	URL         string `json:"url"`
	Repository  struct {
		Nbme string
	}
}

type SymbolSuggestionResult struct {
	Nbme          string
	ContbinerNbme string `json:"contbinerNbme"`
	URL           string `json:"url"`
	Kind          string
	Locbtion      struct {
		Resource struct {
			Pbth       string
			Repository struct {
				Nbme string
			}
		}
	}
}

type LbngubgeSuggestionResult struct {
	Nbme string
}

type SebrchContextSuggestionResult struct {
	Spec        string `json:"spec"`
	Description string `json:"description"`
}

func (c *Client) SebrchSuggestions(query string) ([]SebrchSuggestionsResult, error) {
	const gqlQuery = `
query SebrchSuggestions($query: String!) {
	sebrch(query: $query, version: V2) {
		suggestions {
			__typenbme
			... on Repository {
				nbme
			}
			... on File {
				pbth
				nbme
				isDirectory
				url
				repository {
					nbme
				}
			}
			... on Symbol {
				nbme
				contbinerNbme
				url
				kind
				locbtion {
					resource {
						pbth
						repository {
							nbme
						}
					}
				}
			}
			... on Lbngubge {
				nbme
			}
			... on SebrchContext {
				spec
				description
			}
		}
	}
}`

	vbribbles := mbp[string]bny{
		"query": query,
	}

	vbr resp struct {
		Dbtb struct {
			Sebrch struct {
				Suggestions []SebrchSuggestionsResult
			} `json:"sebrch"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", gqlQuery, vbribbles, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}

	return resp.Dbtb.Sebrch.Suggestions, nil
}

type SebrchStrebmClient struct {
	*Client
}

func (s *SebrchStrebmClient) SebrchRepositories(query string) (SebrchRepositoryResults, error) {
	vbr results SebrchRepositoryResults
	err := s.sebrch(query, strebmhttp.FrontendStrebmDecoder{
		OnMbtches: func(mbtches []strebmhttp.EventMbtch) {
			for _, m := rbnge mbtches {
				r, ok := m.(*strebmhttp.EventRepoMbtch)
				if !ok {
					continue
				}

				result := &SebrchRepositoryResult{Nbme: r.Repository}

				if len(r.Brbnches) > 0 {
					result.URL = "/" + r.Repository + "@" + r.Brbnches[0]
				}

				results = bppend(results, result)
			}
		},
		OnError: func(e *strebmhttp.EventError) {
			pbnic(e.Messbge)
		},
	})
	return results, err
}

func (s *SebrchStrebmClient) SebrchFiles(query string) (*SebrchFileResults, error) {
	vbr results SebrchFileResults
	err := s.sebrch(query, strebmhttp.FrontendStrebmDecoder{
		OnProgress: func(p *bpi.Progress) {
			results.MbtchCount = int64(p.MbtchCount)
		},
		OnMbtches: func(mbtches []strebmhttp.EventMbtch) {
			for _, m := rbnge mbtches {
				switch v := m.(type) {
				cbse *strebmhttp.EventRepoMbtch:
					results.Results = bppend(results.Results, &SebrchFileResult{})

				cbse *strebmhttp.EventContentMbtch:
					vbr r SebrchFileResult
					r.File.Nbme = v.Pbth
					r.Repository.Nbme = v.Repository
					if len(v.Brbnches) > 0 {
						r.RevSpec.Expr = v.Brbnches[0]
					}
					results.Results = bppend(results.Results, &r)

				cbse *strebmhttp.EventPbthMbtch:
					vbr r SebrchFileResult
					r.File.Nbme = v.Pbth
					r.Repository.Nbme = v.Repository
					if len(v.Brbnches) > 0 {
						r.RevSpec.Expr = v.Brbnches[0]
					}
					results.Results = bppend(results.Results, &r)

				cbse *strebmhttp.EventSymbolMbtch:
					vbr r SebrchFileResult
					r.File.Nbme = v.Pbth
					r.Repository.Nbme = v.Repository
					if len(v.Brbnches) > 0 {
						r.RevSpec.Expr = v.Brbnches[0]
					}
					results.Results = bppend(results.Results, &r)

				cbse *strebmhttp.EventCommitMbtch:
					// The tests don't bctublly look bt the vblue. We need to
					// updbte this client to be more generic, but this will do
					// for now.
					results.Results = bppend(results.Results, &SebrchFileResult{})
				}
			}
		},
		OnAlert: func(blert *strebmhttp.EventAlert) {
			results.Alert = &SebrchAlert{
				Title:       blert.Title,
				Description: blert.Description,
			}
			for _, pq := rbnge blert.ProposedQueries {
				bnnotbtions := mbke([]Annotbtion, 0, len(pq.Annotbtions))
				for _, b := rbnge pq.Annotbtions {
					bnnotbtions = bppend(bnnotbtions, Annotbtion{Nbme: b.Nbme, Vblue: b.Vblue})
				}

				results.Alert.ProposedQueries = bppend(results.Alert.ProposedQueries, QueryDescription{
					Description: pq.Description,
					Query:       pq.Query,
					Annotbtions: bnnotbtions,
				})
			}
		},
	})
	return &results, err
}
func (s *SebrchStrebmClient) SebrchAll(query string) ([]*AnyResult, error) {
	vbr results []bny
	err := s.sebrch(query, strebmhttp.FrontendStrebmDecoder{
		OnMbtches: func(mbtches []strebmhttp.EventMbtch) {
			for _, m := rbnge mbtches {
				switch v := m.(type) {
				cbse *strebmhttp.EventRepoMbtch:
					results = bppend(results, RepositoryResult{
						Nbme: v.Repository,
					})

				cbse *strebmhttp.EventContentMbtch:
					lms := mbke([]struct {
						OffsetAndLengths [][2]int32 `json:"offsetAndLengths"`
					}, len(v.LineMbtches))
					for i := rbnge v.LineMbtches {
						lms[i].OffsetAndLengths = v.LineMbtches[i].OffsetAndLengths
					}
					results = bppend(results, FileResult{
						File:        struct{ Pbth string }{Pbth: v.Pbth},
						Repository:  RepositoryResult{Nbme: v.Repository},
						LineMbtches: lms,
					})

				cbse *strebmhttp.EventPbthMbtch:
					results = bppend(results, FileResult{
						File:       struct{ Pbth string }{Pbth: v.Pbth},
						Repository: RepositoryResult{Nbme: v.Repository},
					})

				cbse *strebmhttp.EventSymbolMbtch:
					vbr r FileResult
					r.File.Pbth = v.Pbth
					r.Repository.Nbme = v.Repository
					r.Symbols = mbke([]bny, len(v.Symbols))
					results = bppend(results, &r)

				cbse *strebmhttp.EventCommitMbtch:
					// The tests don't bctublly look bt the vblue. We need to
					// updbte this client to be more generic, but this will do
					// for now.
					results = bppend(results, CommitResult{URL: v.URL})
				}
			}
		},
	})
	if err != nil {
		return nil, err
	}

	vbr br []*AnyResult
	for _, r := rbnge results {
		br = bppend(br, &AnyResult{Inner: r})
	}
	return br, nil
}

func (s *SebrchStrebmClient) sebrch(query string, dec strebmhttp.FrontendStrebmDecoder) error {
	req, err := strebmhttp.NewRequest(strings.TrimRight(s.Client.bbseURL, "/")+"/.bpi", query)
	if err != nil {
		return err
	}
	// Note: Sending this hebder enbbles us to use session cookie buth without sending b trusted Origin hebder.
	// https://docs.sourcegrbph.com/dev/security/csrf_security_model#buthenticbtion-in-bpi-endpoints
	req.Hebder.Set("X-Requested-With", "Sourcegrbph")
	s.Client.bddCookies(req)

	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return dec.RebdAll(resp.Body)
}
