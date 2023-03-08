import { FC, useCallback, useState } from 'react'

import { useApolloClient } from '@apollo/client'
import { mdiCircle, mdiCog, mdiDelete } from '@mdi/js'
import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { asError, isErrorLike, pluralize } from '@sourcegraph/common'
import { Button, Link, LoadingSpinner, Icon, Tooltip, Text, ErrorAlert } from '@sourcegraph/wildcard'

import { AddExternalServiceInput, ExternalServiceKind, ListExternalServiceFields } from '../../graphql-operations'
import { refreshSiteFlags } from '../../site/backend'

import { deleteExternalService } from './backend'
import { defaultExternalServices, EXTERNAL_SERVICE_SYNC_RUNNING_STATUSES } from './externalServices'

import styles from './ExternalServiceNode.module.scss'

export interface ExternalServiceNodeProps {
    node: ListExternalServiceFields
    editingDisabled: boolean
    isSourcegraphApp: boolean
}

export const ExternalServiceNode: FC<ExternalServiceNodeProps> = ({ node, editingDisabled, isSourcegraphApp }) => {
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
            window.location.reload()
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
                                <LoadingSpinner className="mr-2" inline={true} />
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
                        <Icon as={IconComponent} aria-label="Code host logo" className="mr-2" />
                        <strong>
                            <Link to={`/site-admin/external-services/${node.id}`}>{node.displayName}</Link>{' '}
                            <small className="text-muted">
                                ({node.repoCount}
                                {isSourcegraphApp ? (isAppLocalFileService(node) ? ' of ∞' : ' of 10') : ''}{' '}
                                {isSourcegraphApp
                                    ? 'repositories'
                                    : pluralize('repository', node.repoCount, 'repositories')}
                                )
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
                            to={`/site-admin/external-services/${node.id}/edit`}
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

export function isAppLocalFileService(node: ListExternalServiceFields | AddExternalServiceInput): boolean {
    // TODO: Find a better way to reliably detect the auto-generated local file service
    return node.createdAt === '0001-01-01T00:00:00Z'
}
