import React, { useState, useEffect, useCallback } from 'react'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { ExternalAccountFields, ExternalAccountsVariables } from '../../../graphql-operations'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import { ExternalServiceKind, Scalars } from '../../../../../shared/src/graphql-operations'
import { requestGraphQL } from '../../../backend/graphql'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import { SourcegraphContext } from '../../../jscontext'
import { ExternalAccount } from './ExternalAccount'

type MinExternalAccount = Pick<ExternalAccountFields, 'id' | 'serviceID' | 'serviceType' | 'accountData'>
type UserExternalAccount = UserExternalAccountsResult['site']['externalAccounts']['nodes'][0]
type AuthProvider = SourcegraphContext['authProviders'][0]

interface UserExternalAccountsResult {
    site: {
        externalAccounts: {
            nodes: MinExternalAccount[]
        }
    }
}

interface GitHubExternalData {
    name: string
    login: string
    html_url: string
}

interface GitLabExternalData {
    name: string
    username: string
    web_url: string
}

type ServiceType = AuthProvider['serviceType']
type ExternalAccountsByType = Partial<Record<ServiceType, UserExternalAccount>>
type AuthProvidersByType = Partial<Record<ServiceType, AuthProvider>>
type Status = undefined | 'loading' | ErrorLike | ExternalAccountsByType

const isExternalAccountsByType = (status: Status): status is ExternalAccountsByType =>
    typeof status === 'object' && !isErrorLike(status)

export interface NormalizedMinAccount {
    name: string
    icon: React.ComponentType<{ className?: string }>
    // some data may be missing if account is not setup
    external?: {
        id: string
        userName: string
        userLogin: string
        userUrl: string
    }
}

interface Props {
    userID: Scalars['ID']
    kinds: ExternalServiceKind[]
    authProviders: AuthProvider[]
    onNoAccountsFetched: (show: boolean) => void
    onDidError: (error: ErrorLike) => void
}

const getNormalizedAccount = (accounts: ExternalAccountsByType, kind: ExternalServiceKind): NormalizedMinAccount => {
    const type = kind.toLocaleLowerCase() as ServiceType

    const account = accounts[type]
    const accountExternalData = account?.accountData

    // get external service icon and name as they will match external account
    const { icon, title: name } = defaultExternalServices[kind]

    let normalizedAccount: NormalizedMinAccount = {
        icon,
        name,
    }

    // if external account exists - add user specific data to normalizedAccount
    if (account && accountExternalData) {
        switch (type) {
            case 'github':
                {
                    const githubExternalData = accountExternalData as GitHubExternalData
                    normalizedAccount = {
                        ...normalizedAccount,
                        external: {
                            id: account.id,
                            // map github fields
                            userName: githubExternalData.name,
                            userLogin: githubExternalData.login,
                            userUrl: githubExternalData.html_url,
                        },
                    }
                }
                break
            case 'gitlab':
                {
                    const gitlabExternalData = accountExternalData as GitLabExternalData
                    normalizedAccount = {
                        ...normalizedAccount,
                        external: {
                            id: account.id,
                            // map gitlab fields
                            userName: gitlabExternalData.name,
                            userLogin: gitlabExternalData.username,
                            userUrl: gitlabExternalData.web_url,
                        },
                    }
                }
                break
        }
    }

    return normalizedAccount
}

export const ExternalAccountsSignIn: React.FunctionComponent<Props> = ({
    userID,
    kinds,
    authProviders,
    onDidError,
    onNoAccountsFetched,
}) => {
    const [statusOrError, setStatusOrError] = useState<Status>()

    // auth providers by service type
    const authProvidersByType = authProviders.reduce((accumulator: AuthProvidersByType, provider) => {
        accumulator[provider.serviceType] = provider
        return accumulator
    }, {})

    const fetchUserExternalAccounts = useCallback(async (): Promise<void> => {
        setStatusOrError('loading')

        const result = dataOrThrowErrors(
            await requestGraphQL<UserExternalAccountsResult, ExternalAccountsVariables>(
                gql`
                    query aaa($user: ID) {
                        site {
                            externalAccounts(user: $user) {
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
                { user: userID, first: null, serviceType: null, serviceID: null, clientID: null }
            ).toPromise()
        )

        const { nodes } = result.site.externalAccounts

        const externalAccountsByType = nodes.reduce((accumulator: ExternalAccountsByType, account) => {
            accumulator[account.serviceType as ServiceType] = account
            return accumulator
        }, {})

        setStatusOrError(externalAccountsByType)

        // let parent component know to render passwords form
        // if (nodes.length === 0) {
        //     onNoAccountsFetched(true)
        // }
    }, [userID, onNoAccountsFetched])

    const handleError = useCallback(
        (errorLike: ErrorLike): void => {
            const error = asError(errorLike)
            setStatusOrError(error)
            onDidError(error)
        },
        [onDidError]
    )

    useEffect(() => {
        fetchUserExternalAccounts().catch(handleError)
    }, [fetchUserExternalAccounts, handleError])

    return (
        <>
            {statusOrError === 'loading' && (
                <div className="d-flex justify-content-center">
                    <LoadingSpinner className="icon-inline" />
                </div>
            )}

            {/* TODO: send to parent */}
            {isErrorLike(statusOrError) && <div>{JSON.stringify(statusOrError, null, 2)}</div>}

            {isExternalAccountsByType(statusOrError) && (
                <ul className="list-group w-50 mt-3">
                    {kinds.map(kind => {
                        // because of the type guard it's known statusOrError
                        // is external accounts object
                        const externalAccountsByType = statusOrError

                        // TODO: do something about this kind/type conversion
                        const type = kind.toLocaleLowerCase() as ServiceType
                        const authProvider = authProvidersByType[type]

                        // if auth provider for this account doesn't exist -
                        // don't display the account as an option
                        if (authProvider) {
                            const account = getNormalizedAccount(externalAccountsByType, kind)

                            return (
                                <li key={kind} className="list-group-item">
                                    <ExternalAccount
                                        account={account}
                                        authProvider={authProvider}
                                        onDidRemove={fetchUserExternalAccounts}
                                        onDidError={handleError}
                                    />
                                </li>
                            )
                        }

                        return null
                    })}
                </ul>
            )}
        </>
    )
}
