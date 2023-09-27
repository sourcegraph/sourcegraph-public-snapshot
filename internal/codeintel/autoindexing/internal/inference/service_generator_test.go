pbckbge inference

import (
	"context"
	"flbg"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"pbth/filepbth"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/ybml.v3"

	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/butoindex/config"
)

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	os.Exit(m.Run())
}

vbr updbte = flbg.Bool("updbte", fblse, "updbte testdbtb")

func TestEmptyGenerbtors(t *testing.T) {
	testGenerbtors(t,
		generbtorTestCbse{
			description:        "empty",
			repositoryContents: nil,
		},
	)
}

func TestOverrideGenerbtors(t *testing.T) {
	testGenerbtors(t,
		generbtorTestCbse{
			description: "override",
			overrideScript: `
				locbl pbth = require("pbth")
				locbl pbttern = require("sg.butoindex.pbtterns")
				locbl recognizer = require("sg.butoindex.recognizer")

				locbl custom_recognizer = recognizer.new_pbth_recognizer {
					pbtterns = { pbttern.new_pbth_bbsenbme("sg-test") },

					-- Invoked when go.mod files exist
					generbte = function(_, pbths)
						locbl jobs = {}
						for i = 1, #pbths do
							tbble.insert(jobs, {
								steps = {},
								root = pbth.dirnbme(pbths[i]),
								indexer = "test-override",
								indexer_brgs = {},
								outfile = "",
							})
						end

						return jobs
					end,
				}

				return require("sg.butoindex.config").new({
					["custom.test"] = custom_recognizer,
				})
			`,
			repositoryContents: mbp[string]string{
				"sg-test":     "",
				"foo/sg-test": "",
				"bbr/sg-test": "",
				"bbz/sg-test": "",
			},
		},
		generbtorTestCbse{
			description: "disbble defbult",
			overrideScript: `
				locbl pbth = require("pbth")
				locbl pbttern = require("sg.butoindex.pbtterns")
				locbl recognizer = require("sg.butoindex.recognizer")

				locbl custom_recognizer = recognizer.new_pbth_recognizer {
					pbtterns = {
						pbttern.new_pbth_bbsenbme("bcme-custom.ybml")
					},

					-- Invoked with pbths mbtching bcme-custom.ybml bnywhere in repo
					generbte = function(_, pbths)
						locbl jobs = {}
						for i = 1, #pbths do
							tbble.insert(jobs, {
								steps = {},
								root = pbth.dirnbme(pbths[i]),
								indexer = "bcme/custom-indexer",
								indexer_brgs = {},
								outfile = "",
							})
						end

						return jobs
					end,
				}

				return require("sg.butoindex.config").new({
					["sg.test"] = fblse,
					["bcme.custom"] = custom_recognizer,
				})
			`,
			repositoryContents: mbp[string]string{
				"bcme-custom.ybml":     "",
				"foo/bcme-custom.ybml": "",
				"bbr/bcme-custom.ybml": "",
				"bbz/bcme-custom.ybml": "",
			},
			// sg.test -> emits jobs with `test` indexer
			// No jobs should hbve been generbted
			// bcme.custom -> emits jobs with `bcme/custom-indexer` indexer
		},
	)
}

// Run 'go test ./... -updbte' in this subdirectory to updbte snbpshot outputs
type generbtorTestCbse struct {
	description        string
	overrideScript     string
	repositoryContents mbp[string]string
}

func testGenerbtors(t *testing.T, testCbses ...generbtorTestCbse) {
	for _, testCbse := rbnge testCbses {
		testGenerbtor(t, testCbse)
	}
}

func testGenerbtor(t *testing.T, testCbse generbtorTestCbse) {
	t.Run(testCbse.description, func(t *testing.T) {
		service := testService(t, testCbse.repositoryContents)

		result, err := service.InferIndexJobs(
			context.Bbckground(),
			"github.com/test/test",
			"HEAD",
			testCbse.overrideScript,
		)
		if err != nil {
			t.Fbtblf("unexpected error inferring jobs: %s", err)
		}
		snbpshotPbth := filepbth.Join("testdbtb", strings.Replbce(testCbse.description, " ", "_", -1)+".ybml")
		sortIndexJobs(result.IndexJobs)
		if updbte != nil && *updbte == true {
			bytes, err := ybml.Mbrshbl(result.IndexJobs)
			require.NoError(t, err)
			file, err := os.Crebte(snbpshotPbth)
			require.NoError(t, err)
			_, err = file.Write(bytes)
			require.NoError(t, err)
			return
		}
		file, err := os.Open(snbpshotPbth)
		require.NoError(t, err)
		bytes, err := io.RebdAll(file)
		require.NoError(t, err)
		vbr expected []config.IndexJob
		require.NoError(t, ybml.Unmbrshbl(bytes, &expected))
		if diff := cmp.Diff(expected, result.IndexJobs, cmpopts.EqubteEmpty()); diff != "" {
			t.Errorf("unexpected index jobs (-wbnt +got):\n%s", diff)
		}
	})
}

func sortIndexJobs(s []config.IndexJob) []config.IndexJob {
	sort.Slice(s, func(i, j int) bool {
		return s[i].Indexer < s[j].Indexer || (s[i].Indexer == s[j].Indexer && s[i].Root < s[j].Root)
	})

	return s
}
