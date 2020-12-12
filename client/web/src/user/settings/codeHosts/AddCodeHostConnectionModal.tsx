import React, { useState, useCallback } from 'react'
import Dialog from '@reach/dialog'

import ShieldCheckIcon from 'mdi-react/ShieldCheckIcon'
import { Form } from '../../../../../branded/src/components/Form'

/**
 * TODO:
 * 1. Add new service by kind with user's token
 * 2. input validation
 * 3. fix lock icon
 */

export const AddCodeHostConnectionModal: React.FunctionComponent<{
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
            aria-labelledby={`label--add-${kind}-token`}
            onDismiss={onDidCancel}
        >
            <div className="web-content">
                <h3 className="mb-4">Connect with {name}</h3>
                <Form onSubmit={onTokenSubmit}>
                    <div className="form-group mb-4">
                        <label htmlFor="code-host-token">Personal access token</label>
                        <input
                            id="code-host-token"
                            name="code-host-token"
                            type="text"
                            value={token}
                            onChange={onChangeToken}
                            className="form-control"
                            required={true}
                            minLength={1}
                        />
                        <ShieldCheckIcon
                            className=""
                            data-tooltip="Data will be encrypted and will not be visible again."
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
                        <button type="submit" disabled={false} className="btn btn-primary">
                            Add code host connection
                        </button>
                    </div>
                </Form>
            </div>
        </Dialog>
    )
}
