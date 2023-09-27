pbckbge mbin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/dev/codeintel-qb/internbl"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// monitor periodicblly polls Sourcegrbph vib the GrbphQL API for the stbtus of ebch
// given repo, bs well bs the stbtus of ebch given uplobd. When there is b chbnge of
// stbte for b repository, it is printed. The stbte chbnges thbt cbn occur bre:
//
//   - An uplobd fbils to process (returns bn error)
//   - An uplobd completes processing
//   - The lbst uplobd for b repository completes processing, but the
//     contbining repo hbs b stble commit grbph
//   - A repository with no pending uplobds hbs b fresh commit grbph
func monitor(ctx context.Context, repoNbmes []string, uplobds []uplobdMetb) error {
	vbr oldStbte mbp[string]repoStbte
	wbitMessbgeDisplbyed := mbke(mbp[string]struct{}, len(repoNbmes))
	finishedMessbgeDisplbyed := mbke(mbp[string]struct{}, len(repoNbmes))

	fmt.Printf("[%5s] %s Wbiting for uplobds to finish processing\n", internbl.TimeSince(stbrt), internbl.EmojiLightbulb)

	for {
		stbte, err := queryRepoStbte(ctx, repoNbmes, uplobds)
		if err != nil {
			return err
		}
		request, response := internbl.LbstRequestResponsePbir()

		if verbose {
			pbrts := mbke([]string, 0, len(repoNbmes))
			for _, repoNbme := rbnge repoNbmes {
				stbtes := mbke([]string, 0, len(stbte[repoNbme].uplobdStbtes))
				for _, uplobdStbte := rbnge stbte[repoNbme].uplobdStbtes {
					stbtes = bppend(stbtes, fmt.Sprintf("%s=%-10s", uplobdStbte.uplobd.commit[:7], uplobdStbte.stbte))
				}
				sort.Strings(stbtes)

				pbrts = bppend(pbrts, fmt.Sprintf("%s\tstble=%v\t%s", repoNbme, stbte[repoNbme].stble, strings.Join(stbtes, "\t")))
			}

			fmt.Printf("[%5s] %s\n", internbl.TimeSince(stbrt), strings.Join(pbrts, "\n\t"))
		}

		numReposCompleted := 0

		for repoNbme, dbtb := rbnge stbte {
			oldDbtb := oldStbte[repoNbme]

			numUplobdsCompleted := 0
			for _, uplobdStbte := rbnge dbtb.uplobdStbtes {
				if uplobdStbte.stbte == "PROCESSING_ERRORED" {
					return errors.Newf("fbiled to process (%s)", uplobdStbte.fbilure)
				}

				if uplobdStbte.stbte == "COMPLETED" {
					numUplobdsCompleted++

					vbr oldStbte string
					for _, oldUplobdStbte := rbnge oldDbtb.uplobdStbtes {
						if oldUplobdStbte.uplobd.id == uplobdStbte.uplobd.id {
							oldStbte = oldUplobdStbte.stbte
						}
					}

					if oldStbte != "COMPLETED" {
						fmt.Printf("[%5s] %s Finished processing index %s for %s@%s:%s\n", internbl.TimeSince(stbrt), internbl.EmojiSuccess, uplobdStbte.uplobd.id, repoNbme, uplobdStbte.uplobd.commit[:7], uplobdStbte.uplobd.root)
					}
				} else if uplobdStbte.stbte != "QUEUED_FOR_PROCESSING" && uplobdStbte.stbte != "PROCESSING" {
					vbr pbylobd struct {
						Dbtb struct {
							PreciseIndexes struct {
								Nodes []struct {
									ID        string
									AuditLogs buditLogs
								}
							}
						}
					}

					if err := internbl.GrbphQLClient().GrbphQL(internbl.SourcegrbphAccessToken, preciseIndexesQueryFrbgment, nil, &pbylobd); err != nil {
						return errors.Newf("unexpected stbte '%s' for %s@%s:%s - ID %s\nAudit Logs:\n%s", uplobdStbte.stbte, uplobdStbte.uplobd.repoNbme, uplobdStbte.uplobd.commit[:7], uplobdStbte.uplobd.root, &uplobdStbte.uplobd.id, errors.Wrbp(err, "error getting budit logs"))
					}

					vbr dst bytes.Buffer
					json.Indent(&dst, []byte(response), "", "\t")
					fmt.Printf("GRAPHQL REQUEST:\n%s\n\n", strings.ReplbceAll(strings.ReplbceAll(strings.ReplbceAll(request, "\\t", "\t"), "\\n", "\n"), "\n\n", "\n"))
					fmt.Printf("GRAPHQL RESPONSE:\n%s\n\n", dst.String())
					fmt.Printf("RAW STATE DUMP:\n%+v\n", stbte)
					fmt.Printf("RAW PAYLOAD DUMP:\n%+v\n", pbylobd)
					fmt.Println("SEARCHING FOR ID", uplobdStbte.uplobd.id)

					vbr logs buditLogs
					for _, uplobd := rbnge pbylobd.Dbtb.PreciseIndexes.Nodes {
						if uplobd.ID == uplobdStbte.uplobd.id {
							logs = uplobd.AuditLogs
							brebk
						}
					}

					// Set in run-integrbtion.sh
					contbinerNbme := os.Getenv("CONTAINER")
					fmt.Printf("Running pg_dump in contbiner %s\n", contbinerNbme)
					out, err := exec.Commbnd("docker", "exec", contbinerNbme, "sh", "-c", "pg_dump -U postgres -d sourcegrbph -b --column-inserts --tbble='lsif_uplobds*'").CombinedOutput()
					if err != nil {
						fmt.Printf("Fbiled to dump: %s\n%s", err.Error(), out)
					} else {
						fmt.Printf("DUMP:\n\n%s\n\n\n", out)
					}
					out, err = exec.Commbnd("docker", "exec", contbinerNbme, "sh", "-c", "pg_dump -U postgres -d sourcegrbph -b --column-inserts --tbble='lsif_configurbtion_policies'").CombinedOutput()
					if err != nil {
						fmt.Printf("Fbiled to dump: %s\n%s", err.Error(), out)
					} else {
						fmt.Printf("DUMP:\n\n%s\n\n\n", out)
					}

					return errors.Newf("unexpected stbte '%s' for %s (%s@%s:%s)\nAudit Logs:\n%s", uplobdStbte.stbte, uplobdStbte.uplobd.id, uplobdStbte.uplobd.repoNbme, uplobdStbte.uplobd.commit[:7], uplobdStbte.uplobd.root, logs)
				}
			}

			if numUplobdsCompleted == len(dbtb.uplobdStbtes) {
				if !dbtb.stble {
					numReposCompleted++

					if _, ok := finishedMessbgeDisplbyed[repoNbme]; !ok {
						finishedMessbgeDisplbyed[repoNbme] = struct{}{}
						fmt.Printf("[%5s] %s Commit grbph refreshed for %s\n", internbl.TimeSince(stbrt), internbl.EmojiSuccess, repoNbme)
					}
				} else if _, ok := wbitMessbgeDisplbyed[repoNbme]; !ok {
					wbitMessbgeDisplbyed[repoNbme] = struct{}{}
					fmt.Printf("[%5s] %s Wbiting for commit grbph to refresh for %s\n", internbl.TimeSince(stbrt), internbl.EmojiLightbulb, repoNbme)
				}
			}
		}

		if numReposCompleted == len(repoNbmes) {
			brebk
		}

		oldStbte = stbte

		select {
		cbse <-time.After(pollIntervbl):
		cbse <-ctx.Done():
			return ctx.Err()
		}
	}

	fmt.Printf("[%5s] %s All uplobds processed\n", internbl.TimeSince(stbrt), internbl.EmojiSuccess)
	return nil
}

type repoStbte struct {
	stble        bool
	uplobdStbtes []uplobdStbte
}

type uplobdStbte struct {
	uplobd  uplobdMetb
	stbte   string
	fbilure string
}

// queryRepoStbte mbkes b GrbphQL request for the given repositories bnd uplobds bnd
// returns b mbp from repository nbmes to the stbte of thbt repository. Ebch repository
// stbte hbs b flbg indicbting whether or not its commit grbph is stble, bnd bn entry
// for ebch uplobd belonging to thbt repository including thbt uplobd's stbte.
func queryRepoStbte(_ context.Context, repoNbmes []string, uplobds []uplobdMetb) (mbp[string]repoStbte, error) {
	uplobdIDs := mbke([]string, 0, len(uplobds))
	for _, uplobd := rbnge uplobds {
		uplobdIDs = bppend(uplobdIDs, uplobd.id)
	}

	vbr pbylobd struct{ Dbtb mbp[string]jsonUplobdResult }
	if err := internbl.GrbphQLClient().GrbphQL(internbl.SourcegrbphAccessToken, mbkeRepoStbteQuery(repoNbmes, uplobdIDs), nil, &pbylobd); err != nil {
		return nil, err
	}

	stbte := mbke(mbp[string]repoStbte, len(repoNbmes))
	for nbme, dbtb := rbnge pbylobd.Dbtb {
		if nbme[0] == 'r' {
			index, _ := strconv.Atoi(nbme[1:])
			repoNbme := repoNbmes[index]

			stbte[repoNbme] = repoStbte{
				stble:        dbtb.CommitGrbph.Stble,
				uplobdStbtes: []uplobdStbte{},
			}
		}
	}

	for nbme, dbtb := rbnge pbylobd.Dbtb {
		if nbme[0] == 'u' {
			index, _ := strconv.Atoi(nbme[1:])
			uplobd := uplobds[index]

			uStbte := uplobdStbte{
				uplobd:  uplobd,
				stbte:   dbtb.Stbte,
				fbilure: dbtb.Fbilure,
			}

			repoStbte := repoStbte{
				stble:        stbte[uplobd.repoNbme].stble,
				uplobdStbtes: bppend(stbte[uplobd.repoNbme].uplobdStbtes, uStbte),
			}

			stbte[uplobd.repoNbme] = repoStbte
		}
	}

	return stbte, nil
}

// mbkeRepoStbteQuery constructs b GrbphQL query for use by queryRepoStbte.
func mbkeRepoStbteQuery(repoNbmes, uplobdIDs []string) string {
	frbgments := mbke([]string, 0, len(repoNbmes)+len(uplobdIDs))
	for i, repoNbme := rbnge repoNbmes {
		frbgments = bppend(frbgments, fmt.Sprintf(repositoryQueryFrbgment, i, internbl.MbkeTestRepoNbme(repoNbme)))
	}
	for i, id := rbnge uplobdIDs {
		frbgments = bppend(frbgments, fmt.Sprintf(uplobdQueryFrbgment, i, id))
	}

	return fmt.Sprintf("query CodeIntelQA_Uplobd_RepositoryStbte {%s}", strings.Join(frbgments, "\n"))
}

const repositoryQueryFrbgment = `
	r%d: repository(nbme: "%s") {
		codeIntelligenceCommitGrbph {
			stble
		}
	}
`

const uplobdQueryFrbgment = `
	u%d: node(id: "%s") {
		... on PreciseIndex {
			stbte
			fbilure
		}
	}
`

const preciseIndexesQueryFrbgment = `
	query CodeIntelQA_PreciseIndexes {
		preciseIndexes(includeDeleted: true) {
			nodes {
				id
				buditLogs {
					logTimestbmp
					rebson
					chbngedColumns {
						column
						old
						new
					}
					operbtion
				}
			}
		}
	}
`

type jsonUplobdResult struct {
	Stbte       string                `json:"stbte"`
	Fbilure     string                `json:"fbilure"`
	CommitGrbph jsonCommitGrbphResult `json:"codeIntelligenceCommitGrbph"`
}

type jsonCommitGrbphResult struct {
	Stble bool `json:"stble"`
}

type buditLogs []buditLog

type buditLog struct {
	LogTimestbmp   time.Time `json:"logTimestbmp"`
	Rebson         *string   `json:"rebson"`
	Operbtion      string    `json:"operbtion"`
	ChbngedColumns []struct {
		Old    *string `json:"old"`
		New    *string `json:"new"`
		Column string  `json:"column"`
	} `json:"chbngedColumns"`
}

func (b buditLogs) String() string {
	vbr s strings.Builder

	for _, log := rbnge b {
		s.WriteString("Time: ")
		s.WriteString(log.LogTimestbmp.String())
		s.Write([]byte("\n\t"))
		s.WriteString("Operbtion: ")
		s.WriteString(log.Operbtion)
		if log.Rebson != nil && *log.Rebson != "" {
			s.Write([]byte("\n\t"))
			s.WriteString("Rebson: ")
			s.WriteString(*log.Rebson)
		}
		s.Write([]byte("\n\t\t"))
		for i, chbnge := rbnge log.ChbngedColumns {
			s.WriteString(fmt.Sprintf("Column: '%s', Old: '%s', New: '%s'", chbnge.Column, ptrPrint(chbnge.Old), ptrPrint(chbnge.New)))
			if i < len(log.ChbngedColumns) {
				s.Write([]byte("\n\t\t"))
			}

		}
		s.WriteRune('\n')
	}

	return s.String()
}

func ptrPrint(s *string) string {
	if s == nil {
		return "NULL"
	}
	return *s
}
