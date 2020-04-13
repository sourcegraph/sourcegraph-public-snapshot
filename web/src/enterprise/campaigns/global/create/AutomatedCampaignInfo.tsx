import * as React from 'react'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { Collapsible } from '../../../../components/Collapsible'
import LanguageGoIcon from 'mdi-react/LanguageGoIcon'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'

interface Props extends ThemeProps {
    className?: string
}

const combySample = `{
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

/**
 * A tutorial and a list of examples for campaigns
 */
export const AutomatedCampaignInfo: React.FunctionComponent<Props> = ({ className }) => (
    <div className={className}>
        <h1>Create a campaign</h1>
        <div className="card">
            <div className="card-body">
                <p className="alert alert-info">
                    Follow the step-by-step guide to get started with your first campaign. You can also find examples at
                    the bottom of this page.
                </p>
                <h3>1. Install the src cli</h3>
                <p className="pl-2">
                    First, you need to{' '}
                    <a href="https://github.com/sourcegraph/src-cli#setup">Install and set-up the src CLI</a>.
                    <br />
                    <br />
                    <code>
                        export SRC_ENDPOINT={window.location.protocol}//{window.location.host}
                        <br />
                        curl -L ${'{'}SRC_ENDPOINT{'}'}/.api/src-cli/src_dawin_amd64 -o /usr/local/bin/src curl -L $
                        {'{'}SRC_ENDPOINT{'}'}/.api/src-cli/src_linux_amd64 -o /usr/local/bin/src
                    </code>
                </p>
                <h3>2. Run your first action</h3>
                <p className="pl-2">
                    Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut
                    labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores
                    et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.
                    Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod tempor invidunt ut
                    labore et dolore magna aliquyam erat, sed diam voluptua. At vero eos et accusam et justo duo dolores
                    et ea rebum. Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.
                </p>

                <a id="examples" />
                <h3>Examples</h3>
                <ul className="list-group">
                    <li className="list-group-item p-2">
                        <Collapsible
                            title={
                                <h3 className="mb-0">
                                    <LanguageGoIcon className="icon-inline mr-3" /> Add install guide to Go repositories
                                </h3>
                            }
                        >
                            <div>
                                <code>{combySample}</code>
                            </div>
                        </Collapsible>
                    </li>
                    <li className="list-group-item p-2">
                        <Collapsible
                            title={
                                <h3 className="mb-0">
                                    <GithubCircleIcon className="icon-inline mr-3" /> Add a GitHub action to upload LSIF
                                    data to Sourcegraph
                                </h3>
                            }
                        >
                            <div>
                                <code>{combySample}</code>
                            </div>
                        </Collapsible>
                    </li>
                    <li className="list-group-item p-2">
                        <Collapsible
                            title={
                                <h3 className="mb-0">
                                    <LanguageGoIcon className="icon-inline mr-3" /> Refactor Go code with Comby
                                </h3>
                            }
                        >
                            <div>
                                <code>{combySample}</code>
                            </div>
                        </Collapsible>
                    </li>
                </ul>
            </div>
        </div>
    </div>
)
