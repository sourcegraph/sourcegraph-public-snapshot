import * as H from 'history'
import DeleteIcon from 'mdi-react/DeleteIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import UserIcon from 'mdi-react/UserIcon'
import React, { useCallback, useState } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ListExternalServiceFields } from '../../graphql-operations'
import { refreshSiteFlags } from '../../site/backend'
import { ErrorAlert } from '../alerts'

import { deleteExternalService } from './backend'

export interface ExternalServiceNodeProps {
    node: ListExternalServiceFields
    onDidUpdate: () => void
    history: H.History
    routingPrefix: string
    afterDeleteRoute: string
}

export const ExternalServiceNode: React.FunctionComponent<ExternalServiceNodeProps> = ({
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
                            <UserIcon className="icon-inline" />
                            <Link to={node.namespace.url}>{node.namespace.namespaceName}</Link>{' '}
                        </>
                    )}
                    {node.displayName}
                </div>
                <div>
                    <Link
                        className="btn btn-secondary btn-sm test-edit-external-service-button"
                        to={`${routingPrefix}/external-services/${node.id}`}
                        data-tooltip="External service settings"
                    >
                        <SettingsIcon className="icon-inline" /> Edit
                    </Link>{' '}
                    <button
                        type="button"
                        className="btn btn-sm btn-danger test-delete-external-service-button"
                        onClick={onDelete}
                        disabled={isDeleting === true}
                        data-tooltip="Delete external service"
                    >
                        <DeleteIcon className="icon-inline" />
                    </button>
                </div>
            </div>
            {isErrorLike(isDeleting) && <ErrorAlert className="mt-2" error={isDeleting} />}
        </li>
    )
}
