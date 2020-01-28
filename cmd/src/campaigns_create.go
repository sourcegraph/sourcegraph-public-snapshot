package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func init() {
	usage := `
Create a campaign with the given attributes. If -name or -desc are not specified $EDITOR will open a temporary Markdown file to edit both.

Examples:

  Create a campaign with the given name, description and campaign plan:

		$ src campaigns create -name="Format Go code" \
		   -desc="This campaign runs gofmt over all Go repositories" \
		   -plan=Q2FtcGFpZ25QbGFuOjM=

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
		nameFlag        = flagSet.String("name", "", "Name of the campaign. ")
		descriptionFlag = flagSet.String("desc", "", "Description for the campaign in Markdown.")
		namespaceFlag   = flagSet.String("namespace", "", "ID of the namespace under which to create the campaign. The namespace can be the GraphQL ID of a Sourcegraph user or organisation. If not specified, the ID of the authenticated user is queried and used. (Required)")
		planIDFlag      = flagSet.String("plan", "", "ID of campaign plan the campaign should turn into changesets. If no plan is specified, a campaign is created to which changesets can be added manually.")
		draftFlag       = flagSet.Bool("draft", false, "Create the campaign as a draft (which won't create pull requests on code hosts)")

		changesetsFlag = flagSet.Int("changesets", 1000, "Returns the first n changesets per campaign.")

		formatFlag = flagSet.String("f", "{{friendlyCampaignCreatedMessage .}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ID}}: {{.Name}}") or "{{.|json}}")`)
		apiFlags   = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

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

		var namespace string
		if *namespaceFlag != "" {
			namespace = *namespaceFlag
		} else {
			var currentUserResult struct {
				CurrentUser *User
			}

			req := &apiRequest{
				query:  currentUserIDQuery,
				result: &currentUserResult,
				flags:  apiFlags,
			}
			err := req.do()
			if err != nil {
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
			"plan":        nullString(*planIDFlag),
			"draft":       *draftFlag,
		}

		var result struct {
			CreateCampaign Campaign
		}

		return (&apiRequest{
			query: campaignFragment + createcampaignMutation,
			vars: map[string]interface{}{
				"input":           input,
				"changesetsFirst": nullInt(*changesetsFlag),
			},
			result: &result,
			done: func() error {
				return execTemplate(tmpl, result.CreateCampaign)
			},
			flags: apiFlags,
		}).do()
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
