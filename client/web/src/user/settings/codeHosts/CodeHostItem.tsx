import React, { useState, useCallback } from 'react'

import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CircleOutlineIcon from 'mdi-react/CircleOutlineIcon'

import { AddCodeHostConnectionModal } from './AddCodeHostConnectionModal'
import { RemoveCodeHostConnectionModal } from './RemoveCodeHostConnectionModal'
import { hints } from './modalHints'
import { Scalars, ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'
import { ErrorLike } from '../../../../../shared/src/util/errors'

interface CodeHostItemProps {
    userID: Scalars['ID']
    kind: ExternalServiceKind
    name: string
    icon: React.ComponentType<{ className?: string }>
    // optional service object fields when the code host connection is active
    service?: ListExternalServiceFields
    repoCount?: number

    onDidConnect: (service: ListExternalServiceFields) => void
    onDidRemove: () => void
    onDidError: (error: ErrorLike) => void
}

export const CodeHostItem: React.FunctionComponent<CodeHostItemProps> = ({
    userID,
    service,
    repoCount,
    kind,
    name,
    icon: Icon,
    onDidConnect,
    onDidRemove,
    onDidError,
}) => {
    const [showAddConnectionModal, setShowAddConnectionModal] = useState(false)
    const toggleAddConnectionModal = useCallback(() => setShowAddConnectionModal(!showAddConnectionModal), [
        showAddConnectionModal,
    ])

    const [showRemoveConnectionModal, setShowRemoveConnectionModal] = useState(false)
    const toggleRemoveConnectionModal = useCallback(() => setShowRemoveConnectionModal(!showRemoveConnectionModal), [
        showRemoveConnectionModal,
    ])

    return (
        <div className="p-2 d-flex align-items-start">
            {showAddConnectionModal && (
                <AddCodeHostConnectionModal
                    userID={userID}
                    kind={kind}
                    name={name}
                    hintFragment={hints[kind]}
                    onDidAdd={onDidConnect}
                    onDidCancel={toggleAddConnectionModal}
                    onDidError={onDidError}
                />
            )}
            {service && showRemoveConnectionModal && (
                <RemoveCodeHostConnectionModal
                    id={service.id}
                    kind={kind}
                    name={name}
                    repoCount={repoCount}
                    onDidRemove={onDidRemove}
                    onDidCancel={toggleRemoveConnectionModal}
                    onDidError={onDidError}
                />
            )}
            <div className="align-self-center">
                {service?.warning ? (
                    <AlertCircleIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon--danger" />
                ) : service?.id ? (
                    <CheckCircleIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon--success" />
                ) : (
                    <CircleOutlineIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon--outline" />
                )}
                <Icon className="icon-inline mb-0 mr-2" />
            </div>
            <div className="flex-1">
                <h3 className="mt-1 mb-0">{name}</h3>
            </div>
            <div className="align-self-center">
                {service?.id ? (
                    <>
                        <button
                            type="button"
                            className="btn btn-link text-primary p-0 mr-2 shadow-none"
                            onClick={() => {}}
                            disabled={false}
                        >
                            Edit
                        </button>
                        <button
                            type="button"
                            className="btn btn-link text-danger p-0 shadow-none"
                            onClick={toggleRemoveConnectionModal}
                            disabled={false}
                        >
                            Remove
                        </button>
                    </>
                ) : (
                    <button type="button" className="btn btn-success" onClick={toggleAddConnectionModal}>
                        Connect
                    </button>
                )}
            </div>
        </div>
    )
}
