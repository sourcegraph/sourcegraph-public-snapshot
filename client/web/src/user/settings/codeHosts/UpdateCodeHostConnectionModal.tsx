import React, { useState, useCallback } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, ErrorLike } from '@sourcegraph/common'
import { Button, Modal, Link, Alert, H3, Input, Text } from '@sourcegraph/wildcard'

import { updateExternalService } from '../../../components/externalServices/backend'
import { LoaderButton } from '../../../components/LoaderButton'
import { Scalars, ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'

import { EncryptedDataIcon } from './components/EncryptedDataIcon'
import { getMachineUserFragment } from './modalHints'

interface CodeHostConfig {
    url: string
    token: string
}

const updateConfigToken = (config: string, token: string): string => {
    const updatedConfig = JSON.parse(config) as CodeHostConfig
    updatedConfig.token = token
    return JSON.stringify(updatedConfig, null, 2)
}

export const UpdateCodeHostConnectionModal: React.FunctionComponent<
    React.PropsWithChildren<{
        serviceID: Scalars['ID']
        serviceConfig: string
        serviceName: string
        orgName: string
        kind: ExternalServiceKind
        onDidUpdate: (service: ListExternalServiceFields) => void
        onDidCancel: () => void
        onDidError: (error: ErrorLike) => void

        hintFragment?: React.ReactFragment
    }>
> = ({ serviceID, serviceConfig, serviceName, hintFragment, onDidUpdate, onDidCancel, onDidError }) => {
    const [token, setToken] = useState<string>('')
    const [isLoading, setIsLoading] = useState(false)
    const [didAckMachineUserHint, setAckMachineUserHint] = useState(false)

    const onChangeToken: React.ChangeEventHandler<HTMLInputElement> = event => setToken(event.target.value)
    const machineUserFragment = getMachineUserFragment(serviceName)

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
                        input: { id: serviceID, config },
                    })

                    onDidUpdate(newService)
                    onDidCancel()
                }
            } catch (error) {
                handleError(error)
            }
        },
        [serviceConfig, serviceID, token, onDidCancel, handleError, onDidUpdate]
    )

    return (
        <Modal aria-labelledby={`heading--update-${serviceName}-code-host`} onDismiss={onDidCancel}>
            <div className="web-content">
                <H3 id={`heading--update-${serviceName}-code-host`} className="mb-4">
                    Update {serviceName} connection
                </H3>
                <Form onSubmit={onTokenSubmit}>
                    <div className="form-group mb-4">
                        <Alert variant="info" role="alert">
                            Updating the access token may affect which repositories can be synced with Sourcegraph.{' '}
                            <Link
                                to="https://docs.sourcegraph.com/cloud/access_tokens_on_cloud"
                                target="_blank"
                                rel="noopener noreferrer"
                                className="font-weight-normal"
                            >
                                Learn more
                            </Link>
                            .
                        </Alert>
                        {didAckMachineUserHint ? (
                            <>
                                {' '}
                                <div className="position-relative">
                                    <Input
                                        id="code-host-token"
                                        name="code-host-token"
                                        value={token}
                                        onChange={onChangeToken}
                                        inputClassName="pr-4"
                                        autoComplete="off"
                                        className="mb-0"
                                        label="Access token"
                                        inputSymbol={<EncryptedDataIcon />}
                                    />
                                </div>
                                <Text className="mt-1">{hintFragment}</Text>
                            </>
                        ) : (
                            machineUserFragment
                        )}
                    </div>
                    <div className="d-flex justify-content-end">
                        <Button className="mr-2" onClick={onDidCancel} outline={true} variant="secondary">
                            Cancel
                        </Button>

                        {didAckMachineUserHint ? (
                            <LoaderButton
                                type="submit"
                                loading={isLoading}
                                disabled={!token || isLoading}
                                label="Update code host connection"
                                alwaysShowLabel={true}
                                variant="primary"
                            />
                        ) : (
                            <Button
                                onClick={() => setAckMachineUserHint(previousAckStatus => !previousAckStatus)}
                                variant="secondary"
                            >
                                I understand, continue
                            </Button>
                        )}
                    </div>
                </Form>
            </div>
        </Modal>
    )
}
