import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import DeleteIcon from 'mdi-react/DeleteIcon'
import React, { useCallback, useState } from 'react'
import { map } from 'rxjs/operators'
import { NotificationType } from '../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../backend/graphql'

const deleteProject = (args: GQL.IDeleteProjectOnProjectsMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation DeleteProject($project: ID!) {
                projects {
                    deleteProject(project: $project) {
                        alwaysNil
                    }
                }
            }
        `,
        args
    )
        .pipe(
            map(({ data, errors }) => {
                if (!data || !data.projects || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
            })
        )
        .toPromise()

interface Props {
    project: Pick<GQL.IProject, 'id'>
    onDelete: () => void
    className?: string
    buttonClassName?: string
    extensionsController: {
        services: {
            notifications: {
                showMessages: Pick<
                    ExtensionsControllerProps<
                        'services'
                    >['extensionsController']['services']['notifications']['showMessages'],
                    'next'
                >
            }
        }
    }
}

/**
 * A button that permanently deletes a project.
 */
export const ProjectDeleteButton: React.FunctionComponent<Props> = ({
    project,
    onDelete,
    className = '',
    buttonClassName = 'btn-link text-decoration-none',
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            if (!confirm('Are you sure? Deleting will remove all data associated with the project.')) {
                return
            }
            setIsLoading(true)
            try {
                await deleteProject({ project: project.id })
                setIsLoading(false)
                onDelete()
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error deleting project: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, onDelete, project.id]
    )
    return (
        <button type="button" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading ? <LoadingSpinner className="icon-inline" /> : <DeleteIcon className="icon-inline" />} Delete
        </button>
    )
}
