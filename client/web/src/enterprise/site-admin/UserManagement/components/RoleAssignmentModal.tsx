import React, { useId, useState, useMemo } from 'react'

import { mdiBadgeAccount } from '@mdi/js'
import { noop } from 'lodash'

import {
    Button,
    Icon,
    Text,
    Modal,
    H3,
    Form,
    MultiCombobox,
    LoadingSpinner,
    ErrorAlert,
    MultiComboboxInput,
    MultiComboboxPopover,
    MultiComboboxList,
    MultiComboboxOption,
    Link,
} from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../../components/LoaderButton'
import { RoleFields, Scalars } from '../../../../graphql-operations'
import { prettifySystemRole } from '../../../../util/settings'
import { useGetUserRolesAndAllRoles, useSetRoles } from '../backend'

export interface RoleAssignmentModalProps {
    onCancel: () => void
    onSuccess: (user: { username: string }) => void
    user: { id: Scalars['ID']; username: string }
}

type Role = Pick<RoleFields, 'id' | 'system' | 'name'>

const prepareDisplayRole = (role: Pick<RoleFields, 'id' | 'system' | 'name'>): Role => ({
    ...role,
    name: role.system ? prettifySystemRole(role.name) : role.name,
})

export const RoleAssignmentModal: React.FunctionComponent<RoleAssignmentModalProps> = ({
    onCancel,
    onSuccess,
    user,
}) => {
    const labelID = 'RoleAssignment'

    const id = useId()
    const [searchTerm, setSearchTerm] = useState('')

    const {
        data,
        loading,
        error: getUserRolesError,
    } = useGetUserRolesAndAllRoles(user.id, data => {
        if (data.node?.__typename !== 'User') {
            throw new Error('User not found')
        }
        const userRoles = data.node.roles.nodes.map(prepareDisplayRole)
        const allRoles = data.roles.nodes.map(prepareDisplayRole)
        setSelectedRoles(userRoles)
        setAllRoles(allRoles)
    })

    const [selectedRoles, setSelectedRoles] = useState<Role[]>([])
    // Use roles from cached data if it's available, as these will change infrequently.
    const [allRoles, setAllRoles] = useState<Role[]>((data?.roles.nodes || []).map(prepareDisplayRole))

    const selectedRoleNames = useMemo(() => selectedRoles.map(role => role.name), [selectedRoles])
    const [setRoles, { loading: setRolesLoading, error: setRolesError }] = useSetRoles(() => onSuccess(user))
    const suggestions = useMemo(
        () =>
            allRoles.filter(
                role =>
                    !selectedRoleNames.includes(role.name) &&
                    !role.system &&
                    role.name.toLowerCase().includes(searchTerm.toLowerCase())
            ),
        [allRoles, selectedRoleNames, searchTerm]
    )

    const handleSubmit: React.FormEventHandler = (event): void => {
        event.preventDefault()
        const roleIDs = selectedRoles.map(role => role.id)
        setRoles({
            variables: {
                user: user.id,
                roles: roleIDs,
            },
        }).catch(noop)
    }

    const error = getUserRolesError || setRolesError

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelID} as={Form} onSubmit={handleSubmit}>
            <div className="d-flex align-items-center mb-2">
                <Icon className="icon mr-1" svgPath={mdiBadgeAccount} inline={false} aria-hidden={true} />{' '}
                <H3 id={labelID} className="mb-0">
                    Manage roles for {user.username}
                </H3>
            </div>
            <Text>
                Roles determine which permissions are granted to this user.{' '}
                <Link to="/site-admin/roles">View roles settings</Link> to manage available roles and permissions.
            </Text>

            {loading && allRoles.length === 0 && (
                <div className="d-flex align-items-center">
                    <Text className="d-block font-italic m-0 mr-2">Loading roles</Text> <LoadingSpinner />
                </div>
            )}
            {error && !loading && <ErrorAlert error={error} />}

            {(!loading || allRoles.length > 0) && (
                <MultiCombobox
                    selectedItems={selectedRoles}
                    getItemKey={item => item.id}
                    getItemName={item => item.name}
                    getItemIsPermanent={item => item.system}
                    onSelectedItemsChange={setSelectedRoles}
                    aria-label="Select role(s) to assign to user"
                >
                    <MultiComboboxInput
                        id={id}
                        value={searchTerm}
                        autoFocus={false}
                        autoCorrect="false"
                        autoComplete="off"
                        placeholder="Select role..."
                        onChange={event => setSearchTerm(event.target.value)}
                    />
                    <small className="text-muted pl-2">
                        System roles cannot be revoked or assigned via this modal.
                    </small>

                    <MultiComboboxPopover>
                        <MultiComboboxList items={suggestions}>
                            {items =>
                                items.map((item, index) => (
                                    <MultiComboboxOption value={item.name} key={item.id} index={index}>
                                        <small>{item.name}</small>
                                    </MultiComboboxOption>
                                ))
                            }
                        </MultiComboboxList>
                    </MultiComboboxPopover>
                </MultiCombobox>
            )}

            <div className="d-flex my-2 justify-content-end">
                <Button variant="secondary" className="mr-2" onClick={onCancel}>
                    Cancel
                </Button>
                <LoaderButton
                    variant="primary"
                    loading={setRolesLoading}
                    label="Update"
                    alwaysShowLabel={true}
                    disabled={loading || setRolesLoading}
                    type="submit"
                />
            </div>
        </Modal>
    )
}
