import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import React, { useState, useCallback } from 'react'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { ErrorLike } from '@sourcegraph/shared/src/util/errors'

import { CircleDashedIcon } from '../../../components/CircleDashedIcon'
import { LoaderButton } from '../../../components/LoaderButton'
import { ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'

import { AddCodeHostConnectionModal } from './AddCodeHostConnectionModal'
import { hints } from './modalHints'
import { RemoveCodeHostConnectionModal } from './RemoveCodeHostConnectionModal'
import { ifNotNavigated } from './UserAddCodeHostsPage'

interface CodeHostItemProps {
    kind: ExternalServiceKind
    // TODO: export this
    owner: { id: Scalars['ID']; tags?: string[]; type: 'user' | 'org' }
    name: string
    icon: React.ComponentType<{ className?: string }>
    navigateToAuthProvider: (kind: ExternalServiceKind) => void
    isTokenUpdateRequired: boolean | undefined
    // optional service object fields when the code host connection is active
    service?: ListExternalServiceFields

    onDidAdd?: (service: ListExternalServiceFields) => void
    onDidRemove: () => void
    onDidError: (error: ErrorLike) => void
}

export const CodeHostItem: React.FunctionComponent<CodeHostItemProps> = ({
    owner,
    service,
    kind,
    name,
    isTokenUpdateRequired,
    icon: Icon,
    navigateToAuthProvider,
    onDidRemove,
    onDidError,
    onDidAdd,
}) => {
    // PAT >>>

    const [isAddConnectionModalOpen, setIsAddConnectionModalOpen] = useState(false)
    const toggleAddConnectionModal = useCallback(() => setIsAddConnectionModalOpen(!isAddConnectionModalOpen), [
        isAddConnectionModalOpen,
    ])

    // <<< PAT

    const [isRemoveConnectionModalOpen, setIsRemoveConnectionModalOpen] = useState(false)
    const toggleRemoveConnectionModal = useCallback(
        () => setIsRemoveConnectionModalOpen(!isRemoveConnectionModalOpen),
        [isRemoveConnectionModalOpen]
    )

    const [oauthInFlight, setOauthInFlight] = useState(false)

    const toAuthProvider = useCallback((): void => {
        setOauthInFlight(true)
        ifNotNavigated(() => {
            setOauthInFlight(false)
        })
        navigateToAuthProvider(kind)
    }, [kind, navigateToAuthProvider])

    const connectAction = owner.type === 'user' ? toAuthProvider : toggleAddConnectionModal

    return (
        <div className="d-flex align-items-start">
            {onDidAdd && isAddConnectionModalOpen && (
                <AddCodeHostConnectionModal
                    ownerID={owner.id}
                    kind={kind}
                    name={name}
                    hintFragment={hints[kind]}
                    onDidAdd={onDidAdd}
                    onDidCancel={toggleAddConnectionModal}
                    onDidError={onDidError}
                />
            )}
            {service && isRemoveConnectionModalOpen && (
                <RemoveCodeHostConnectionModal
                    id={service.id}
                    kind={kind}
                    name={name}
                    repoCount={service.repoCount}
                    onDidRemove={onDidRemove}
                    onDidCancel={toggleRemoveConnectionModal}
                    onDidError={onDidError}
                />
            )}
            <div className="align-self-center">
                {service?.warning || service?.lastSyncError ? (
                    <AlertCircleIcon className="icon-inline mb-0 mr-2 text-danger" />
                ) : service?.id ? (
                    <CheckCircleIcon className="icon-inline mb-0 mr-2 text-success" />
                ) : (
                    <CircleDashedIcon className="icon-inline mb-0 mr-2 user-code-hosts-page__icon--dashed" />
                )}
                <Icon className="mb-0 mr-1" />
            </div>
            <div className="flex-1 align-self-center">
                <h3 className="m-0">{name}</h3>
            </div>
            <div className="align-self-center">
                {/* always show remove button when the service exists */}
                {service?.id && (
                    <button
                        type="button"
                        className="btn btn-link text-danger shadow-none"
                        onClick={toggleRemoveConnectionModal}
                    >
                        Remove
                    </button>
                )}

                {/* Show one of: update, updating, connect, connecting buttons */}
                {!service?.id ? (
                    oauthInFlight ? (
                        <LoaderButton
                            type="button"
                            className="btn btn-success"
                            loading={true}
                            disabled={true}
                            label="Connecting..."
                            alwaysShowLabel={true}
                        />
                    ) : (
                        <button type="button" className="btn btn-success" onClick={connectAction}>
                            Connect
                        </button>
                    )
                ) : (
                    isTokenUpdateRequired &&
                    (oauthInFlight ? (
                        <LoaderButton
                            type="button"
                            className="btn btn-merged"
                            loading={true}
                            disabled={true}
                            label="Updating..."
                            alwaysShowLabel={true}
                        />
                    ) : (
                        <button type="button" className="btn btn-merged" onClick={toAuthProvider}>
                            Update
                        </button>
                    ))
                )}
            </div>
        </div>
    )
}
