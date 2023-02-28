import React, { useCallback, useState } from 'react'

import { mdiChevronUp, mdiChevronDown, mdiDelete } from '@mdi/js'

import { logger } from '@sourcegraph/common'
import { Button, Icon, Text, Tooltip } from '@sourcegraph/wildcard'

import { RoleFields } from '../../../graphql-operations'
import { PermissionsMap, useDeleteRole } from '../backend'

import { ConfirmDeleteRoleModal } from './ConfirmDeleteRoleModal'
import { PermissionList } from './Permissions'

import styles from './RoleNode.module.scss'

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
    const openModal = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setShowConfirmDeleteModal(true)
    }, [])
    const closeModal = useCallback(() => {
        setShowConfirmDeleteModal(false)
    }, [])
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
        [deleteRole, name, afterDelete, closeModal, node.id]
    )

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
                        content="System roles are required by Sourcegraph. They cannot be deleted."
                        placement="topStart"
                    >
                        <Text className={styles.roleNodeSystemText}>System</Text>
                    </Tooltip>
                )}
            </div>

            {!node.system && (
                <Tooltip content="Deleting a role is an irreversible action.">
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
