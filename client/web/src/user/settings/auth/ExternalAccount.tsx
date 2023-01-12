import React, { useState, useCallback } from 'react'

import { ErrorLike } from '@sourcegraph/common'
import { Button, Link, H3 } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../components/LoaderButton'
import { AuthProvider } from '../../../jscontext'

import type { NormalizedExternalAccount } from './ExternalAccountsSignIn'
import { RemoveExternalAccountModal } from './RemoveExternalAccountModal'

interface Props {
    account: NormalizedExternalAccount
    authProvider: AuthProvider
    onDidRemove: (id: string, name: string) => void
    onDidError: (error: ErrorLike) => void
}

export const ExternalAccount: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    account,
    authProvider,
    onDidRemove,
    onDidError,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const [isRemoveAccountModalOpen, setIsRemoveAccountModalOpen] = useState(false)
    const toggleRemoveAccountModal = useCallback(
        () => setIsRemoveAccountModalOpen(!isRemoveAccountModalOpen),
        [isRemoveAccountModalOpen]
    )

    const navigateToAuthProvider = useCallback((): void => {
        setIsLoading(true)

        if (authProvider.serviceType === 'saml') {
            window.location.assign(authProvider.authenticationURL)
        } else {
            window.location.assign(`${authProvider.authenticationURL}&redirect=${window.location.href}`)
        }
    }, [authProvider.serviceType, authProvider.authenticationURL])

    const { icon: AccountIcon } = account

    let accountConnection: JSX.Element | string
    switch (authProvider.serviceType) {
        case 'openidconnect':
        case 'saml':
            accountConnection = account.external?.displayName || 'Not connected'
            break
        default:
            accountConnection = (
                <>
                    {account.external?.url ? (
                        <>
                            {account.external.displayName} (
                            <Link to={account.external.url} target="_blank" rel="noopener noreferrer">
                                @{account.external.login}
                            </Link>
                            )
                        </>
                    ) : (
                        'Not connected'
                    )}
                </>
            )
    }

    return (
        <div className="d-flex align-items-start">
            {isRemoveAccountModalOpen && account.external && (
                <RemoveExternalAccountModal
                    id={account.external.id}
                    name={account.name}
                    onDidCancel={toggleRemoveAccountModal}
                    onDidRemove={onDidRemove}
                    onDidError={onDidError}
                />
            )}
            <div className="align-self-center">
                <AccountIcon className="mb-0 mr-2" />
            </div>
            <div className="flex-1 flex-column">
                <H3 className="m-0">{authProvider.displayName}</H3>
                <div className="text-muted">{accountConnection}</div>
            </div>
            <div className="align-self-center">
                {account.external ? (
                    <Button className="text-danger px-0" onClick={toggleRemoveAccountModal} variant="link">
                        Remove
                    </Button>
                ) : (
                    <LoaderButton
                        loading={isLoading}
                        label="Add"
                        display="block"
                        onClick={navigateToAuthProvider}
                        variant="success"
                    />
                )}
            </div>
        </div>
    )
}
