import React, { useState, useCallback } from 'react'
import { Link } from '../../../../../shared/src/components/Link'
// import { ErrorLike } from '../../../../../shared/src/util/errors'
import { SourcegraphContext } from '../../../jscontext'
import { RemoveExternalAccountModal } from './RemoveExternalAccountModal'

import type { NormalizedMinAccount } from './ExternalAccountsSignIn'
import { ErrorLike } from '../../../../../shared/src/util/errors'

type AuthProvider = SourcegraphContext['authProviders'][0]

interface Props {
    account: NormalizedMinAccount
    authProvider: AuthProvider
    onDidRemove: () => void
    onDidError: (error: ErrorLike) => void
}

export const ExternalAccount: React.FunctionComponent<Props> = ({ account, authProvider, onDidRemove, onDidError }) => {
    const [isRemoveAccountModalOpen, setIsRemoveAccountModalOpen] = useState(false)
    const toggleRemoveAccountModal = useCallback(() => setIsRemoveAccountModalOpen(!isRemoveAccountModalOpen), [
        isRemoveAccountModalOpen,
    ])

    const { icon: AccountIcon } = account

    return (
        <div className="p-2 d-flex align-items-start ">
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
                <h3 className="m-0">{account.name}</h3>
                <div className="text-muted">
                    {account.external ? (
                        <>
                            {account.external.userName} (
                            <Link to={account.external.userUrl} target="_blank" rel="noopener noreferrer">
                                @{account.external.userLogin}
                            </Link>
                            )
                        </>
                    ) : (
                        'Not connected'
                    )}
                </div>
            </div>
            <div className="align-self-center">
                {account.external ? (
                    <button type="button" className="btn btn-link text-danger px-0" onClick={toggleRemoveAccountModal}>
                        Remove
                    </button>
                ) : (
                    <a
                        // authenticationURL should always be there
                        href={`${authProvider.authenticationURL as string}&redirect=${window.location.href}`}
                        rel="noopener noreferrer"
                        className="btn btn-secondary btn-block"
                    >
                        Add
                    </a>
                )}
            </div>
        </div>
    )
}
