import React, { useCallback, useState } from 'react'

import classNames from 'classnames'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { logger } from '@sourcegraph/common'
import { Button, Modal, Link, Code, Label, Text, Input } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../components/LoaderButton'
import { ExternalServiceKind, Scalars } from '../../../graphql-operations'

import { useCreateBatchChangesCredential } from './backend'
import { ModalHeader } from './ModalHeader'

import styles from './AddSecretModal.module.scss'

export interface AddSecretModalProps {
    onCancel: () => void
    afterCreate: () => void
    userID: Scalars['ID'] | null
    externalServiceKind: ExternalServiceKind
    externalServiceURL: string
    requiresSSH: boolean
    requiresUsername: boolean
}

const HELP_TEXT_LINK_URL = 'https://docs.sourcegraph.com/batch_changes/quickstart#configure-code-host-credentials'

export const AddSecretModal: React.FunctionComponent<React.PropsWithChildren<AddSecretModalProps>> = ({
    onCancel,
    afterCreate,
    userID,
    externalServiceKind,
    externalServiceURL,
    requiresSSH,
    requiresUsername,
}) => {
    const labelId = 'addCredential'
    const [credential, setCredential] = useState<string>('')
    const [sshPublicKey, setSSHPublicKey] = useState<string>()
    const [username, setUsername] = useState<string>('')

    const onChangeCredential = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setCredential(event.target.value)
    }, [])

    const onChangeUsername = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setUsername(event.target.value)
    }, [])

    const [createBatchChangesCredential, { loading, error }] = useCreateBatchChangesCredential()

    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()

            try {
                const { data } = await createBatchChangesCredential({
                    variables: {
                        user: userID,
                        credential,
                        username: requiresUsername ? username : null,
                        externalServiceKind,
                        externalServiceURL,
                    },
                })

                afterCreate()
            } catch (error) {
                logger.error(error)
            }
        },
        [
            createBatchChangesCredential,
            userID,
            credential,
            requiresUsername,
            username,
            externalServiceKind,
            externalServiceURL,
            requiresSSH,
            afterCreate,
        ]
    )

    const patLabel =
        externalServiceKind === ExternalServiceKind.BITBUCKETCLOUD ? 'App password' : 'Personal access token'

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <div className="test-add-credential-modal">
                <ModalHeader
                    id={labelId}
                    externalServiceKind={externalServiceKind}
                    externalServiceURL={externalServiceURL}
                />
                {requiresSSH && (
                    <div className="d-flex w-100 justify-content-between mb-4">
                        <div className="flex-grow-1 mr-2">
                            <Text className={classNames('mb-0 py-2', step === 'get-ssh-key' && 'text-muted')}>
                                1. Add token
                            </Text>
                            <div
                                className={classNames(
                                    styles.addSecretModalModalStepRuler,
                                    styles.addSecretModalModalStepRulerPurple
                                )}
                            />
                        </div>
                        <div className="flex-grow-1 ml-2">
                            <Text className={classNames('mb-0 py-2', step === 'add-token' && 'text-muted')}>
                                2. Get SSH Key
                            </Text>
                            <div
                                className={classNames(
                                    styles.addSecretModalModalStepRuler,
                                    step === 'add-token' && styles.addSecretModalModalStepRulerGray,
                                    step === 'get-ssh-key' && styles.addSecretModalModalStepRulerBlue
                                )}
                            />
                        </div>
                    </div>
                )}
                {error && <ErrorAlert error={error} />}
                <Form onSubmit={onSubmit}>
                    <div className="form-group">
                        {requiresUsername && (
                            <>
                                <Input
                                    id="username"
                                    name="username"
                                    autoComplete="off"
                                    inputClassName="mb-2"
                                    className="mb-0"
                                    required={true}
                                    spellCheck="false"
                                    minLength={1}
                                    value={username}
                                    onChange={onChangeUsername}
                                    label="Username"
                                />
                            </>
                        )}
                        <Label htmlFor="token">{patLabel}</Label>
                        <Input
                            id="token"
                            name="token"
                            type="password"
                            autoComplete="off"
                            data-testid="test-add-credential-modal-input"
                            required={true}
                            spellCheck="false"
                            minLength={1}
                            value={credential}
                            onChange={onChangeCredential}
                        />
                        <Text className="form-text">
                            <Link
                                to={HELP_TEXT_LINK_URL}
                                rel="noreferrer noopener"
                                target="_blank"
                                aria-label={`Follow our docs to learn how to create a new ${patLabel.toLocaleLowerCase()} on this code host`}
                            >
                                Create a new {patLabel.toLocaleLowerCase()}
                            </Link>{' '}
                            {scopeRequirements[externalServiceKind]}
                        </Text>
                    </div>
                    <div className="d-flex justify-content-end">
                        <Button
                            disabled={loading}
                            className="mr-2"
                            onClick={onCancel}
                            outline={true}
                            variant="secondary"
                        >
                            Cancel
                        </Button>
                        <LoaderButton
                            type="submit"
                            disabled={loading || credential.length === 0}
                            className="test-add-credential-modal-submit"
                            variant="primary"
                            loading={loading}
                            alwaysShowLabel={true}
                            label={requiresSSH ? 'Next' : 'Add credential'}
                        />
                    </div>
                </Form>
            </div>
        </Modal>
    )
}
