import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'
import { ChangesetForm, ChangesetFormData } from '../form/ChangesetForm'

const updateChangeset = (input: GQL.IUpdateChangesetInput): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation UpdateChangeset($input: UpdateChangesetInput!) {
                updateChangeset(input: $input) {
                    id
                }
            }
        `,
        { input }
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(void 0)
        )
        .toPromise()

interface Props {
    changeset: Pick<GQL.IChangeset, 'id'> & ChangesetFormData

    /** Called when the form is dismissed. */
    onDismiss: () => void

    /** Called after the changeset is updated successfully. */
    onChangesetUpdate: () => void

    className?: string
}

/**
 * A form to edit a changeset.
 */
export const EditChangesetForm: React.FunctionComponent<Props> = ({
    changeset,
    onDismiss,
    onChangesetUpdate,
    className = '',
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(
        async ({ name }: ChangesetFormData) => {
            setIsLoading(true)
            try {
                await updateChangeset({ id: changeset.id, name })
                setIsLoading(false)
                onDismiss()
                onChangesetUpdate()
            } catch (err) {
                setIsLoading(false)
                alert(err.message) // TODO!(sqs)
            }
        },
        [onDismiss, onChangesetUpdate, changeset.id]
    )

    return (
        <ChangesetForm
            initialValue={changeset}
            onDismiss={onDismiss}
            onSubmit={onSubmit}
            buttonText="Save changes"
            isLoading={isLoading}
            className={className}
        />
    )
}
