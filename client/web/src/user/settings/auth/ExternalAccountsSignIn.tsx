import React from 'react'

import classNames from 'classnames'
import {AuthProvider} from 'src/jscontext'

import {ErrorLike} from '@sourcegraph/common'

import {defaultExternalAccounts, ExternalAccountKind} from '../../../components/externalAccounts/externalAccounts'

import {ExternalAccount} from './ExternalAccount'
import {AccountByServiceID, UserExternalAccount} from './UserSettingsSecurityPage'
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
    Values: Map<string, Attribute>
}

interface Attribute {
    Name: string
    Values: AttributeValue[]
}

interface AttributeValue {
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

const emailAttrs = new Set(['http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress', 'emailaddress', '"http://schemas.xmlsoap.org/claims/EmailAddress'])
const usernameAttrs = new Set(['nickname', 'login', 'username', 'http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name'])

const getNormalizedAccount = (
    accounts: Partial<Record<string, UserExternalAccount>>,
    authProvider: AuthProvider
): NormalizedMinAccount => {
    // kind and type match except for the casing
    const kind = authProvider.serviceType as ExternalAccountKind

    const account = accounts[authProvider.serviceID]
    const accountExternalData = account?.accountData

    const {icon, title: name} = defaultExternalAccounts[kind]

    let normalizedAccount: NormalizedMinAccount = {
        icon,
        name,
    }

    // if external account exists - add user specific data to normalizedAccount
    if (account && accountExternalData) {
        switch (kind) {
            case 'github': {
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
            case 'gitlab': {
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
            case 'saml': {
                // The data stored in user_external_account has this structure: “Values”[“<attribute names for email or username>"] > Values[0] > “Value”: string
                let email = ''
                let username = ''

                const samlExternalData = accountExternalData as SamlExternalData
                if (samlExternalData.Values) {
                    const entries = Object.entries(samlExternalData.Values)
                    for (const [name, att] of entries) {
                        if (usernameAttrs.has(name)) {
                            username = att.Values.length > 0 && att.Values[0].Value || ''

                            break
                        }

                        if (emailAttrs.has(name)) {
                            email = att.Values.find((val: AttributeValue) =>
                                val.Value.includes('@')
                            )?.Value || ''
                        }
                    }
                }

                normalizedAccount = {
                    ...normalizedAccount,
                    external: {
                        id: account.id,
                        userName: username || email,
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
