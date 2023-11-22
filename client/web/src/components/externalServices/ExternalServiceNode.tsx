import { type FC, useCallback, useState } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiCircle, mdiCog, mdiDelete } from '@mdi/js'
import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { asError, isErrorLike, pluralize } from '@sourcegraph/common'
import { Button, Link, LoadingSpinner, Icon, Tooltip, Text, ErrorAlert } from '@sourcegraph/wildcard'

import type { ListExternalServiceFields } from '../../graphql-operations'
import { refreshSiteFlags } from '../../site/backend'

import { deleteExternalService } from './backend'
import { defaultExternalServices, EXTERNAL_SERVICE_SYNC_RUNNING_STATUSES } from './externalServices'

import styles from './ExternalServiceNode.module.scss'

export interface ExternalServiceNodeProps {
    node: ListExternalServiceFields
    editingDisabled: boolean
}

export const ExternalServiceNode: FC<ExternalServiceNodeProps> = ({ node, editingDisabled }) => {
    const [isDeleting, setIsDeleting] = useState<boolean | Error>(false)
    const client = useApolloClient()
    const onDelete = useCallback<React.MouseEventHandler>(async () => {
        if (!window.confirm(`Delete the external service ${node.displayName}?`)) {
            return
        }
        setIsDeleting(true)
        try {
            await deleteExternalService(node.id)
            setIsDeleting(false)
            await refreshSiteFlags(client)
        } catch (error) {
            setIsDeleting(asError(error))
        } finally {
            const deletedCodeHostId = client.cache.identify({
                __typename: 'ExternalService',
                id: node.id,
            })

            // Remove deleted code host from the apollo cache.
            client.cache.evict({ id: deletedCodeHostId })
        }
    }, [node, client])

    const IconComponent = defaultExternalServices[node.kind].icon

    return (
        <li
            className={classNames(styles.listNode, 'external-service-node list-group-item')}
            data-test-external-service-name={node.displayName}
        >
            <div className="d-flex align-items-center justify-content-between">
                <div className="align-self-start">
                    {EXTERNAL_SERVICE_SYNC_RUNNING_STATUSES.has(node.syncJobs?.nodes[0]?.state) ? (
                        <Tooltip content="Sync is running">
                            <div aria-label="Sync is running">
                                <LoadingSpinner className="m-0 mr-2" inline={true} />
                            </div>
                        </Tooltip>
                    ) : node.lastSyncError === null ? (
                        <Tooltip content="All good, no errors!">
                            <Icon
                                svgPath={mdiCircle}
                                aria-label="Code host integration is healthy"
                                className="text-success mr-2"
                            />
                        </Tooltip>
                    ) : (
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
                        <Icon as={IconComponent} aria-label="Code host logo" className="code-host-logo mr-2" />
                        <strong>
                            <Link to={`/site-admin/external-services/${encodeURIComponent(node.id)}`}>
                                {node.displayName}
                            </Link>{' '}
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
                    <Tooltip
                        content={
                            editingDisabled
                                ? 'Editing code host connections through the UI is disabled when the EXTSVC_CONFIG_FILE environment variable is set.'
                                : 'Edit code host connection settings'
                        }
                    >
                        <Button
                            className="test-edit-external-service-button"
                            to={`/site-admin/external-services/${encodeURIComponent(node.id)}/edit`}
                            variant="secondary"
                            size="sm"
                            as={Link}
                            disabled={editingDisabled}
                        >
                            <Icon aria-hidden={true} svgPath={mdiCog} /> Edit
                        </Button>
                    </Tooltip>{' '}
                    <Tooltip
                        content={
                            editingDisabled
                                ? 'Deleting code host connections through the UI is disabled when the EXTSVC_CONFIG_FILE environment variable is set.'
                                : 'Delete code host connection'
                        }
                    >
                        <Button
                            aria-label="Delete"
                            className="test-delete-external-service-button"
                            onClick={onDelete}
                            disabled={isDeleting === true || editingDisabled}
                            variant="danger"
                            size="sm"
                        >
                            <Icon aria-hidden={true} svgPath={mdiDelete} />
                            {' Delete'}
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
