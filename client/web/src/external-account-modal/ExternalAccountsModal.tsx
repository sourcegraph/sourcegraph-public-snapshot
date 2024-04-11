import { useEffect, useState } from 'react'

import type { AuthenticatedUser } from 'src/auth'
import { BrandLogo } from 'src/components/branding/BrandLogo'
import type { UserExternalAccountsWithAccountDataVariables } from 'src/graphql-operations'
import type { AuthProvider, SourcegraphContext } from 'src/jscontext'
import { ExternalAccountsSignIn } from 'src/user/settings/auth/ExternalAccountsSignIn'
import type { UserExternalAccount, UserExternalAccountsResult } from 'src/user/settings/auth/UserSettingsSecurityPage'
import { USER_EXTERNAL_ACCOUNTS } from 'src/user/settings/backend'

import type { ErrorLike } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import { Button, ErrorAlert, H2, LoadingSpinner, Modal, Text } from '@sourcegraph/wildcard'

import styles from './ExternalAccountsModal.module.scss'

export interface ExternalAccountsModalProps {
    authenticatedUser: AuthenticatedUser
    context: Pick<SourcegraphContext, 'authProviders'>
}

const EXTERNAL_ACCOUNTS_DISMISSED_KEY = 'user.externalaccounts.seen'

// getSeenAuthProviders returns a list of auth providers that have already
// been seen by the user.
export const getSeenAuthProviders = (): AuthProvider[] => {
    const seenAuthProviders = localStorage.getItem(EXTERNAL_ACCOUNTS_DISMISSED_KEY)
    if (seenAuthProviders) {
        try {
            return JSON.parse(seenAuthProviders)
        } catch {
            localStorage.removeItem(EXTERNAL_ACCOUNTS_DISMISSED_KEY)
            return []
        }
    }
    return []
}

// setSeenAuthProviders sets a list of auth providers that have already
// been seen by the user in local storage.
export const setSeenAuthProviders = (authProviders: AuthProvider[]): void => {
    localStorage.setItem(EXTERNAL_ACCOUNTS_DISMISSED_KEY, JSON.stringify(authProviders))
}

// shouldShowExternalAccountsModal compares a list of active auth providers
// with a list of auth providers that have already been seen by the user.
// If the active auth providers contain an auth provider that has not yet
// been seen by the user, true is returned. Otherwise false is returned.
export const shouldShowExternalAccountsModal = (
    activeAuthProviders: AuthProvider[],
    seenAuthProviders: AuthProvider[]
): boolean => {
    for (const activeProvider of activeAuthProviders) {
        // Skip the builtin provider
        if (activeProvider.isBuiltin) {
            continue
        }

        if (
            !seenAuthProviders.find(
                seenProvider =>
                    seenProvider.serviceType === activeProvider.serviceType &&
                    seenProvider.clientID === activeProvider.clientID
            )
        ) {
            return true
        }
    }

    return false
}

export const ExternalAccountsModal: React.FunctionComponent<ExternalAccountsModalProps> = props => {
    const [accounts, setAccounts] = useState<{ fetched?: UserExternalAccount[]; lastRemoved?: string }>({
        fetched: [],
        lastRemoved: '',
    })

    const { data, loading, refetch } = useQuery<
        UserExternalAccountsResult,
        UserExternalAccountsWithAccountDataVariables
    >(USER_EXTERNAL_ACCOUNTS, {
        variables: { username: props.authenticatedUser.username },
    })

    const [error, setError] = useState<ErrorLike>()

    const handleError = (error: ErrorLike): [] => {
        setError(error)
        return []
    }

    useEffect(() => {
        setAccounts({ fetched: data?.user?.externalAccounts.nodes, lastRemoved: '' })
    }, [data])

    const onAccountRemoval = (removeId: string, name: string): void => {
        // keep every account that doesn't match removeId
        setAccounts({ fetched: accounts.fetched?.filter(({ id }) => id !== removeId), lastRemoved: name })
    }

    const onAccountAdd = (): void => {
        refetch({ username: props.authenticatedUser.username })
            .then(() => {})
            .catch(handleError)
    }

    const [isModalOpen, setIsModalOpen] = useState(true)

    const onDismiss = (): void => {
        if (confirm('You can always review your external account connections in your user settings.')) {
            setSeenAuthProviders(props.context.authProviders)
            setIsModalOpen(false)
        }
    }

    return (
        <Modal
            aria-label="Connect your external accounts"
            isOpen={isModalOpen}
            onDismiss={onDismiss}
            className={styles.modal}
            position="center"
        >
            <div className={styles.title}>
                <BrandLogo variant="symbol" isLightTheme={true} />
                <div>
                    <H2>Sourcegraph setup: permissions & security</H2>
                    <Text>Connect external identity providers to access private repositories.</Text>
                </div>
            </div>
            <hr />
            {loading && <LoadingSpinner />}
            {error && <ErrorAlert className="mb-3" error={error} />}
            {accounts.fetched && (
                <ExternalAccountsSignIn
                    onDidAdd={onAccountAdd}
                    onDidError={handleError}
                    onDidRemove={onAccountRemoval}
                    accounts={accounts.fetched}
                    authProviders={props.context.authProviders}
                />
            )}
            <hr />
            <Button onClick={onDismiss} className={styles.skip} size="lg" variant="secondary" outline={true}>
                Done
            </Button>
        </Modal>
    )
}
