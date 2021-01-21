import React, { useState, useEffect } from 'react'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import {
    ExternalAccountFields,
    ExternalAccountsVariables,
    DeleteExternalAccountResult,
    DeleteExternalAccountVariables,
} from '../../../graphql-operations'

import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'

import { Link } from '../../../../../shared/src/components/Link'

import { requestGraphQL } from '../../../backend/graphql'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import { defaultExternalServices } from '../../../components/externalServices/externalServices'

import { ExternalServiceKind, Scalars, Maybe } from '../../../../../shared/src/graphql-operations'
import { SourcegraphContext } from '../../../jscontext'

const deleteUserExternalAccount = async (externalAccount: Scalars['ID']): Promise<void> => {
    dataOrThrowErrors(
        await requestGraphQL<DeleteExternalAccountResult, DeleteExternalAccountVariables>(
            gql`
                mutation DeleteExternalAccount($externalAccount: ID!) {
                    deleteExternalAccount(externalAccount: $externalAccount) {
                        alwaysNil
                    }
                }
            `,
            { externalAccount }
        ).toPromise()
    )
}

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

interface Props {
    userID: Scalars['ID']
    kinds: ExternalServiceKind[]
    authProviders: AuthProvider[]
    onDidError: (error: ErrorLike) => void
}

type ServiceType = AuthProvider['serviceType']
type ExternalAccountsByType = Partial<Record<ServiceType, UserExternalAccount>>
type AuthProvidersByType = Partial<Record<ServiceType, AuthProvider>>
type Status = undefined | 'loading' | ErrorLike | ExternalAccountsByType

// TODO: narrow more
const isExternalAccountsByType = (status: Status): status is ExternalAccountsByType =>
    typeof status === 'object' && !isErrorLike(status)

export const CodeHostsSignIn: React.FunctionComponent<Props> = ({ userID, kinds, authProviders, onDidError }) => {
    const [statusOrError, setStatusOrError] = useState<Status>()

    // auth providers by service type
    const authProvidersByType = authProviders.reduce((accumulator: AuthProvidersByType, provider) => {
        accumulator[provider.serviceType] = provider
        return accumulator
    }, {})

    const fetchUserExternalAccounts = async (userID: Scalars['ID']): Promise<void> => {
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
    }

    useEffect(() => {
        fetchUserExternalAccounts(userID).catch(error => {
            // TODO: send to parent
            setStatusOrError(asError(error))
        })
    }, [userID])

    interface NormalizedMinExternalAccount {
        name: string
        login: string
        url: string
    }

    const getNormalizedAccount = (
        accounts: ExternalAccountsByType,
        type: ServiceType
    ): Maybe<NormalizedMinExternalAccount> => {
        const accountData = accounts[type]?.accountData

        if (!accountData) {
            return null
        }

        switch (type) {
            case 'github': {
                return {
                    url: accountData.html_url,
                    name: accountData.name,
                    login: accountData.login,
                }
            }
            case 'gitlab':
            default: {
                return null
            }
        }
    }

    return (
        <>
            {statusOrError === 'loading' && (
                <div className="d-flex justify-content-center">
                    <LoadingSpinner className="icon-inline" />
                </div>
            )}

            {isExternalAccountsByType(statusOrError) && (
                <ul className="list-group w-50 mt-3">
                    {kinds.map(kind => {
                        const { icon: ServiceIcon, title } = defaultExternalServices[kind]
                        // we know it's external accounts object now, "renaming"
                        const externalAccountsByType = statusOrError
                        const type = kind.toLocaleLowerCase() as ServiceType
                        const account = getNormalizedAccount(externalAccountsByType, type)
                        debugger

                        return (
                            <li key={kind} className="list-group-item">
                                <div className="p-2 d-flex align-items-start ">
                                    <div className="align-self-center">
                                        <ServiceIcon className="mb-0 mr-2" />
                                    </div>
                                    <div className="flex-1 flex-column">
                                        <h3 className="m-0">{title}</h3>
                                        <div className="text-muted">
                                            {account ? (
                                                <>
                                                    {account.name} (
                                                    <Link to={account.url} target="_blank" rel="noopener noreferrer">
                                                        @{account.login}
                                                    </Link>
                                                    )
                                                </>
                                            ) : (
                                                'Not connected'
                                            )}
                                        </div>
                                    </div>
                                    <div className="align-self-center">
                                        {account ? (
                                            <button
                                                type="button"
                                                className="btn btn-link text-danger px-0"
                                                onClick={() => {}}
                                            >
                                                Remove
                                            </button>
                                        ) : (
                                            <a
                                                href={authProvidersByType[type]?.authenticationURL}
                                                className="btn btn-secondary btn-block"
                                                target="_blank"
                                                rel="noopener noreferrer"
                                            >
                                                Add
                                            </a>
                                        )}
                                    </div>
                                </div>
                            </li>
                        )
                    })}
                </ul>
            )}
        </>
    )
}
