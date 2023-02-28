import React from 'react'

import { ApolloError } from '@apollo/client'
import { mdiAlert } from '@mdi/js'

import { Button, Icon, Text, Modal, H3, ErrorAlert, Form } from '@sourcegraph/wildcard'

import { RoleFields } from '../../../graphql-operations'
import { LoaderButton } from '../../../components/LoaderButton'

interface ConfirmDeleteRoleModalProps {
    onCancel: () => void
    onConfirm: (event: React.FormEvent) => void
    role: RoleFields
    error: ApolloError | undefined
    loading: boolean
}

export const ConfirmDeleteRoleModal: React.FunctionComponent<React.PropsWithChildren<ConfirmDeleteRoleModalProps>> = ({
    onCancel,
    onConfirm,
    role,
    loading,
    error,
}) => {
    const labelID = 'DeleteRole'

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelID}>
            <div className="d-flex align-items-center mb-2">
                <Icon className="icon mr-1" svgPath={mdiAlert} inline={false} aria-hidden={true} />{' '}
                <H3 id={labelID} className="mb-0">
                    Delete role
                </H3>
            </div>
            <Text>
                Once deleted, all users assigned the <span className="font-weight-bold">"{role.name}"</span> role will
                lose access to the permissions associated with the role.
            </Text>
            {error && <ErrorAlert error={error} />}
            <Form onSubmit={onConfirm}>
                <div className="d-flex justify-content-end">
                    <Button disabled={loading} className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                        Cancel
                    </Button>
                    <LoaderButton
                        type="submit"
                        disabled={loading}
                        variant="primary"
                        loading={loading}
                        alwaysShowLabel={true}
                        label="Delete"
                    />
                </div>
            </Form>
        </Modal>
    )
}
