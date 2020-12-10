import React, { useState, useCallback } from 'react'
import * as H from 'history'

import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CircleOutlineIcon from 'mdi-react/CircleOutlineIcon'

import { Link } from '../../../../../shared/src/components/Link'
import { ExternalServiceKind } from '../../../graphql-operations'
import { AddCodeHostTokenModal } from './AddCodeHostTokenModal'

interface CodeHostItemProps {
    onDidConnect: () => void
    onDidEdit: () => void
    onDidRemove: () => void
    name: string
    /**
     * Title to show in the external service "button"
     */
    title: string

    /**
     * Icon to show in the external service "button"
     */
    icon: React.ComponentType<{ className?: string }>

    /**
     * A short description that will appear in the external service "button" under the title
     */
    shortDescription?: string
    kind: ExternalServiceKind
    className?: string
}

const MODAL_HINTS: Record<ExternalServiceKind, React.ReactFragment> = {
    [ExternalServiceKind.GITHUB]: (
        <small>
            <Link to="">Create a new access token</Link>
            <span className="text-muted"> on GitHub.com with repo or public_repo scope.</span>
        </small>
    ),
    [ExternalServiceKind.GITLAB]: (
        <small>
            <Link to="">Create a new access token</Link>
            <span className="text-muted"> on GitLab.com with repo or TODO scope.</span>
        </small>
    ),

    // As of now the following types of user code hosts are not supported - stubs
    [ExternalServiceKind.BITBUCKETSERVER]: <></>,
    [ExternalServiceKind.BITBUCKETCLOUD]: <></>,
    [ExternalServiceKind.GITOLITE]: <></>,
    [ExternalServiceKind.PHABRICATOR]: <></>,
    [ExternalServiceKind.AWSCODECOMMIT]: <></>,
    [ExternalServiceKind.OTHER]: <></>,
}

export const CodeHostItem: React.FunctionComponent<CodeHostItemProps> = ({
    kind,
    name,
    icon: Icon,
    shortDescription,
}) => {
    const [showAddTokenModal, setShowAddTokenModal] = useState(false)
    const toggleAddTokenModal = useCallback(() => setShowAddTokenModal(!showAddTokenModal), [showAddTokenModal])

    const onCodeHostConnect = useCallback((token: string): void => {
        console.log(`Addind token: ${token}`)
    }, [])
    // const onEdit
    // const onRemove

    return (
        <div className="p-2 d-flex align-items-start">
            {showAddTokenModal && (
                <AddCodeHostTokenModal
                    kind={kind}
                    name={name}
                    onDidAdd={onCodeHostConnect}
                    onDidCancel={toggleAddTokenModal}
                    hintFragment={MODAL_HINTS[kind]}
                />
            )}
            <div className="align-self-center">
                <CircleOutlineIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon_off" />
                <AlertCircleIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon_warn" />
                <CheckCircleIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon_check" />
                <Icon className="icon-inline mb-0 mr-2" />
            </div>
            <div className="flex-1">
                <h3 className={shortDescription ? 'mb-0' : 'mt-1 mb-0'}>{name}</h3>
                {shortDescription && <p className="mb-0 text-muted">{shortDescription}</p>}
            </div>
            <div className="align-self-center">
                {true && (
                    <button type="button" className="btn btn-success" onClick={toggleAddTokenModal}>
                        Connect
                    </button>
                )}
                {true && (
                    <button
                        type="button"
                        className="btn btn-link text-primary p-0 mr-2"
                        onClick={() => {}}
                        disabled={false}
                    >
                        Edit
                    </button>
                )}
                {true && (
                    <button type="button" className="btn btn-link text-danger p-0" disabled={false}>
                        Remove
                    </button>
                )}
            </div>
        </div>
    )
}
