import React, { useMemo } from 'react'
import { Collapsible } from '../../../../components/Collapsible'
import LanguageGoIcon from 'mdi-react/LanguageGoIcon'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import { highlightCode as _highlightCode } from '../../../../search/backend'
import { useObservable } from '../../../../../../shared/src/util/useObservable'
import { ThemeProps } from '../../../../../../shared/src/theme'
import LanguageTypescriptIcon from 'mdi-react/LanguageTypescriptIcon'

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
    return (
        <div className={className}>
            <h1>Create a campaign</h1>
            <div className="card">
                <div className="card-body">
                    <p className="alert alert-info">
                        Follow the step-by-step guide to get started with your first campaign. You can also find
                        examples at the bottom of this page.
                    </p>
                    <h3>1. Install the src cli</h3>
                    <div className="ml-2">
                        <p>
                            If you have not already, first install and configure the src CLI to point to your
                            Sourcegraph instance. This guide will get you the most recent compatible version with your
                            Sourcegraph instance.
                        </p>
                        <h4>Configure the endpoint of your Sourcegraph instance</h4>
                        <div>
                            <code>
                                export SRC_ENDPOINT={window.location.protocol}//{window.location.host}
                            </code>
                            <br />
                            <p>
                                <strong>Tip:</strong> You might want to put this in your shell config.
                            </p>
                        </div>
                        <h4>Download the src CLI</h4>
                        <h4>macOS</h4>
                        <p>
                            <code>
                                curl -L ${'{'}SRC_ENDPOINT{'}'}/.api/src-cli/src_dawin_amd64 -o /usr/local/bin/src
                            </code>
                        </p>
                        <h4>Linux</h4>
                        <p>
                            <code>
                                curl -L ${'{'}SRC_ENDPOINT{'}'}/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
                            </code>
                        </p>
                        <p>Also, make sure that git is installed and accessible by src.</p>
                        <h4>Grating access to src CLI</h4>
                        <p>
                            To acquire the access token, visit your Sourcegraph instance, click your username in the top
                            right to open the user menu, select Settings, and then select <strong>Access tokens</strong>{' '}
                            in the left hand menu. Then expose it to SRC_CLI using the following command:
                        </p>
                        <div>
                            <code>export SRC_ACCESS_TOKEN=&lt;YOUR TOKEN&gt;</code>
                            <br />
                            <p>
                                <strong>Tip:</strong> You might want to put this in your shell config.
                            </p>
                        </div>
                    </div>
                    <h3>2. Create an action JSON file (e.g. action.json) that contains an action definition</h3>
                    <div className="ml-2 mb-1">
                        See below for examples of those files and{' '}
                        <a
                            href="https://docs.sourcegraph.com/user/campaigns#defining-an-action"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            read the docs on what they can do
                        </a>
                        .
                    </div>
                    <h3>3. Optional: See repositories the action would run over</h3>
                    <div className="ml-2 mb-2">
                        <code>src actions scope-query -f action.json</code>
                    </div>
                    <h3>4. Create a set of patches by executing the action over repositories</h3>
                    <div className="ml-2 mb-2">
                        <code>src actions exec -f action.json &gt; patches.json</code>
                    </div>
                    <h3>5. Save the patches in Sourcegraph by creating a patch set</h3>
                    <div className="ml-2 mb-2">
                        <code>src campaign patchset create-from-patches &lt; patches.json</code>
                    </div>
                    <h3>5. Create a campaign based on the patch set</h3>
                    <div className="ml-2 mb-2">
                        <code>
                            src campaigns create -branch=&lt;branch-name&gt;
                            -patchset=&lt;patchset-ID-returned-by-previous-command&gt;
                        </code>
                    </div>
                    <div className="ml-2">
                        <a
                            href="https://docs.sourcegraph.com/user/campaigns#creating-a-campaign-using-the-src-cli"
                            rel="noopener noreferrer"
                            target="_blank"
                        >
                            Read on detailed steps and documentation in the docs.{' '}
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
