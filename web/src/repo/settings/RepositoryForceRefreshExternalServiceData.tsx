import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import SyncIcon from 'mdi-react/SyncIcon'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../backend/graphql'

const forceRefreshRepositoryThreads = (args: GQL.IForceRefreshRepositoryThreadsOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation ForceRefreshRepositoryThreads($repository: ID!) {
                forceRefreshRepositoryThreads(repository: $repository) {
                    id
                }
            }
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(undefined)
        )
        .toPromise()

interface Props extends ExtensionsControllerNotificationProps {
    repository: Pick<GQL.IRepository, 'id'>
    className?: string
    buttonClassName?: string
}

/**
 * A button that force-refreshes a repository's data from its external services.
 */
export const RepositoryForceRefreshExternalServiceDataButton: React.FunctionComponent<Props> = ({
    repository,
    className = '',
    buttonClassName = 'btn-link text-decoration-none',
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                await forceRefreshRepositoryThreads({ repository: repository.id })
                setIsLoading(false)
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error force-refreshing repository: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, repository.id]
    )
    return (
        <button type="button" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading ? <LoadingSpinner className="icon-inline" /> : <SyncIcon className="icon-inline" />} Refresh
            external service data
        </button>
    )
}
