import React, { useState, useCallback } from 'react'
import Dialog from '@reach/dialog'

import { Form } from '../../../../../branded/src/components/Form'

export const AddCodeHostTokenModal: React.FunctionComponent<{
    name: string
    kind: string
    onDidAdd: (token: string) => void
    onDidCancel: () => void

    hintFragment?: React.ReactFragment
}> = ({ onDidAdd, onDidCancel, name, kind, hintFragment }) => {
    const [token, setToken] = useState<string>()

    const onChangeToken = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => setToken(event.target.value),
        []
    )

    const onTokenSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        event => {
            event.preventDefault()
            if (token) {
                onDidAdd(token)
            }
        },
        [token, onDidAdd]
    )

    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            aria-labelledby={`label--${kind}-token`}
            onDismiss={onDidCancel}
        >
            <div className="web-content test-add-credential-modal">
                <h3 className="mb-4">Connect with {name}</h3>
                <Form onSubmit={onTokenSubmit}>
                    <div className="form-group mb-4">
                        <label htmlFor="token">Personal access token</label>
                        <input
                            id="token"
                            name="token"
                            type=" "
                            className="form-control test-add-credential-modal-input"
                            required={true}
                            minLength={1}
                            value={token}
                            onChange={onChangeToken}
                        />
                        {hintFragment}
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
                        <button
                            type="submit"
                            disabled={false}
                            className="btn btn-primary test-add-credential-modal-submit"
                        >
                            Add code host connection
                        </button>
                    </div>
                </Form>
            </div>
        </Dialog>
    )
}
