import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { combineLatest, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators'
import * as GQL from '../backend/graphqlschema'
import { HeroPage } from '../components/HeroPage'
import { fetchSettings } from '../configuration/backend'
import { ErrorLike, isErrorLike } from '../util/errors'
import { SettingsPage } from './SettingsPage'

const NotFoundPage = () => <HeroPage icon={DirectionalSignIcon} title="404: Not Found" />

/** Props shared by SettingsArea and its sub-pages. */
interface SettingsAreaPageCommonProps {
    /** The subject whose settings to edit. */
    subject: Pick<GQL.ConfigurationSubject, '__typename' | 'id'>

    /**
     * The currently authenticated user, NOT (necessarily) the user who is the subject of the page.
     */
    authenticatedUser: GQL.IUser | null

    isLightTheme: boolean
}

/** Properties passed to all pages in the settings area. */
export interface SettingsAreaPageProps extends SettingsAreaPageCommonProps {
    /** The settings, or null if the subject has no settings yet. */
    settings: GQL.ISettings | null

    /** Called when the page updates the subject's settings. */
    onUpdate: () => void
}

interface Props extends SettingsAreaPageCommonProps, RouteComponentProps<{}> {
    extraHeader?: JSX.Element
}

const LOADING: 'loading' = 'loading'

interface State {
    /** The settings, null if there are no settings yet for the subject, loading, or an error. */
    settingsOrError: typeof LOADING | GQL.ISettings | null | ErrorLike
}

/**
 * A settings area with a top-level JSON editor and sub-pages for editing nested settings values.
 */
export class SettingsArea extends React.Component<Props, State> {
    public state: State = { settingsOrError: LOADING }

    private subjectChanges = new Subject<Pick<GQL.IConfigurationSubject, 'id'>>()
    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Load settings.
        this.subscriptions.add(
            combineLatest(this.subjectChanges, this.refreshRequests.pipe(startWith<void>(void 0)))
                .pipe(
                    distinctUntilChanged(),
                    switchMap(([{ id }]) =>
                        fetchSettings(id).pipe(
                            catchError(error => [error]),
                            map(c => ({ settingsOrError: c }))
                        )
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
        if (this.state.settingsOrError === LOADING) {
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

        const transferProps: SettingsAreaPageProps = {
            settings: this.state.settingsOrError,
            subject: this.props.subject,
            authenticatedUser: this.props.authenticatedUser,
            onUpdate: this.onUpdate,
            isLightTheme: this.props.isLightTheme,
        }

        return (
            <div>
                <h2>{term} settings</h2>
                {this.props.extraHeader}
                <Switch>
                    <Route
                        path={this.props.match.url}
                        key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                        exact={true}
                        // tslint:disable-next-line:jsx-no-lambda
                        render={routeComponentProps => <SettingsPage {...routeComponentProps} {...transferProps} />}
                    />
                    <Route key="hardcoded-key" component={NotFoundPage} />
                </Switch>
            </div>
        )
    }

    private onUpdate = () => this.refreshRequests.next()
}
