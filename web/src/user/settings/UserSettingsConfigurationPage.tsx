import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { catchError, concat, distinctUntilChanged, map, mergeMap, switchMap } from 'rxjs/operators'
import * as GQL from '../../backend/graphqlschema'
import { PageTitle } from '../../components/PageTitle'
import { fetchSettings, updateSettings } from '../../configuration/backend'
import { SettingsFile } from '../../settings/SettingsFile'
import { eventLogger } from '../../tracking/eventLogger'
import { ErrorLike, isErrorLike } from '../../util/errors'
import { UserAreaPageProps } from '../area/UserArea'
import { refreshConfiguration } from './backend'

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

    private userChanges = new Subject<{ id: GQL.ID /* user ID */ }>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSettingsConfiguration')

        // Load settings.
        this.subscriptions.add(
            this.userChanges
                .pipe(
                    distinctUntilChanged(),
                    switchMap(({ id }) =>
                        fetchSettings(id).pipe(catchError(error => [error]), map(c => ({ settingsOrError: c })))
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        this.userChanges.next(this.props.user)
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.user !== this.props.user) {
            this.userChanges.next(props.user)
        }
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
                <p>User settings override global and organization settings.</p>
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
            </div>
        )
    }

    private onDidCommit = (lastID: number | null, contents: string): void => {
        this.setState({ commitError: undefined })
        updateSettings(this.props.user.id, lastID, contents)
            .pipe(mergeMap(() => refreshConfiguration().pipe(concat([null]))))
            .subscribe(
                () => {
                    this.setState({ commitError: undefined })
                    this.userChanges.next({ id: this.props.user.id })
                },
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
