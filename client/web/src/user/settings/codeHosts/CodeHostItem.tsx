import React, { useState, useCallback } from 'react'
import { noop } from 'lodash'
// import * as H from 'history'

import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CircleOutlineIcon from 'mdi-react/CircleOutlineIcon'

import { Link } from '../../../../../shared/src/components/Link'
import { ExternalServiceKind, ExternalServiceFields } from '../../../graphql-operations'
import { AddCodeHostConnectionModal } from './AddCodeHostConnectionModal'
import { RemoveCodeHostConnectionModal } from './RemoveCodeHostConnectionModal'

interface CodeHostItemProps {
    kind: ExternalServiceKind
    icon: React.ComponentType<{ className?: string }>
    name: string
    serviceIds: ExternalServiceFields['id'][] | undefined

    onDidConnect: () => void
    onDidEdit: () => void
    onDidRemove: () => void
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

    // // As of now the following types of user code hosts are not supported - stubs
    // [ExternalServiceKind.BITBUCKETSERVER]: <></>,
    // [ExternalServiceKind.BITBUCKETCLOUD]: <></>,
    // [ExternalServiceKind.GITOLITE]: <></>,
    // [ExternalServiceKind.PHABRICATOR]: <></>,
    // [ExternalServiceKind.AWSCODECOMMIT]: <></>,
    // [ExternalServiceKind.OTHER]: <></>,
}

export const CodeHostItem: React.FunctionComponent<CodeHostItemProps> = ({
    kind,
    name,
    icon: Icon,
    serviceIds = [],
}) => {
    const [showAddConnectionModal, setShowAddConnectionModal] = useState(false)
    const toggleAddConnectionModal = useCallback(() => setShowAddConnectionModal(!showAddConnectionModal), [
        showAddConnectionModal,
    ])

    const [showRemoveConnectionModal, setShowRemoveConnectionModal] = useState(false)
    const toggleRemoveConnectionModal = useCallback(() => setShowRemoveConnectionModal(!showRemoveConnectionModal), [
        showRemoveConnectionModal,
    ])

    const onCodeHostConnect = useCallback((token: string): void => {
        console.log(`Adding token: ${token}`)
    }, [])

    const hasServices = serviceIds.length !== 0

    return (
        <div className="p-2 d-flex align-items-start">
            {showAddConnectionModal && (
                <AddCodeHostConnectionModal
                    kind={kind}
                    name={name}
                    hintFragment={MODAL_HINTS[kind]}
                    onDidAdd={onCodeHostConnect}
                    onDidCancel={toggleAddConnectionModal}
                />
            )}
            {showRemoveConnectionModal && (
                <RemoveCodeHostConnectionModal
                    kind={kind}
                    name={name}
                    servicesCount={serviceIds.length}
                    onDidRemove={noop}
                    onDidCancel={toggleRemoveConnectionModal}
                />
            )}
            <div className="align-self-center">
                {hasServices ? (
                    <CheckCircleIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon_check" />
                ) : (
                    <CircleOutlineIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon_off" />
                )}
                {/* <AlertCircleIcon className="icon-inline mb-0 mr-2 add-user-code-hosts-page__icon_warn" /> */}
                <Icon className="icon-inline mb-0 mr-2" />
            </div>
            <div className="flex-1">
                <h3 className="mt-1 mb-0">{name}</h3>
            </div>
            <div className="align-self-center">
                {!hasServices && (
                    <button type="button" className="btn btn-success" onClick={toggleAddConnectionModal}>
                        Connect
                    </button>
                )}
                {hasServices && (
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
                )}
            </div>
        </div>
    )
}
