import React, { useState, useCallback } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, ErrorLike } from '@sourcegraph/common'
import { Button, Modal, H3, Input, Text } from '@sourcegraph/wildcard'

import { addExternalService } from '../../../components/externalServices/backend'
import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import { LoaderButton } from '../../../components/LoaderButton'
import { Scalars, ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'

import { EncryptedDataIcon } from './components/EncryptedDataIcon'
import { getMachineUserFragment } from './modalHints'

import styles from './AddCodeHostConnectionModal.module.scss'

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

export const AddCodeHostConnectionModal: React.FunctionComponent<
    React.PropsWithChildren<{
        ownerID: Scalars['ID']
        serviceName: string
        serviceKind: ExternalServiceKind
        onDidAdd: (service: ListExternalServiceFields) => void
        onDidCancel: () => void
        onDidError: (error: ErrorLike) => void

        hintFragment?: React.ReactFragment
    }>
> = ({ ownerID, serviceName, serviceKind, hintFragment, onDidAdd, onDidCancel, onDidError }) => {
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
                    const config = getServiceConfig(serviceKind, token)

                    const { webhookURL, ...newService } = await addExternalService(
                        { input: { kind: serviceKind, config, displayName: serviceName, namespace: ownerID } },
                        eventLogger
                    )

                    onDidAdd(newService)
                    onDidCancel()
                }
            } catch (error) {
                handleError(error)
            }
        },
        [ownerID, token, serviceKind, serviceName, onDidCancel, handleError, onDidAdd]
    )

    return (
        <Modal
            className={styles.modalPlain}
            aria-labelledby={`heading--connect-with-${serviceName}`}
            onDismiss={onDidCancel}
        >
            <div className="web-content">
                <H3 id={`heading--connect-with-${serviceName}`} className="mb-4">
                    Connect with {serviceName}
                </H3>
                <Form onSubmit={onTokenSubmit}>
                    <div className="form-group mb-4">
                        {didAckMachineUserHint ? (
                            <>
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
                                label="Add code host connection"
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
