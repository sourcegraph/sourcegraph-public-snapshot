import copy from 'copy-to-clipboard'
import * as React from 'react'
import { PageTitle } from '../../components/PageTitle'
import { events } from '../../tracking/events'
import { sourcegraphContext } from '../../util/sourcegraphContext'

interface Props { }
interface State {
    copiedLink: boolean
}

/**
 * Page to enable users to authenticate/link to their editors
 */
export class EditorAuthPage extends React.Component<Props, State> {
    public state: State = { copiedLink: false }
    private sessionId = sourcegraphContext.sessionID.slice(sourcegraphContext.sessionID.indexOf(' ') + 1)

    public render(): JSX.Element | null {
        return (
            <div className='ui-section'>
                <PageTitle title='authenticate editor' />
                <h1>
                    Authenticate your editor
                </h1>
                <p>
                    Your session ID is:
                </p>
                <p className='ui-text-box'>
                    <textarea readOnly className='editor-auth-page__session-id' value={this.sessionId} />
                    <input type='button' className='ui-button' onClick={this.copySessionId}
                        value={this.state.copiedLink ? 'Copied to clipboard!' : 'Copy to clipboard'} />
                </p>
                <p>
                    Paste this value into the input box in your editor, and click 'Save.'
                </p>
                <p>
                    This will allow you to sync your favorite editor settings across devices,
                    join a team, make inline comments, and share your workspace with teammates!
                </p>
            </div>
        )
    }

    private copySessionId = (): void => {
        events.EditorAuthIdCopied.log()
        copy(this.sessionId)
        this.setState({ copiedLink: true })

        setTimeout(() => {
            this.setState({ copiedLink: false })
        }, 3000)
    }
}
