import { useEffect, useState } from 'react'

import type { ErrorLike } from '@sourcegraph/common'
import { useQuery } from '@sourcegraph/http-client'
import type { SeenAuthProvider } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { Button, ErrorAlert, H2, LoadingSpinner, Modal, Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../auth'
import { BrandLogo } from '../components/branding/BrandLogo'
import type {
    AuthzProvidersResult,
    AuthzProvidersVariables,
    UserExternalAccountsWithAccountDataVariables,
} from '../graphql-operations'
import type { AuthProvider, SourcegraphContext } from '../jscontext'
import { ExternalAccountsSignIn } from '../user/settings/auth/ExternalAccountsSignIn'
import type { UserExternalAccount, UserExternalAccountsResult } from '../user/settings/auth/UserSettingsSecurityPage'
import { USER_EXTERNAL_ACCOUNTS } from '../user/settings/backend'

import { AUTHZ_PROVIDERS } from './backend'

import styles from './ExternalAccountsModal.module.scss'

export interface ExternalAccountsModalProps {
    authenticatedUser: AuthenticatedUser
    isLightTheme: boolean
    setSeenAuthProvidersFunc: (seenAuthProviders: SeenAuthProvider[]) => void
    context: Pick<SourcegraphContext, 'authProviders'>
}

// shouldShowExternalAccountsModal compares a list of active auth providers
// with a list of auth providers that have already been seen by the user.
// If the active auth providers contain an auth provider that has not yet
// been seen by the user, true is returned. Otherwise false is returned.
export const shouldShowExternalAccountsModal = (
    activeAuthProviders: AuthProvider[],
    seenAuthProviders: SeenAuthProvider[] | undefined
): boolean => {
    for (const activeProvider of activeAuthProviders) {
        // Skip the builtin provider
        if (activeProvider.isBuiltin) {
            continue
        }

        // Do this after checking for isBuiltin
        if (!seenAuthProviders) {
            return true
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

function filterAuthProviders(
    authProviders: AuthProvider[],
    authzProviders: AuthzProvidersResult['authzProviders'],
    userExternalAccounts: UserExternalAccount[]
): AuthProvider[] {
    const filteredProviders = authProviders.filter(provider => {
        if (
            authzProviders.find(
                authzProvider =>
                    authzProvider.serviceID === provider.serviceID && authzProvider.serviceType === provider.serviceType
            ) &&
            !userExternalAccounts.find(
                userExternalAccount =>
                    userExternalAccount.serviceType === provider.serviceType &&
                    userExternalAccount.serviceID === provider.serviceID
            )
        ) {
            return true
        }

        return false
    })

    return filteredProviders
}

export const ExternalAccountsModal: React.FunctionComponent<ExternalAccountsModalProps> = props => {
    const [userExternalAccounts, setUserExternalAccounts] = useState<{
        fetched?: UserExternalAccount[]
        lastRemoved?: string
    }>({
        fetched: [],
        lastRemoved: '',
    })

    const [authzProviders, setAuthzProviders] = useState<AuthProvider[]>([])

    const {
        data: userAccountsData,
        loading: userAccountsLoading,
        refetch: userAccountsRefetch,
    } = useQuery<UserExternalAccountsResult, UserExternalAccountsWithAccountDataVariables>(USER_EXTERNAL_ACCOUNTS, {
        variables: { username: props.authenticatedUser.username },
    })

    const { data: authzProvidersData } = useQuery<AuthzProvidersResult, AuthzProvidersVariables>(AUTHZ_PROVIDERS, {})

    const [error, setError] = useState<ErrorLike>()

    const handleError = (error: ErrorLike): [] => {
        setError(error)
        return []
    }

    useEffect(() => {
        setUserExternalAccounts({ fetched: userAccountsData?.user?.externalAccounts.nodes, lastRemoved: '' })
    }, [userAccountsData])

    const [isModalOpen, setIsModalOpen] = useState(false)

    useEffect(() => {
        if (authzProvidersData?.authzProviders && userExternalAccounts.fetched) {
            const filteredProviders = filterAuthProviders(
                props.context.authProviders,
                authzProvidersData.authzProviders,
                userExternalAccounts.fetched
            )
            setAuthzProviders(filteredProviders)
            if (filteredProviders.length > 0) {
                setIsModalOpen(true)
            }
        }
    }, [authzProvidersData, props.context.authProviders, userExternalAccounts])

    const onAccountRemoval = (removeId: string, name: string): void => {
        // keep every account that doesn't match removeId
        setUserExternalAccounts({
            fetched: userExternalAccounts.fetched?.filter(({ id }) => id !== removeId),
            lastRemoved: name,
        })
    }

    const onAccountAdd = (): void => {
        userAccountsRefetch({ username: props.authenticatedUser.username })
            .then(() => {})
            .catch(handleError)
    }

    const onDismiss = (): void => {
        if (confirm('You can always review your external account connections in your user settings.')) {
            props.setSeenAuthProvidersFunc(props.context.authProviders)
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
                <BrandLogo variant="symbol" isLightTheme={props.isLightTheme} disableSymbolSpin={true} />
                <div>
                    <H2>Sourcegraph setup: permissions & security</H2>
                    <Text>Connect external identity providers to access private repositories.</Text>
                </div>
            </div>
            <hr />
            {userAccountsLoading && <LoadingSpinner />}
            {error && <ErrorAlert className="mb-3" error={error} />}
            {userExternalAccounts.fetched && (
                <ExternalAccountsSignIn
                    onDidAdd={onAccountAdd}
                    onDidError={handleError}
                    onDidRemove={onAccountRemoval}
                    accounts={userExternalAccounts.fetched}
                    authProviders={authzProviders}
                />
            )}
            <hr />
            <Button onClick={onDismiss} className={styles.skip} size="lg" variant="secondary" outline={true}>
                Done
            </Button>
        </Modal>
    )
}
