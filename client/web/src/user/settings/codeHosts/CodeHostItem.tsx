import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import React, { useState, useCallback } from 'react'
import { ButtonDropdown, DropdownToggle, DropdownMenu, DropdownItem } from 'reactstrap'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ErrorLike } from '@sourcegraph/shared/src/util/errors'

import { CircleDashedIcon } from '../../../components/CircleDashedIcon'
import { Scalars, ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'

import { AddCodeHostConnectionModal } from './AddCodeHostConnectionModal'
import { hints } from './modalHints'
import { RemoveCodeHostConnectionModal } from './RemoveCodeHostConnectionModal'
import { ifNotNavigated } from './UserAddCodeHostsPage'

interface CodeHostItemProps {
    userID: Scalars['ID']
    kind: ExternalServiceKind
    name: string
    icon: React.ComponentType<{ className?: string }>
    navigateToAuthProvider: (kind: ExternalServiceKind) => void

    // optional service object fields when the code host connection is active
    service?: ListExternalServiceFields

    onDidAdd: (service: ListExternalServiceFields) => void
    onDidRemove: () => void
    onDidError: (error: ErrorLike) => void
}

export const CodeHostItem: React.FunctionComponent<CodeHostItemProps> = ({
    userID,
    service,
    kind,
    name,
    icon: Icon,
    navigateToAuthProvider,
    onDidAdd,
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

    const [dropdownOpen, setOpen] = useState(false)
    const toggleDropdown = useCallback((): void => setOpen(!dropdownOpen), [dropdownOpen])

    const [oauthInFlight, setOauthInFlight] = useState(false)

    const toAuthProvider = useCallback((): void => {
        setOauthInFlight(true)
        ifNotNavigated(() => {
            setOauthInFlight(false)
        })
        navigateToAuthProvider(kind)
    }, [kind, navigateToAuthProvider])

    return (
        <div className="p-2 d-flex align-items-start">
            {isAddConnectionModalOpen && (
                <AddCodeHostConnectionModal
                    userID={userID}
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
                <Icon className="icon-inline mb-0 mr-1" />
            </div>
            <div className="flex-1 align-self-center">
                <h3 className="m-0">{name}</h3>
            </div>
            <div className="align-self-center">
                {service?.id ? (
                    <button
                        type="button"
                        className="btn btn-link btn-sm text-danger px-0 shadow-none"
                        onClick={toggleRemoveConnectionModal}
                    >
                        Remove
                    </button>
                ) : (
                    <ButtonDropdown isOpen={dropdownOpen} toggle={toggleDropdown} direction="down">
                        <DropdownToggle className="btn-sm" color="outline-secondary" caret={true}>
                            Connect
                        </DropdownToggle>
                        <DropdownMenu right={true}>
                            <DropdownItem toggle={false} onClick={toAuthProvider}>
                                Connect with {name}
                                {oauthInFlight && <LoadingSpinner className="icon-inline ml-2" />}
                            </DropdownItem>
                            <DropdownItem onClick={toggleAddConnectionModal}>Connect with access token</DropdownItem>
                        </DropdownMenu>
                    </ButtonDropdown>
                )}
            </div>
        </div>
    )
}
