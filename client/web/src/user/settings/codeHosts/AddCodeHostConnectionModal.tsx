import React, { useState, useCallback } from 'react'
import Dialog from '@reach/dialog'
import ShieldCheckIcon from 'mdi-react/ShieldCheckIcon'

import { Form } from '../../../../../branded/src/components/Form'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { addExternalService } from '../../../components/externalServices/backend'
import { ExternalServiceKind } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'
import { defaultExternalServices } from '../../../components/externalServices/externalServices'

interface CodeHostConfig {
    url: string
    token: string
}

export const AddCodeHostConnectionModal: React.FunctionComponent<{
    name: string
    kind: ExternalServiceKind
    onDidAdd: () => void
    onDidCancel: () => void
    onDidError: (error: ErrorLike) => void

    hintFragment?: React.ReactFragment
}> = ({ onDidAdd, onDidCancel, onDidError, name, kind, hintFragment }) => {
    const [token, setToken] = useState<string>('')
    const [isLoading, setIsLoading] = useState(false)

    const onChangeToken: React.ChangeEventHandler<HTMLInputElement> = event => setToken(event.target.value)

    const onTokenSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async event => {
            event.preventDefault()
            setIsLoading(true)
            try {
                if (token) {
                    const { defaultConfig } = defaultExternalServices[kind]
                    const config: CodeHostConfig = JSON.parse(defaultConfig)
                    config.token = token
                    const finalConfig = JSON.stringify(config)

                    await addExternalService({ input: { kind, displayName: name, config: finalConfig } }, eventLogger)
                    onDidAdd()
                }
            } catch (error) {
                setIsLoading(false)
                onDidError(asError(error))
            }
            // } finally {
            //     setIsLoading(false)
            //     onDidCancel()
            // }
        },
        [token, kind, name, onDidAdd, onDidError, onDidCancel]
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
                            className="form-control pr-4"
                            required={true}
                            minLength={1}
                        />
                        <ShieldCheckIcon
                            className="icon-inline add-user-code-hosts-page__icon--inside add-user-code-hosts-page__icon--muted"
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
                        <button type="submit" disabled={!token || isLoading} className="btn btn-primary">
                            Add code host connection
                        </button>
                    </div>
                </Form>
            </div>
        </Dialog>
    )
}
