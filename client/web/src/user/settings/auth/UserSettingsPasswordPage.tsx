import * as React from 'react'

import { Subject, Subscription } from 'rxjs'
import { catchError, filter, mergeMap, tap } from 'rxjs/operators'

import { logger } from '@sourcegraph/common'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { EVENT_LOGGER } from '@sourcegraph/shared/src/telemetry/web/eventLogger'
import {
    Button,
    Container,
    PageHeader,
    LoadingSpinner,
    Link,
    Alert,
    Input,
    Label,
    ErrorAlert,
    Form,
} from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../../auth'
import { PasswordInput } from '../../../auth/SignInSignUpCommon'
import { PageTitle } from '../../../components/PageTitle'
import type { UserAreaUserFields } from '../../../graphql-operations'
import { validatePassword, getPasswordRequirements } from '../../../util/security'
import { updatePassword } from '../backend'

import styles from './UserSettingsPasswordPage.module.scss'

interface Props extends TelemetryV2Props {
    user: UserAreaUserFields
    authenticatedUser: AuthenticatedUser
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
    private setNewPasswordConfirmationField = (element: HTMLInputElement | null): void => {
        this.newPasswordConfirmationField = element
    }

    public componentDidMount(): void {
        EVENT_LOGGER.logViewEvent('UserSettingsPassword')
        this.props.telemetryRecorder.recordEvent('settings.password', 'view')

        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(event => {
                        event.preventDefault()
                        EVENT_LOGGER.log('UpdatePasswordClicked')
                        this.props.telemetryRecorder.recordEvent('settings.password', 'update')
                    }),
                    filter(event => event.currentTarget.checkValidity()),
                    tap(() => this.setState({ loading: true })),
                    mergeMap(() =>
                        updatePassword({
                            args: {
                                oldPassword: this.state.oldPassword,
                                newPassword: this.state.newPassword,
                            },
                            telemetryRecorder: this.props.telemetryRecorder,
                        }).pipe(
                            // Sign the user out after their password is changed.
                            // We do this because the backend will no longer accept their current session
                            // and failing to sign them out will leave them in a confusing state
                            tap(() => (window.location.href = '/-/sign-out')),
                            catchError(error => this.handleError(error))
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
                    error => this.handleError(error)
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
                <PageHeader headingElement="h2" path={[{ text: 'Change password' }]} className="mb-3" />
                {this.props.authenticatedUser.id !== this.props.user.id ? (
                    <Alert variant="danger">
                        Only the user may change their password. Site admins may{' '}
                        <Link to={`/site-admin/users?query=${encodeURIComponent(this.props.user.username)}`}>
                            reset a user's password
                        </Link>
                        .
                    </Alert>
                ) : (
                    <>
                        {this.state.error && <ErrorAlert className="mb-3" error={this.state.error} />}
                        {this.state.saved && (
                            <Alert className="mb-3" variant="success">
                                Password changed!
                            </Alert>
                        )}
                        <Form onSubmit={this.handleSubmit}>
                            <Container className="mb-3">
                                {/* Include a username field as a hint for password managers to update the saved password. */}
                                <Input
                                    value={this.props.user.username}
                                    name="username"
                                    autoComplete="username"
                                    readOnly={true}
                                    hidden={true}
                                />
                                <div className="form-group">
                                    <Label htmlFor="oldPassword">Old password</Label>
                                    <PasswordInput
                                        value={this.state.oldPassword}
                                        onChange={this.onOldPasswordFieldChange}
                                        disabled={this.state.loading}
                                        id="oldPassword"
                                        name="oldPassword"
                                        aria-label="old password"
                                        placeholder=" "
                                        autoComplete="current-password"
                                        className={styles.userSettingsPasswordPageInput}
                                    />
                                </div>

                                <div className="form-group">
                                    <Label htmlFor="newPassword">New password</Label>
                                    <PasswordInput
                                        value={this.state.newPassword}
                                        onChange={this.onNewPasswordFieldChange}
                                        disabled={this.state.loading}
                                        id="newPassword"
                                        name="newPassword"
                                        aria-label="new password"
                                        minLength={window.context.authMinPasswordLength}
                                        placeholder=" "
                                        autoComplete="new-password"
                                        className={styles.userSettingsPasswordPageInput}
                                    />
                                    <small>{getPasswordRequirements(window.context)}</small>
                                </div>
                                <div className="form-group mb-0">
                                    <Label htmlFor="newPasswordConfirmation">Confirm new password</Label>
                                    <PasswordInput
                                        value={this.state.newPasswordConfirmation}
                                        onChange={this.onNewPasswordConfirmationFieldChange}
                                        disabled={this.state.loading}
                                        id="newPasswordConfirmation"
                                        name="newPasswordConfirmation"
                                        aria-label="new password confirmation"
                                        placeholder=" "
                                        minLength={window.context.authMinPasswordLength}
                                        inputRef={this.setNewPasswordConfirmationField}
                                        autoComplete="new-password"
                                        className={styles.userSettingsPasswordPageInput}
                                    />
                                </div>
                            </Container>
                            <Button
                                className="user-settings-password-page__button"
                                type="submit"
                                disabled={this.state.loading}
                                variant="primary"
                            >
                                {this.state.loading && (
                                    <>
                                        <LoadingSpinner />{' '}
                                    </>
                                )}
                                Update password
                            </Button>
                        </Form>
                    </>
                )}
            </div>
        )
    }

    private onOldPasswordFieldChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
        this.setState({ oldPassword: event.target.value })
    }

    private onNewPasswordFieldChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
        this.setState({ newPassword: event.target.value }, () => this.validateForm())
    }

    private onNewPasswordConfirmationFieldChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
        this.setState({ newPasswordConfirmation: event.target.value }, () => this.validateForm())
    }

    private validateForm(): void {
        if (this.newPasswordConfirmationField) {
            if (this.state.newPassword === this.state.newPasswordConfirmation) {
                this.validatePassword(this.state.newPassword)
            } else {
                this.newPasswordConfirmationField.setCustomValidity("New passwords don't match.")
            }
        }
    }

    private validatePassword(password: string): void {
        const message = validatePassword(window.context, password)

        if (message !== undefined) {
            this.newPasswordConfirmationField?.setCustomValidity(message)
        } else {
            this.newPasswordConfirmationField?.setCustomValidity('')
        }
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        this.submits.next(event)
    }

    private handleError = (error: Error): [] => {
        logger.error(error)
        this.setState({ loading: false, saved: false, error })
        return []
    }
}
