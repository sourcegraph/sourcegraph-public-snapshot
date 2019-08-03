import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import DeleteIcon from 'mdi-react/DeleteIcon'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'

const deleteChangeset = (args: GQL.IDeleteChangesetOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation DeleteChangeset($changeset: ID!) {
                deleteChangeset(changeset: $changeset) {
                    alwaysNil
                }
            }
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(void 0)
        )
        .toPromise()

interface Props extends ExtensionsControllerNotificationProps {
    changeset: Pick<GQL.IChangeset, 'id'>
    onDelete?: () => void
    className?: string
    buttonClassName?: string
}

/**
 * A button that permanently deletes a changeset.
 */
export const ChangesetDeleteButton: React.FunctionComponent<Props> = ({
    changeset,
    onDelete,
    className = '',
    buttonClassName = 'btn-link text-decoration-none',
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            if (!confirm('Are you sure? Deleting will remove all data associated with the changeset.')) {
                return
            }
            setIsLoading(true)
            try {
                await deleteChangeset({ changeset: changeset.id })
                setIsLoading(false)
                if (onDelete) {
                    onDelete()
                }
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error deleting changeset: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, onDelete, changeset.id]
    )
    return (
        <button type="button" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading ? <LoadingSpinner className="icon-inline" /> : <DeleteIcon className="icon-inline" />} Delete
            changeset
        </button>
    )
}
