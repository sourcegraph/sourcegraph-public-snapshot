import React, { useMemo } from 'react'
import { Collapsible } from '../../../../components/Collapsible'
import LanguageGoIcon from 'mdi-react/LanguageGoIcon'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import { highlightCode as _highlightCode } from '../../../../search/backend'
import { useObservable } from '../../../../../../shared/src/util/useObservable'
import { ThemeProps } from '../../../../../../shared/src/theme'
import LanguageTypescriptIcon from 'mdi-react/LanguageTypescriptIcon'
import { highlight } from 'highlight.js/lib/highlight'

interface Props extends ThemeProps {
    className?: string
    highlightCode?: typeof _highlightCode
}

const lsifAction = `{
  "scopeQuery": "repohasfile:go.mod repo:github -repohasfile:.github/workflows/lsif.yml",
  "steps": [
    {
      "type": "docker",
      "image": "add-lsif-to-build-pipeline-action"
    }
  ]
}`

const lsifGHAction = `name: LSIF
on:
  - push
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v1
      - name: Generate LSIF data
        uses: sourcegraph/lsif-go-action@master
        with:
          verbose: "true"
      - name: Upload LSIF data
        uses: sourcegraph/lsif-upload-action@master
        with:
          github_token: \${{ secrets.GITHUB_TOKEN }}`

const lsifDockerfile = `FROM alpine:3
ADD ./github-action-workflow-golang.yml /tmp/workflows/

CMD mkdir -p .github/workflows && \\
  DEST=.github/workflows/lsif.yml; \\
  if [ ! -f .github/workflows/lsif.yml ]; then \\
    cp /tmp/workflows/github-action-workflow-golang.yml $DEST; \\
  else \\
    echo Doing nothing because existing LSIF workflow found at $DEST; \\
  fi`

const combyAction = `{
  "scopeQuery": "lang:go fmt.Sprintf",
  "steps": [
    {
      "type": "docker",
      "image": "comby/comby",
      "args": ["-in-place", "fmt.Sprintf(\\"%d\\", :[v])", "strconv.Itoa(:[v])", "-matcher", ".go", "-d", "/work"]
    },
    {
      "type": "docker",
      "image": "cytopia/goimports",
      "args": ["-w", "/work"]
    }
  ]
}`

const eslintAction = `{
  "scopeQuery": "repohasfile:yarn\\\\.lock repohasfile:tsconfig\\\\.json",
  "steps": [
    {
      "type": "docker",
      "image": "eslint-fix-action"
    }
  ]
}`

const eslintDockerfile = `FROM node:12-alpine3.10
CMD package_json_bkup=$(mktemp) && \\
  cp package.json $package_json_bkup && \\
  yarn -s --non-interactive --pure-lockfile --ignore-optional --ignore-scripts --ignore-engines --ignore-platform --no-progress && \\
  yarn add --ignore-workspace-root-test --non-interactive --ignore-optional --ignore-scripts --ignore-engines --ignore-platform --pure-lockfile -D @typescript-eslint/parser @typescript-eslint/eslint-plugin --no-progress eslint && \\
  node_modules/.bin/eslint \\
  --fix \\
  --plugin @typescript-eslint \\
  --parser @typescript-eslint/parser \\
  --parser-options '{"ecmaVersion": 8, "sourceType": "module", "project": "tsconfig.json"}' \\
  --rule '@typescript-eslint/prefer-optional-chain: 2' \\
  --rule '@typescript-eslint/no-unnecessary-type-assertion: 2' \\
  --rule '@typescript-eslint/no-unnecessary-type-arguments: 2' \
  --rule '@typescript-eslint/no-unnecessary-condition: 2' \\
  --rule '@typescript-eslint/no-unnecessary-type-arguments: 2' \\
  --rule '@typescript-eslint/prefer-includes: 2' \\
  --rule '@typescript-eslint/prefer-readonly: 2' \\
  --rule '@typescript-eslint/prefer-string-starts-ends-with: 2' \\
  --rule '@typescript-eslint/prefer-nullish-coalescing: 2' \\
  --rule '@typescript-eslint/no-non-null-assertion: 2' \\
  '**/*.ts'; \\
  mv $package_json_bkup package.json; \\
  rm -rf node_modules; \\
  yarn upgrade --latest --ignore-workspace-root-test --non-interactive --ignore-optional --ignore-scripts --ignore-engines --ignore-platform --no-progress typescript && \\
  rm -rf node_modules`

/**
 * A tutorial and a list of examples for campaigns using src CLI
 */
export const CampaignCLIHelp: React.FunctionComponent<Props> = ({
    className,
    isLightTheme,
    highlightCode = _highlightCode,
}) => {
    const formattedLsifAction = useObservable(
        useMemo(
            () =>
                highlightCode({
                    code: lsifAction,
                    isLightTheme,
                    fuzzyLanguage: 'json',
                    disableTimeout: false,
                }),
            [isLightTheme, highlightCode]
        )
    )
    const formattedLsifGHAction = useObservable(
        useMemo(
            () =>
                highlightCode({
                    code: lsifGHAction,
                    isLightTheme,
                    fuzzyLanguage: 'yaml',
                    disableTimeout: false,
                }),
            [isLightTheme, highlightCode]
        )
    )
    const formattedLsifDockerfile = useObservable(
        useMemo(
            () =>
                highlightCode({
                    code: lsifDockerfile,
                    isLightTheme,
                    fuzzyLanguage: 'Dockerfile',
                    disableTimeout: false,
                }),
            [isLightTheme, highlightCode]
        )
    )
    const formattedCombyAction = useObservable(
        useMemo(
            () =>
                highlightCode({
                    code: combyAction,
                    isLightTheme,
                    fuzzyLanguage: 'json',
                    disableTimeout: false,
                }),
            [isLightTheme, highlightCode]
        )
    )
    const formattedEslintAction = useObservable(
        useMemo(
            () =>
                highlightCode({
                    code: eslintAction,
                    isLightTheme,
                    fuzzyLanguage: 'json',
                    disableTimeout: false,
                }),
            [isLightTheme, highlightCode]
        )
    )
    const formattedEslintDockerfile = useObservable(
        useMemo(
            () =>
                highlightCode({
                    code: eslintDockerfile,
                    isLightTheme,
                    fuzzyLanguage: 'Dockerfile',
                    disableTimeout: false,
                }),
            [isLightTheme, highlightCode]
        )
    )

    const srcInstall = `# Configure your Sourcegraph instance:
$ export SRC_ENDPOINT=${window.location.protocol}//${window.location.host}

# Download the src binary for macOS:
$ curl -L $SRC_ENDPOINT/.api/src-cli/src_dawin_amd64 -o /usr/local/bin/src
# Download the src binary for Linux:
$ curl -L $SRC_ENDPOINT/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src

# Set your personal access token:
$ export SRC_ACCESS_TOKEN=<YOUR TOKEN>
`

    return (
        <div className={className}>
            <h1>Create a campaign</h1>
            <div className="card">
                <div className="card-body p-3">
                    <h3>
                        1. Install the{' '}
                        <a href="https://github.com/sourcegraph/src-cli">
                            <code>src</code> CLI
                        </a>
                    </h3>
                    <div className="ml-2">
                        <p>Install and configure the src CLI:</p>
                        <pre className="ml-3">
                            <code
                                dangerouslySetInnerHTML={{
                                    __html: highlight('bash', srcInstall, true).value,
                                }}
                            />
                        </pre>
                        <p>
                            Make sure that <code>git</code> is installed and accessible by the src CLI.
                        </p>
                        <p>
                            To create and manage access tokens, click your username in the top right to open the user
                            menu, select <strong>Settings</strong>, and then <strong>Access tokens</strong>.
                        </p>
                    </div>
                    <h3>2. Create an action definition</h3>
                    <div className="ml-2 mb-1">
                        See below for examples and{' '}
                        <a
                            href="https://docs.sourcegraph.com/user/campaigns/creating_campaign_from_patches"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            read the documentation on what actions can can do.
                        </a>
                        .
                    </div>
                    <h3>3. Optional: See repositories the action would run over</h3>
                    <div className="ml-2 mb-2">
                        <pre className="ml-3">
                            <code
                                dangerouslySetInnerHTML={{
                                    __html: highlight('bash', '$ src actions scope-query -f action.json', true).value,
                                }}
                            />
                        </pre>
                    </div>
                    <h3>4. Create a set of patches by executing the action over repositories</h3>
                    <div className="ml-2 mb-2">
                        <pre className="ml-3">
                            <code
                                dangerouslySetInnerHTML={{
                                    __html: highlight(
                                        'bash',
                                        '$ src actions exec -f action.json -create-patchset',
                                        true
                                    ).value,
                                }}
                            />
                        </pre>
                    </div>
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
                </div>
            </div>
            <a id="examples" />
            <h2 className="mt-2">Examples</h2>
            <ul className="list-group mb-3">
                <li className="list-group-item p-2">
                    <Collapsible
                        title={
                            <h3 className="mb-0">
                                <GithubCircleIcon className="icon-inline ml-1 mr-2" /> Add a GitHub action to upload
                                LSIF data to Sourcegraph
                            </h3>
                        }
                        titleClassName="flex-grow-1"
                    >
                        <div>
                            <p>
                                Our goal for this campaign is to add a GitHub Action that generates and uploads LSIF
                                data to Sourcegraph by adding a .github/workflows/lsif.yml file to each repository that
                                doesn’t have it yet.
                            </p>
                            <p>
                                In order to build the Docker image, we first need to create a file called
                                <code>github-action-workflow-golang.yml</code> with the following content:
                            </p>

                            {formattedLsifGHAction && (
                                <code dangerouslySetInnerHTML={{ __html: formattedLsifGHAction }} />
                            )}
                            <p>And a Dockerfile:</p>
                            {formattedLsifDockerfile && (
                                <code dangerouslySetInnerHTML={{ __html: formattedLsifDockerfile }} />
                            )}
                            <p className="mt-2">Build the docker image:</p>
                            <p>
                                <code>docker build -t add-lsif-to-build-pipeline-action .</code>
                            </p>
                            {formattedLsifAction && <code dangerouslySetInnerHTML={{ __html: formattedLsifAction }} />}
                        </div>
                    </Collapsible>
                </li>
                <li className="list-group-item p-2">
                    <Collapsible
                        title={
                            <h3 className="mb-0">
                                <LanguageGoIcon className="icon-inline ml-1 mr-2" /> Refactor Go code using Comby
                            </h3>
                        }
                        titleClassName="flex-grow-1"
                    >
                        <div>
                            <p>
                                Our goal for this campaign is to simplify Go code by using Comby to rewrite calls to
                                <code>fmt.Sprintf("%d", arg)</code> with <code>strconv.Itoa(:[v])</code>. The semantics
                                are the same, but one more cleanly expresses the intention behind the code.
                            </p>
                            <p className="text-muted">
                                Note: Learn more about Comby and what it’s capable of at{' '}
                                <a href="https://comby.dev" target="_blank" rel="noopener noreferrer">
                                    comby.dev
                                </a>
                            </p>
                            <p>
                                To do that we use two Docker containers. One container launches Comby to rewrite the the
                                code in Go files and the other runs goimports to update the import statements in the
                                updated Go code so that strconv is correctly imported and, possibly, fmt is removed.
                            </p>
                            {formattedCombyAction && (
                                <code dangerouslySetInnerHTML={{ __html: formattedCombyAction }} />
                            )}
                        </div>
                    </Collapsible>
                </li>
                <li className="list-group-item p-2">
                    <Collapsible
                        title={
                            <h3 className="mb-0">
                                <LanguageTypescriptIcon className="icon-inline ml-1 mr-2" /> Upgrade to a new TypeScript
                                version
                            </h3>
                        }
                        titleClassName="flex-grow-1"
                    >
                        <div>
                            <p>
                                Our goal for this campaign is to convert all TypeScript code synced to our Sourcegraph
                                instance to make use of new TypeScript features. To do this we convert the code, then
                                update the TypeScript version.
                            </p>
                            <p>
                                Our goal for this campaign is to convert all TypeScript code synced to ouTo convert the
                                code we install and run ESLint with the desired typescript-eslint rules, using the --fix
                                flag to automatically fix problems. We then update the TypeScript version using yarn
                                upgrade.
                            </p>
                            <p>
                                The first thing we need is a Docker container in which we can freely install and run
                                ESLint. Here is the Dockerfile:
                            </p>
                            {formattedEslintDockerfile && (
                                <code dangerouslySetInnerHTML={{ __html: formattedEslintDockerfile }} />
                            )}
                            <p className="mt-2">
                                When turned into an image and run as a container, the instructions in this Dockerfile
                                will do the following:
                            </p>
                            <ol>
                                <li>
                                    Copy the current package.json to a backup location so that we can install ESLint
                                    without changes to the original package.json
                                </li>
                                <li>Install all dependencies &amp; add eslint with the typescript-eslint plugin</li>
                                <li>
                                    Run eslint --fix with a set of TypeScript rules to detect and fix problems over all
                                    *.ts files
                                </li>
                                <li>Restore the original package.json from its backup location</li>
                                <li>Run yarn upgrade to update the typescript version</li>
                            </ol>
                            <p>
                                Before we can run it as an action we need to turn it into a Docker image, by running the
                                following command in the directory where the Dockerfile was saved:
                            </p>
                            <code>docker build -t eslint-fix-action .</code>
                            <p>That builds a Docker image and names it eslint-fix-action.</p>
                            <p>Once that is done we’re ready to define our action:</p>
                            {formattedEslintAction && (
                                <code dangerouslySetInnerHTML={{ __html: formattedEslintAction }} />
                            )}
                        </div>
                    </Collapsible>
                </li>
            </ul>
        </div>
    )
}
