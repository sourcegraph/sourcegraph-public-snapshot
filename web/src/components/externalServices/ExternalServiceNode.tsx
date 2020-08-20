import * as H from 'history'
import React, { useCallback } from 'react'
import { useEventObservable } from '../../../../shared/src/util/useObservable'
import { Observable, concat, from } from 'rxjs'
import { switchMap, mapTo, catchError, filter, tap } from 'rxjs/operators'
import { deleteExternalService } from './backend'
import { ErrorLike, asError, isErrorLike } from '../../../../shared/src/util/errors'
import { refreshSiteFlags } from '../../site/backend'
import { Link } from '../../../../shared/src/components/Link'
import SettingsIcon from 'mdi-react/SettingsIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import { ErrorAlert } from '../alerts'
import { ListExternalServiceFields } from '../../graphql-operations'

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
    const [nextDeleteClick, deletedOrError] = useEventObservable(
        useCallback(
            (clicks: Observable<React.MouseEvent>) =>
                clicks.pipe(
                    filter(() => window.confirm(`Delete the external service ${node.displayName}?`)),
                    switchMap(() =>
                        concat(
                            from(deleteExternalService(node.id)).pipe(
                                mapTo(true as const),
                                catchError((error): [ErrorLike] => [asError(error)])
                            )
                        )
                    ),
                    tap(onDidUpdate),
                    tap(deletedOrError => {
                        // eslint-disable-next-line rxjs/no-ignored-subscription
                        refreshSiteFlags().subscribe()
                        if (deletedOrError === true) {
                            history.push(afterDeleteRoute)
                        }
                    })
                ),
            [history, node.displayName, node.id, onDidUpdate, afterDeleteRoute]
        )
    )

    return (
        <li className="external-service-node list-group-item py-2" data-test-external-service-name={node.displayName}>
            <div className="d-flex align-items-center justify-content-between">
                <div>{node.displayName}</div>
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
                        onClick={nextDeleteClick}
                        disabled={deletedOrError === undefined}
                        data-tooltip="Delete external service"
                    >
                        <DeleteIcon className="icon-inline" />
                    </button>
                </div>
            </div>
            {isErrorLike(deletedOrError) && <ErrorAlert className="mt-2" error={deletedOrError} history={history} />}
        </li>
    )
}
