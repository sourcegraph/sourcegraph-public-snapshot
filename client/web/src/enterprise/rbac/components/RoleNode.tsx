import React, { useCallback, useMemo, useState } from 'react'

import { mdiChevronUp, mdiChevronDown, mdiDelete } from '@mdi/js'
import { startCase } from 'lodash'

import {
    Button,
    Icon,
    Text,
    Tooltip,
    ErrorAlert,
    Collapse,
    CollapseHeader,
    CollapsePanel,
    H4,
    useCheckboxes,
    useForm,
    Form,
    SubmissionResult,
} from '@sourcegraph/wildcard'

import { RoleFields } from '../../../graphql-operations'
import { PermissionsMap, useDeleteRole, useSetPermissions } from '../backend'

import { LoaderButton } from '../../../components/LoaderButton'
import { ConfirmDeleteRoleModal } from './ConfirmDeleteRoleModal'
import { PermissionsList } from './Permissions'

import styles from './RoleNode.module.scss'

interface RoleNodeProps {
    node: RoleFields
    refetchAll: () => void
    allPermissions: PermissionsMap
}

interface RoleNodePermissionsFormValues {
    permissions: string[]
}

export const RoleNode: React.FunctionComponent<RoleNodeProps> = ({ node, refetchAll, allPermissions }) => {
    const [isExpanded, setIsExpanded] = useState<boolean>(false)
    const [showConfirmDeleteModal, setShowConfirmDeleteModal] = useState<boolean>(false)

    const handleOpenChange = (isOpen: boolean): void => {
        setIsExpanded(isOpen)
    }
    const openModal = useCallback<React.MouseEventHandler>(event => {
        event.stopPropagation()
        setShowConfirmDeleteModal(true)
    }, [])
    const closeModal = useCallback(() => {
        setShowConfirmDeleteModal(false)
    }, [])

    const [deleteRole, { loading: deleteRoleLoading, error: deleteRoleError }] = useDeleteRole(() => {
        closeModal()
        refetchAll()
    }, closeModal)

    const roleName = useMemo(() => {
        const lowerCaseName = node.name.replace(/_/g, ' ').toLowerCase()
        return startCase(lowerCaseName)
    }, [node.name])

    const onDelete = useCallback<React.FormEventHandler>(
        async event => {
            event.preventDefault()
            await deleteRole({ variables: { role: node.id } })
        },
        [deleteRole, node.id]
    )

    const { nodes: permissionNodes } = node.permissions
    const rolePermissionIDs = useMemo(() => permissionNodes.map(permission => permission.id), [permissionNodes])

    const [setPermissions, { loading: setPermissionsLoading, error: setPermissionsError }] = useSetPermissions(refetchAll)

    const onSubmit = (values: RoleNodePermissionsFormValues): SubmissionResult => {
        setPermissions({ variables: { role: node.id, permissions: values.permissions } })
    }
    const defaultFormValues: RoleNodePermissionsFormValues = { permissions: rolePermissionIDs }
    const { formAPI, ref, handleSubmit } = useForm({
        initialValues: defaultFormValues,
        onSubmit,
    })
    const {
        input: { isChecked, onBlur, onChange },
    } = useCheckboxes('permissions', formAPI)

    const error = deleteRoleError || setPermissionsError
    return (
        <li className={styles.roleNode}>
            {showConfirmDeleteModal && (
                <ConfirmDeleteRoleModal onCancel={closeModal} role={node} onConfirm={onDelete} />
            )}

            <Collapse isOpen={isExpanded} onOpenChange={handleOpenChange}>
                <CollapseHeader
                    as={Button}
                    className={styles.roleNodeCollapsibleHeader}
                    aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                    outline={true}
                    variant="icon"
                >
                    <Icon
                        data-caret={true}
                        className="mr-1 bg-red"
                        aria-hidden={true}
                        svgPath={isExpanded ? mdiChevronUp : mdiChevronDown}
                    />

                    <header className="d-flex flex-column justify-content-center mr-2">
                        <div className="d-flex align-items-center">
                            <H4 className="m-0">{roleName}</H4>

                            {node.system && (
                                <Tooltip
                                    content="System roles are predefined by Sourcegraph. They cannot be deleted."
                                    placement="topStart"
                                >
                                    <Text className={styles.roleNodeSystemText}>System</Text>
                                </Tooltip>
                            )}
                        </div>
                        {error && <ErrorAlert className="mt-2" error={error} />}
                    </header>

                    {!node.system && (
                        <Tooltip content="Deleting a role is an irreversible action.">
                            <Button
                                aria-label="Delete"
                                onClick={openModal}
                                disabled={deleteRoleLoading}
                                variant="danger"
                                size="sm"
                                className={styles.roleNodeDeleteBtn}
                            >
                                <Icon aria-hidden={true} svgPath={mdiDelete} className={styles.roleNodeDeleteBtnIcon} />
                            </Button>
                        </Tooltip>
                    )}
                </CollapseHeader>

                <CollapsePanel
                    className={styles.roleNodePermissions}
                    forcedRender={false}
                    as={Form}
                    ref={ref}
                    onSubmit={handleSubmit}
                >
                    <PermissionsList
                        allPermissions={allPermissions}
                        isChecked={isChecked}
                        onBlur={onBlur}
                        onChange={onChange}
                    />
                    <LoaderButton alwaysShowLabel={true} variant="primary" type="submit" loading={setPermissionsLoading} label="Update" />
                </CollapsePanel>
            </Collapse>
        </li>
    )
}
