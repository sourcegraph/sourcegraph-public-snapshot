import * as React from 'react'

import { Subject, Subscription } from 'rxjs'
import { catchError, filter, mergeMap, tap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { ErrorLike, asError } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { Container, PageHeader, LoadingSpinner, Button, Link, Alert, H3, Input, Label } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { PasswordInput } from '../../../auth/SignInSignUpCommon'
import { requestGraphQL } from '../../../backend/graphql'
import { PageTitle } from '../../../components/PageTitle'
import {
    UserAreaUserFields,
    ExternalServiceKind,
    ExternalAccountFields,
    MinExternalAccountsVariables,
} from '../../../graphql-operations'
import { AuthProvider, SourcegraphContext } from '../../../jscontext'
import { eventLogger } from '../../../tracking/eventLogger'
import { updatePassword, createPassword } from '../backend'

import { ExternalAccountsSignIn } from './ExternalAccountsSignIn'

// pick only the fields we need
type MinExternalAccount = Pick<ExternalAccountFields, 'id' | 'serviceID' | 'serviceType' | 'accountData'>
type UserExternalAccount = UserExternalAccountsResult['user']['externalAccounts']['nodes'][0]
type ServiceType = AuthProvider['serviceType']

export type ExternalAccountsByType = Partial<Record<ServiceType, UserExternalAccount>>
export type AuthProvidersByType = Partial<Record<ServiceType, AuthProvider>>

interface UserExternalAccountsResult {
    user: {
        externalAccounts: {
            nodes: MinExternalAccount[]
        }
    }
}

interface Props {
    user: UserAreaUserFields
    authenticatedUser: AuthenticatedUser
    context: Pick<SourcegraphContext, 'authProviders'>
}

interface State {
    error?: ErrorLike
    loading?: boolean
    saved?: boolean
    accounts: { fetched?: MinExternalAccount[]; lastRemoved?: string }
    oldPassword: string
    newPassword: string
    newPasswordConfirmation: string
}

const fetchUserExternalAccountsByType = async (username: string): Promise<MinExternalAccount[]> => {
    const result = dataOrThrowErrors(
        await requestGraphQL<UserExternalAccountsResult, MinExternalAccountsVariables>(
            gql`
                query MinExternalAccounts($username: String!) {
                    user(username: $username) {
                        externalAccounts {
                            nodes {
                                id
                                serviceID
                                serviceType
                                accountData
                            }
                        }
                    }
                }
            `,
            { username }
        ).toPromise()
    )
    // if user doesn't have external accounts API will return an empty array
    return result.user.externalAccounts.nodes
}

const accountsByType = (accounts: MinExternalAccount[]): ExternalAccountsByType =>
    accounts.reduce((accumulator: ExternalAccountsByType, account) => {
        accumulator[account.serviceType as ServiceType] = account
        return accumulator
    }, {})

export class UserSettingsSecurityPage extends React.Component<Props, State> {
    public state: State = {
        oldPassword: '',
        newPassword: '',
        newPasswordConfirmation: '',
        accounts: {},
    }

    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private subscriptions = new Subscription()

    private newPasswordConfirmationField: HTMLInputElement | null = null
    private setNewPasswordConfirmationField = (element: HTMLInputElement | null): void => {
        this.newPasswordConfirmationField = element
    }

    // auth providers by service type
    private authProvidersByType = this.props.context.authProviders.reduce(
        (accumulator: AuthProvidersByType, provider) => {
            accumulator[provider.serviceType] = provider
            return accumulator
        },
        {}
    )

    private shouldShowOldPasswordInput = (): boolean =>
        /**
         * Show old password form only when all items are true
         * 1. user has a password set
         * 2. user doesn't have external accounts
         */
        this.props.user.builtinAuth && this.state.accounts.fetched?.length === 0

    private fetchAccounts = (): void => {
        fetchUserExternalAccountsByType(this.props.user.username)
            .then(accounts => {
                this.setState({ accounts: { fetched: accounts } })

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
                                (this.shouldShowOldPasswordInput()
                                    ? updatePassword({
                                          oldPassword: this.state.oldPassword,
                                          newPassword: this.state.newPassword,
                                      })
                                    : createPassword({
                                          newPassword: this.state.newPassword,
                                      })
                                ).pipe(
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
                                    accounts: {},
                                }),
                            error => this.handleError(error)
                        )
                )
            })
            .catch(error => {
                this.setState({ error: asError(error) })
            })
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSettingsPassword')
        this.fetchAccounts()
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

        return <small className="form-help text-muted">{requirements}</small>
    }

    public render(): JSX.Element | null {
        return (
            <>
                <PageTitle title="Account security" />

                {this.props.authenticatedUser.id !== this.props.user.id && (
                    <Alert variant="danger">
                        Only the user may change their password. Site admins may{' '}
                        <Link to={`/site-admin/users?query=${encodeURIComponent(this.props.user.username)}`}>
                            reset a user's password
                        </Link>
                        .
                    </Alert>
                )}

                {this.state.accounts.lastRemoved && (
                    <Alert role="alert" variant="warning">
                        Sign in connection for {this.state.accounts.lastRemoved} removed. Please set a new password for
                        your account.
                    </Alert>
                )}

                {this.state.error && <ErrorAlert className="mb-3" error={this.state.error} />}

                {this.state.saved && (
                    <Alert className="mb-3" variant="success">
                        Password changed!
                    </Alert>
                )}

                <PageHeader
                    headingElement="h2"
                    path={[{ text: 'Account security' }]}
                    description="Connect your account with a third-party login service to make signing in easier."
                    className="mb-3"
                />

                {/* external accounts not fetched yet */}
                {!this.state.accounts.fetched && this.state.error && (
                    <div className="d-flex justify-content-center mt-4">
                        <LoadingSpinner />
                    </div>
                )}

                {/* fetched external accounts */}
                {this.state.accounts.fetched && (
                    <Container>
                        <ExternalAccountsSignIn
                            supported={[ExternalServiceKind.GITHUB, ExternalServiceKind.GITLAB]}
                            accounts={accountsByType(this.state.accounts.fetched)}
                            authProviders={this.authProvidersByType}
                            onDidError={this.handleError}
                            onDidRemove={this.onAccountRemoval}
                        />
                    </Container>
                )}

                {/* fetched external accounts but user doesn't have any */}
                {this.state.accounts.fetched?.length === 0 && (
                    <>
                        <hr className="my-4" />
                        <H3 className="mb-3">Password</H3>
                        <Container>
                            <Form onSubmit={this.handleSubmit}>
                                {/* Include a username field as a hint for password managers to update the saved password. */}
                                <Input
                                    value={this.props.user.username}
                                    name="username"
                                    autoComplete="username"
                                    readOnly={true}
                                    hidden={true}
                                />
                                {this.shouldShowOldPasswordInput() && (
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
                                        />
                                    </div>
                                )}

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
                                    />
                                    {this.getPasswordRequirements()}
                                </div>
                                <div className="form-group">
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
                                    />
                                </div>
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
                                    {this.shouldShowOldPasswordInput() ? 'Update password' : 'Set password'}
                                </Button>
                            </Form>
                        </Container>
                    </>
                )}
            </>
        )
    }

    private onAccountRemoval = (removeId: string, name: string): void => {
        // keep every account that doesn't match removeId
        this.setState(previousState => ({
            accounts: {
                fetched: previousState.accounts.fetched?.filter(({ id }) => id !== removeId),
                lastRemoved: name,
            },
        }))
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
                this.newPasswordConfirmationField.setCustomValidity('') // valid
            } else {
                this.newPasswordConfirmationField.setCustomValidity("New passwords don't match.")
            }
        }
    }

    private handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        this.submits.next(event)
    }

    private handleError = (error: ErrorLike): [] => {
        this.setState({ loading: false, saved: false, error })
        return []
    }
}
