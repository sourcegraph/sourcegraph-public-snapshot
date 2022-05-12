import React, { useState, useCallback } from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'

import { ErrorLike } from '@sourcegraph/common'
import { Button, Badge, Icon, Typography } from '@sourcegraph/wildcard'

import { CircleDashedIcon } from '../../../components/CircleDashedIcon'
import { LoaderButton } from '../../../components/LoaderButton'
import { ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'
import { Owner } from '../cloud-ga'

import { AddCodeHostConnectionModal } from './AddCodeHostConnectionModal'
import { scopes } from './modalHints'
import { RemoveCodeHostConnectionModal } from './RemoveCodeHostConnectionModal'
import { UpdateCodeHostConnectionModal } from './UpdateCodeHostConnectionModal'
import { ifNotNavigated, ServiceConfig } from './UserAddCodeHostsPage'

import styles from './CodeHostItem.module.scss'

interface CodeHostItemProps {
    kind: ExternalServiceKind
    owner: Owner
    name: string
    icon: React.ComponentType<React.PropsWithChildren<{ className?: string }>>
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
    loading?: boolean
    useGitHubApp?: boolean
}

export interface ParentWindow extends Window {
    onSuccess?: (reason: string | null) => void
}

export const CodeHostItem: React.FunctionComponent<React.PropsWithChildren<CodeHostItemProps>> = ({
    owner,
    service,
    kind,
    name,
    isTokenUpdateRequired,
    icon: ItemIcon,
    navigateToAuthProvider,
    onDidRemove,
    onDidError,
    onDidAdd,
    isUpdateModalOpen,
    toggleUpdateModal,
    onDidUpsert,
    loading = false,
    useGitHubApp = false,
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

    let serviceConfig: ServiceConfig = { pending: false }
    if (service) {
        serviceConfig = JSON.parse(service.config) as ServiceConfig
    }

    return (
        <div className="d-flex align-items-start">
            {onDidAdd && isAddConnectionModalOpen && (
                <AddCodeHostConnectionModal
                    ownerID={owner.id}
                    serviceKind={kind}
                    serviceName={name}
                    hintFragment={scopes[kind]}
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
            {service && toggleUpdateModal && onDidUpsert && isUpdateModalOpen && !serviceConfig.pending && (
                <UpdateCodeHostConnectionModal
                    serviceID={service.id}
                    serviceConfig={service.config}
                    serviceName={service.displayName}
                    orgName={owner.name || 'organization'}
                    kind={kind}
                    hintFragment={scopes[kind]}
                    onDidCancel={toggleUpdateModal}
                    onDidUpdate={onDidUpsert}
                    onDidError={onDidError}
                />
            )}
            <div className="align-self-center">
                {serviceConfig.pending ? (
                    <Icon className="mb-0 mr-2 text-info" as={AlertCircleIcon} />
                ) : service?.warning || service?.lastSyncError ? (
                    <Icon className="mb-0 mr-2 text-warning" as={AlertCircleIcon} />
                ) : service?.id ? (
                    <Icon className="mb-0 mr-2 text-success" as={CheckCircleIcon} />
                ) : (
                    <Icon className={classNames('mb-0 mr-2', styles.iconDashed)} as={CircleDashedIcon} />
                )}
                <Icon className="mb-0 mr-1" as={ItemIcon} />
            </div>
            <div className="flex-1 align-self-center">
                <Typography.H3 className="m-0">
                    {name} {serviceConfig.pending ? <Badge color="secondary">Pending</Badge> : null}
                </Typography.H3>
            </div>
            <div className="align-self-center">
                {/* Show one of: update, updating, connect, connecting buttons */}
                {!service?.id || serviceConfig.pending ? (
                    oauthInFlight ? (
                        <LoaderButton
                            loading={true}
                            disabled={true}
                            label="Connecting..."
                            alwaysShowLabel={true}
                            variant="primary"
                        />
                    ) : loading ? (
                        <LoaderButton
                            type="button"
                            className="btn btn-primary"
                            loading={true}
                            disabled={true}
                            alwaysShowLabel={false}
                        />
                    ) : (
                        <Button onClick={useGitHubApp ? toAuthProvider : connectAction} variant="primary">
                            Connect
                        </Button>
                    )
                ) : (
                    (isTokenUpdateRequired || !isUserOwner) &&
                    (oauthInFlight ? (
                        <LoaderButton
                            loading={true}
                            disabled={true}
                            label="Updating..."
                            alwaysShowLabel={true}
                            variant="merged"
                        />
                    ) : (
                        !useGitHubApp && (
                            <Button
                                className={classNames(!isUserOwner && 'p-0 shadow-none font-weight-normal')}
                                variant={isUserOwner ? 'merged' : 'link'}
                                onClick={updateAction}
                            >
                                Update
                            </Button>
                        )
                    ))
                )}

                {/* always show remove button when the service exists */}
                {service?.id && (
                    <Button
                        className="text-danger font-weight-normal shadow-none px-0 ml-3"
                        onClick={toggleRemoveConnectionModal}
                        variant="link"
                    >
                        Remove
                    </Button>
                )}
            </div>
        </div>
    )
}
