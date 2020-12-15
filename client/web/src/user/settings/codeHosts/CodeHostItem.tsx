import React, { useState, useCallback } from 'react'

import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CircleOutlineIcon from 'mdi-react/CircleOutlineIcon'

import { AddCodeHostConnectionModal } from './AddCodeHostConnectionModal'
import { RemoveCodeHostConnectionModal } from './RemoveCodeHostConnectionModal'
import { Link } from '../../../../../shared/src/components/Link'
import { ExternalServiceKind, ExternalServiceFields } from '../../../graphql-operations'
import { ErrorLike } from '../../../../../shared/src/util/errors'

interface CodeHostItemProps {
    kind: ExternalServiceKind
    name: string
    icon: React.ComponentType<{ className?: string }>
    // optional service object fields when the code host connection is active
    id?: ExternalServiceFields['id']
    repoCount?: number
    warning?: string

    onDidConnect: () => void
    onDidRemove: () => void
    onDidError: (error: ErrorLike) => void
}

const MODAL_HINTS: Partial<Record<ExternalServiceKind, React.ReactFragment>> = {
    [ExternalServiceKind.GITHUB]: (
        <small>
            <Link to="" target="_blank" rel="noopener noreferrer">
                Create a new access token
            </Link>
            <span className="text-muted"> on GitHub.com with repo or public_repo scope.</span>
        </small>
    ),
    [ExternalServiceKind.GITLAB]: (
        <small>
            <Link
                to="https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#creating-a-personal-access-token"
                target="_blank"
                rel="noopener noreferrer"
            >
                Create a new access token
            </Link>
            <span className="text-muted"> on GitLab.com with read_user, read_api, and read_repository scope.</span>
        </small>
    ),
}

export const CodeHostItem: React.FunctionComponent<CodeHostItemProps> = ({
    id,
    repoCount,
    warning,
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
                    kind={kind}
                    name={name}
                    hintFragment={MODAL_HINTS[kind]}
                    onDidAdd={onDidConnect}
                    onDidCancel={toggleAddConnectionModal}
                    onDidError={onDidError}
                />
            )}
            {id && showRemoveConnectionModal && (
                <RemoveCodeHostConnectionModal
                    id={id}
                    kind={kind}
                    name={name}
                    repoCount={repoCount}
                    onDidRemove={onDidRemove}
                    onDidCancel={toggleRemoveConnectionModal}
                    onDidError={onDidError}
                />
            )}
            <div className="align-self-center">
                {id ? (
                    <CheckCircleIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon--success" />
                ) : warning ? (
                    <AlertCircleIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon--danger" />
                ) : (
                    <CircleOutlineIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon--outline" />
                )}
                <Icon className="icon-inline mb-0 mr-2" />
            </div>
            <div className="flex-1">
                <h3 className="mt-1 mb-0">{name}</h3>
            </div>
            <div className="align-self-center">
                {id ? (
                    <>
                        <button
                            type="button"
                            className="btn btn-link text-primary p-0 mr-2"
                            onClick={() => {}}
                            disabled={false}
                        >
                            Edit
                        </button>
                        <button
                            type="button"
                            className="btn btn-link text-danger p-0"
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
