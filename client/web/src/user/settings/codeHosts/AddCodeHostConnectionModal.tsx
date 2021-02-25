import React, { useState, useCallback } from 'react'
import Dialog from '@reach/dialog'
import ShieldCheckIcon from 'mdi-react/ShieldCheckIcon'

import { Form } from '../../../../../branded/src/components/Form'
import { LoaderButton } from '../../../components/LoaderButton'
import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { addExternalService } from '../../../components/externalServices/backend'
import { Scalars, ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'

interface CodeHostConfig {
    url: string
    token: string
}

const getServiceConfig = (kind: ExternalServiceKind, token: string): string => {
    const { defaultConfig } = defaultExternalServices[kind]
    const config = JSON.parse(defaultConfig) as CodeHostConfig
    config.token = token
    return JSON.stringify(config, null, 2)
}

export const AddCodeHostConnectionModal: React.FunctionComponent<{
    userID: Scalars['ID']
    name: string
    kind: ExternalServiceKind
    onDidAdd: (service: ListExternalServiceFields) => void
    onDidCancel: () => void
    onDidError: (error: ErrorLike) => void

    hintFragment?: React.ReactFragment
}> = ({ userID, name, kind, hintFragment, onDidAdd, onDidCancel, onDidError }) => {
    const [token, setToken] = useState<string>('')
    const [isLoading, setIsLoading] = useState(false)

    const onChangeToken: React.ChangeEventHandler<HTMLInputElement> = event => setToken(event.target.value)

    const handleError = useCallback(
        (error: ErrorLike | string): void => {
            setIsLoading(false)
            onDidCancel()
            onDidError(asError(error))
        },
        [onDidCancel, onDidError]
    )

    const onTokenSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async event => {
            event.preventDefault()
            setIsLoading(true)

            try {
                if (token) {
                    const config = getServiceConfig(kind, token)

                    const { webhookURL, ...newService } = await addExternalService(
                        { input: { kind, config, displayName: name, namespace: userID } },
                        eventLogger
                    )

                    onDidAdd(newService)
                    onDidCancel()
                }
            } catch (error) {
                handleError(error)
            }
        },
        [userID, token, kind, name, onDidCancel, handleError, onDidAdd]
    )

    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            aria-labelledby={`heading--connect-with-${name}`}
            onDismiss={onDidCancel}
        >
            <div className="web-content">
                <h3 id={`heading--connect-with-${name}`} className="mb-4">
                    Connect with {name}
                </h3>
                <Form onSubmit={onTokenSubmit}>
                    <div className="form-group mb-4">
                        <label htmlFor="code-host-token">Personal access token</label>
                        <div className="position-relative">
                            <input
                                id="code-host-token"
                                name="code-host-token"
                                type="text"
                                value={token}
                                onChange={onChangeToken}
                                className="form-control pr-4"
                                autoComplete="off"
                            />
                            <ShieldCheckIcon
                                className="icon-inline add-user-code-hosts-page__icon--inside text-muted"
                                data-tooltip="Data will be encrypted and will not be visible again."
                            />
                        </div>

                        {hintFragment}
                    </div>
                    <div className="d-flex justify-content-end">
                        <button type="button" className="btn btn-outline-secondary mr-2" onClick={onDidCancel}>
                            Cancel
                        </button>
                        <LoaderButton
                            type="submit"
                            className="btn btn-primary"
                            loading={isLoading}
                            disabled={!token || isLoading}
                            label="Add code host connection"
                            alwaysShowLabel={true}
                        />
                    </div>
                </Form>
            </div>
        </Dialog>
    )
}
