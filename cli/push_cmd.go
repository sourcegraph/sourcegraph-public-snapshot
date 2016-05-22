package cli

import (
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	srclib "sourcegraph.com/sourcegraph/srclib/cli"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
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
	Repo     string `long:"repo" description:"repo URI (on server) to import into"`
	CommitID string `long:"commit" description:"commit ID of data to import"`
}

func (c *pushCmd) Execute(args []string) error {
	lrepo, err := srclib.OpenLocalRepo()
	if err != nil {
		return err
	}

	commitID := lrepo.CommitID
	if c.CommitID != "" {
		commitID = c.CommitID
	}

	repoSpec := sourcegraph.RepoSpec{URI: c.Repo}
	repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: repoSpec, CommitID: commitID}

	appURL, err := getRemoteAppURL(cliContext)
	if err != nil {
		return err
	}

	if err := c.do(cliContext, repoRevSpec); err != nil {
		return err
	}

	log.Printf("# Success! View the repository at: %s", appURL.ResolveReference(router.Rel.URLToRepoRev(repoRevSpec.URI, repoRevSpec.CommitID)))

	return nil
}

func (c *pushCmd) do(ctx context.Context, repoRevSpec sourcegraph.RepoRevSpec) (err error) {
	cl := cliClient

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

	return nil
}
