package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	usage := `
Create a campaign with the given attributes. If -name or -desc are not specified $EDITOR will open a temporary Markdown file to edit both.

Examples:

  Create a campaign with the given name, branch, description and campaign patch set:

		$ src campaigns create -name="Format Go code" \
		   -desc="This campaign runs gofmt over all Go repositories" \
		   -branch=run-go-fmt \
		   -patchset=Q2FtcGFpZ25QbGFuOjM=

  Create a manual campaign with the given name and description and adds two GitHub pull requests to it:

		$ src campaigns create -name="Migrate to Python 3" \
		   -desc="This campaign tracks all Python 3 migration PRs"
		$ src campaigns add-changesets -campaign=<id-returned-by-previous-command> \
		   -repo-name=github.com/our-org/a-python-repo 5612 7321

`

	flagSet := flag.NewFlagSet("create", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src campaigns create %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		nameFlag        = flagSet.String("name", "", "Name of the campaign.")
		descriptionFlag = flagSet.String("desc", "", "Description for the campaign in Markdown.")
		namespaceFlag   = flagSet.String("namespace", "", "ID of the namespace under which to create the campaign. The namespace can be the GraphQL ID of a Sourcegraph user or organisation. If not specified, the ID of the authenticated user is queried and used. (Required)")
		patchsetIDFlag  = flagSet.String("patchset", "", "ID of patch set the campaign should turn into changesets. If no patch set is specified, a campaign is created to which changesets can be added manually.")
		branchFlag      = flagSet.String("branch", "", "Name of the branch that will be created in each repository on the code host. Required for Sourcegraph >= 3.13 when 'patchset' is specified.")

		changesetsFlag = flagSet.Int("changesets", 1000, "Returns the first n changesets per campaign.")

		formatFlag = flagSet.String("f", "{{friendlyCampaignCreatedMessage .}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ID}}: {{.Name}}") or "{{.|json}}")`)
		apiFlags   = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		err := flagSet.Parse(args)
		if err != nil {
			return err
		}

		var name, description string

		if *nameFlag == "" || *descriptionFlag == "" {
			editor := &CampaignEditor{
				Name:        *nameFlag,
				Description: *descriptionFlag,
			}

			var err error
			name, description, err = editor.EditAndExtract()
			if err != nil {
				return err
			}
		} else {
			name = *nameFlag
			description = *descriptionFlag
		}

		if name == "" {
			return &usageError{errors.New("campaign name cannot be blank")}
		}

		if description == "" {
			return &usageError{errors.New("campaign description cannot be blank")}
		}

		ctx := context.Background()
		client := cfg.apiClient(apiFlags, flagSet.Output())

		if *patchsetIDFlag != "" {
			// We only need to check for -branch if the Sourcegraph version is >= 3.13
			version, err := getSourcegraphVersion(ctx, client)
			if err != nil {
				return err
			}
			needsBranch, err := sourcegraphVersionCheck(version, ">= 3.13-0", "2020-02-13")
			if err != nil {
				return err
			}

			if needsBranch && *branchFlag == "" {
				return &usageError{errors.New("branch cannot be blank for campaigns with a patch set")}
			}
		}

		var namespace string
		if *namespaceFlag != "" {
			namespace = *namespaceFlag
		} else {
			var currentUserResult struct {
				CurrentUser *User
			}

			if _, err := client.NewQuery(currentUserIDQuery).Do(ctx, &currentUserResult); err != nil {
				return err
			}
			if currentUserResult.CurrentUser.ID == "" {
				return errors.New("Failed to query authenticated user's ID")
			}
			namespace = currentUserResult.CurrentUser.ID
		}

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		input := map[string]interface{}{
			"name":        name,
			"description": description,
			"namespace":   namespace,
			"patchSet":    api.NullString(*patchsetIDFlag),
			"branch":      *branchFlag,
		}

		var result struct {
			CreateCampaign Campaign
		}

		if ok, err := client.NewRequest(campaignFragment+createcampaignMutation, map[string]interface{}{
			"input":           input,
			"changesetsFirst": api.NullInt(*changesetsFlag),
		}).Do(ctx, &result); err != nil || !ok {
			return err
		}

		return execTemplate(tmpl, result.CreateCampaign)
	}

	// Register the command.
	campaignsCommands = append(campaignsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

const currentUserIDQuery = `query CurrentUser { currentUser { id } }`

const createcampaignMutation = `mutation CreateCampaign($input: CreateCampaignInput!, $changesetsFirst: Int) {
  createCampaign(input: $input) {
	... campaign
  }
}
`

const (
	sep    = "------- EVERYTHING BELOW THIS LINE WILL BE IGNORED -------"
	notice = `You are creating a new campaign.
Write a name and description for this campaign in this file.
The first line of text is the name and the rest is the description.`
)

type CampaignEditor struct {
	Name        string
	Description string
}

func (e *CampaignEditor) EditAndExtract() (string, string, error) {
	f, err := ioutil.TempFile("", "new-campaign*.md")
	if err != nil {
		return "", "", err
	}
	defer os.Remove(f.Name())

	err = e.writeTemplate(f)
	if err != nil {
		return "", "", err
	}

	err = openInEditor(f.Name())
	if err != nil {
		return "", "", errors.Wrap(err, "Failed to open text editor to edit campaign")
	}

	content, err := extractContent(f.Name())
	if err != nil {
		return "", "", err
	}

	var name, description string

	parts := strings.SplitN(content, "\n\n", 2)
	if len(parts) >= 1 {
		name = strings.TrimSpace(strings.Replace(parts[0], "\n", " ", -1))
	}
	if len(parts) >= 2 {
		description = strings.TrimSpace(parts[1])
	}

	return name, description, nil
}

func (e *CampaignEditor) writeTemplate(f *os.File) error {
	template := e.Name + "\n\n" + e.Description
	template += "\n\n" + sep
	template += "\n\n" + notice

	_, err := f.WriteString(template)
	return err
}

func extractContent(file string) (string, error) {
	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	trimmed := bytes.TrimSpace(fileContent)

	scanner := bufio.NewScanner(bytes.NewReader(trimmed))

	content := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		if line == sep {
			break
		}
		content = append(content, line)
	}
	if err = scanner.Err(); err != nil {
		return "", err
	}

	return strings.Join(content, "\n"), nil
}

func openInEditor(file string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return errors.New("$EDITOR is not set")
	}

	cmd := exec.Command(editor, file)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	tty, err := os.OpenFile("/dev/tty", os.O_RDONLY, 0660)
	if err == nil {
		cmd.Stdin = tty
	}

	return cmd.Run()
}

const sourcegraphVersionQuery = `query SourcegraphVersion {
  site {
    productVersion
  }
}
`

func getSourcegraphVersion(ctx context.Context, client api.Client) (string, error) {
	var sourcegraphVersion struct {
		Site struct {
			ProductVersion string
		}
	}

	_, err := client.NewQuery(sourcegraphVersionQuery).Do(ctx, &sourcegraphVersion)
	return sourcegraphVersion.Site.ProductVersion, err
}

func sourcegraphVersionCheck(version, constraint, minDate string) (bool, error) {
	if version == "dev" || version == "0.0.0+dev" {
		return true, nil
	}

	buildDate := regexp.MustCompile(`^\d+_(\d{4}-\d{2}-\d{2})_[a-z0-9]{7}$`)
	matches := buildDate.FindStringSubmatch(version)
	if len(matches) > 1 {
		return matches[1] >= minDate, nil
	}

	c, err := semver.NewConstraint(constraint)
	if err != nil {
		return false, nil
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return false, err
	}
	return c.Check(v), nil
}
