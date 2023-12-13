import * as React from 'react'

import classNames from 'classnames'
import { Subject, Subscription } from 'rxjs'
import { catchError, mergeMap, tap } from 'rxjs/operators'

import { asError, logger } from '@sourcegraph/common'
import { Button, Link, Label, H2, Text, ErrorAlert, Form } from '@sourcegraph/wildcard'

import { EmailInput, UsernameInput } from '../auth/SignInSignUpCommon'
import { PageTitle } from '../components/PageTitle'
import type { CreateUserResult } from '../graphql-operations'
import { eventLogger } from '../tracking/eventLogger'

import { createUser } from './backend'
import { AccountCreatedAlert } from './components/AccountCreatedAlert'

import styles from './SiteAdminCreateUserPage.module.scss'

interface State {
    errorDescription?: string
    loading: boolean

    /**
     * The result of creating the user.
     */
    createUserResult?: CreateUserResult['createUser']

    // Form
    username: string
    email: string
}

/**
 * A page with a form to create a user account.
 */
export class SiteAdminCreateUserPage extends React.Component<{}, State> {
    public state: State = {
        loading: false,
        username: '',
        email: '',
    }

    private submits = new Subject<{ username: string; email: string }>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        window.context.telemetryRecorder?.recordEvent('siteAdminCreateUser', 'viewed')
        eventLogger.logViewEvent('SiteAdminCreateUser')

        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(() =>
                        this.setState({
                            createUserResult: undefined,
                            loading: true,
                            errorDescription: undefined,
                        })
                    ),
                    mergeMap(({ username, email }) =>
                        createUser(username, email).pipe(
                            catchError(error => {
                                logger.error(error)
                                this.setState({
                                    createUserResult: undefined,
                                    loading: false,
                                    errorDescription: asError(error).message,
                                })
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    createUserResult =>
                        this.setState({
                            loading: false,
                            errorDescription: undefined,
                            createUserResult,
                        }),
                    error => logger.error(error)
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-create-user-page">
                <PageTitle title="Create user - Admin" />
                <H2>Create user account</H2>
                <Text>
                    Create a new user account
                    {window.context.resetPasswordEnabled
                        ? ' and generate a password reset link. If sending emails is not configured, you must manually send the link to the new user.'
                        : '. New users must authenticate using a configured authentication provider.'}
                </Text>
                <Text className="mb-4">
                    For information about configuring SSO authentication, see{' '}
                    <Link to="/help/admin/auth">User authentication</Link> in the Sourcegraph documentation.
                </Text>
                {this.state.createUserResult ? (
                    <AccountCreatedAlert
                        username={this.state.username}
                        email={this.state.email}
                        resetPasswordURL={this.state.createUserResult.resetPasswordURL}
                    >
                        <Button className="mt-2" onClick={this.dismissAlert} autoFocus={true} variant="primary">
                            Create another user
                        </Button>
                    </AccountCreatedAlert>
                ) : (
                    <Form onSubmit={this.onSubmit} className="site-admin-create-user-page__form">
                        <div className={classNames('form-group', styles.formGroup)}>
                            <Label htmlFor="site-admin-create-user-page__form-username">Username</Label>
                            <UsernameInput
                                id="site-admin-create-user-page__form-username"
                                onChange={this.onUsernameFieldChange}
                                value={this.state.username}
                                required={true}
                                disabled={this.state.loading}
                                autoFocus={true}
                            />
                            <small className="form-text text-muted">
                                A username consists of letters, numbers, hyphens (-), dots (.) and may not begin or end
                                with a dot, nor begin with a hyphen.
                            </small>
                        </div>
                        <div className={classNames('form-group', styles.formGroup)}>
                            <Label htmlFor="site-admin-create-user-page__form-email">Email</Label>
                            <EmailInput
                                id="site-admin-create-user-page__form-email"
                                onChange={this.onEmailFieldChange}
                                value={this.state.email}
                                disabled={this.state.loading}
                                aria-describedby="site-admin-create-user-page__form-email-help"
                            />
                            <small id="site-admin-create-user-page__form-email-help" className="form-text text-muted">
                                Optional email for the user{' '}
                                {window.context.emailEnabled
                                    ? 'that must be verified by the user'
                                    : 'that will be marked as verified'}
                                .
                            </small>
                        </div>
                        {this.state.errorDescription && (
                            <ErrorAlert className="my-2" error={this.state.errorDescription} />
                        )}
                        <Button disabled={this.state.loading} type="submit" variant="primary">
                            {window.context.resetPasswordEnabled
                                ? 'Create account & generate password reset link'
                                : 'Create account'}
                        </Button>
                    </Form>
                )}
            </div>
        )
    }

    private onEmailFieldChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
        this.setState({ email: event.target.value, errorDescription: undefined })
    }

    private onUsernameFieldChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
        this.setState({ username: event.target.value, errorDescription: undefined })
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        event.stopPropagation()
        this.submits.next({ username: this.state.username, email: this.state.email })
    }

    private dismissAlert = (): void =>
        this.setState({
            createUserResult: undefined,
            errorDescription: undefined,
            username: '',
            email: '',
        })
}
