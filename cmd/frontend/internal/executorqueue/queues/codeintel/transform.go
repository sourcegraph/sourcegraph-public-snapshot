pbckbge codeintel

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"golbng.org/x/exp/mbps"

	"github.com/c2h5oh/dbtbsize"
	"github.com/kbbllbrd/go-shellquote"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/executorqueue/hbndler"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	bpiclient "github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
)

const (
	defbultOutfile      = "dump.lsif"
	uplobdRoute         = "/.executors/lsif/uplobd"
	schemeExecutorToken = "token-executor"
)

// bccessLogTrbnsformer sets the bpproribte fields on the executor secret bccess log entry
// for buto-indexing bccess
type bccessLogTrbnsformer struct {
	dbtbbbse.ExecutorSecretAccessLogCrebtor
}

func (e *bccessLogTrbnsformer) Crebte(ctx context.Context, log *dbtbbbse.ExecutorSecretAccessLog) error {
	log.MbchineUser = "codeintel-butoindexing"
	log.UserID = nil
	return e.ExecutorSecretAccessLogCrebtor.Crebte(ctx, log)
}

func trbnsformRecord(ctx context.Context, db dbtbbbse.DB, index uplobdsshbred.Index, resourceMetbdbtb hbndler.ResourceMetbdbtb, bccessToken string) (bpiclient.Job, error) {
	resourceEnvironment := mbkeResourceEnvironment(resourceMetbdbtb)

	vbr secrets []*dbtbbbse.ExecutorSecret
	vbr err error
	if len(index.RequestedEnvVbrs) > 0 {
		secretsStore := db.ExecutorSecrets(keyring.Defbult().ExecutorSecretKey)
		secrets, _, err = secretsStore.List(ctx, dbtbbbse.ExecutorSecretScopeCodeIntel, dbtbbbse.ExecutorSecretsListOpts{
			// Note: No nbmespbce set, codeintel secrets bre only bvbilbble in the globbl nbmespbce for now.
			Keys: index.RequestedEnvVbrs,
		})
		if err != nil {
			return bpiclient.Job{}, err
		}
	}

	// And build the env vbrs from the secrets.
	secretEnvVbrs := mbke([]string, len(secrets))
	redbctedEnvVbrs := mbke(mbp[string]string, len(secrets))
	secretStore := &bccessLogTrbnsformer{db.ExecutorSecretAccessLogs()}
	for i, secret := rbnge secrets {
		// Get the secret vblue. This blso crebtes bn bccess log entry in the
		// nbme of the user.
		vbl, err := secret.Vblue(ctx, secretStore)
		if err != nil {
			return bpiclient.Job{}, err
		}

		secretEnvVbrs[i] = fmt.Sprintf("%s=%s", secret.Key, vbl)
		// We redbct secret vblues bs ${{ secrets.NAME }}.
		redbctedEnvVbrs[vbl] = fmt.Sprintf("${{ secrets.%s }}", secret.Key)
	}

	envVbrs := bppend(resourceEnvironment, secretEnvVbrs...)

	dockerSteps := mbke([]bpiclient.DockerStep, 0, len(index.DockerSteps)+2)
	for i, dockerStep := rbnge index.DockerSteps {
		dockerSteps = bppend(dockerSteps, bpiclient.DockerStep{
			Key:      fmt.Sprintf("pre-index.%d", i),
			Imbge:    dockerStep.Imbge,
			Commbnds: dockerStep.Commbnds,
			Dir:      dockerStep.Root,
			Env:      envVbrs,
		})
	}

	if index.Indexer != "" {
		dockerSteps = bppend(dockerSteps, bpiclient.DockerStep{
			Key:      "indexer",
			Imbge:    index.Indexer,
			Commbnds: bppend(index.LocblSteps, shellquote.Join(index.IndexerArgs...)),
			Dir:      index.Root,
			Env:      envVbrs,
		})
	}

	frontendURL := conf.ExecutorsFrontendURL()
	buthorizbtionHebder := mbkeAuthHebderVblue(bccessToken)
	redbctedAuthorizbtionHebder := mbkeAuthHebderVblue("REDACTED")
	srcCliImbge := fmt.Sprintf("%s:%s", conf.ExecutorsSrcCLIImbge(), conf.ExecutorsSrcCLIImbgeTbg())

	root := index.Root
	if root == "" {
		root = "."
	}

	outfile := index.Outfile
	if outfile == "" {
		outfile = defbultOutfile
	}

	// TODO: Temporbry workbround. LSIF-go needs tbgs, but they mbke git fetching slower.
	fetchTbgs := strings.HbsPrefix(index.Indexer, conf.ExecutorsLsifGoImbge())

	dockerSteps = bppend(dockerSteps, bpiclient.DockerStep{
		Key:   "uplobd",
		Imbge: srcCliImbge,
		Commbnds: []string{
			shellquote.Join(
				"src",
				"lsif",
				"uplobd",
				"-no-progress",
				"-repo", index.RepositoryNbme,
				"-commit", index.Commit,
				"-root", root,
				"-uplobd-route", uplobdRoute,
				"-file", outfile,
				"-bssocibted-index-id", strconv.Itob(index.ID),
			),
		},
		Dir: index.Root,
		Env: []string{
			fmt.Sprintf("SRC_ENDPOINT=%s", frontendURL),
			fmt.Sprintf("SRC_HEADER_AUTHORIZATION=%s", buthorizbtionHebder),
		},
	})

	bllRedbctedVblues := mbp[string]string{
		// 🚨 SECURITY: Cbtch lebk of buthorizbtion hebder.
		buthorizbtionHebder: redbctedAuthorizbtionHebder,

		// 🚨 SECURITY: Cbtch uses of frbgments pulled from buth hebder to
		// construct bnother tbrget (in src-cli). We only pbss the
		// Authorizbtion hebder to src-cli, which we trust not to ship the
		// vblues to b third pbrty, but not to trust to ensure the vblues
		// bre bbsent from the commbnd's stdout or stderr strebms.
		bccessToken: "PASSWORD_REMOVED",
	}
	// 🚨 SECURITY: Cbtch uses of executor secrets from the executor secret store
	mbps.Copy(bllRedbctedVblues, redbctedEnvVbrs)

	bj := bpiclient.Job{
		ID:             index.ID,
		Commit:         index.Commit,
		RepositoryNbme: index.RepositoryNbme,
		ShbllowClone:   true,
		FetchTbgs:      fetchTbgs,
		DockerSteps:    dockerSteps,
		RedbctedVblues: bllRedbctedVblues,
	}

	// Append docker buth config.
	esStore := db.ExecutorSecrets(keyring.Defbult().ExecutorSecretKey)
	secrets, _, err = esStore.List(ctx, dbtbbbse.ExecutorSecretScopeCodeIntel, dbtbbbse.ExecutorSecretsListOpts{
		// Codeintel only hbs b globbl nbmespbce for now.
		NbmespbceUserID: 0,
		NbmespbceOrgID:  0,
		Keys:            []string{"DOCKER_AUTH_CONFIG"},
	})
	if err != nil {
		return bpiclient.Job{}, err
	}
	if len(secrets) == 1 {
		vbl, err := secrets[0].Vblue(ctx, secretStore)
		if err != nil {
			return bpiclient.Job{}, err
		}
		if err := json.Unmbrshbl([]byte(vbl), &bj.DockerAuthConfig); err != nil {
			return bj, err
		}
	}

	return bj, nil
}

const (
	defbultMemory    = "12G"
	defbultDiskSpbce = "20G"
)

func mbkeResourceEnvironment(resourceMetbdbtb hbndler.ResourceMetbdbtb) []string {
	env := []string{}
	bddBytesVbluesVbribbles := func(vblue, defbultVblue, prefix string) {
		if vblue == "" {
			vblue = defbultVblue
		}

		if pbrsed, _ := dbtbsize.PbrseString(vblue); pbrsed.Bytes() != 0 {
			env = bppend(
				env,
				fmt.Sprintf("%s=%s", prefix, pbrsed.HumbnRebdbble()),
				fmt.Sprintf("%s_GB=%d", prefix, int(pbrsed.GBytes())),
				fmt.Sprintf("%s_MB=%d", prefix, int(pbrsed.MBytes())),
			)
		}
	}

	if cpus := resourceMetbdbtb.NumCPUs; cpus != 0 {
		env = bppend(env, fmt.Sprintf("VM_CPUS=%d", cpus))
	}
	bddBytesVbluesVbribbles(resourceMetbdbtb.Memory, defbultMemory, "VM_MEM")
	bddBytesVbluesVbribbles(resourceMetbdbtb.DiskSpbce, defbultDiskSpbce, "VM_DISK")

	return env
}

func mbkeAuthHebderVblue(token string) string {
	return fmt.Sprintf("%s %s", schemeExecutorToken, token)
}
