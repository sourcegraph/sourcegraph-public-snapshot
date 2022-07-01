import React, { useCallback, useState } from 'react'

import { mdiAccount, mdiCog, mdiDelete } from '@mdi/js'
import * as H from 'history'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, isErrorLike } from '@sourcegraph/common'
import { Button, Link, Icon } from '@sourcegraph/wildcard'

import { ListExternalServiceFields } from '../../graphql-operations'
import { refreshSiteFlags } from '../../site/backend'

import { deleteExternalService } from './backend'

export interface ExternalServiceNodeProps {
    node: ListExternalServiceFields
    onDidUpdate: () => void
    history: H.History
    routingPrefix: string
    afterDeleteRoute: string
}

export const ExternalServiceNode: React.FunctionComponent<React.PropsWithChildren<ExternalServiceNodeProps>> = ({
    node,
    onDidUpdate,
    history,
    routingPrefix,
    afterDeleteRoute,
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

    return (
        <li className="external-service-node list-group-item py-2" data-test-external-service-name={node.displayName}>
            <div className="d-flex align-items-center justify-content-between">
                <div>
                    {node.namespace && (
                        <>
                            <Icon aria-hidden={true} svgPath={mdiAccount} />
                            <Link to={node.namespace.url}>{node.namespace.namespaceName}</Link>{' '}
                        </>
                    )}
                    {node.displayName}
                </div>
                <div>
                    <Button
                        className="test-edit-external-service-button"
                        to={`${routingPrefix}/external-services/${node.id}`}
                        data-tooltip="External service settings"
                        variant="secondary"
                        size="sm"
                        as={Link}
                    >
                        <Icon aria-hidden={true} svgPath={mdiCog} /> Edit
                    </Button>{' '}
                    <Button
                        className="test-delete-external-service-button"
                        onClick={onDelete}
                        disabled={isDeleting === true}
                        data-tooltip="Delete external service"
                        aria-label="Delete external service"
                        variant="danger"
                        size="sm"
                    >
                        <Icon aria-hidden={true} svgPath={mdiDelete} />
                    </Button>
                </div>
            </div>
            {isErrorLike(isDeleting) && <ErrorAlert className="mt-2" error={isDeleting} />}
        </li>
    )
}
