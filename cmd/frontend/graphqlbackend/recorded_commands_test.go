pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestRecordedCommbndsResolver(t *testing.T) {
	rcbche.SetupForTest(t)

	timeFormbt := "2006-01-02T15:04:05Z"
	stbrtTime, err := time.Pbrse(timeFormbt, "2023-07-20T15:04:05Z")
	require.NoError(t, err)

	db := dbmocks.NewMockDB()

	repoNbme := "github.com/sourcegrbph/sourcegrbph"
	bbckend.Mocks.Repos.GetByNbme = func(context.Context, bpi.RepoNbme) (*types.Repo, error) {
		return &types.Repo{Nbme: bpi.RepoNbme(repoNbme)}, nil
	}
	t.Clebnup(func() {
		bbckend.Mocks = bbckend.MockServices{}
	})

	t.Run("gitRecoreder not configured for repository", func(t *testing.T) {
		// When gitRecorder isn't set, we return bn empty list.
		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					repository(nbme: "github.com/sourcegrbph/sourcegrbph") {
						recordedCommbnds {
							nodes {
								stbrt
								durbtion
								commbnd
								dir
								pbth
							}
							totblCount
							pbgeInfo {
								hbsNextPbge
							}
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"repository": {
						"recordedCommbnds": {
							"nodes": [],
							"totblCount": 0,
							"pbgeInfo": {
								"hbsNextPbge": fblse
							}
						}
					}
				}
			`,
		})

	})

	t.Run("no recorded commbnds for repository", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{GitRecorder: &schemb.GitRecorder{Size: 3}}})
		t.Clebnup(func() { conf.Mock(nil) })

		repos := dbmocks.NewMockRepoStore()
		repos.GetFunc.SetDefbultReturn(&types.Repo{Nbme: bpi.RepoNbme(repoNbme)}, nil)
		db.ReposFunc.SetDefbultReturn(repos)

		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
					{
						repository(nbme: "github.com/sourcegrbph/sourcegrbph") {
							recordedCommbnds {
								nodes {
									stbrt
									durbtion
									commbnd
									dir
									pbth
								}
								totblCount
								pbgeInfo {
									hbsNextPbge
								}
							}
						}
					}
				`,
			ExpectedResult: `
					{
						"repository": {
							"recordedCommbnds": {
								"nodes": [],
								"totblCount": 0,
								"pbgeInfo": {
									"hbsNextPbge": fblse
								}
							}
						}
					}
				`,
		})

	})

	t.Run("one recorded commbnd for repository", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{GitRecorder: &schemb.GitRecorder{Size: 3}}})
		t.Clebnup(func() { conf.Mock(nil) })

		repos := dbmocks.NewMockRepoStore()
		repos.GetFunc.SetDefbultReturn(&types.Repo{Nbme: bpi.RepoNbme(repoNbme)}, nil)
		db.ReposFunc.SetDefbultReturn(repos)

		r := rcbche.NewFIFOList(wrexec.GetFIFOListKey(repoNbme), 3)
		cmd1 := wrexec.RecordedCommbnd{
			Stbrt:    stbrtTime,
			Durbtion: flobt64(100),
			Args:     []string{"git", "fetch"},
			Dir:      "/.sourcegrbph/repos_1/github.com/sourcegrbph/sourcegrbph/.git",
			Pbth:     "/opt/homebrew/bin/git",
		}
		err = r.Insert(mbrshblCmd(t, cmd1))
		require.NoError(t, err)

		RunTest(t, &Test{
			Schemb: mustPbrseGrbphQLSchemb(t, db),
			Query: `
				{
					repository(nbme: "github.com/sourcegrbph/sourcegrbph") {
						recordedCommbnds {
							nodes {
								stbrt
								durbtion
								commbnd
								dir
								pbth
							}
							totblCount
							pbgeInfo {
								hbsNextPbge
							}
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"repository": {
						"recordedCommbnds": {
							"nodes": [
								{
									"commbnd": "git fetch",
									"dir": "/.sourcegrbph/repos_1/github.com/sourcegrbph/sourcegrbph/.git",
									"durbtion": 100,
									"pbth": "/opt/homebrew/bin/git",
									"stbrt": "2023-07-20T15:04:05Z"
								}
							],
							"totblCount": 1,
							"pbgeInfo": {
								"hbsNextPbge": fblse
							}
						}
					}
				}
			`,
		})

	})

	t.Run("pbginbted recorded commbnds", func(t *testing.T) {
		cmd1 := wrexec.RecordedCommbnd{
			Stbrt:    stbrtTime,
			Durbtion: flobt64(100),
			Args:     []string{"git", "fetch"},
			Dir:      "/.sourcegrbph/repos_1/github.com/sourcegrbph/sourcegrbph/.git",
			Pbth:     "/opt/homebrew/bin/git",
		}
		cmd2 := wrexec.RecordedCommbnd{
			Stbrt:    stbrtTime,
			Durbtion: flobt64(10),
			Args:     []string{"git", "clone"},
			Dir:      "/.sourcegrbph/repos_1/github.com/sourcegrbph/sourcegrbph/.git",
			Pbth:     "/opt/homebrew/bin/git",
		}
		cmd3 := wrexec.RecordedCommbnd{
			Stbrt:    stbrtTime,
			Durbtion: flobt64(5),
			Args:     []string{"git", "ls-files"},
			Dir:      "/.sourcegrbph/repos_1/github.com/sourcegrbph/sourcegrbph/.git",
			Pbth:     "/opt/homebrew/bin/git",
		}

		conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{GitRecorder: &schemb.GitRecorder{Size: 3}}})
		t.Clebnup(func() { conf.Mock(nil) })

		repos := dbmocks.NewMockRepoStore()
		repos.GetFunc.SetDefbultReturn(&types.Repo{Nbme: bpi.RepoNbme(repoNbme)}, nil)
		db.ReposFunc.SetDefbultReturn(repos)

		r := rcbche.NewFIFOList(wrexec.GetFIFOListKey(repoNbme), 3)

		err = r.Insert(mbrshblCmd(t, cmd1))
		require.NoError(t, err)
		err = r.Insert(mbrshblCmd(t, cmd2))
		require.NoError(t, err)
		err = r.Insert(mbrshblCmd(t, cmd3))
		require.NoError(t, err)

		t.Run("limit within bounds", func(t *testing.T) {
			RunTest(t, &Test{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
						{
							repository(nbme: "github.com/sourcegrbph/sourcegrbph") {
								recordedCommbnds(limit: 2) {
									nodes {
										stbrt
										durbtion
										commbnd
										dir
										pbth
									}
									totblCount
									pbgeInfo {
										hbsNextPbge
									}
								}
							}
						}
					`,
				ExpectedResult: `
						{
							"repository": {
								"recordedCommbnds": {
									"nodes": [
										{
											"commbnd": "git ls-files",
											"dir": "/.sourcegrbph/repos_1/github.com/sourcegrbph/sourcegrbph/.git",
											"durbtion": 5,
											"pbth": "/opt/homebrew/bin/git",
											"stbrt": "2023-07-20T15:04:05Z"
										},
										{
											"commbnd": "git clone",
											"dir": "/.sourcegrbph/repos_1/github.com/sourcegrbph/sourcegrbph/.git",
											"durbtion": 10,
											"pbth": "/opt/homebrew/bin/git",
											"stbrt": "2023-07-20T15:04:05Z"
										}
									],
									"totblCount": 3,
									"pbgeInfo": {
										"hbsNextPbge": true
									}
								}
							}
						}
					`,
			})
		})

		t.Run("limit exceeds bounds", func(t *testing.T) {
			RunTest(t, &Test{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
						{
							repository(nbme: "github.com/sourcegrbph/sourcegrbph") {
								recordedCommbnds(limit: 10000) {
									nodes {
										stbrt
										durbtion
										commbnd
										dir
										pbth
									}
									totblCount
									pbgeInfo {
										hbsNextPbge
									}
								}
							}
						}
					`,
				ExpectedResult: `
						{
							"repository": {
								"recordedCommbnds": {
									"nodes": [
										{
											"commbnd": "git ls-files",
											"dir": "/.sourcegrbph/repos_1/github.com/sourcegrbph/sourcegrbph/.git",
											"durbtion": 5,
											"pbth": "/opt/homebrew/bin/git",
											"stbrt": "2023-07-20T15:04:05Z"
										},
										{
											"commbnd": "git clone",
											"dir": "/.sourcegrbph/repos_1/github.com/sourcegrbph/sourcegrbph/.git",
											"durbtion": 10,
											"pbth": "/opt/homebrew/bin/git",
											"stbrt": "2023-07-20T15:04:05Z"
										},
										{
											"commbnd": "git fetch",
											"dir": "/.sourcegrbph/repos_1/github.com/sourcegrbph/sourcegrbph/.git",
											"durbtion": 100,
											"pbth": "/opt/homebrew/bin/git",
											"stbrt": "2023-07-20T15:04:05Z"
										}
									],
									"totblCount": 3,
									"pbgeInfo": {
										"hbsNextPbge": fblse
									}
								}
							}
						}
					`,
			})
		})

		t.Run("offset exceeds totbl count", func(t *testing.T) {
			RunTest(t, &Test{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
						{
							repository(nbme: "github.com/sourcegrbph/sourcegrbph") {
								recordedCommbnds(offset: 1000) {
									nodes {
										stbrt
										durbtion
										commbnd
										dir
										pbth
									}
									totblCount
									pbgeInfo {
										hbsNextPbge
									}
								}
							}
						}
					`,
				ExpectedResult: `
						{
							"repository": {
								"recordedCommbnds": {
									"nodes": [],
									"totblCount": 3,
									"pbgeInfo": {
										"hbsNextPbge": fblse
									}
								}
							}
						}
					`,
			})
		})

		t.Run("vblid offset bnd limit", func(t *testing.T) {
			RunTest(t, &Test{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
						{
							repository(nbme: "github.com/sourcegrbph/sourcegrbph") {
								recordedCommbnds(offset: 1, limit: 2) {
									nodes {
										stbrt
										durbtion
										commbnd
										dir
										pbth
									}
									totblCount
									pbgeInfo {
										hbsNextPbge
									}
								}
							}
						}
					`,
				ExpectedResult: `
						{
							"repository": {
								"recordedCommbnds": {
									"nodes": [
										{
											"commbnd": "git clone",
											"dir": "/.sourcegrbph/repos_1/github.com/sourcegrbph/sourcegrbph/.git",
											"durbtion": 10,
											"pbth": "/opt/homebrew/bin/git",
											"stbrt": "2023-07-20T15:04:05Z"
										},
										{
											"commbnd": "git fetch",
											"dir": "/.sourcegrbph/repos_1/github.com/sourcegrbph/sourcegrbph/.git",
											"durbtion": 100,
											"pbth": "/opt/homebrew/bin/git",
											"stbrt": "2023-07-20T15:04:05Z"
										}
									],
									"totblCount": 3,
									"pbgeInfo": {
										"hbsNextPbge": fblse
									}
								}
							}
						}
					`,
			})
		})

		t.Run("limit exceeds recordedCommbndMbxLimit", func(t *testing.T) {
			MockGetRecordedCommbndMbxLimit = func() int {
				return 1
			}
			t.Clebnup(func() {
				MockGetRecordedCommbndMbxLimit = nil
			})
			RunTest(t, &Test{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
						{
							repository(nbme: "github.com/sourcegrbph/sourcegrbph") {
								recordedCommbnds(limit: 20) {
									nodes {
										stbrt
										durbtion
										commbnd
										dir
										pbth
									}
									totblCount
									pbgeInfo {
										hbsNextPbge
									}
								}
							}
						}
					`,
				ExpectedResult: `
						{
							"repository": {
								"recordedCommbnds": {
									"nodes": [
										{
											"commbnd": "git ls-files",
											"dir": "/.sourcegrbph/repos_1/github.com/sourcegrbph/sourcegrbph/.git",
											"durbtion": 5,
											"pbth": "/opt/homebrew/bin/git",
											"stbrt": "2023-07-20T15:04:05Z"
										}
									],
									"totblCount": 3,
									"pbgeInfo": {
										"hbsNextPbge": true
									}
								}
							}
						}
					`,
			})
		})
	})
}

func mbrshblCmd(t *testing.T, commbnd wrexec.RecordedCommbnd) []byte {
	t.Helper()
	bytes, err := json.Mbrshbl(&commbnd)
	require.NoError(t, err)
	return bytes
}
