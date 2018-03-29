import CopyIcon from '@sourcegraph/icons/lib/Copy'
import copy from 'copy-to-clipboard'
import * as React from 'react'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'

interface State {
    copied: boolean
}

/**
 * Displays a user's session token.
 */
export class UserSettingsTokensPage extends React.Component<{}, State> {
    public state: State = { copied: false }
    private sessionToken = window.context.sessionID.slice(window.context.sessionID.indexOf(' ') + 1)

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSettingsTokens')
    }

    public render(): JSX.Element | null {
        const exampleCommand = `curl \\
  -H 'Authorization: session ${this.sessionToken}' \\
  -d '{"query":"query { currentUser { username } }"}' \\
  ${window.context.appURL}/.api/graphql`

        return (
            <div className="user-settings-tokens-page">
                <PageTitle title="Tokens" />
                <h2>Session token</h2>
                <p>This access token is valid for the duration of your user session.</p>
                <div className="form-control user-settings-tokens-page__textarea-container">
                    <textarea
                        readOnly={true}
                        className="user-settings-tokens-page__textarea"
                        rows={2}
                        value={this.sessionToken}
                    />
                </div>
                <button
                    type="button"
                    className="btn btn-primary mt-2 mb-3"
                    onClick={this.copySessionId}
                    disabled={this.state.copied}
                >
                    <CopyIcon className="icon-inline" />{' '}
                    {this.state.copied ? 'Copied to clipboard!' : 'Copy to clipboard'}
                </button>
                <h3 className="mt-3">Usage</h3>
                <p>
                    To make an authenticated HTTP request to the Sourcegraph API, set <code>Authorization</code> HTTP
                    request header as shown in the example below.
                </p>
                <div className="form-control user-settings-tokens-page__textarea-container">
                    <textarea
                        readOnly={true}
                        className="user-settings-tokens-page__textarea user-settings-tokens-page__textarea-pre"
                        value={exampleCommand}
                        rows={5}
                    />
                </div>
            </div>
        )
    }

    private copySessionId = (): void => {
        eventLogger.log('UserTokenCopied')
        copy(this.sessionToken)
        this.setState({ copied: true })

        setTimeout(() => {
            this.setState({ copied: false })
        }, 1500)
    }
}
