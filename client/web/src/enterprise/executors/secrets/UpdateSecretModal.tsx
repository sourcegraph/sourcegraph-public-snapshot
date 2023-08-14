import React, { useCallback, useState } from 'react'

import { logger } from '@sourcegraph/common'
import { Button, Modal, Input, H3, Text, Alert, Link, ErrorAlert, Form } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../components/LoaderButton'
import type { ExecutorSecretFields } from '../../../graphql-operations'

import { useUpdateExecutorSecret } from './backend'

interface UpdateSecretModalProps {
    secret: ExecutorSecretFields

    onCancel: () => void
    afterUpdate: () => void
}

export const UpdateSecretModal: React.FunctionComponent<React.PropsWithChildren<UpdateSecretModalProps>> = ({
    secret,
    onCancel,
    afterUpdate,
}) => {
    const labelId = 'updateSecret'

    const [value, setValue] = useState<string>('')
    const onChangeValue = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setValue(event.target.value)
    }, [])

    const [updateExecutorSecret, { loading, error }] = useUpdateExecutorSecret()

    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()

            try {
                await updateExecutorSecret({
                    variables: {
                        secret: secret.id,
                        scope: secret.scope,
                        value,
                    },
                })

                afterUpdate()
            } catch (error) {
                // Non-request error. API errors will be available under `error` above.
                logger.error(error)
            }
        },
        [updateExecutorSecret, secret.id, secret.scope, value, afterUpdate]
    )
    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Update secret value for {secret.key}</H3>
            <Text>
                Executor secrets are available to executor jobs as environment variables. They will never appear in
                logs.
            </Text>
            {secret.key === 'DOCKER_AUTH_CONFIG' && (
                <Alert variant="info" className="mt-2">
                    This secret value will be used to{' '}
                    <Link
                        to="/help/admin/executors/deploy_executors#using-private-registries"
                        rel="noopener"
                        target="_blank"
                    >
                        configure docker client authentication with private registries
                    </Link>
                    .
                </Alert>
            )}

            {error && <ErrorAlert error={error} />}

            <Form onSubmit={onSubmit}>
                <div className="form-group">
                    <Input
                        id="value"
                        name="value"
                        type="password"
                        autoComplete="off"
                        required={true}
                        spellCheck="false"
                        minLength={1}
                        label="Value"
                        placeholder="******"
                        value={value}
                        onChange={onChangeValue}
                    />
                </div>
                <div className="d-flex justify-content-end">
                    <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                        Cancel
                    </Button>
                    <LoaderButton
                        type="submit"
                        disabled={loading || value.length === 0}
                        variant="primary"
                        loading={loading}
                        alwaysShowLabel={true}
                        label="Update secret"
                    />
                </div>
            </Form>
        </Modal>
    )
}
