import React, { useCallback, useState } from 'react'

import { logger } from '@sourcegraph/common'
import { Button, Modal, Input, H3, Text, Link, ErrorAlert, Form } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../components/LoaderButton'
import { useCreateRole } from './backend'

export interface AddRoleModalProps {
    onCancel: () => void
    afterCreate: () => void

    /** For testing only */
    initialKey?: string
}

export const AddRoleModal: React.FunctionComponent<React.PropsWithChildren<AddRoleModalProps>> = ({
    onCancel,
    afterCreate,
    initialKey = '',
}) => {
    const labelId = 'addRole'

    const [name, setName] = useState<string>(initialKey)
    const onChangeName = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setName(event.target.value)
    }, [])

    const [createRole, { loading, error }] = useCreateRole()

    const onSubmit = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()

            try {
                await createRole({ variables: { name } })
                afterCreate()
            } catch (error) {
                logger.error(error)
            }
        },
        [createRole, name, afterCreate]
    )

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelId}>
            <H3 id={labelId}>Add new role</H3>
            <Text>
                Enter a descriptive name for the role to be created. The name should encapsulate the job function or attributes of users that'll be assigned this role.
            </Text>
            {error && <ErrorAlert error={error} />}
            <Form onSubmit={onSubmit}>
                <div className="form-group">
                    <Input
                        id="name"
                        name="name"
                        autoComplete="off"
                        inputClassName="mb-2"
                        className="mb-0"
                        required={true}
                        spellCheck="false"
                        minLength={1}
                        value={name}
                        onChange={onChangeName}
                        pattern="^[A-Za-z0-9_-]*$"
                    />
                </div>
                <div className="d-flex justify-content-end">
                    <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                        Cancel
                    </Button>
                    <LoaderButton
                        type="submit"
                        disabled={loading || name.length === 0}
                        variant="primary"
                        loading={loading}
                        alwaysShowLabel={true}
                        label="Add role"
                    />
                </div>
            </Form>
        </Modal>
    )
}
