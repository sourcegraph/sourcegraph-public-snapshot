import copy from 'copy-to-clipboard'
import * as React from 'react'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'

interface Props {}
interface State {
    copiedLink: boolean
}

/**
 * Page to enable users to authenticate/link to their editors
 */
export class EditorAuthPage extends React.Component<Props, State> {
    public state: State = { copiedLink: false }
    private sessionId = window.context.sessionID.slice(window.context.sessionID.indexOf(' ') + 1)

    public componentDidMount(): void {
        eventLogger.logViewEvent('EditorAuth')
    }

    public render(): JSX.Element | null {
        return (
            <div className="editor-auth-page">
                <PageTitle title="Authenticate editor" />
                <h1>Authenticate your editor</h1>
                <p>Your session ID is:</p>
                <p className="form-control editor-auth-page__session-id-container">
                    <textarea readOnly={true} className="editor-auth-page__session-id" value={this.sessionId} />
                    <input
                        type="button"
                        className="btn btn-primary"
                        onClick={this.copySessionId}
                        value={this.state.copiedLink ? 'Copied to clipboard!' : 'Copy to clipboard'}
                    />
                </p>
                <p>Paste this value into the input box in your editor, and click 'Save.'</p>
                <p>
                    Learn more about{' '}
                    <a
                        href="https://about.sourcegraph.com/products/editor"
                        target="_blank"
                        className="editor-auth-page__link"
                    >
                        Sourcegraph Editor
                    </a>, and{' '}
                    <a
                        href="https://about.sourcegraph.com/products/editor#beta"
                        target="_blank"
                        className="editor-auth-page__link"
                    >
                        sign up for beta access
                    </a>{' '}
                    today.
                </p>
            </div>
        )
    }

    private copySessionId = (): void => {
        eventLogger.log('EditorAuthIdCopied')
        copy(this.sessionId)
        this.setState({ copiedLink: true })

        setTimeout(() => {
            this.setState({ copiedLink: false })
        }, 3000)
    }
}
