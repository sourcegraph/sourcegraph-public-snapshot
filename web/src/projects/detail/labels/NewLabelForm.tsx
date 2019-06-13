import React, { useCallback, useState } from 'react'
import { map } from 'rxjs/operators'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../backend/graphql'
import { LabelForm, LabelFormData } from './LabelForm'

const createLabel = (input: GQL.ICreateLabelInput): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation CreateLabel($input: CreateLabelInput!) {
                labels {
                    createLabel(input: $input) {
                        id
                    }
                }
            }
        `,
        { input }
    )
        .pipe(
            map(({ data, errors }) => {
                if (!data || !data.labels || !data.labels.createLabel || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
            })
        )
        .toPromise()

interface Props {
    project: Pick<GQL.IProject, 'id'>

    /** Called when the form is dismissed. */
    onDismiss: () => void

    /** Called after the label is created successfully. */
    onLabelCreate: () => void

    className?: string
}

/**
 * A form to create a new label.
 */
export const NewLabelForm: React.FunctionComponent<Props> = ({ project, onDismiss, onLabelCreate, className = '' }) => {
    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(
        async ({ name, color, description }: LabelFormData) => {
            setIsLoading(true)
            try {
                await createLabel({ name, color, description, project: project.id })
                setIsLoading(false)
                onDismiss()
                onLabelCreate()
            } catch (err) {
                setIsLoading(false)
                alert(err.message) // TODO!(sqs)
            }
        },
        [onDismiss, onLabelCreate, project.id]
    )

    return (
        <LabelForm
            onDismiss={onDismiss}
            onSubmit={onSubmit}
            buttonText="Create label"
            isLoading={isLoading}
            className={className}
        />
    )
}
