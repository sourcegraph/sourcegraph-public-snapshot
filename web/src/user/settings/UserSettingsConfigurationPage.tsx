import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { merge, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, switchMap, tap } from 'rxjs/operators'
import * as GQL from '../../backend/graphqlschema'
import { PageTitle } from '../../components/PageTitle'
import { SettingsFile } from '../../settings/SettingsFile'
import { eventLogger } from '../../tracking/eventLogger'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'
import { UserAreaPageProps } from '../area/UserArea'
import { fetchUserSettings, updateUserSettings } from './backend'

interface Props extends UserAreaPageProps, RouteComponentProps<{}> {
    isLightTheme: boolean
}

interface State {
    /** The user's settings, or an error, or undefined while loading. */
    settingsOrError?: GQL.ISettings | null | ErrorLike

    /** An error that occurred while saving the settings. */
    commitError?: Error
}

export class UserSettingsConfigurationPage extends React.Component<Props, State> {
    public state: State = {}

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSettingsConfiguration')

        const userChanges = this.componentUpdates.pipe(
            distinctUntilChanged((a, b) => a.user.id === b.user.id),
            map(({ user }) => user)
        )

        // Fetch the user settings.
        this.subscriptions.add(
            userChanges
                .pipe(
                    tap(() => this.setState({ commitError: undefined })),
                    switchMap(user =>
                        merge(
                            of({ settingsOrError: undefined }),
                            fetchUserSettings(user.id).pipe(
                                catchError(error => [asError(error)]),
                                map(c => ({ settingsOrError: c } as Pick<State, 'settingsOrError'>))
                            )
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="user-settings-configuration-page">
                <PageTitle title="User configuration" />
                <h2>Configuration</h2>
                {isErrorLike(this.state.settingsOrError) && (
                    <p className="alert alert-danger">Error: {upperFirst(this.state.settingsOrError.message)}</p>
                )}
                <p>View and edit user search scopes and saved queries.</p>
                {this.state.settingsOrError !== undefined &&
                    !isErrorLike(this.state.settingsOrError) && (
                        <SettingsFile
                            settings={this.state.settingsOrError}
                            onDidCommit={this.onDidCommit}
                            onDidDiscard={this.onDidDiscard}
                            commitError={this.state.commitError}
                            history={this.props.history}
                            isLightTheme={this.props.isLightTheme}
                        />
                    )}
                <small className="form-text">
                    Documentation:{' '}
                    <a target="_blank" href="https://about.sourcegraph.com/docs/server/config/search-scopes">
                        Customizing search scopes
                    </a>
                </small>
            </div>
        )
    }

    private onDidCommit = (lastKnownSettingsID: number | null, contents: string): void => {
        this.setState({ commitError: undefined })
        updateUserSettings(this.props.user.id, lastKnownSettingsID, contents).subscribe(
            settings =>
                this.setState({
                    commitError: undefined,
                    settingsOrError: settings,
                }),
            error => {
                this.setState({ commitError: error })
                console.error(error)
            }
        )
    }

    private onDidDiscard = (): void => {
        this.setState({ commitError: undefined })
    }
}
