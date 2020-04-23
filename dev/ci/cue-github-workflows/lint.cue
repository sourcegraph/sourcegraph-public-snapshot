import (
	"strings"
	"path"
)

Schema :: _ @jsonschema(schema="http://json-schema.org/draft-07/schema")
Schema :: number | null | bool | string | [...] | {
	// The name of your workflow. GitHub displays the names of your
	// workflows on your repository's actions page. If you omit this
	// field, GitHub sets the name to the workflow's filename.
	name?: string

	// A map of environment variables that are available to all jobs
	// and steps in the workflow.
	env?: def.env

	// A map of default settings that will apply to all jobs in the
	// workflow.
	defaults?: def.defaults

	// The name of the GitHub event that triggers the workflow. You
	// can provide a single event string, array of events, array of
	// event types, or an event configuration map that schedules a
	// workflow or restricts the execution of a workflow to specific
	// files, tags, or branch changes. For a list of available
	// events, see
	// https://help.github.com/en/github/automating-your-workflow-with-github-actions/events-that-trigger-workflows.
	on: def.event | [...def.event] & [_, ...] | {
		// Runs your workflow anytime the status of a Git commit changes,
		// which triggers the status event. For information about the
		// REST API, see https://developer.github.com/v3/repos/statuses/.
		status?: def.eventObject

		// Runs your workflow anytime the check_run event occurs. More
		// than one activity type triggers this event. For information
		// about the REST API, see
		// https://developer.github.com/v3/checks/runs.
		check_run?: def.eventObject & {
			types?: def.types & [..."created" | "rerequested" | "completed" | "requested_action"] | *["created", "rerequested", "completed", "requested_action"]
			...
		}

		// Runs your workflow anytime the check_suite event occurs. More
		// than one activity type triggers this event. For information
		// about the REST API, see
		// https://developer.github.com/v3/checks/suites/.
		check_suite?: def.eventObject & {
			types?: def.types & [..."completed" | "requested" | "rerequested"] | *["completed", "requested", "rerequested"]
			...
		}

		// Runs your workflow anytime someone creates a branch or tag,
		// which triggers the create event. For information about the
		// REST API, see
		// https://developer.github.com/v3/git/refs/#create-a-reference.
		create?: def.eventObject

		// Runs your workflow anytime someone deletes a branch or tag,
		// which triggers the delete event. For information about the
		// REST API, see
		// https://developer.github.com/v3/git/refs/#delete-a-reference.
		delete?: def.eventObject

		// Runs your workflow anytime someone creates a deployment, which
		// triggers the deployment event. Deployments created with a
		// commit SHA may not have a Git ref. For information about the
		// REST API, see
		// https://developer.github.com/v3/repos/deployments/.
		deployment?: def.eventObject

		// Runs your workflow anytime a third party provides a deployment
		// status, which triggers the deployment_status event.
		// Deployments created with a commit SHA may not have a Git ref.
		// For information about the REST API, see
		// https://developer.github.com/v3/repos/deployments/#create-a-deployment-status.
		deployment_status?: def.eventObject

		// Runs your workflow anytime when someone forks a repository,
		// which triggers the fork event. For information about the REST
		// API, see
		// https://developer.github.com/v3/repos/forks/#create-a-fork.
		fork?: def.eventObject

		// Runs your workflow when someone creates or updates a Wiki page,
		// which triggers the gollum event.
		gollum?: def.eventObject

		// Runs your workflow anytime the issue_comment event occurs. More
		// than one activity type triggers this event. For information
		// about the REST API, see
		// https://developer.github.com/v3/issues/comments/.
		issue_comment?: def.eventObject & {
			types?: def.types & [..."created" | "edited" | "deleted"] | *["created", "edited", "deleted"]
			...
		}

		// Runs your workflow anytime the issues event occurs. More than
		// one activity type triggers this event. For information about
		// the REST API, see https://developer.github.com/v3/issues.
		issues?: def.eventObject & {
			types?: def.types & [..."opened" | "edited" | "deleted" | "transferred" | "pinned" | "unpinned" | "closed" | "reopened" | "assigned" | "unassigned" | "labeled" | "unlabeled" | "locked" | "unlocked" | "milestoned" | "demilestoned"] | *["opened", "edited", "deleted", "transferred", "pinned", "unpinned", "closed", "reopened", "assigned", "unassigned", "labeled", "unlabeled", "locked", "unlocked", "milestoned", "demilestoned"]
			...
		}

		// Runs your workflow anytime the label event occurs. More than
		// one activity type triggers this event. For information about
		// the REST API, see
		// https://developer.github.com/v3/issues/labels/.
		label?: def.eventObject & {
			types?: def.types & [..."created" | "edited" | "deleted"] | *["created", "edited", "deleted"]
			...
		}

		// Runs your workflow anytime the member event occurs. More than
		// one activity type triggers this event. For information about
		// the REST API, see
		// https://developer.github.com/v3/repos/collaborators/.
		member?: def.eventObject & {
			types?: def.types & [..."added" | "edited" | "deleted"] | *["added", "edited", "deleted"]
			...
		}

		// Runs your workflow anytime the milestone event occurs. More
		// than one activity type triggers this event. For information
		// about the REST API, see
		// https://developer.github.com/v3/issues/milestones/.
		milestone?: def.eventObject & {
			types?: def.types & [..."created" | "closed" | "opened" | "edited" | "deleted"] | *["created", "closed", "opened", "edited", "deleted"]
			...
		}

		// Runs your workflow anytime someone pushes to a GitHub
		// Pages-enabled branch, which triggers the page_build event. For
		// information about the REST API, see
		// https://developer.github.com/v3/repos/pages/.
		page_build?: def.eventObject

		// Runs your workflow anytime the project event occurs. More than
		// one activity type triggers this event. For information about
		// the REST API, see https://developer.github.com/v3/projects/.
		project?: def.eventObject & {
			types?: def.types & [..."created" | "updated" | "closed" | "reopened" | "edited" | "deleted"] | *["created", "updated", "closed", "reopened", "edited", "deleted"]
			...
		}

		// Runs your workflow anytime the project_card event occurs. More
		// than one activity type triggers this event. For information
		// about the REST API, see
		// https://developer.github.com/v3/projects/cards.
		project_card?: def.eventObject & {
			types?: def.types & [..."created" | "moved" | "converted" | "edited" | "deleted"] | *["created", "moved", "converted", "edited", "deleted"]
			...
		}

		// Runs your workflow anytime the project_column event occurs.
		// More than one activity type triggers this event. For
		// information about the REST API, see
		// https://developer.github.com/v3/projects/columns.
		project_column?: def.eventObject & {
			types?: def.types & [..."created" | "updated" | "moved" | "deleted"] | *["created", "updated", "moved", "deleted"]
			...
		}

		// Runs your workflow anytime someone makes a private repository
		// public, which triggers the public event. For information about
		// the REST API, see https://developer.github.com/v3/repos/#edit.
		public?: def.eventObject

		// Runs your workflow anytime the pull_request event occurs. More
		// than one activity type triggers this event. For information
		// about the REST API, see https://developer.github.com/v3/pulls.
		// Note: Workflows do not run on private base repositories when
		// you open a pull request from a forked repository.
		// When you create a pull request from a forked repository to the
		// base repository, GitHub sends the pull_request event to the
		// base repository and no pull request events occur on the forked
		// repository.
		// Workflows don't run on forked repositories by default. You must
		// enable GitHub Actions in the Actions tab of the forked
		// repository.
		// The permissions for the GITHUB_TOKEN in forked repositories is
		// read-only. For more information about the GITHUB_TOKEN, see
		// https://help.github.com/en/articles/virtual-environments-for-github-actions.
		pull_request?: def.ref & {
			types?: def.types & [..."assigned" | "unassigned" | "labeled" | "unlabeled" | "opened" | "edited" | "closed" | "reopened" | "synchronize" | "ready_for_review" | "locked" | "unlocked" | "review_requested" | "review_request_removed"] | *["opened", "synchronize", "reopened"]

			[=~"^(branche|tag|path)s(-ignore)?$" & !~"^(types)$"]: _
		}

		// Runs your workflow anytime the pull_request_review event
		// occurs. More than one activity type triggers this event. For
		// information about the REST API, see
		// https://developer.github.com/v3/pulls/reviews.
		// Note: Workflows do not run on private base repositories when
		// you open a pull request from a forked repository.
		// When you create a pull request from a forked repository to the
		// base repository, GitHub sends the pull_request event to the
		// base repository and no pull request events occur on the forked
		// repository.
		// Workflows don't run on forked repositories by default. You must
		// enable GitHub Actions in the Actions tab of the forked
		// repository.
		// The permissions for the GITHUB_TOKEN in forked repositories is
		// read-only. For more information about the GITHUB_TOKEN, see
		// https://help.github.com/en/articles/virtual-environments-for-github-actions.
		pull_request_review?: def.eventObject & {
			types?: def.types & [..."submitted" | "edited" | "dismissed"] | *["submitted", "edited", "dismissed"]
			...
		}

		// Runs your workflow anytime a comment on a pull request's
		// unified diff is modified, which triggers the
		// pull_request_review_comment event. More than one activity type
		// triggers this event. For information about the REST API, see
		// https://developer.github.com/v3/pulls/comments.
		// Note: Workflows do not run on private base repositories when
		// you open a pull request from a forked repository.
		// When you create a pull request from a forked repository to the
		// base repository, GitHub sends the pull_request event to the
		// base repository and no pull request events occur on the forked
		// repository.
		// Workflows don't run on forked repositories by default. You must
		// enable GitHub Actions in the Actions tab of the forked
		// repository.
		// The permissions for the GITHUB_TOKEN in forked repositories is
		// read-only. For more information about the GITHUB_TOKEN, see
		// https://help.github.com/en/articles/virtual-environments-for-github-actions.
		pull_request_review_comment?: def.eventObject & {
			types?: def.types & [..."created" | "edited" | "deleted"] | *["created", "edited", "deleted"]
			...
		}

		// Runs your workflow when someone pushes to a repository branch,
		// which triggers the push event.
		// Note: The webhook payload available to GitHub Actions does not
		// include the added, removed, and modified attributes in the
		// commit object. You can retrieve the full commit object using
		// the REST API. For more information, see
		// https://developer.github.com/v3/repos/commits/#get-a-single-commit.
		push?: def.ref &
			{
				[=~"^(branche|tag|path)s(-ignore)?$" & !~"^()$"]: _
			}

		// Runs your workflow anytime the release event occurs. More than
		// one activity type triggers this event. For information about
		// the REST API, see
		// https://developer.github.com/v3/repos/releases/.
		release?: def.eventObject & {
			types?: def.types & [..."published" | "unpublished" | "created" | "edited" | "deleted" | "prereleased"] | *["published", "unpublished", "created", "edited", "deleted", "prereleased"]
			...
		}

		// Runs your workflow anytime the watch event occurs. More than
		// one activity type triggers this event. For information about
		// the REST API, see
		// https://developer.github.com/v3/activity/starring/.
		watch?: def.eventObject

		// You can use the GitHub API to trigger a webhook event called
		// repository_dispatch when you want to trigger a workflow for
		// activity that happens outside of GitHub. For more information,
		// see
		// https://developer.github.com/v3/repos/#create-a-repository-dispatch-event.
		// To trigger the custom repository_dispatch webhook event, you
		// must send a POST request to a GitHub API endpoint and provide
		// an event_type name to describe the activity type. To trigger a
		// workflow run, you must also configure your workflow to use the
		// repository_dispatch event.
		repository_dispatch?: def.eventObject

		// You can schedule a workflow to run at specific UTC times using
		// POSIX cron syntax
		// (https://pubs.opengroup.org/onlinepubs/9699919799/utilities/crontab.html#tag_20_25_07).
		// Scheduled workflows run on the latest commit on the default or
		// base branch. The shortest interval you can run scheduled
		// workflows is once every 5 minutes.
		// Note: GitHub Actions does not support the non-standard syntax
		// @yearly, @monthly, @weekly, @daily, @hourly, and @reboot.
		// You can use crontab guru (https://crontab.guru/). to help
		// generate your cron syntax and confirm what time it will run.
		// To help you get started, there is also a list of crontab guru
		// examples (https://crontab.guru/examples.html).
		schedule?: [...number | null | bool | string | [...] | {
			cron?: =~"^(((\\d+,)+\\d+|((\\d+|\\*)\\/\\d+)|(\\d+-\\d+)|\\d+|\\*) ?){5,7}$"
		}] & [_, ...]
	}

	// A workflow run is made up of one or more jobs. Jobs run in
	// parallel by default. To run jobs sequentially, you can define
	// dependencies on other jobs using the jobs.<job_id>.needs
	// keyword.
	// Each job runs in a fresh instance of the virtual environment
	// specified by runs-on.
	// You can run an unlimited number of jobs as long as you are
	// within the workflow usage limits. For more information, see
	// https://help.github.com/en/github/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#usage-limits.
	jobs: [=~"^[_a-zA-Z][a-zA-Z0-9_-]*$" & !~"^()$"]: {
		// The name of the job displayed on GitHub.
		name?: string

		// A map of environment variables that are available to all steps
		// in the job.
		env?: def.env

		// A container to run any steps in a job that don't already
		// specify a container. If you have steps that use both script
		// and container actions, the container actions will run as
		// sibling containers on the same network with the same volume
		// mounts.
		// If you do not set a container, all steps will run directly on
		// the host specified by runs-on unless a step refers to an
		// action configured to run in a container.
		container?: [string]: string | def.container

		// A map of default settings that will apply to all steps in the
		// job.
		defaults?: def.defaults

		// Identifies any jobs that must complete successfully before this
		// job will run. It can be a string or array of strings. If a job
		// fails, all jobs that need it are skipped unless the jobs use a
		// conditional statement that causes the job to continue.
		needs?: [...def.name] & [_, ...] | def.name

		// The type of machine to run the job on. The machine can be
		// either a GitHub-hosted runner, or a self-hosted runner.
		"runs-on": "${{ matrix.os }}" | "macos-latest" | "macos-10.15" | "self-hosted" | "ubuntu-16.04" | "ubuntu-18.04" | "ubuntu-latest" | "windows-latest" | "windows-2019" | (["self-hosted"] | ["self-hosted", def.machine] | ["self-hosted", def.architecture] | ["self-hosted", def.machine, def.architecture] | ["self-hosted", def.architecture, def.machine])

		// A map of outputs for a job. Job outputs are available to all
		// downstream jobs that depend on this job.
		outputs?: [string]: string

		// You can use the if conditional to prevent a job from running
		// unless a condition is met. You can use any supported context
		// and expression to create a conditional.
		// Expressions in an if conditional do not require the ${{ }}
		// syntax. For more information, see
		// https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.
		if?: string

		// A job contains a sequence of tasks called steps. Steps can run
		// commands, run setup tasks, or run an action in your
		// repository, a public repository, or an action published in a
		// Docker registry. Not all steps run actions, but all actions
		// run as a step. Each step runs in its own process in the
		// virtual environment and has access to the workspace and
		// filesystem. Because steps run in their own process, changes to
		// environment variables are not preserved between steps. GitHub
		// provides built-in steps to set up and complete a job.
		steps?: [...{
			// A name for your step to display on GitHub.
			name?: string

			// Sets environment variables for steps to use in the virtual
			// environment. You can also set environment variables for the
			// entire workflow or a job.
			env?: def.env

			// Runs command-line programs using the operating system's shell.
			// If you do not provide a name, the step name will default to
			// the text specified in the run command.
			// Commands run using non-login shells by default. You can choose
			// a different shell and customize the shell used to run
			// commands. For more information, see
			// https://help.github.com/en/actions/automating-your-workflow-with-github-actions/workflow-syntax-for-github-actions#using-a-specific-shell.
			// Each run keyword represents a new process and shell in the
			// virtual environment. When you provide multi-line commands,
			// each line runs in the same shell.
			run?:                 string
			shell?:               def.shell
			"working-directory"?: def.workingDirectory

			// You can use the if conditional to prevent a step from running
			// unless a condition is met. You can use any supported context
			// and expression to create a conditional.
			// Expressions in an if conditional do not require the ${{ }}
			// syntax. For more information, see
			// https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.
			if?: string

			// A unique identifier for the step. You can use the id to
			// reference the step in contexts. For more information, see
			// https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.
			id?: string

			// Selects an action to run as part of a step in your job. An
			// action is a reusable unit of code. You can use an action
			// defined in the same repository as the workflow, a public
			// repository, or in a published Docker container image
			// (https://hub.docker.com/).
			// We strongly recommend that you include the version of the
			// action you are using by specifying a Git ref, SHA, or Docker
			// tag number. If you don't specify a version, it could break
			// your workflows or cause unexpected behavior when the action
			// owner publishes an update.
			// - Using the commit SHA of a released action version is the
			// safest for stability and security.
			// - Using the specific major action version allows you to receive
			// critical fixes and security patches while still maintaining
			// compatibility. It also assures that your workflow should still
			// work.
			// - Using the master branch of an action may be convenient, but
			// if someone releases a new major version with a breaking
			// change, your workflow could break.
			// Some actions require inputs that you must set using the with
			// keyword. Review the action's README file to determine the
			// inputs required.
			// Actions are either JavaScript files or Docker containers. If
			// the action you're using is a Docker container you must run the
			// job in a Linux virtual environment. For more details, see
			// https://help.github.com/en/articles/virtual-environments-for-github-actions.
			uses?: string

			// A map of the input parameters defined by the action. Each input
			// parameter is a key/value pair. Input parameters are set as
			// environment variables. The variable is prefixed with INPUT_
			// and converted to upper case.
			with?: def.env & {
				args?:       string
				entrypoint?: string
				...
			}

			// Prevents a job from failing when a step fails. Set to true to
			// allow a job to pass when this step fails.
			"continue-on-error"?: bool | *false

			// The maximum number of minutes to run the step before killing
			// the process.
			"timeout-minutes"?: number
		}] & [_, ...]

		// Prevents a workflow run from failing when a job fails. Set to
		// true to allow a workflow run to pass when this job fails.
		"continue-on-error"?: bool | string

		// The maximum number of minutes to let a workflow run before
		// GitHub automatically cancels it. Default: 360
		"timeout-minutes"?: number | *360

		// A strategy creates a build matrix for your jobs. You can define
		// different variations of an environment to run each job in.
		strategy?: {
			// A build matrix is a set of different configurations of the
			// virtual environment. For example you might run a job against
			// more than one supported version of a language, operating
			// system, or tool. Each configuration is a copy of the job that
			// runs and reports a status.
			// You can specify a matrix by supplying an array for the
			// configuration options. For example, if the GitHub virtual
			// environment supports Node.js versions 6, 8, and 10 you could
			// specify an array of those versions in the matrix.
			// When you define a matrix of operating systems, you must set the
			// required runs-on keyword to the operating system of the
			// current job, rather than hard-coding the operating system
			// name. To access the operating system name, you can use the
			// matrix.os context parameter to set runs-on. For more
			// information, see
			// https://help.github.com/en/articles/contexts-and-expression-syntax-for-github-actions.
			matrix: {
				[=~"^(in|ex)clude$" & !~"^()$"]: [...{
					[string]: def.configuration
				}] & [_, ...]

				[!~"^(in|ex)clude$" & !~"^()$"]: [...def.configuration] & [_, ...]
			}

			// When set to true, GitHub cancels all in-progress jobs if any
			// matrix job fails. Default: true
			"fail-fast"?: bool | *true

			// The maximum number of jobs that can run simultaneously when
			// using a matrix job strategy. By default, GitHub will maximize
			// the number of jobs run in parallel depending on the available
			// runners on GitHub-hosted virtual machines.
			"max-parallel"?: number
		}

		// Additional containers to host services for a job in a workflow.
		// These are useful for creating databases or cache services like
		// redis. The runner on the virtual machine will automatically
		// create a network and manage the life cycle of the service
		// containers.
		// When you use a service container for a job or your step uses
		// container actions, you don't need to set port information to
		// access the service. Docker automatically exposes all ports
		// between containers on the same network.
		// When both the job and the action run in a container, you can
		// directly reference the container by its hostname. The hostname
		// is automatically mapped to the service name.
		// When a step does not use a container action, you must access
		// the service using localhost and bind the ports.
		services?: [string]: def.container
	}
}

def: path :: def.globs

def: name :: =~"^[_a-zA-Z][a-zA-Z0-9_-]*$"

def: env ::
	[string]: number | bool | string

def: architecture :: "ARM32" | "x64" | "x86"

def: branch :: def.globs

def: configuration :: string | number | {
	[string]: def.configuration
} | [...def.configuration]

def: container :: {
	// Sets an array of environment variables in the container.
	env?: def.env

	// The Docker image to use as the container to run the action. The
	// value can be the Docker Hub image name or a public docker
	// registry name.
	image: string

	// Sets an array of ports to expose on the container.
	ports?: [...number | string] & [_, ...]

	// Sets an array of volumes for the container to use. You can use
	// volumes to share data between services or other steps in a
	// job. You can specify named Docker volumes, anonymous Docker
	// volumes, or bind mounts on the host.
	// To specify a volume, you specify the source and destination
	// path: <source>:<destinationPath>
	// The <source> is a volume name or an absolute path on the host
	// machine, and <destinationPath> is an absolute path in the
	// container.
	volumes?: [...=~"^[^:]+:[^:]+$"] & [_, ...]

	// Additional Docker container resource options. For a list of
	// options, see
	// https://docs.docker.com/engine/reference/commandline/create/#options.
	options?: string
}

def: defaults ::
	run?: {
		shell?:               def.shell
		"working-directory"?: def.workingDirectory
	}

def: shell :: string | ("bash" | "pwsh" | "python" | "sh" | "cmd" | "powershell")

def: event :: "check_run" | "check_suite" | "create" | "delete" | "deployment" | "deployment_status" | "fork" | "gollum" | "issue_comment" | "issues" | "label" | "member" | "milestone" | "page_build" | "project" | "project_card" | "project_column" | "public" | "pull_request" | "pull_request_review" | "pull_request_review_comment" | "push" | "registry_package" | "release" | "status" | "watch" | "repository_dispatch"

def: eventObject :: null

def: globs :: [...strings.MinRunes(1)] & [_, ...]

def: machine :: "linux" | "macos" | "windows"

def: ref :: null | {
	branches?:          def.branch
	"branches-ignore"?: def.branch
	tags?:              def.branch
	"tags-ignore"?:     def.branch
	paths?:             def.path
	"paths-ignore"?:    def.path
	...
}

def: types :: [_, ...]

def: workingDirectory :: string

setup: [
	{
		name: "[setup] checkout repository"
		uses: "actions/checkout@v2"
	},
	{
		name: "[setup] install asdf"
		uses: "asdf-vm/actions/setup@v1.0.0"
	},
	{
		name: "[setup] (asdf) configure .nvmrc"
		run: """
		bash -c \"echo 'legacy_version_file = yes' > ~/.asdfrc\"
		"""
	},
	{
		name: "[setup] (asdf) install nodejs plugin"
		run:  "asdf plugin-add nodejs"
	}, {
		name: "[setup] (asdf) import nodejs keyring"
		run: """
		bash ~/.asdf/plugins/nodejs/bin/import-release-team-keyring
		"""
	}, {
		name: "[setup] (asdf) add plugins"
		uses: "asdf-vm/actions/plugins-add@v1.0.0"
	}, {

		name: "[setup] (asdf) install tools"
		run: """
		asdf install
		"""
	},
]

check_job :: {
	Name = name
	name:      string
	"runs-on": "ubuntu-latest"
	steps:     setup + [
			{
			name: path.Base(Name)
			run:  Name
		},
	]
	env: {
		"PGHOST":     "localhost"
		"PGPORT":     "5416"
		"PGUSER":     "postgres"
		"PGPASSWORD": "postgres"
		"PGSSLMODE":  "disable"
	}
	services: {
		postgres: {
			image: "postgres:9.6.17"
			env: {
				"POSTGRES_PASSWORD": "postgres"
			}
			options: strings.Join([
					"--health-cmd pg_isready",
					"--health-interval 10s",
					"--health-timeout 5s",
					"--health-retries 10",
			], " ")

			ports: [
				"5416:5432",
			]
		}
	}

}

workflow: Schema
workflow: {
	name: "Check"
	on: [
		"push",
	]
	jobs: [string]: check_job
	jobs: {
		"bash-syntax": name:          "./dev/check/bash-syntax.sh"
		"build": name:                "./dev/check/build.sh"
		"docsite": name:              "./dev/check/docsite.sh"
		"go-db-conn-import": name:    "./dev/check/go-dbconn-import.sh"
		"broken-urls": name:          "./dev/check/broken-urls.bash"
		"check-owners": name:         "./dev/check/check-owners.sh"
		"go-enterprise-import": name: "./dev/check/go-enterprise-import.sh"
		"go-generate": name:          "./dev/check/go-generate.sh"
		"go-lint": name:              "./dev/check/go-lint.sh"
		"gofmt": name:                "./dev/check/gofmt.sh"
		"no-localhost-guard": name:   "./dev/check/no-localhost-guard.sh"
		"template-inlines": name:     "./dev/check/template-inlines.sh"
		"todo-security": name:        "./dev/check/todo-security.sh"
		"yarn-deduplicate": name:     "./dev/check/yarn-deduplicate.sh"
		"shfmt": name:                "./dev/check/shfmt.sh"
		"shellcheck": name:           "./dev/check/shellcheck.sh"
		"licenses": name:             "./dev/check/licenses.sh"
	}
}
