import React, { useCallback, useState } from 'react'

import { mdiAccount, mdiCircle, mdiCog, mdiDelete } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, isErrorLike, pluralize } from '@sourcegraph/common'
import { Button, Link, Icon, Tooltip, Text } from '@sourcegraph/wildcard'

import { ListExternalServiceFields } from '../../graphql-operations'
import { refreshSiteFlags } from '../../site/backend'
import { Timestamp } from '../time/Timestamp'

import { deleteExternalService } from './backend'
import { defaultExternalServices } from './externalServices'

import styles from './ExternalServiceNode.module.scss'

export interface ExternalServiceNodeProps {
    node: ListExternalServiceFields
    onDidUpdate: () => void
    history: H.History
    routingPrefix: string
    afterDeleteRoute: string
    editingDisabled: boolean
}

export const ExternalServiceNode: React.FunctionComponent<React.PropsWithChildren<ExternalServiceNodeProps>> = ({
    node,
    onDidUpdate,
    history,
    routingPrefix,
    afterDeleteRoute,
    editingDisabled,
}) => {
    const [isDeleting, setIsDeleting] = useState<boolean | Error>(false)
    const onDelete = useCallback<React.MouseEventHandler>(async () => {
        if (!window.confirm(`Delete the external service ${node.displayName}?`)) {
            return
        }
        setIsDeleting(true)
        try {
            await deleteExternalService(node.id)
            setIsDeleting(false)
            onDidUpdate()
            // eslint-disable-next-line rxjs/no-ignored-subscription
            refreshSiteFlags().subscribe()
            history.push(afterDeleteRoute)
        } catch (error) {
            setIsDeleting(asError(error))
        }
    }, [afterDeleteRoute, history, node.displayName, node.id, onDidUpdate])

    const IconComponent = defaultExternalServices[node.kind].icon

    return (
        <li
            className={classNames(styles.listNode, 'external-service-node list-group-item')}
            data-test-external-service-name={node.displayName}
        >
            <div className="d-flex align-items-center justify-content-between">
                <div className="align-self-start">
                    {node.lastSyncError === null && (
                        <Tooltip content="All good, no errors!">
                            <Icon
                                svgPath={mdiCircle}
                                aria-label="Code host integration is healthy"
                                className="text-success mr-2"
                            />
                        </Tooltip>
                    )}
                    {node.lastSyncError !== null && (
                        <Tooltip content="Syncing failed, check the error message for details!">
                            <Icon
                                svgPath={mdiCircle}
                                aria-label="Code host integration is unhealthy"
                                className="text-danger mr-2"
                            />
                        </Tooltip>
                    )}
                </div>
                <div className="flex-grow-1">
                    <div>
                        <Icon as={IconComponent} aria-label="Code host logo" className="mr-2" />
                        <strong>
                            {node.namespace && (
                                <>
                                    <Icon aria-hidden={true} svgPath={mdiAccount} />
                                    <Link to={node.namespace.url}>{node.namespace.namespaceName}</Link>{' '}
                                </>
                            )}
                            {node.displayName}{' '}
                            <small className="text-muted">
                                ({node.repoCount} {pluralize('repository', node.repoCount, 'repositories')})
                            </small>
                        </strong>
                        <br />
                        <Text className="mb-0 text-muted">
                            <small>
                                {node.lastSyncAt === null ? (
                                    <>Never synced.</>
                                ) : (
                                    <>
                                        Last synced <Timestamp date={node.lastSyncAt} />.
                                    </>
                                )}{' '}
                                {node.nextSyncAt !== null && (
                                    <>
                                        Next sync scheduled <Timestamp date={node.nextSyncAt} />.
                                    </>
                                )}
                                {node.nextSyncAt === null && <>No next sync scheduled.</>}
                            </small>
                        </Text>
                    </div>
                </div>
                <div className="flex-shrink-0 ml-3">
                    <Tooltip content={`${editingDisabled ? 'View' : 'Edit'} code host connection settings`}>
                        <Button
                            className="test-edit-external-service-button"
                            to={`${routingPrefix}/external-services/${node.id}`}
                            variant="secondary"
                            size="sm"
                            as={Link}
                        >
                            <Icon aria-hidden={true} svgPath={mdiCog} /> {editingDisabled ? 'View' : 'Edit'}
                        </Button>
                    </Tooltip>{' '}
                    <Tooltip content="Delete code host connection">
                        <Button
                            aria-label="Delete"
                            className="test-delete-external-service-button"
                            onClick={onDelete}
                            disabled={isDeleting === true}
                            variant="danger"
                            size="sm"
                        >
                            <Icon aria-hidden={true} svgPath={mdiDelete} />
                        </Button>
                    </Tooltip>
                </div>
            </div>
            {node.lastSyncError !== null && (
                <ErrorAlert error={node.lastSyncError} variant="danger" className="mt-2 mb-0" />
            )}
            {isErrorLike(isDeleting) && <ErrorAlert className="mt-2" error={isDeleting} />}
        </li>
    )
}
