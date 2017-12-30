import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../../components/PageTitle'
import { SettingsFile } from '../../settings/SettingsFile'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchUserSettings, updateUserSettings } from './backend'

interface Props extends RouteComponentProps<any> {
    user: GQL.IUser
}

interface State {
    settings?: GQL.ISettings | null
    error?: string
    commitError?: Error
}

export class UserSettingsConfigurationPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSettingsConfiguration')

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
        const userInEditorBeta = this.props.user.tags && this.props.user.tags.some(tag => tag.name === 'editor-beta')

        return (
            <div className="user-settings-configuration-page">
                <PageTitle title="User configuration" />
                <h2>Configuration</h2>
                {this.state.error && <div className="alert alert-danger">{this.state.error}</div>}
                {this.state.settings !== undefined && (
                    <SettingsFile
                        settings={this.state.settings}
                        onDidCommit={this.onDidCommit}
                        commitError={this.state.commitError}
                        history={this.props.history}
                    />
                )}
                <small className="form-text">
                    Documentation:{' '}
                    <a target="_blank" href="https://about.sourcegraph.com/docs/search#scope">
                        Customizing search scopes
                    </a>
                </small>
                {userInEditorBeta && (
                    <small className="form-text">
                        Editor beta users: This configuration does not yet take effect in Sourcegraph Editor, unlike org
                        config (which does). It can only be used to configure the Sourcegraph web app.
                    </small>
                )}
            </div>
        )
    }

    private onDidCommit = (lastKnownSettingsID: number | null, contents: string): void => {
        this.setState({
            error: undefined,
            commitError: undefined,
        })
        updateUserSettings(lastKnownSettingsID, contents).subscribe(
            settings =>
                this.setState({
                    error: undefined,
                    commitError: undefined,
                    settings,
                }),
            error => {
                this.setState({ error: undefined, commitError: error.message })
                console.error(error)
            }
        )
    }
}
