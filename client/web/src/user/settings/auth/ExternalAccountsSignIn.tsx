import React from 'react'

import classNames from 'classnames'
import type { AuthProvider } from 'src/jscontext'

import type { ErrorLike } from '@sourcegraph/common'

import { defaultExternalAccounts } from '../../../components/externalAccounts/externalAccounts'

import { ExternalAccount } from './ExternalAccount'
import type { AccountsByServiceID, UserExternalAccount } from './UserSettingsSecurityPage'

import styles from './ExternalAccountsSignIn.module.scss'

export interface NormalizedExternalAccount {
    name: string
    icon: React.ComponentType<React.PropsWithChildren<{ className?: string }>>
    authProvider: AuthProvider
    // some data may be missing if account is not setup
    external?: UserExternalAccount['publicAccountData'] & {
        id: string
    }
}

interface Props {
    accounts: AccountsByServiceID
    authProviders: AuthProvider[]
    onDidRemove: (id: string, name: string) => void
    onDidError: (error: ErrorLike) => void
    onDidAdd: () => void
}

const getNormalizedAccounts = (
    accounts: Partial<Record<string, UserExternalAccount[]>>,
    authProvider: AuthProvider
): NormalizedExternalAccount[] => {
    if (
        authProvider.serviceType === 'builtin' ||
        authProvider.serviceType === 'http-header' ||
        authProvider.serviceType === 'sourcegraph-operator'
    ) {
        return []
    }

    const { icon, title: name } = defaultExternalAccounts[authProvider.serviceType]

    let normalizedAccounts: NormalizedExternalAccount[] = []
    const providerAccounts = accounts[authProvider.serviceID]?.filter(acc => acc.clientID === authProvider.clientID)
    for (const providerAccount of providerAccounts || []) {
        let normalizedAccount: NormalizedExternalAccount = {
            icon,
            name,
            authProvider,
        }

        if (providerAccount?.publicAccountData) {
            normalizedAccount.external = {
                id: providerAccount.id,
                ...providerAccount.publicAccountData,
            }
        }

        normalizedAccounts.push(normalizedAccount)
    }

    if (normalizedAccounts.length === 0) {
        return [
            {
                icon,
                name,
                authProvider,
            },
        ]
    }

    return normalizedAccounts
}

export const ExternalAccountsSignIn: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    accounts,
    authProviders,
    onDidRemove,
    onDidError,
    onDidAdd,
}) => {
    const accountGroups = authProviders.map(authProvider => {
        return {
            authProvider,
            accounts: getNormalizedAccounts(accounts, authProvider),
        }
    })
    const accountsList = accountGroups
        .flatMap(group => group.accounts)
        .map(account => (
            <li
                key={account.external ? account.external.id : account.authProvider.serviceID}
                className={classNames('list-group-item', styles.externalAccount)}
            >
                <ExternalAccount
                    account={account}
                    authProvider={account.authProvider}
                    onDidRemove={onDidRemove}
                    onDidError={onDidError}
                    onDidAdd={onDidAdd}
                />
            </li>
        ))
    return <>{authProviders && <ul className="list-group">{accountsList}</ul>}</>
}
