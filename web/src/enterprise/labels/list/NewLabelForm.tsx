import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'
import { LabelForm, LabelFormData } from './LabelForm'

const createLabel = (input: GQL.ICreateLabelInput): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation CreateLabel($input: CreateLabelInput!) {
                createLabel(input: $input) {
                    id
                }
            }
        `,
        { input }
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(undefined)
        )
        .toPromise()

interface Props {
    repository: Pick<GQL.IRepository, 'id'>

    /** Called when the form is dismissed. */
    onDismiss: () => void

    /** Called after the label is created successfully. */
    onLabelCreate: () => void

    className?: string
}

/**
 * A form to create a new label.
 */
export const NewLabelForm: React.FunctionComponent<Props> = ({
    repository,
    onDismiss,
    onLabelCreate,
    className = '',
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(
        async ({ name, color, description }: LabelFormData) => {
            setIsLoading(true)
            try {
                await createLabel({ name, color, description, repository: repository.id })
                setIsLoading(false)
                onDismiss()
                onLabelCreate()
            } catch (err) {
                setIsLoading(false)
                alert(err.message) // TODO!(sqs)
            }
        },
        [onDismiss, onLabelCreate, repository.id]
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
