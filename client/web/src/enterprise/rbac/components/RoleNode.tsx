import React, { useCallback, useMemo, useState } from 'react'

import { mdiChevronUp, mdiChevronDown, mdiDelete } from '@mdi/js'
import { startCase } from 'lodash'

import { Button, Icon, Text, Tooltip, ErrorAlert } from '@sourcegraph/wildcard'

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

    const [deleteRole, { loading, error }] = useDeleteRole(() => {
        closeModal()
        afterDelete()
    }, closeModal)

    const roleName = useMemo(() => {
        const lowerCaseName = node.name.replace(/_/g, ' ').toLowerCase()
        return startCase(lowerCaseName)
    }, [node.name])

    const onDelete = useCallback<React.FormEventHandler>(
        event => {
            event.preventDefault()
            deleteRole({ variables: { role: node.id } })
        },
        [deleteRole, node.id]
    )

    return (
        <li className={styles.roleNode}>
            {showConfirmDeleteModal && (
                <ConfirmDeleteRoleModal
                    onCancel={closeModal}
                    role={node}
                    onConfirm={onDelete}
                />
            )}
            <Button
                variant="icon"
                aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                onClick={toggleIsExpanded}
            >
                <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} />
            </Button>

            <div className="d-flex flex-column">
                <div className="d-flex align-items-center">
                    <Text className="font-weight-bold m-0">{roleName}</Text>

                {node.system && (
                    <Tooltip
                        content="System roles are predefined by Sourcegraph. They cannot be deleted."
                        placement="topStart"
                    >
                        <Text className={styles.roleNodeSystemText}>System</Text>
                    </Tooltip>
                )}
                </div>
                {error && <ErrorAlert error={error} />}
            </div>

            {!node.system && (
                <Tooltip content="Deleting a role is an irreversible action.">
                    <Button aria-label="Delete" onClick={openModal} disabled={loading} variant="danger" size="sm" className={styles.roleNodeDeleteBtn}>
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
