import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { concat } from 'rxjs/operators/concat'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { switchMap } from 'rxjs/operators/switchMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { SettingsFile } from '../settings/SettingsFile'
import { eventLogger } from '../tracking/eventLogger'
import { refreshConfiguration } from '../user/settings/backend'
import { fetchOrg, updateOrgSettings } from './backend'

interface Props extends RouteComponentProps<any> {
    org: GQL.IOrg
    user: GQL.IUser
    isLightTheme: boolean
}

interface State {
    settings?: GQL.ISettings | null
    error?: string
    commitError?: Error
}

export class OrgSettingsConfigurationPage extends React.PureComponent<Props, State> {
    public state: State = {}

    private orgChanges = new Subject<GQL.IOrg | undefined>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('OrgSettingsConfiguration')

        this.subscriptions.add(
            this.orgChanges
                .pipe(switchMap(org => fetchOrg((org || this.props.org).id)))
                .subscribe(
                    org => this.setState({ settings: org && org.latestSettings }),
                    error => this.setState({ error: error.message })
                )
        )
        this.orgChanges.next(this.props.org)
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.org !== this.props.org) {
            this.orgChanges.next(props.org)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const orgInEditorBeta = this.props.org.tags.some(tag => tag.name === 'editor-beta')

        return (
            <div className="settings-file-container">
                <PageTitle title="Organization configuration" />
                <h2>Configuration</h2>
                <p>View and edit your organization's search scopes and saved queries.</p>
                {this.state.settings !== undefined && (
                    <SettingsFile
                        settings={this.state.settings}
                        commitError={this.state.commitError}
                        onDidCommit={this.onDidCommit}
                        history={this.props.history}
                        isLightTheme={this.props.isLightTheme}
                    />
                )}
                <small className="form-text">
                    Documentation:{' '}
                    <a target="_blank" href="https://about.sourcegraph.com/docs/server/config/search-scopes">
                        Customizing search scopes for org members
                    </a>
                </small>
                {orgInEditorBeta && (
                    <small className="form-text">
                        This configuration applies to all org members and takes effect in Sourcegraph Editor and on the
                        web. You can also run the 'Preferences: Open Organization Settings' command inside of
                        Sourcegraph Editor to change this configuration.
                    </small>
                )}
            </div>
        )
    }

    private onDidCommit = (lastKnownSettingsID: number | null, contents: string) =>
        updateOrgSettings(this.props.org.id, lastKnownSettingsID, contents)
            .pipe(mergeMap(() => refreshConfiguration().pipe(concat([null]))))
            .subscribe(
                () => {
                    this.setState({ commitError: undefined })
                    this.orgChanges.next(undefined)
                },
                err => {
                    this.setState({ commitError: err })
                    console.error(err)
                }
            )
}
