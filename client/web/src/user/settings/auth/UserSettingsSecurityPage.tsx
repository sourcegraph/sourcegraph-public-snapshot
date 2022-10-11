import React, { useState, useEffect } from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { ErrorLike } from '@sourcegraph/common'
import { gql, useQuery } from '@sourcegraph/http-client'
import { Container, PageHeader, LoadingSpinner, Button, Link, Alert, H3, Input, Label } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { PasswordInput } from '../../../auth/SignInSignUpCommon'
import { PageTitle } from '../../../components/PageTitle'
import {
    UserAreaUserFields,
    ExternalServiceKind,
    ExternalAccountFields,
    MinExternalAccountsVariables,
    MinExternalAccountsResult,
} from '../../../graphql-operations'
import { AuthProvider, SourcegraphContext } from '../../../jscontext'
import { eventLogger } from '../../../tracking/eventLogger'
import { getPasswordRequirements } from '../../../util/security'
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

const accountsByType = (accounts: MinExternalAccount[]): ExternalAccountsByType =>
    accounts.reduce((accumulator: ExternalAccountsByType, account) => {
        accumulator[account.serviceType as ServiceType] = account
        return accumulator
    }, {})

export const UserSettingsSecurityPage: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const [oldPassword, setOldPassword] = useState<string>('')
    const [newPassword, setNewPassword] = useState<string>('')
    const [newPasswordConfirmation, setNewPasswordConfirmation] = useState<string>('')
    const [accounts, setAccounts] = useState<{ fetched?: MinExternalAccount[]; lastRemoved?: string }>({
        fetched: [],
        lastRemoved: '',
    })
    const [saved, setSaved] = useState<boolean>(false)
    const [error, setError] = useState<ErrorLike>()

    const { data, loading } = useQuery<MinExternalAccountsResult, MinExternalAccountsVariables>(
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
        {
            variables: { username: props.user.username },
            onError: (error): void => {
                handleError(error)
            },
        }
    )

    let newPasswordConfirmationField: HTMLInputElement | null = null
    const setNewPasswordConfirmationField = (element: HTMLInputElement | null): void => {
        newPasswordConfirmationField = element
    }

    // auth providers by service type
    const authProvidersByType = props.context.authProviders.reduce((accumulator: AuthProvidersByType, provider) => {
        accumulator[provider.serviceType] = provider
        return accumulator
    }, {})

    const shouldShowOldPasswordInput = (): boolean =>
        /**
         * Show old password form only when all items are true
         * 1. user has a password set
         * 2. user doesn't have external accounts
         */
        props.user.builtinAuth && accounts.fetched?.length === 0

    useEffect(() => {
        eventLogger.logPageView('UserSettingsPassword')

        setAccounts({ fetched: data?.user?.externalAccounts.nodes, lastRemoved: '' })
    }, [data])

    const onAccountRemoval = (removeId: string, name: string): void => {
        // keep every account that doesn't match removeId
        setAccounts({ fetched: accounts.fetched?.filter(({ id }) => id !== removeId), lastRemoved: name })
    }

    const onOldPasswordFieldChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
        setOldPassword(event.target.value)
    }

    const onNewPasswordFieldChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
        setNewPassword(event.target.value)
        validateForm()
    }

    const onNewPasswordConfirmationFieldChange = (event: React.ChangeEvent<HTMLInputElement>): void => {
        setNewPasswordConfirmation(event.target.value)
        validateForm()
    }

    function validateForm(): void {
        if (newPasswordConfirmationField) {
            if (newPassword === newPasswordConfirmation) {
                newPasswordConfirmationField.setCustomValidity('') // valid
            } else {
                newPasswordConfirmationField.setCustomValidity("New passwords don't match.")
            }
        }
    }

    const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        shouldShowOldPasswordInput()
            ? updatePassword({
                  oldPassword: oldPassword,
                  newPassword: newPassword,
              })
            : createPassword({
                  newPassword: newPassword,
              })
        setSaved(true)
    }

    const handleError = (error: ErrorLike): [] => {
        setError(error)
        setSaved(false)
        return []
    }

    return (
        <>
            <PageTitle title="Account security" />

            {props.authenticatedUser.id !== props.user.id && (
                <Alert variant="danger">
                    Only the user may change their password. Site admins may{' '}
                    <Link to={`/site-admin/users?query=${encodeURIComponent(props.user.username)}`}>
                        reset a user's password
                    </Link>
                    .
                </Alert>
            )}

            {accounts.lastRemoved && (
                <Alert role="alert" variant="warning">
                    Sign in connection for {accounts.lastRemoved} removed. Please set a new password for your account.
                </Alert>
            )}

            {error && <ErrorAlert className="mb-3" error={error} />}

            {saved && (
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
            {!accounts.fetched && error && (
                <div className="d-flex justify-content-center mt-4">
                    <LoadingSpinner />
                </div>
            )}

            {/* fetched external accounts */}
            {accounts.fetched && (
                <Container>
                    <ExternalAccountsSignIn
                        supported={[ExternalServiceKind.GITHUB, ExternalServiceKind.GITLAB]}
                        accounts={accountsByType(accounts.fetched)}
                        authProviders={authProvidersByType}
                        onDidError={handleError}
                        onDidRemove={onAccountRemoval}
                    />
                </Container>
            )}

            {/* fetched external accounts but user doesn't have any */}
            {accounts.fetched?.length === 0 && (
                <>
                    <hr className="my-4" />
                    <H3 className="mb-3">Password</H3>
                    <Container>
                        <Form onSubmit={handleSubmit}>
                            {/* Include a username field as a hint for password managers to update the saved password. */}
                            <Input
                                value={props.user.username}
                                name="username"
                                autoComplete="username"
                                readOnly={true}
                                hidden={true}
                            />
                            {shouldShowOldPasswordInput() && (
                                <div className="form-group">
                                    <Label htmlFor="oldPassword">Old password</Label>
                                    <PasswordInput
                                        value={oldPassword}
                                        onChange={onOldPasswordFieldChange}
                                        disabled={loading}
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
                                    value={newPassword}
                                    onChange={onNewPasswordFieldChange}
                                    disabled={loading}
                                    id="newPassword"
                                    name="newPassword"
                                    aria-label="new password"
                                    minLength={window.context.authMinPasswordLength}
                                    placeholder=" "
                                    autoComplete="new-password"
                                />
                                <small className="form-help text-muted">
                                    {getPasswordRequirements(window.context)}
                                </small>
                            </div>
                            <div className="form-group">
                                <Label htmlFor="newPasswordConfirmation">Confirm new password</Label>
                                <PasswordInput
                                    value={newPasswordConfirmation}
                                    onChange={onNewPasswordConfirmationFieldChange}
                                    disabled={loading}
                                    id="newPasswordConfirmation"
                                    name="newPasswordConfirmation"
                                    aria-label="new password confirmation"
                                    placeholder=" "
                                    minLength={window.context.authMinPasswordLength}
                                    inputRef={setNewPasswordConfirmationField}
                                    autoComplete="new-password"
                                />
                            </div>
                            <Button
                                className="user-settings-password-page__button"
                                type="submit"
                                disabled={loading}
                                variant="primary"
                            >
                                {loading && (
                                    <>
                                        <LoadingSpinner />{' '}
                                    </>
                                )}
                                {shouldShowOldPasswordInput() ? 'Update password' : 'Set password'}
                            </Button>
                        </Form>
                    </Container>
                </>
            )}
        </>
    )
}
