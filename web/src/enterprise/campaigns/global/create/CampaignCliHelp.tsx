import React from 'react'
import LanguageGoIcon from 'mdi-react/LanguageGoIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import LanguageTypescriptIcon from 'mdi-react/LanguageTypescriptIcon'
import { highlightCodeSafe } from '../../../../../../shared/src/util/markdown'

const sourceInstall = `# Configure your Sourcegraph instance:
$ export SRC_ENDPOINT=${window.location.protocol}//${window.location.host}

# Download the src binary for macOS:
$ curl -L $SRC_ENDPOINT/.api/src-cli/src_darwin_amd64 -o /usr/local/bin/src
# Download the src binary for Linux:
$ curl -L $SRC_ENDPOINT/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src

# Set your personal access token:
$ export SRC_ACCESS_TOKEN=<YOUR TOKEN>
`

const actionFileExample = `$ echo '{
  "scopeQuery": "repohasfile:main.go",
  "steps": [
    {
      "type": "docker",
      "image": "golang:1.14-alpine",
      "args": ["sh", "-c", "cd /work && gofmt -w ./"]
    }
  ]
}
' > action.json
`

interface Props {
    className?: string
}

/**
 * A tutorial and a list of examples for campaigns using src CLI
 */
export const CampaignCliHelp: React.FunctionComponent<Props> = ({ className }) => (
    <div className={className}>
        <h1>
            Create a campaign{' '}
            <sup>
                <span className="badge badge-info">Beta</span>
            </sup>
        </h1>
        <div className="card">
            <div className="card-body p-3">
                <div className="alert alert-info mt-2">
                    <a
                        href=" https://docs.sourcegraph.com/user/campaigns/creating_campaign_from_patches"
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        Take a look at the documentation for more detailed steps and additional information.{' '}
                        <small>
                            <ExternalLinkIcon className="icon-inline" />
                        </small>
                    </a>
                </div>
                <h3>
                    1. Install the{' '}
                    <a href="https://github.com/sourcegraph/src-cli">
                        <code>src</code> CLI
                    </a>
                </h3>
                <div className="ml-2">
                    <pre className="alert alert-secondary ml-3">
                        <code
                            dangerouslySetInnerHTML={{
                                __html: highlightCodeSafe(sourceInstall, 'bash'),
                            }}
                        />
                    </pre>
                    <p>
                        Make sure that <code>git</code> is installed and accessible by the src CLI.
                    </p>
                    <p>
                        To create and manage access tokens, click your username in the top right to open the user menu,
                        select <strong>Settings</strong>, and then <strong>Access tokens</strong>.
                    </p>
                </div>
                <h3>2. Create an action definition</h3>
                <div className="ml-2 mb-1">
                    <p>
                        Here is a short example definition to run <code>gofmt</code> over all repositories that have a{' '}
                        <code>main.go</code> file:
                    </p>
                    <pre className="alert alert-secondary ml-3">
                        <code
                            dangerouslySetInnerHTML={{
                                __html: highlightCodeSafe(actionFileExample, 'bash'),
                            }}
                        />
                    </pre>
                    <p>
                        See the examples below for more real-world use cases and read "{' '}
                        <a
                            href="https://docs.sourcegraph.com/user/campaigns/creating_campaign_from_patches"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            Creating a campaign from patches
                        </a>
                        " in the documentation for more information about what actions can do.
                    </p>
                </div>
                <h3>3. Create a set of patches by executing the action over repositories</h3>
                <div className="ml-2 mb-2">
                    <pre className="alert alert-secondary ml-3">
                        <code
                            dangerouslySetInnerHTML={{
                                __html: highlightCodeSafe('$ src action exec -f action.json -create-patchset', 'bash'),
                            }}
                        />
                    </pre>
                    <p>
                        After executing the command, follow the printed instructions to create a campaign from the
                        patches and turn the patches into changesets (pull requests) on your code hosts.
                    </p>
                    <p>
                        Run <code>src action exec -h</code> to see more usage information about this command.
                    </p>
                </div>
            </div>
        </div>
        <a id="examples" />
        <h2 className="mt-2">Examples</h2>
        <ul className="list-group mb-3">
            <li className="list-group-item p-2">
                <h3 className="mb-0">
                    <GithubIcon className="icon-inline ml-1 mr-2" />{' '}
                    <a
                        href="https://docs.sourcegraph.com/user/campaigns/examples/lsif_action"
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        Add a GitHub action to upload LSIF data to Sourcegraph
                    </a>
                </h3>
            </li>
            <li className="list-group-item p-2">
                <h3 className="mb-0">
                    <LanguageGoIcon className="icon-inline ml-1 mr-2" />{' '}
                    <a
                        href="https://docs.sourcegraph.com/user/campaigns/examples/refactor_go_comby"
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        Refactor Go code using Comby
                    </a>
                </h3>
            </li>
            <li className="list-group-item p-2">
                <h3 className="mb-0">
                    <LanguageTypescriptIcon className="icon-inline ml-1 mr-2" />{' '}
                    <a
                        href="https://docs.sourcegraph.com/user/campaigns/examples/eslint_typescript_version"
                        rel="noopener noreferrer"
                        target="_blank"
                    >
                        Migrate to a new TypeScript version
                    </a>
                </h3>
            </li>
        </ul>
    </div>
)
