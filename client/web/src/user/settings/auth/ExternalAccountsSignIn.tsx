import React from 'react'

import classNames from 'classnames'
import { AuthProvider } from 'src/jscontext'

import { ErrorLike } from '@sourcegraph/common'

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

export interface SamlExternalData {
    Values: {
        emailaddress?: Attribute
        'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress'?: Attribute
        'http://schemas.xmlsoap.org/claims/EmailAddress'?: Attribute
        username?: Attribute
        nickname?: Attribute
        login?: Attribute
        'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name'?: Attribute
    }
}

export interface Attribute {
    Values: AttributeValue[]
}

export interface AttributeValue {
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

export function getUsernameOrEmail(data: SamlExternalData): string {
    return (
        data.Values.nickname?.Values[0]?.Value ||
        data.Values.login?.Values[0]?.Value ||
        data.Values.username?.Values[0]?.Value ||
        data.Values['http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name']?.Values[0]?.Value ||
        data.Values.emailaddress?.Values[0]?.Value ||
        data.Values['http://schemas.xmlsoap.org/claims/EmailAddress']?.Values[0]?.Value ||
        data.Values['http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress']?.Values[0]?.Value ||
        ''
    )
}

const getNormalizedAccount = (
    accounts: Partial<Record<string, UserExternalAccount>>,
    authProvider: AuthProvider
): NormalizedMinAccount | null => {
    if (
        authProvider.serviceType === 'builtin' ||
        authProvider.serviceType === 'http-header' ||
        authProvider.serviceType === 'openidconnect'
    ) {
        return null
    }

    const account = accounts[authProvider.serviceID]
    const accountExternalData = account?.accountData

    const { icon, title: name } = defaultExternalAccounts[authProvider.serviceType]

    let normalizedAccount: NormalizedMinAccount = {
        icon,
        name,
    }

    // if external account exists - add user specific data to normalizedAccount
    if (account && accountExternalData) {
        switch (authProvider.serviceType) {
            case 'github':
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
            case 'saml':
                {
                    const samlExternalData = accountExternalData as SamlExternalData
                    // In case the SAML values don't have a username, we get the user email.
                    normalizedAccount = {
                        ...normalizedAccount,
                        external: {
                            id: account.id,
                            userName: getUsernameOrEmail(samlExternalData),
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
                    const normAccount = getNormalizedAccount(accounts, authProvider)
                    if (normAccount) {
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
