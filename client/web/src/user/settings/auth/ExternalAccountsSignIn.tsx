import React from 'react'

import classNames from 'classnames'
import { AuthProvider } from 'src/jscontext'

import { ErrorLike } from '@sourcegraph/common'
import { ExternalAccountKind } from '@sourcegraph/shared/src/graphql-operations'

import { defaultExternalAccounts } from '../../../components/externalAccounts/externalAccounts'

import { ExternalAccount } from './ExternalAccount'
import { AccountByServiceID, UserExternalAccount } from './UserSettingsSecurityPage'

import styles from './ExternalAccountsSignIn.module.scss'

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

interface SamlExternalData {
    username: string
    Assertions: assertion[]
}

interface assertion {
    AttributeStatement: attributeStatement
}

interface attributeStatement {
    Attributes: attribute[]
}

interface attribute {
    Name: string
    Values: value[]
}

interface value {
    Value: string
}

export interface NormalizedMinAccount {
    name: string
    icon: React.ComponentType<React.PropsWithChildren<{ className?: string }>>
    // some data may be missing if account is not setup
    external?: {
        id: string
        userName: string
        userLogin?: string
        userUrl?: string
    }
}

interface Props {
    accounts: AccountByServiceID
    authProviders: AuthProvider[]
    onDidRemove: (id: string, name: string) => void
    onDidError: (error: ErrorLike) => void
}

const getNormalizedAccount = (
    accounts: Partial<Record<string, UserExternalAccount>>,
    authProvider: AuthProvider
): NormalizedMinAccount => {
    // kind and type match except for the casing
    const kind = authProvider.serviceType.toLocaleUpperCase() as ExternalAccountKind

    const account = accounts[authProvider.serviceID]
    const accountExternalData = account?.accountData

    const { icon, title: name } = defaultExternalAccounts[kind]

    let normalizedAccount: NormalizedMinAccount = {
        icon,
        name,
    }

    // if external account exists - add user specific data to normalizedAccount
    if (account && accountExternalData) {
        switch (kind) {
            case 'GITHUB':
                {
                    const githubExternalData = accountExternalData as GitHubExternalData
                    normalizedAccount = {
                        ...normalizedAccount,
                        external: {
                            id: account.id,
                            // map GitHub fields
                            userName: githubExternalData.name,
                            userLogin: githubExternalData.login,
                            userUrl: githubExternalData.html_url,
                        },
                    }
                }
                break
            case 'GITLAB':
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
            case 'SAML':
                {
                    const samlExternalData = accountExternalData as SamlExternalData
                    const data = samlExternalData.Assertions[0].AttributeStatement.Attributes
                    let userName = ''

                    for (const item of data) {
                        if (item.Values && item.Name) {
                            const key = item.Name.split('/').pop()
                            if (key === 'nickname') {
                                userName = item.Values[0].Value
                            }
                        }
                    }

                    normalizedAccount = {
                        ...normalizedAccount,
                        external: {
                            id: account.id,
                            userName,
                            userLogin: userName,
                        },
                    }
                }
                break
        }
    }

    return normalizedAccount
}

export const ExternalAccountsSignIn: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    accounts,
    authProviders,
    onDidRemove,
    onDidError,
}) => (
    <>
        {authProviders && (
            <ul className="list-group">
                {authProviders.map(authProvider => {
                    // if auth provider for this account doesn't exist -
                    // don't display the account as an option
                    if (authProvider && authProvider.serviceType !== 'builtin') {
                        const normAccount = getNormalizedAccount(accounts, authProvider)

                        return (
                            <li
                                key={authProvider.serviceID}
                                className={classNames('list-group-item', styles.externalAccount)}
                            >
                                <ExternalAccount
                                    account={normAccount}
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
