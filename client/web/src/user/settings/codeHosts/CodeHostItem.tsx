import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import React, { useState, useCallback } from 'react'

import { ErrorLike } from '@sourcegraph/shared/src/util/errors'

import { CircleDashedIcon } from '../../../components/CircleDashedIcon'
import { LoaderButton } from '../../../components/LoaderButton'
import { ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'
import { Owner } from '../cloud-ga'

import { AddCodeHostConnectionModal } from './AddCodeHostConnectionModal'
import { hints } from './modalHints'
import { RemoveCodeHostConnectionModal } from './RemoveCodeHostConnectionModal'
import { UpdateCodeHostConnectionModal } from './UpdateCodeHostConnectionModal'
import { ifNotNavigated } from './UserAddCodeHostsPage'

interface CodeHostItemProps {
    kind: ExternalServiceKind
    owner: Owner
    name: string
    icon: React.ComponentType<{ className?: string }>
    navigateToAuthProvider: (kind: ExternalServiceKind) => void
    isTokenUpdateRequired: boolean | undefined
    // optional service object fields when the code host connection is active
    service?: ListExternalServiceFields
    isUpdateModalOpen?: boolean
    toggleUpdateModal?: () => void
    onDidUpsert?: (service: ListExternalServiceFields) => void
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
    isUpdateModalOpen,
    toggleUpdateModal,
    onDidUpsert,
}) => {
    const [isAddConnectionModalOpen, setIsAddConnectionModalOpen] = useState(false)
    const toggleAddConnectionModal = useCallback(() => setIsAddConnectionModalOpen(!isAddConnectionModalOpen), [
        isAddConnectionModalOpen,
    ])

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

    const isUserOwner = owner.type === 'user'
    const connectAction = isUserOwner ? toAuthProvider : toggleAddConnectionModal
    const updateAction = isUserOwner ? toAuthProvider : toggleUpdateModal

    return (
        <div className="d-flex align-items-start">
            {onDidAdd && isAddConnectionModalOpen && (
                <AddCodeHostConnectionModal
                    ownerID={owner.id}
                    serviceKind={kind}
                    serviceName={name}
                    hintFragment={hints[kind]}
                    onDidAdd={onDidAdd}
                    onDidCancel={toggleAddConnectionModal}
                    onDidError={onDidError}
                />
            )}
            {service && isRemoveConnectionModalOpen && (
                <RemoveCodeHostConnectionModal
                    serviceID={service.id}
                    serviceName={name}
                    serviceKind={kind}
                    orgName={owner.name || 'organization'}
                    repoCount={service.repoCount}
                    onDidRemove={onDidRemove}
                    onDidCancel={toggleRemoveConnectionModal}
                    onDidError={onDidError}
                />
            )}
            {service && toggleUpdateModal && onDidUpsert && isUpdateModalOpen && (
                <UpdateCodeHostConnectionModal
                    serviceID={service.id}
                    serviceConfig={service.config}
                    serviceName={service.displayName}
                    orgName={owner.name || 'organization'}
                    kind={kind}
                    hintFragment={hints[kind]}
                    onDidCancel={toggleUpdateModal}
                    onDidUpdate={onDidUpsert}
                    onDidError={onDidError}
                />
            )}
            <div className="align-self-center">
                {service?.warning || service?.lastSyncError ? (
                    <AlertCircleIcon className="icon-inline mb-0 mr-2 text-warning" />
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
                {/* Show one of: update, updating, connect, connecting buttons */}
                {!service?.id ? (
                    oauthInFlight ? (
                        <LoaderButton
                            type="button"
                            className="btn btn-primary"
                            loading={true}
                            disabled={true}
                            label="Connecting..."
                            alwaysShowLabel={true}
                        />
                    ) : (
                        <button type="button" className="btn btn-primary" onClick={connectAction}>
                            Connect
                        </button>
                    )
                ) : (
                    (isTokenUpdateRequired || !isUserOwner) &&
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
                        <button
                            type="button"
                            className={`btn ${
                                !isUserOwner ? 'btn-link p-0 shadow-none font-weight-normal' : 'btn-merged'
                            }`}
                            onClick={updateAction}
                        >
                            Update
                        </button>
                    ))
                )}

                {/* always show remove button when the service exists */}
                {service?.id && (
                    <button
                        type="button"
                        className="btn btn-link text-danger font-weight-normal shadow-none px-0 ml-3"
                        onClick={toggleRemoveConnectionModal}
                    >
                        Remove
                    </button>
                )}
            </div>
        </div>
    )
}
