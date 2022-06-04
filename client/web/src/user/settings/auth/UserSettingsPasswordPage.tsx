import * as React from 'react'

import { RouteComponentProps } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { catchError, filter, mergeMap, tap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { Button, Container, PageHeader, LoadingSpinner, Link, Alert, Input, Label } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { PasswordInput } from '../../../auth/SignInSignUpCommon'
import { PageTitle } from '../../../components/PageTitle'
import { UserAreaUserFields } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { updatePassword } from '../backend'

import styles from './UserSettingsPasswordPage.module.scss'

interface Props extends RouteComponentProps<{}> {
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

    public getPasswordRequirements(): JSX.Element {
        let requirements = ''
        const passwordPolicyReference = window.context.experimentalFeatures.passwordPolicy

        if (passwordPolicyReference && passwordPolicyReference.enabled === true) {
            if (passwordPolicyReference.minimumLength && passwordPolicyReference.minimumLength > 0) {
                requirements +=
                    'Your password must include at least ' +
                    passwordPolicyReference.minimumLength.toString() +
                    ' characters'
            }
            if (
                passwordPolicyReference.numberOfSpecialCharacters &&
                passwordPolicyReference.numberOfSpecialCharacters > 0
            ) {
                requirements +=
                    ', ' + passwordPolicyReference.numberOfSpecialCharacters.toString() + ' special characters'
            }
            if (
                passwordPolicyReference.requireAtLeastOneNumber &&
                passwordPolicyReference.requireAtLeastOneNumber === true
            ) {
                requirements += ', at least one number'
            }
            if (
                passwordPolicyReference.requireUpperandLowerCase &&
                passwordPolicyReference.requireUpperandLowerCase === true
            ) {
                requirements += ', at least one uppercase letter'
            }
        } else {
            requirements += 'At least 12 characters.'
        }

        return <small>{requirements}</small>
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
                                        minLength={
                                            window.context.experimentalFeatures.passwordPolicy?.enabled &&
                                            window.context.experimentalFeatures.passwordPolicy.minimumLength !==
                                                undefined
                                                ? window.context.experimentalFeatures.passwordPolicy.minimumLength
                                                : 12
                                        }
                                        placeholder=" "
                                        autoComplete="new-password"
                                        className={styles.userSettingsPasswordPageInput}
                                    />
                                    {this.getPasswordRequirements()}
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
                                        minLength={
                                            window.context.experimentalFeatures.passwordPolicy?.enabled &&
                                            window.context.experimentalFeatures.passwordPolicy.minimumLength !==
                                                undefined
                                                ? window.context.experimentalFeatures.passwordPolicy.minimumLength
                                                : 12
                                        }
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
        if (window.context.experimentalFeatures.passwordPolicy?.enabled) {
            if (
                window.context.experimentalFeatures.passwordPolicy.minimumLength &&
                password.length < window.context.experimentalFeatures.passwordPolicy.minimumLength
            ) {
                this.newPasswordConfirmationField?.setCustomValidity(
                    'Password must be greater than ' +
                        window.context.experimentalFeatures.passwordPolicy.minimumLength.toString() +
                        ' characters.'
                )
            }
            if (
                window.context.experimentalFeatures.passwordPolicy?.numberOfSpecialCharacters &&
                window.context.experimentalFeatures.passwordPolicy.numberOfSpecialCharacters > 0
            ) {
                const specialCharacters = /[!"#$%&'()*+,./:;<=>?@[\]^_`{|}~-]/
                const count = (password.match(specialCharacters) || []).length
                if (
                    window.context.experimentalFeatures.passwordPolicy.numberOfSpecialCharacters &&
                    count < window.context.experimentalFeatures.passwordPolicy.numberOfSpecialCharacters
                ) {
                    this.newPasswordConfirmationField?.setCustomValidity(
                        'Password must contain ' +
                            window.context.experimentalFeatures.passwordPolicy.numberOfSpecialCharacters.toString() +
                            ' special character(s).'
                    )
                }
            }

            if (
                window.context.experimentalFeatures.passwordPolicy.requireAtLeastOneNumber &&
                window.context.experimentalFeatures.passwordPolicy.requireAtLeastOneNumber
            ) {
                const validRequireAtLeastOneNumber = /\d+/
                if (password.match(validRequireAtLeastOneNumber) === null) {
                    this.newPasswordConfirmationField?.setCustomValidity('Password must contain at least one number.')
                }
            }

            if (
                window.context.experimentalFeatures.passwordPolicy.requireUpperandLowerCase &&
                window.context.experimentalFeatures.passwordPolicy.requireUpperandLowerCase
            ) {
                const validUseUpperCase = new RegExp('[A-Z]+')
                if (!validUseUpperCase.test(password)) {
                    this.newPasswordConfirmationField?.setCustomValidity(
                        'Password must contain at least one uppercase letter.'
                    )
                }
            }

            this.newPasswordConfirmationField?.setCustomValidity('')
        }

        if (password.length < 12) {
            this.newPasswordConfirmationField?.setCustomValidity('Password must be at least 12 characters.')
        }

        this.newPasswordConfirmationField?.setCustomValidity('')
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        this.submits.next(event)
    }

    private handleError = (error: Error): [] => {
        console.error(error)
        this.setState({ loading: false, saved: false, error })
        return []
    }
}
