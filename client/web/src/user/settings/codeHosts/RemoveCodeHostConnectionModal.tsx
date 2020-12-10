import React, { useCallback } from 'react'
import Dialog from '@reach/dialog'

import { Form } from '../../../../../branded/src/components/Form'

export const RemoveCodeHostConnectionModal: React.FunctionComponent<{
    name: string
    kind: string
    repoCount: string
    onDidRemove: () => void
    onDidCancel: () => void
}> = ({ onDidRemove, onDidCancel, name, kind, repoCount }) => {
    const onConnectionRemove = useCallback<React.FormEventHandler<HTMLFormElement>>(
        event => {
            event.preventDefault()
            onDidRemove()
        },
        [onDidRemove]
    )

    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            aria-labelledby={`label--remove-${kind}-token`}
            onDismiss={onDidCancel}
        >
            <div className="web-content">
                <h3 className="text-danger mb-4">Remove connection with {name}?</h3>
                <Form onSubmit={onConnectionRemove}>
                    <div className="form-group mb-4">
                        There are {repoCount} repositories synced to Sourcegraph from {name}. If the connection with
                        {name} is removed, these repositories will no longer be synced with Sourcegraph.
                    </div>
                    <div className="d-flex justify-content-end">
                        <button
                            type="button"
                            disabled={false}
                            className="btn btn-outline-secondary mr-2"
                            onClick={onDidCancel}
                        >
                            Cancel
                        </button>
                        <button type="submit" disabled={false} className="btn btn-danger">
                            Yes, remove connection
                        </button>
                    </div>
                </Form>
            </div>
        </Dialog>
    )
}
