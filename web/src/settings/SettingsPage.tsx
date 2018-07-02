import ErrorIcon from '@sourcegraph/icons/lib/Error'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { catchError, concat, distinctUntilChanged, map, mergeMap, switchMap } from 'rxjs/operators'
import * as GQL from '../backend/graphqlschema'
import { HeroPage } from '../components/HeroPage'
import { overwriteSettings } from '../configuration/backend'
import { fetchSettings } from '../configuration/backend'
import { SettingsFile } from '../settings/SettingsFile'
import { refreshConfiguration } from '../user/settings/backend'
import { ErrorLike, isErrorLike } from '../util/errors'

interface Props extends Pick<RouteComponentProps<{}>, 'history' | 'location'> {
    /** The subject whose settings to edit. */
    subject: Pick<GQL.ConfigurationSubject, '__typename' | 'id'>

    /** Optional description to render above the editor. */
    description?: React.ReactNode

    isLightTheme: boolean
}

interface State {
    settingsOrError?: GQL.ISettings | null | ErrorLike
    commitError?: Error
}

/**
 * Displays a page where the settings for a subject can be edited.
 */
export class SettingsPage extends React.PureComponent<Props, State> {
    public state: State = {}

    private subjectChanges = new Subject<Pick<GQL.IConfigurationSubject, 'id'>>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Load settings.
        this.subscriptions.add(
            this.subjectChanges
                .pipe(
                    distinctUntilChanged(),
                    switchMap(({ id }) =>
                        fetchSettings(id).pipe(catchError(error => [error]), map(c => ({ settingsOrError: c })))
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        this.subjectChanges.next(this.props.subject)
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.subject !== this.props.subject) {
            this.subjectChanges.next(props.subject)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.settingsOrError === undefined) {
            return null // loading
        }
        if (isErrorLike(this.state.settingsOrError)) {
            return <HeroPage icon={ErrorIcon} title="Error" subtitle={upperFirst(this.state.settingsOrError.message)} />
        }

        let term: string
        switch (this.props.subject.__typename) {
            case 'User':
                term = 'User'
                break
            case 'Org':
                term = 'Organization'
                break
            case 'Site':
                term = 'Global'
                break
            default:
                term = 'Unknown'
                break
        }

        return (
            <div>
                <h2>{term} settings</h2>
                {this.props.description}
                <SettingsFile
                    settings={this.state.settingsOrError}
                    commitError={this.state.commitError}
                    onDidCommit={this.onDidCommit}
                    onDidDiscard={this.onDidDiscard}
                    history={this.props.history}
                    isLightTheme={this.props.isLightTheme}
                />
            </div>
        )
    }

    private onDidCommit = (lastID: number | null, contents: string) => {
        this.setState({ commitError: undefined })
        overwriteSettings(this.props.subject.id, lastID, contents)
            .pipe(mergeMap(() => refreshConfiguration().pipe(concat([null]))))
            .subscribe(
                () => {
                    this.setState({ commitError: undefined })
                    this.subjectChanges.next({ id: this.props.subject.id }) // refresh
                },
                err => {
                    this.setState({ commitError: err })
                    console.error(err)
                }
            )
    }

    private onDidDiscard = (): void => {
        this.setState({ commitError: undefined })
    }
}
