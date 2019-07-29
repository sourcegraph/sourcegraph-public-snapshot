import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'
import { ThreadForm, ThreadFormData } from '../form/ThreadForm'

const updateThread = (input: GQL.IUpdateThreadInput): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation UpdateThread($input: UpdateThreadInput!) {
                updateThread(input: $input) {
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
    thread: Pick<GQL.IThread, 'id'> & ThreadFormData

    /** Called when the form is dismissed. */
    onDismiss: () => void

    /** Called after the thread is updated successfully. */
    onThreadUpdate: () => void

    className?: string
}

/**
 * A form to edit a thread.
 */
export const EditThreadForm: React.FunctionComponent<Props> = ({
    thread,
    onDismiss,
    onThreadUpdate,
    className = '',
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(
        async ({ name }: ThreadFormData) => {
            setIsLoading(true)
            try {
                await updateThread({ id: thread.id, name })
                setIsLoading(false)
                onDismiss()
                onThreadUpdate()
            } catch (err) {
                setIsLoading(false)
                alert(err.message) // TODO!(sqs)
            }
        },
        [onDismiss, onThreadUpdate, thread.id]
    )

    return (
        <ThreadForm
            initialValue={thread}
            onDismiss={onDismiss}
            onSubmit={onSubmit}
            buttonText="Save changes"
            isLoading={isLoading}
            className={className}
        />
    )
}
