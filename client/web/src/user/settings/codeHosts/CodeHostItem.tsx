import React, { useState, useCallback } from 'react'

import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { CircleDashedIcon } from '../../../components/CircleDashedIcon'

import { AddCodeHostConnectionModal } from './AddCodeHostConnectionModal'
import { RemoveCodeHostConnectionModal } from './RemoveCodeHostConnectionModal'
import { UpdateCodeHostConnectionModal } from './UpdateCodeHostConnectionModal'
import { hints } from './modalHints'
import { Scalars, ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'
import { ErrorLike } from '../../../../../shared/src/util/errors'

interface CodeHostItemProps {
    userID: Scalars['ID']
    kind: ExternalServiceKind
    name: string
    icon: React.ComponentType<{ className?: string }>
    isUpdateModalOpen: boolean
    toggleUpdateModal: () => void
    // optional service object fields when the code host connection is active
    service?: ListExternalServiceFields

    onDidUpsert: (service: ListExternalServiceFields) => void
    onDidRemove: () => void
    onDidError: (error: ErrorLike) => void
}

export const CodeHostItem: React.FunctionComponent<CodeHostItemProps> = ({
    userID,
    service,
    kind,
    name,
    icon: Icon,
    isUpdateModalOpen,
    toggleUpdateModal,
    onDidUpsert,
    onDidRemove,
    onDidError,
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

    return (
        <div className="p-2 d-flex align-items-start">
            {isAddConnectionModalOpen && (
                <AddCodeHostConnectionModal
                    userID={userID}
                    kind={kind}
                    name={name}
                    hintFragment={hints[kind]}
                    onDidAdd={onDidUpsert}
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
            {service && isUpdateModalOpen && (
                <UpdateCodeHostConnectionModal
                    serviceId={service.id}
                    serviceConfig={service.config}
                    name={service.displayName}
                    kind={kind}
                    hintFragment={hints[kind]}
                    onDidCancel={toggleUpdateModal}
                    onDidUpdate={onDidUpsert}
                    onDidError={onDidError}
                />
            )}
            <div className="align-self-center">
                {service?.warning || service?.lastSyncError ? (
                    <AlertCircleIcon className="icon-inline mb-0 mr-2 text-danger" />
                ) : service?.id ? (
                    <CheckCircleIcon className="icon-inline mb-0 mr-2 text-success" />
                ) : (
                    <CircleDashedIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon--dashed" />
                )}
                <Icon className="icon-inline mb-0 mr-1" />
            </div>
            <div className="flex-1 align-self-center">
                <h3 className="m-0">{name}</h3>
            </div>
            <div className="align-self-center">
                {service?.id ? (
                    <>
                        <button
                            type="button"
                            className="btn btn-link text-primary px-0 mr-2 shadow-none"
                            onClick={toggleUpdateModal}
                        >
                            Edit
                        </button>
                        <button
                            type="button"
                            className="btn btn-link text-danger px-0 shadow-none"
                            onClick={toggleRemoveConnectionModal}
                        >
                            Remove
                        </button>
                    </>
                ) : (
                    <button type="button" className="btn btn-success shadow-none" onClick={toggleAddConnectionModal}>
                        Connect
                    </button>
                )}
            </div>
        </div>
    )
}
