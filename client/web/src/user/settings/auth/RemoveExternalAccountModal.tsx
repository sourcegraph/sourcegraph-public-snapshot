import Dialog from '@reach/dialog'
import React, { useCallback, useState } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { gql, dataOrThrowErrors } from '@sourcegraph/shared/src/graphql/graphql'
import { asError, ErrorLike } from '@sourcegraph/shared/src/util/errors'

import { requestGraphQL } from '../../../backend/graphql'
import { Scalars, DeleteExternalAccountResult, DeleteExternalAccountVariables } from '../../../graphql-operations'

const deleteUserExternalAccount = async (externalAccount: Scalars['ID']): Promise<void> => {
    dataOrThrowErrors(
        await requestGraphQL<DeleteExternalAccountResult, DeleteExternalAccountVariables>(
            gql`
                mutation DeleteExternalAccount($externalAccount: ID!) {
                    deleteExternalAccount(externalAccount: $externalAccount) {
                        alwaysNil
                    }
                }
            `,
            { externalAccount }
        ).toPromise()
    )
}

export const RemoveExternalAccountModal: React.FunctionComponent<{
    id: Scalars['ID']
    name: string

    onDidRemove: (id: string, name: string) => void
    onDidCancel: () => void
    onDidError: (error: ErrorLike) => void
}> = ({ id, name, onDidRemove, onDidCancel, onDidError }) => {
    const [isLoading, setIsLoading] = useState(false)

    const onAccountRemove = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async event => {
            event.preventDefault()
            setIsLoading(true)

            try {
                await deleteUserExternalAccount(id)
                onDidRemove(id, name)
            } catch (error) {
                setIsLoading(false)
                onDidError(asError(error))
                onDidCancel()
            }
        },
        [id, name, onDidRemove, onDidError, onDidCancel]
    )

    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            aria-labelledby={`heading--disconnect-${name}`}
            aria-describedby={`description--disconnect-${name}`}
            onDismiss={onDidCancel}
        >
            <h3 id={`heading--disconnect-${name}`} className="text-danger mb-4">
                Disconnect {name}?
            </h3>
            <Form onSubmit={onAccountRemove}>
                <div id={`description--disconnect-${name}`} className="form-group mb-4">
                    You are about to remove the sign in connection with {name}. After removing it, you won’t be able to
                    use {name} to sign in to Sourcegraph.
                </div>
                <div className="d-flex justify-content-end">
                    <button
                        type="button"
                        disabled={isLoading}
                        className="btn btn-outline-secondary mr-2"
                        onClick={onDidCancel}
                    >
                        Cancel
                    </button>
                    <button type="submit" disabled={isLoading} className="btn btn-danger">
                        Yes, disconnect {name}
                    </button>
                </div>
            </Form>
        </Dialog>
    )
}
