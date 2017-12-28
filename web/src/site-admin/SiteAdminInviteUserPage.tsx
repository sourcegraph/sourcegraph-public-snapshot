import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { catchError } from 'rxjs/operators/catchError'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { EmailInput, UsernameInput } from '../auth/SignInSignUpCommon'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { createUserBySiteAdmin } from './backend'

interface Props extends RouteComponentProps<any> {}

export interface State {
    errorDescription?: string
    loading: boolean

    /**
     * The password reset URL generated for the new user account.
     */
    newUserPasswordResetURL?: string

    // Form
    username: string
    email: string
}

/**
 * A page with a form to invite a user to the site.
 */
export class SiteAdminInviteUserPage extends React.Component<Props, State> {
    public state: State = {
        loading: false,
        username: '',
        email: '',
    }

    private submits = new Subject<{ username: string; email: string }>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminInviteUser')

        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(() =>
                        this.setState({
                            loading: true,
                            errorDescription: undefined,
                        })
                    ),
                    mergeMap(({ username, email }) =>
                        createUserBySiteAdmin(username, email).pipe(
                            catchError(error => {
                                console.error(error)
                                this.setState({ loading: false, errorDescription: error.message })
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    ({ resetPasswordURL }) =>
                        this.setState({
                            loading: false,
                            errorDescription: undefined,
                            newUserPasswordResetURL: resetPasswordURL,
                        }),
                    error => console.error(error)
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-invite-user-page">
                <PageTitle title="Invite user - Admin" />
                <h2>Invite user</h2>
                <p>
                    Create a new user account and generate a password reset link. You must manually send the link to the
                    new user.
                </p>
                {this.state.newUserPasswordResetURL ? (
                    <div className="alert alert-success">
                        <p>
                            Account created for <strong>{this.state.username}</strong>. You must manually send this
                            password reset link to the new user:
                        </p>
                        <p>
                            <code className="site-admin-invite-user-page__url">
                                {this.state.newUserPasswordResetURL}
                            </code>
                        </p>
                        <button
                            className="btn btn-secondary btn-sm site-admin-invite-user-page__alert-button"
                            onClick={this.dismissAlert}
                        >
                            Invite another user
                        </button>
                    </div>
                ) : (
                    <form onSubmit={this.onSubmit} className="site-admin-invite-user-page__form">
                        <dl className="form-group">
                            <dl className="form-group">
                                <dt className="input-label">Username</dt>
                                <dd>
                                    <UsernameInput
                                        className="form-control"
                                        onChange={this.onUsernameFieldChange}
                                        value={this.state.username}
                                        required={true}
                                        disabled={this.state.loading}
                                        autoFocus={true}
                                    />
                                </dd>
                            </dl>
                            <dl className="form-group">
                                <dt className="input-label">Email</dt>
                                <dd>
                                    <EmailInput
                                        className="form-control"
                                        onChange={this.onEmailFieldChange}
                                        required={true}
                                        value={this.state.email}
                                        disabled={this.state.loading}
                                    />
                                </dd>
                            </dl>
                            {this.state.errorDescription && (
                                <div className="alert alert-danger form-group site-admin-invite-user-page__error">
                                    {this.state.errorDescription}
                                </div>
                            )}
                            <button className="btn btn-primary" disabled={this.state.loading} type="submit">
                                Generate password reset link
                            </button>
                        </dl>
                    </form>
                )}
                <hr />
                <p>
                    See <a href="https://about.sourcegraph.com/docs/server/config/">Sourcegraph documentation</a> for
                    information about configuring user accounts and SSO authentication.
                </p>
            </div>
        )
    }

    private onEmailFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ email: e.target.value, errorDescription: undefined })
    }

    private onUsernameFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ username: e.target.value, errorDescription: undefined })
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        event.stopPropagation()
        this.submits.next({ username: this.state.username, email: this.state.email })
    }

    private dismissAlert = () =>
        this.setState({
            newUserPasswordResetURL: undefined,
            errorDescription: undefined,
            username: '',
            email: '',
        })
}
