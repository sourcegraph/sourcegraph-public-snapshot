import classNames from 'classnames'
import React from 'react'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'
import { ErrorLike } from '@sourcegraph/shared/src/util/errors'

import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import { AuthProvider } from '../../../jscontext'

import { ExternalAccount } from './ExternalAccount'
import styles from './ExternalAccountsSignIn.module.scss'
import { ExternalAccountsByType, AuthProvidersByType } from './UserSettingsSecurityPage'

type ServiceType = AuthProvider['serviceType']

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
    supported: ExternalServiceKind[]
    accounts: ExternalAccountsByType
    authProviders: AuthProvidersByType
    onDidRemove: (id: string, name: string) => void
    onDidError: (error: ErrorLike) => void
}

const getNormalizedAccount = (accounts: ExternalAccountsByType, kind: ExternalServiceKind): NormalizedMinAccount => {
    // kind and type match except for the casing
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
    supported,
    accounts,
    authProviders,
    onDidRemove,
    onDidError,
}) => (
    <>
        {accounts && (
            <ul className="list-group">
                {supported.map(kind => {
                    const type = kind.toLocaleLowerCase() as ServiceType
                    const authProvider = authProviders[type]

                    // if auth provider for this account doesn't exist -
                    // don't display the account as an option
                    if (authProvider) {
                        const account = getNormalizedAccount(accounts, kind)

                        return (
                            <li key={kind} className={classNames('list-group-item', styles.externalAccount)}>
                                <ExternalAccount
                                    account={account}
                                    authProvider={authProvider}
                                    onDidRemove={onDidRemove}
                                    onDidError={onDidError}
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
