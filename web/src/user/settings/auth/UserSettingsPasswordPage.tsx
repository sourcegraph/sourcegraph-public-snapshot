import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import { catchError, filter, mergeMap, tap } from 'rxjs/operators'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PasswordInput } from '../../../auth/SignInSignUpCommon'
import { Form } from '../../../components/Form'
import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'
import { updatePassword } from '../backend'

interface Props extends RouteComponentProps<any> {
    user: GQL.IUser
    authenticatedUser: GQL.IUser
}

interface State {
    error?: Error
    loading?: boolean
    saved?: boolean
    oldPassword: string
    newPassword: string
    newPasswordConfirmation: string
}

export class UserSettingsPasswordPage extends React.Component<Props, State> {
    public state: State = {
        oldPassword: '',
        newPassword: '',
        newPasswordConfirmation: '',
    }

    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private subscriptions = new Subscription()

    private newPasswordConfirmationField: HTMLInputElement | null = null
    private setNewPasswordConfirmationField = (e: HTMLInputElement | null) => (this.newPasswordConfirmationField = e)

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSettingsPassword')
        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(event => {
                        event.preventDefault()
                        eventLogger.log('UpdatePasswordClicked')
                    }),
                    filter(event => event.currentTarget.checkValidity()),
                    tap(() => this.setState({ loading: true })),
                    mergeMap(() =>
                        updatePassword({
                            oldPassword: this.state.oldPassword,
                            newPassword: this.state.newPassword,
                        }).pipe(
                            // Change URL after updating to trigger Chrome to show "Update password?" dialog.
                            tap(() => this.props.history.replace({ hash: 'updated' })),
                            catchError(err => this.handleError(err))
                        )
                    )
                )
                .subscribe(
                    () =>
                        this.setState({
                            loading: false,
                            error: undefined,
                            oldPassword: '',
                            newPassword: '',
                            newPasswordConfirmation: '',
                            saved: true,
                        }),
                    err => this.handleError(err)
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="user-settings-password-page">
                <PageTitle title="Change password" />
                <h2>Change password</h2>
                {this.props.authenticatedUser.id !== this.props.user.id ? (
                    <div className="alert alert-danger">
                        Only the user may change their password. Site admins may{' '}
                        <Link to={`/site-admin/users?query=${encodeURIComponent(this.props.user.username)}`}>
                            reset a user's password
                        </Link>
                        .
                    </div>
                ) : (
                    <>
                        {this.state.error && (
                            <p className="alert alert-danger">{upperFirst(this.state.error.message)}</p>
                        )}
                        {this.state.saved && <p className="alert alert-success">Password changed!</p>}
                        <Form onSubmit={this.handleSubmit}>
                            {/* Include a username field as a hint for password managers to update the saved password. */}
                            <input
                                type="text"
                                value={this.props.user.username}
                                name="username"
                                autoComplete="username"
                                readOnly={true}
                                hidden={true}
                            />
                            <div className="form-group">
                                <label>Old password</label>
                                <PasswordInput
                                    value={this.state.oldPassword}
                                    onChange={this.onOldPasswordFieldChange}
                                    disabled={this.state.loading}
                                    name="oldPassword"
                                    placeholder=" "
                                    autoComplete="current-password"
                                />
                            </div>
                            <div className="form-group">
                                <label>New password</label>
                                <PasswordInput
                                    value={this.state.newPassword}
                                    onChange={this.onNewPasswordFieldChange}
                                    disabled={this.state.loading}
                                    name="newPassword"
                                    placeholder=" "
                                    autoComplete="new-password"
                                />
                            </div>
                            <div className="form-group">
                                <label>Confirm new password</label>
                                <PasswordInput
                                    value={this.state.newPasswordConfirmation}
                                    onChange={this.onNewPasswordConfirmationFieldChange}
                                    disabled={this.state.loading}
                                    name="newPasswordConfirmation"
                                    placeholder=" "
                                    inputRef={this.setNewPasswordConfirmationField}
                                    autoComplete="new-password"
                                />
                            </div>
                            <button
                                className="btn btn-primary user-settings-password-page__button"
                                type="submit"
                                disabled={this.state.loading}
                            >
                                Update password
                            </button>
                            {this.state.loading && (
                                <div className="icon-inline">
                                    <LoadingSpinner className="icon-inline" />
                                </div>
                            )}
                        </Form>
                    </>
                )}
            </div>
        )
    }

    private onOldPasswordFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ oldPassword: e.target.value })
    }

    private onNewPasswordFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ newPassword: e.target.value }, () => this.validateForm())
    }

    private onNewPasswordConfirmationFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ newPasswordConfirmation: e.target.value }, () => this.validateForm())
    }

    private validateForm(): void {
        if (this.newPasswordConfirmationField) {
            if (this.state.newPassword === this.state.newPasswordConfirmation) {
                this.newPasswordConfirmationField.setCustomValidity('') // valid
            } else {
                this.newPasswordConfirmationField.setCustomValidity("New passwords don't match.")
            }
        }
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
