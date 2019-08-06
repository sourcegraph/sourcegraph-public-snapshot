import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'
import { IssueForm, IssueFormData } from '../form/IssueForm'

export const updateIssue = (input: GQL.IUpdateIssueInput): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation UpdateIssue($input: UpdateIssueInput!) {
                updateIssue(input: $input) {
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
    issue: Pick<GQL.IIssue, 'id'> & IssueFormData

    /** Called when the form is dismissed. */
    onDismiss: () => void

    /** Called after the issue is updated successfully. */
    onIssueUpdate: () => void

    className?: string
}

/**
 * A form to edit a issue.
 */
export const EditIssueForm: React.FunctionComponent<Props> = ({
    issue,
    onDismiss,
    onIssueUpdate,
    className = '',
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback(
        async ({ title }: IssueFormData) => {
            setIsLoading(true)
            try {
                await updateIssue({ id: issue.id, title })
                setIsLoading(false)
                onDismiss()
                onIssueUpdate()
            } catch (err) {
                setIsLoading(false)
                alert(err.message) // TODO!(sqs)
            }
        },
        [onDismiss, onIssueUpdate, issue.id]
    )

    return (
        <IssueForm
            initialValue={issue}
            onDismiss={onDismiss}
            onSubmit={onSubmit}
            buttonText="Save changes"
            isLoading={isLoading}
            className={className}
        />
    )
}
