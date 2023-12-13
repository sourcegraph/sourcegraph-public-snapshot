import React, { useState, useEffect } from 'react'

import type { ErrorLike } from '@sourcegraph/common'
import { useMutation, useQuery } from '@sourcegraph/http-client'
import {
    Container,
    PageHeader,
    LoadingSpinner,
    Button,
    Link,
    Alert,
    H3,
    Input,
    Label,
    Text,
    ErrorAlert,
    Form,
} from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../../auth'
import { PasswordInput } from '../../../auth/SignInSignUpCommon'
import { PageTitle } from '../../../components/PageTitle'
import type {
    CreatePasswordResult,
    CreatePasswordVariables,
    UpdatePasswordResult,
    UpdatePasswordVariables,
    UserAreaUserFields,
    UserExternalAccountFields,
    UserExternalAccountsWithAccountDataVariables,
} from '../../../graphql-operations'
import type { AuthProvider, SourcegraphContext } from '../../../jscontext'
import { eventLogger } from '../../../tracking/eventLogger'
import { getPasswordRequirements } from '../../../util/security'
import { CREATE_PASSWORD, USER_EXTERNAL_ACCOUNTS, UPDATE_PASSWORD } from '../backend'

import { ExternalAccountsSignIn } from './ExternalAccountsSignIn'

// pick only the fields we need
export type UserExternalAccount = Pick<
    UserExternalAccountFields,
    'id' | 'serviceID' | 'serviceType' | 'publicAccountData' | 'clientID'
>
type ServiceType = AuthProvider['serviceType']

export type ExternalAccountsByType = Partial<Record<ServiceType, UserExternalAccount>>
export type AuthProvidersByBaseURL = Partial<Record<string, AuthProvider>>
export type AccountsByServiceID = Partial<Record<string, UserExternalAccount[]>>

interface UserExternalAccountsResult {
    user: {
        externalAccounts: {
            nodes: UserExternalAccount[]
        }
    }
}

interface Props {
    user: UserAreaUserFields
    authenticatedUser: AuthenticatedUser
    context: Pick<SourcegraphContext, 'authProviders'>
}

export const UserSettingsSecurityPage: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const [oldPassword, setOldPassword] = useState<string>('')
    const [newPassword, setNewPassword] = useState<string>('')
    const [newPasswordConfirmation, setNewPasswordConfirmation] = useState<string>('')
    const [accounts, setAccounts] = useState<{ fetched?: UserExternalAccount[]; lastRemoved?: string }>({
        fetched: [],
        lastRemoved: '',
    })
    const [saved, setSaved] = useState<boolean>(false)
    const [error, setError] = useState<ErrorLike>()

    const handleError = (error: ErrorLike): [] => {
        setError(error)
        setSaved(false)
        return []
    }

    const { data, loading, refetch } = useQuery<
        UserExternalAccountsResult,
        UserExternalAccountsWithAccountDataVariables
    >(USER_EXTERNAL_ACCOUNTS, {
        variables: { username: props.user.username },
        onError: handleError,
    })

    let newPasswordConfirmationField: HTMLInputElement | null = null
    const setNewPasswordConfirmationField = (element: HTMLInputElement | null): void => {
        newPasswordConfirmationField = element
    }

    // auth providers by service ID
    const accountsByServiceID = accounts.fetched?.reduce((accumulator: AccountsByServiceID, account) => {
        const accountArray = accumulator[account.serviceID] ?? []
        accountArray.push(account)
        accumulator[account.serviceID] = accountArray

        return accumulator
    }, {})

    useEffect(() => {
        window.context.telemetryRecorder?.recordEvent('userSettingsPassword', 'viewed')
        eventLogger.logPageView('UserSettingsPassword')

        setAccounts({ fetched: data?.user?.externalAccounts.nodes, lastRemoved: '' })
    }, [data, window.context.telemetryRecorder])

    const onAccountRemoval = (removeId: string, name: string): void => {
        // keep every account that doesn't match removeId
        setAccounts({ fetched: accounts.fetched?.filter(({ id }) => id !== removeId), lastRemoved: name })
    }

    const onAccountAdd = (): void => {
        refetch({ username: props.user.username })
            .then(() => {})
            .catch(handleError)
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

    const [updatePassword] = useMutation<UpdatePasswordResult, UpdatePasswordVariables>(UPDATE_PASSWORD, {
        variables: {
            oldPassword,
            newPassword,
        },
        onError: handleError,
    })

    const [createPassword] = useMutation<CreatePasswordResult, CreatePasswordVariables>(CREATE_PASSWORD, {
        variables: {
            newPassword,
        },
        onError: handleError,
    })

    const handleSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        if (props.user.builtinAuth) {
            updatePassword().catch(handleError)
        } else {
            createPassword().catch(handleError)
        }
        setSaved(true)
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
                className="mb-3 user-settings-account-security-page"
            />

            {/* external accounts not fetched yet */}
            {!accounts.fetched && error && (
                <div className="d-flex justify-content-center mt-4">
                    <LoadingSpinner />
                </div>
            )}

            {/* fetched external accounts */}
            {accountsByServiceID && (
                <Container>
                    <ExternalAccountsSignIn
                        accounts={accountsByServiceID}
                        authProviders={props.context.authProviders}
                        onDidError={handleError}
                        onDidRemove={onAccountRemoval}
                        onDidAdd={onAccountAdd}
                    />
                </Container>
            )}

            {/* Only display password creation/update if builtin auth is enabled */}
            {props.context.authProviders.some(provider => provider.isBuiltin) && (
                <>
                    <hr className="my-4" />
                    <H3 className="mb-3">{props.user.builtinAuth ? 'Update ' : 'Create '}Password</H3>
                    {props.user.builtinAuth ? (
                        <Text>Change your account password.</Text>
                    ) : (
                        <Text>Create a password to enable sign-in using a username/password combination.</Text>
                    )}
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
                            {props.user.builtinAuth && (
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
                                {props.user.builtinAuth ? 'Update password' : 'Create password'}
                            </Button>
                        </Form>
                    </Container>
                </>
            )}
        </>
    )
}
