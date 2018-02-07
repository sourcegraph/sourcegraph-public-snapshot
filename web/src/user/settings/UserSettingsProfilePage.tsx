import { Loader } from '@sourcegraph/icons/lib/Loader'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { catchError } from 'rxjs/operators/catchError'
import { concat } from 'rxjs/operators/concat'
import { filter } from 'rxjs/operators/filter'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { currentUser, refreshCurrentUser } from '../../auth'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { VALID_USERNAME_REGEXP } from '../index'
import { UserAvatar } from '../UserAvatar'
import { updateUser } from './backend'

interface Props extends RouteComponentProps<any> {}

interface State {
    user: GQL.IUser | null
    error?: Error
    loading?: boolean
    saved?: boolean
    username: string
    displayName: string
}

export class UserSettingsProfilePage extends React.Component<Props, State> {
    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        const redirectError = this.getRedirectError()
        this.state = {
            user: null,
            username: '',
            displayName: '',
            error: redirectError ? new Error(redirectError) : undefined,
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserProfile')
        this.subscriptions.add(
            currentUser.subscribe(
                user =>
                    this.setState({
                        user,
                        username: (user && user.username) || '',
                        displayName: (user && user.displayName) || '',
                    }),
                error => this.setState({ error })
            )
        )
        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(event => {
                        event.preventDefault()
                        eventLogger.log('UpdateUserClicked')
                    }),
                    filter(event => event.currentTarget.checkValidity()),
                    tap(() => this.setState({ loading: true })),
                    mergeMap(event =>
                        updateUser({
                            username: this.state.username,
                            displayName: this.state.displayName,
                        }).pipe(catchError(this.handleError))
                    ),
                    tap(() => {
                        this.setState({ loading: false, error: undefined, saved: true })
                        setTimeout(() => this.setState({ saved: false }), 500)
                    }),
                    mergeMap(() => refreshCurrentUser().pipe(concat([null])))
                )
                .subscribe(undefined, this.handleError)
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="user-settings-profile-page">
                <PageTitle title="Profile" />
                <h2>Profile</h2>
                {this.state.error && <p className="alert alert-danger">{upperFirst(this.state.error.message)}</p>}
                <form className="user-settings-profile-page__form" onSubmit={this.handleSubmit}>
                    <div className="user-settings-profile-page__avatar-row">
                        <div className="form-group">
                            <label>Username</label>
                            <input
                                type="text"
                                className="form-control"
                                value={this.state.username}
                                onChange={this.onUsernameFieldChange}
                                pattern={VALID_USERNAME_REGEXP.toString().slice(1, -1)}
                                required={true}
                                disabled={this.state.loading}
                                spellCheck={false}
                                placeholder="Username"
                            />
                            <small className="form-text">
                                A username consists of letters, numbers, hyphens (-) and may not begin or end with a
                                hyphen
                            </small>
                        </div>
                        <div className="user-settings-profile-page__avatar-column">
                            <UserAvatar />
                        </div>
                    </div>
                    <div className="form-group">
                        <label>Display name (optional)</label>
                        <input
                            type="text"
                            className="form-control"
                            value={this.state.displayName}
                            onChange={this.onDisplayNameFieldChange}
                            disabled={this.state.loading}
                            spellCheck={false}
                            placeholder="Display name"
                        />
                    </div>
                    <button
                        className="btn btn-primary user-settings-profile-page__button"
                        type="submit"
                        disabled={this.state.loading}
                    >
                        Update profile
                    </button>
                    {this.state.loading && (
                        <div className="icon-inline">
                            <Loader className="icon-inline" />
                        </div>
                    )}
                    {this.state.saved && (
                        <p className="alert alert-success user-settings-profile-page__alert">Profile saved!</p>
                    )}
                </form>
            </div>
        )
    }

    private onUsernameFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ username: e.target.value })
    }

    private onDisplayNameFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ displayName: e.target.value })
    }

    private getRedirectError(): string | undefined {
        const code = new URLSearchParams(this.props.location.search).get('error')
        if (!code) {
            return undefined
        }
        switch (code) {
            case 'err_username_exists':
                return 'The username you selected is already taken, please try again.'
        }
        return 'There was an error creating your profile, please try again.'
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
        this.submits.next(event)
    }

    private handleError = (err: Error) => {
        console.error(err)
        this.setState({ loading: false, saved: false, error: err })
        return []
    }
}
