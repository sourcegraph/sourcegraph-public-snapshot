package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/kballard/go-shellquote"
	"github.com/mattn/go-isatty"
)

func init() {
	usage := `
Examples:

  Upload an LSIF dump:

    	$ src lsif upload -repo=FOO -commit=BAR -upload-token=BAZ -file=data.lsif

  Upload an LSIF dump for a subproject:

    	$ src lsif upload -repo=FOO -commit=BAR -upload-token=BAZ -file=data.lsif -root=cmd/

`

	flagSet := flag.NewFlagSet("upload", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src lsif %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		repoFlag           = flagSet.String("repo", "", `The name of the repository. (required)`)
		commitFlag         = flagSet.String("commit", "", `The 40-character hash of the commit. (required)`)
		fileFlag           = flagSet.String("file", "", `The path to the LSIF dump file. (required)`)
		githubTokenFlag    = flagSet.String("github-token", "", `A GitHub access token with 'public_repo' scope that Sourcegraph uses to verify you have access to the repository.`)
		rootFlag           = flagSet.String("root", "", `The path in the repository that matches the LSIF projectRoot (e.g. cmd/)`)
		skipValidationFlag = flagSet.Bool("skip-validation", false, `Whether or not to perform input validation on the server (much faster)`)
		apiFlags           = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		ensureSet := func(value *string, argName string) {
			if value == nil || *value == "" {
				fmt.Printf("src lsif: no %s supplied\n", argName)
				fmt.Printf("Run 'src lsif help' for usage.\n")
				os.Exit(1)
			}
		}

		ensureSet(repoFlag, "repository")
		ensureSet(commitFlag, "commit")
		ensureSet(fileFlag, "dump file")

		// First, build the URL which is used to both make the request
		// and to emit a cURL command. This is a little different than
		// the rest of the commands as it does not use a GraphQL endpoint,
		// using the path and query string instead of the body.

		qs := url.Values{}
		qs.Add("repository", *repoFlag)
		qs.Add("commit", *commitFlag)
		if *githubTokenFlag != "" {
			qs.Add("github_token", *githubTokenFlag)
		}
		if *rootFlag != "" {
			qs.Add("root", *rootFlag)
		}
		if *skipValidationFlag == true {
			qs.Add("skipValidation", "true")
		}

		url, err := url.Parse(cfg.Endpoint + "/.api/lsif/upload")
		if err != nil {
			return err
		}
		url.RawQuery = qs.Encode()

		// Emit a cURL command. This is also a bit different than the rest
		// of the commands as it uploads a file rather than just sending a
		// JSON-encoded body.
		//
		// Because we compress the body before sending it to the API below,
		// we need to pipe the output of gzip into the cURL command.

		if *apiFlags.getCurl {
			curl := fmt.Sprintf("gzip -c %s | curl \\\n", shellquote.Join(*fileFlag))
			curl += fmt.Sprintf("   -X POST \\\n")

			curl += fmt.Sprintf("   %s \\\n", shellquote.Join("-H", "Content-Type: application/x-ndjson+lsif"))
			curl += fmt.Sprintf("   %s \\\n", shellquote.Join(url.String()))
			curl += fmt.Sprintf("   %s", shellquote.Join("--data-binary", "@-"))

			fmt.Println(curl)
			return nil
		}

		f, err := os.Open(*fileFlag)
		if err != nil {
			return err
		}
		defer f.Close()

		// compress the file
		pr, ch := gzipReader(f)

		// Create the HTTP request.
		req, err := http.NewRequest("POST", url.String(), pr)
		if err != nil {
			return err
		}

		req.Header.Set("Content-Type", "application/x-ndjson+lsif")

		// Perform the request.
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// See if we had a reader error
		if err := <-ch; err != nil {
			return err
		}

		// Our request may have failed before the reaching GraphQL endpoint, so
		// confirm the status code. You can test this easily with e.g. an invalid
		// endpoint like -endpoint=https://google.com
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			if resp.StatusCode == http.StatusUnauthorized && string(body) == "Must provide github_token.\n" {
				return fmt.Errorf("error: you have to provide -github-token with 'public_repo' scope")
			}

			if resp.StatusCode == http.StatusUnauthorized && isatty.IsTerminal(os.Stdout.Fd()) {
				fmt.Println("You may need to specify or update your GitHub access token to use this endpoint.")
				fmt.Println("See https://github.com/sourcegraph/src-cli#authentication")
				fmt.Println("")
			}
			return fmt.Errorf("error: %s\n\n%s", resp.Status, body)
		}

		fmt.Printf("LSIF dump uploaded.\n")
		return nil
	}

	// Register the command.
	lsifCommands = append(lsifCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

func gzipReader(r io.Reader) (io.Reader, <-chan error) {
	ch := make(chan error)
	br := bufio.NewReader(r)
	pr, pw := io.Pipe()
	gw := gzip.NewWriter(pw)

	go func() {
		defer close(ch)
		defer pw.Close() // must be closed 2nd
		defer gw.Close() // must be closed 1st

		if _, err := br.WriteTo(gw); err != nil {
			ch <- err
		}
	}()

	return pr, ch
}
