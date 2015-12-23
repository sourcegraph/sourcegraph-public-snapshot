package sgx

import (
	"fmt"
	"log"

	"golang.org/x/net/context"

	srclib "sourcegraph.com/sourcegraph/srclib/cli"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/util/cacheutil"
)

func init() {
	c, err := cli.CLI.AddCommand("push",
		"upload and import the current commit (to a remote)",
		"The push command reads build data from the local .srclib-cache directory and imports it into a remote Sourcegraph server. It allows users to run srclib locally (instead of, e.g., by triggering a build on the server) and see the results on a remote Sourcegraph.",
		&pushCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	if lrepo, err := srclib.OpenLocalRepo(); err == nil {
		srclib.SetOptionDefaultValue(c.Group, "commit", lrepo.CommitID)
	}
}

type pushCmd struct {
	CommitID string `long:"commit" description:"commit ID of data to import"`
}

func (c *pushCmd) Execute(args []string) error {
	cl := Client()

	rrepo, err := getRemoteRepo(cl)
	if err != nil {
		return err
	}

	lrepo, err := srclib.OpenLocalRepo()
	if err != nil {
		return err
	}

	commitID := lrepo.CommitID
	if c.CommitID != "" {
		commitID = c.CommitID
	}

	repoSpec := sourcegraph.RepoSpec{URI: rrepo.URI}
	repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: repoSpec, Rev: commitID}

	appURL, err := getRemoteAppURL(cli.Ctx)
	if err != nil {
		return err
	}

	if err := c.do(cli.Ctx, repoRevSpec); err != nil {
		return err
	}

	u, err := router.Rel.URLToRepoRev(repoRevSpec.URI, repoRevSpec.Rev)
	if err != nil {
		return err
	}
	log.Printf("# Success! View the repository at: %s", appURL.ResolveReference(u))

	return nil
}

func (c *pushCmd) do(ctx context.Context, repoRevSpec sourcegraph.RepoRevSpec) (err error) {
	cl := Client()

	// Resolve to the full commit ID, and ensure that the remote
	// server knows about the commit.
	commit, err := cl.Repos.GetCommit(ctx, &repoRevSpec)
	if err != nil {
		return err
	}
	repoRevSpec.CommitID = string(commit.ID)

	if globalOpt.Verbose {
		log.Printf("Pushing srclib data for %s@%s to server at %s...", repoRevSpec.URI, repoRevSpec.CommitID, sourcegraph.GRPCEndpoint(ctx))
	}

	// Perform the import.
	srcstore := pb.Client(ctx, pb.NewMultiRepoImporterClient(cl.Conn))

	bdfs, err := srclib.GetBuildDataFS(repoRevSpec.CommitID)
	if err != nil {
		return fmt.Errorf("getting local build data FS for %s@%s: %s", repoRevSpec.URI, repoRevSpec.CommitID, err)
	}

	importOpt := srclib.ImportOpt{
		Repo:     repoRevSpec.URI,
		CommitID: repoRevSpec.CommitID,
		Verbose:  globalOpt.Verbose,
	}
	if err := srclib.Import(bdfs, srcstore, importOpt); err != nil {
		return fmt.Errorf("import failed: %s", err)
	}

	// Precache root dir
	appURL, err := getRemoteAppURL(cli.Ctx)
	if err != nil {
		return err
	}
	cacheutil.HTTPAddr = appURL.String()
	cacheutil.PrecacheRoot(repoRevSpec.URI)

	return nil
}
