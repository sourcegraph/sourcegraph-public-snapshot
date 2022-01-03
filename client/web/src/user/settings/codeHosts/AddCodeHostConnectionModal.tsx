import Dialog from '@reach/dialog'
import classNames from 'classnames'
import React, { useState, useCallback } from 'react'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, ErrorLike } from '@sourcegraph/common'

import { addExternalService } from '../../../components/externalServices/backend'
import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import { LoaderButton } from '../../../components/LoaderButton'
import { Scalars, ExternalServiceKind, ListExternalServiceFields } from '../../../graphql-operations'
import { eventLogger } from '../../../tracking/eventLogger'

import styles from './AddCodeHostConnectionModal.module.scss'
import { EncryptedDataIcon } from './components/EncryptedDataIcon'
import { getMachineUserFragment } from './modalHints'

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
    ownerID: Scalars['ID']
    serviceName: string
    serviceKind: ExternalServiceKind
    onDidAdd: (service: ListExternalServiceFields) => void
    onDidCancel: () => void
    onDidError: (error: ErrorLike) => void

    hintFragment?: React.ReactFragment
}> = ({ ownerID, serviceName, serviceKind, hintFragment, onDidAdd, onDidCancel, onDidError }) => {
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
        <Dialog
            className={classNames('modal-body modal-body--top-third p-4 rounded border', styles.modalPlain)}
            aria-labelledby={`heading--connect-with-${serviceName}`}
            onDismiss={onDidCancel}
        >
            <div className="web-content">
                <h3 id={`heading--connect-with-${serviceName}`} className="mb-4">
                    Connect with {serviceName}
                </h3>
                <Form onSubmit={onTokenSubmit}>
                    <div className="form-group mb-4">
                        {didAckMachineUserHint ? (
                            <>
                                <label htmlFor="code-host-token">Access token</label>
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
                                    <small>
                                        <EncryptedDataIcon />
                                    </small>
                                </div>
                                <p className="mt-1">{hintFragment}</p>
                            </>
                        ) : (
                            machineUserFragment
                        )}
                    </div>
                    <div className="d-flex justify-content-end">
                        <button type="button" className="btn btn-outline-secondary mr-2" onClick={onDidCancel}>
                            Cancel
                        </button>
                        {didAckMachineUserHint ? (
                            <LoaderButton
                                type="submit"
                                className="btn btn-primary"
                                loading={isLoading}
                                disabled={!token || isLoading}
                                label="Add code host connection"
                                alwaysShowLabel={true}
                            />
                        ) : (
                            <button
                                type="button"
                                className="btn btn-secondary"
                                onClick={() => setAckMachineUserHint(previousAckStatus => !previousAckStatus)}
                            >
                                I understand, continue
                            </button>
                        )}
                    </div>
                </Form>
            </div>
        </Dialog>
    )
}
