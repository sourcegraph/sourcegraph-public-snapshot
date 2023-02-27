import React, { useCallback, useState } from 'react'

import { ApolloError } from '@apollo/client'
import { mdiChevronUp, mdiChevronDown, mdiDelete, mdiAlert } from '@mdi/js'

import { logger } from '@sourcegraph/common'
import { Button, Icon, Text, Tooltip, Modal, H3, ErrorAlert, Form } from '@sourcegraph/wildcard'

import { PermissionList } from './Permissions'
import { RoleFields } from '../../../graphql-operations'
import { PermissionsMap, useDeleteRole } from '../backend'
import { LoaderButton } from '../../../components/LoaderButton'

import styles from './Roles.module.scss'

interface RoleNodeProps {
    node: RoleFields
    afterDelete: () => void
    allPermissions: PermissionsMap
}

export const RoleNode: React.FunctionComponent<RoleNodeProps> = ({ node, afterDelete, allPermissions }) => {
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    const [showConfirmDeleteModal, setShowConfirmDeleteModal] = useState<boolean>(false)

    const [deleteRole, { loading, error }] = useDeleteRole()

    const toggleIsExpanded = useCallback<React.MouseEventHandler<HTMLButtonElement>>(
        event => {
            event.preventDefault()
            setIsExpanded(!isExpanded)
        },
        [isExpanded]
    )
    const onDelete = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()

            try {
                await deleteRole({ variables: { role: node.id } })
                closeModal()
                afterDelete()
            } catch (error) {
                logger.error(error)
            }
        },
        [deleteRole, name, afterDelete]
    )
    const openModal = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setShowConfirmDeleteModal(true)
    }, [])
    const closeModal = useCallback(() => {
        setShowConfirmDeleteModal(false)
    }, [])

    return (
        <li className={styles.roleNode}>
            {showConfirmDeleteModal && (
                <ConfirmDeleteRoleModal
                    onCancel={closeModal}
                    role={node}
                    onConfirm={onDelete}
                    loading={loading}
                    error={error}
                />
            )}
            <Button
                variant="icon"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} />
            </Button>

            <div className="d-flex align-items-center">
                <Text className="font-weight-bold m-0">{node.name}</Text>

                {node.system && (
                    <Tooltip
                        content="System roles are required by Sourcegraph instance. They cannot be deleted."
                        placement="topStart"
                    >
                        <Text className={styles.roleNodeSystemText}>System</Text>
                    </Tooltip>
                )}
            </div>

            {!node.system && (
                <Tooltip content={node.system ? 'System roles cannot be deleted.' : 'Delete this role.'}>
                    <Button aria-label="Delete" onClick={openModal} disabled={loading} variant="danger" size="sm">
                        <Icon aria-hidden={true} svgPath={mdiDelete} />
                    </Button>
                </Tooltip>
            )}

            {isExpanded ? (
                <div className={styles.roleNodePermissions}>
                    <PermissionList role={node} allPermissions={allPermissions} />
                </div>
            ) : (
                <span />
            )}
        </li>
    )
}

interface ConfirmDeleteRoleModalProps {
    onCancel: () => void
    onConfirm: (event: React.FormEvent) => void
    role: RoleFields
    error: ApolloError | undefined
    loading: boolean
}

const ConfirmDeleteRoleModal: React.FunctionComponent<React.PropsWithChildren<ConfirmDeleteRoleModalProps>> = ({
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
