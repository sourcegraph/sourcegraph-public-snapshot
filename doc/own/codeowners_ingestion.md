## Codeowners ingestion

Sourcegraph Own allows you to surface ownership data using `CODEOWNERS` files. 
This is done automatically if there is a committed `CODEOWNERS` file in your repository at any of the following locations:

```md
CODEOWNERS
.github/CODEOWNERS
.gitlab/CODEOWNERS
docs/CODEOWNERS
```

However, it might be you do not want to have a `CODEOWNERS` file committed to your repository (for example, to avoid automatic review requests), or you would like to overwrite the existing one.

Sourcegraph provides a UI and CLI to ingest a `CODEOWNERS` file per-repository, that overrides any existing committed file.

### Ingesting a file through the UI

Navigating to any repository page, clicking the Ownership button will surface information about any ingested `CODEOWNERS` file, and will allow you to upload or update an existing one.

![img](pending)

### Ingesting a file with src-cli

There is the option to ingest data with the Sourcegraph [src-cli](../cli/quickstart.md).
The CLI provides `add`, `update`, `delete`, and `list` functionality.

```bash
'src codeowners' is a tool that manages ingested code ownership data in a Sourcegraph instance.

Usage:

	src codeowners command [command options]

The commands are:

	get	returns the codeowners file for a repository, if exists
	create	create a codeowners file
	update	update a codeowners file
	delete	delete a codeowners file

Use "src codeowners [command] -h" for more information about a command.
```

The input file can be written inline or passed in. 

### Limitations 

- The file should respect `CODEOWNERS` formatting for Sourcegraph Own to surface useful information. No formatting validation is done at upload time.
- Only site admins can add, update or delete a `CODEOWNERS` file through the ingestion API.
- Ingested `CODEOWNERS` files are limited to a size of 10Mb if uploaded through the client.