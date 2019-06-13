import React, { useCallback, useState } from 'react'
import { map } from 'rxjs/operators'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../backend/graphql'
import { ProjectForm, ProjectFormData } from './ProjectForm'

const updateProject = (input: GQL.IUpdateProjectInput): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation UpdateProject($input: UpdateProjectInput!) {
                projects {
                    updateProject(input: $input) {
                        id
                    }
                }
            }
        `,
        { input }
    )
        .pipe(
            map(({ data, errors }) => {
                if (!data || !data.projects || !data.projects.updateProject || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
            })
        )
        .toPromise()

interface Props {
    project: Pick<GQL.IProject, 'id'> & ProjectFormData

    /** Called when the form is dismissed. */
    onDismiss: () => void

    /** Called after the project is updated successfully. */
    onProjectUpdate: () => void

    className?: string
}

/**
 * A form to update a project.
 */
export const UpdateProjectForm: React.FunctionComponent<Props> = ({
    project,
    onDismiss,
    onProjectUpdate,
    className = '',
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(
        async ({ name }: ProjectFormData) => {
            setIsLoading(true)
            try {
                await updateProject({ id: project.id, name })
                setIsLoading(false)
                onDismiss()
                onProjectUpdate()
            } catch (err) {
                setIsLoading(false)
                alert(err.message) // TODO!(sqs)
            }
        },
        [onDismiss, onProjectUpdate, project.id]
    )

    return (
        <ProjectForm
            initialValue={project}
            onDismiss={onDismiss}
            onSubmit={onSubmit}
            buttonText="Save changes"
            isLoading={isLoading}
            className={className}
        />
    )
}
