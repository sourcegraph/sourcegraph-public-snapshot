import React from 'react'

import { mdiAlert } from '@mdi/js'

import { Button, Icon, Text, Modal, H3, Form } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../components/LoaderButton'
import { RoleFields } from '../../../graphql-operations'

interface ConfirmDeleteRoleModalProps {
    onCancel: () => void
    onConfirm: (event: React.FormEvent) => void
    role: RoleFields
}

export const ConfirmDeleteRoleModal: React.FunctionComponent<React.PropsWithChildren<ConfirmDeleteRoleModalProps>> = ({
    onCancel,
    onConfirm,
    role,
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
                Please confirm that you want to delete the role. Once deleted, all users assigned the{' '}
                <span className="font-weight-bold">"{role.name}"</span> role will lose access to the permissions
                associated with the role.
            </Text>
            <Form onSubmit={onConfirm}>
                <div className="d-flex justify-content-end">
                    <Button className="mr-2" onClick={onCancel} outline={true} variant="secondary">
                        Cancel
                    </Button>
                    <LoaderButton type="submit" variant="danger" alwaysShowLabel={true} label="Delete" />
                </div>
            </Form>
        </Modal>
    )
}
