import React, { useCallback, useState } from 'react'

import * as H from 'history'
import AccountIcon from 'mdi-react/AccountIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'

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
                            <Icon as={AccountIcon} />
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
                        <Icon as={SettingsIcon} /> Edit
                    </Button>{' '}
                    <Button
                        className="test-delete-external-service-button"
                        onClick={onDelete}
                        disabled={isDeleting === true}
                        data-tooltip="Delete external service"
                        variant="danger"
                        size="sm"
                    >
                        <Icon as={DeleteIcon} />
                    </Button>
                </div>
            </div>
            {isErrorLike(isDeleting) && <ErrorAlert className="mt-2" error={isDeleting} />}
        </li>
    )
}
