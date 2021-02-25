import React, { useState, useCallback } from 'react'
import Dialog from '@reach/dialog'
import ShieldCheckIcon from 'mdi-react/ShieldCheckIcon'

import { Form } from '../../../../../branded/src/components/Form'
import { LoaderButton } from '../../../components/LoaderButton'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { updateExternalService } from '../../../components/externalServices/backend'
import { Scalars, ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'

interface CodeHostConfig {
    url: string
    token: string
}

const updateConfigToken = (config: string, token: string): string => {
    const updatedConfig = JSON.parse(config) as CodeHostConfig
    updatedConfig.token = token
    return JSON.stringify(updatedConfig, null, 2)
}

export const UpdateCodeHostConnectionModal: React.FunctionComponent<{
    serviceId: Scalars['ID']
    serviceConfig: string
    name: string
    kind: ExternalServiceKind
    onDidUpdate: (service: ListExternalServiceFields) => void
    onDidCancel: () => void
    onDidError: (error: ErrorLike) => void

    hintFragment?: React.ReactFragment
}> = ({ serviceId, serviceConfig, name, hintFragment, onDidUpdate, onDidCancel, onDidError }) => {
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
                    const config = updateConfigToken(serviceConfig, token)

                    const { webhookURL, ...newService } = await updateExternalService({
                        input: { id: serviceId, config },
                    })

                    onDidUpdate(newService)
                    onDidCancel()
                }
            } catch (error) {
                handleError(error)
            }
        },
        [serviceConfig, serviceId, token, onDidCancel, handleError, onDidUpdate]
    )

    return (
        <Dialog
            className="modal-body modal-body--top-third p-4 rounded border"
            aria-labelledby={`heading--update-${name}-code-host`}
            onDismiss={onDidCancel}
        >
            <div className="web-content">
                <h3 id={`heading--update-${name}-code-host`} className="mb-4">
                    Update {name} token
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
                            label="Update token"
                            alwaysShowLabel={true}
                        />
                    </div>
                </Form>
            </div>
        </Dialog>
    )
}
