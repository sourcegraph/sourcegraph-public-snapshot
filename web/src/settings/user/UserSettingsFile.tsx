import * as H from 'history'
import * as React from 'react'
import { Subscription } from 'rxjs/Subscription'
import { fetchUserSettings, updateUserSettings } from '../backend'
import { SettingsFile } from '../SettingsFile'

interface Props {
    userInEditorBeta: boolean

    history: H.History
}

interface State {
    error?: Error
    settings?: GQL.ISettings | null
}

export class UserSettingsFile extends React.PureComponent<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            fetchUserSettings().subscribe(
                settings => this.setState({ settings }),
                error => this.setState({ error: error.message })
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="settings-file-container">
                {this.state.error && <p className="settings-file-container__error">{this.state.error}</p>}
                {this.state.settings !== undefined && (
                    <SettingsFile
                        settings={this.state.settings}
                        onDidCommit={this.onDidCommit}
                        history={this.props.history}
                    />
                )}
                <small className="form-text">
                    Documentation:{' '}
                    <a target="_blank" href="https://about.sourcegraph.com/docs/search#scope">
                        Customizing search scopes
                    </a>
                </small>
                {this.props.userInEditorBeta && (
                    <small className="form-text">
                        Editor beta users: This configuration does not yet take effect in Sourcegraph Editor, unlike org
                        config (which does). It can only be used to configure the Sourcegraph web app.
                    </small>
                )}
            </div>
        )
    }

    private onDidCommit = (lastKnownSettingsID: number | null, contents: string): void => {
        this.setState({ error: undefined })
        updateUserSettings(lastKnownSettingsID, contents).subscribe(
            settings =>
                this.setState({
                    error: undefined,
                    settings,
                }),
            error => {
                this.setState({ error: error.message })
                console.error(error)
            }
        )
    }
}
