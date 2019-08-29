import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'
import { LabelForm, LabelFormData } from './LabelForm'

const updateLabel = (input: GQL.IUpdateLabelInput): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation UpdateLabel($input: UpdateLabelInput!) {
                updateLabel(input: $input) {
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
    label: Pick<GQL.ILabel, 'id'> & LabelFormData

    /** Called when the form is dismissed. */
    onDismiss: () => void

    /** Called after the label is updated successfully. */
    onLabelUpdate: () => void

    className?: string
}

/**
 * A form to edit a label.
 */
export const EditLabelForm: React.FunctionComponent<Props> = ({ label, onDismiss, onLabelUpdate, className = '' }) => {
    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(
        async ({ name, color, description }: LabelFormData) => {
            setIsLoading(true)
            try {
                await updateLabel({ id: label.id, name, color, description })
                setIsLoading(false)
                onDismiss()
                onLabelUpdate()
            } catch (err) {
                setIsLoading(false)
                alert(err.message) // TODO!(sqs)
            }
        },
        [label.id, onDismiss, onLabelUpdate]
    )

    return (
        <LabelForm
            initialValue={label}
            onDismiss={onDismiss}
            onSubmit={onSubmit}
            buttonText="Save changes"
            isLoading={isLoading}
            className={className}
        />
    )
}
