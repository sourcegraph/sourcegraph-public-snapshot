import React, { useId, useState, useMemo, useCallback } from 'react'

import { mdiCogOutline } from '@mdi/js'
import { differenceBy } from 'lodash'

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
} from '@sourcegraph/wildcard'

import { useGetUserRolesAndAllRoles, useSetRoles } from '../backend'
import { LoaderButton } from '../../../../components/LoaderButton'

import { RoleFields } from '../../../../graphql-operations'

export interface RoleAssignmentModalProps {
    onCancel: () => void
    onSuccess: () => void
}

type Role = Pick<RoleFields, 'id' | 'system' | 'name'>

export const RoleAssignmentModal: React.FunctionComponent<RoleAssignmentModalProps> = ({ onCancel, onSuccess }) => {
    const labelID = 'RoleAssignment'
    const userID = 'VXNlcjoy'

    const id = useId()
    const [searchTerm, setSearchTerm] = useState('')
    const [selectedRoles, setSelectedRoles] = useState<Role[]>([])
    const [allRoles, setAllRoles] = useState<Role[]>([])

    const { loading, error: getUserRolesError } = useGetUserRolesAndAllRoles(userID, data => {
        if (data.node?.__typename !== 'User') {
            throw new Error('User not found')
        }
        const { nodes: userRoles } = data.node.roles
        const { nodes: allRoles } = data.roles
        setSelectedRoles(userRoles)
        setAllRoles(allRoles)
    })

    const selectedRoleNames = useMemo(() => selectedRoles.map(role => role.name), [selectedRoles])
    const [setRoles, { loading: setRolesLoading, error: setRolesError }] = useSetRoles(onSuccess)
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

    const handleSelectedRolesChange = useCallback(
        (roles: Role[]) => {
            // Site admins will not be able to revoke system permissions from this modal. To assign or revoke
            // the site administrator role, the site admin would have to use the `Promote to site admin` option
            // available in the `UsersList` options.
            // The MultiComboBox doesn't support disabling options displayed in `Input` field, so using this hack
            // to ensure system roles can never be removed.
            const diff = differenceBy(selectedRoles, roles, 'id')
            if (diff.findIndex(role => role.system) < 0) {
                setSelectedRoles(roles)
            }
        },
        [selectedRoles]
    )

    const handleSubmit: React.FormEventHandler = (event): void => {
        event.preventDefault()
        const roleIDs = selectedRoles.map(role => role.id)
        setRoles({
            variables: {
                user: userID,
                roles: roleIDs,
            },
        })
    }

    const error = getUserRolesError || setRolesError

    return (
        <Modal onDismiss={onCancel} aria-labelledby={labelID} as={Form} onSubmit={handleSubmit}>
            <div className="d-flex align-items-center mb-2">
                <Icon className="icon mr-1" svgPath={mdiCogOutline} inline={false} aria-hidden={true} />{' '}
                <H3 id={labelID} className="mb-0">
                    Assign roles
                </H3>
            </div>
            <Text>Select roles to be assigned to the user.</Text>

            {loading && <LoadingSpinner />}
            {error && !loading && <ErrorAlert error={error} />}

            <MultiCombobox
                selectedItems={selectedRoles}
                getItemKey={item => item.id}
                getItemName={item => item.name}
                onSelectedItemsChange={handleSelectedRolesChange}
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
                <small className="text-muted pl-2">System roles cannot be revoked or assigned via this modal.</small>

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

            <div className="d-flex my-2 justify-content-end">
                <Button variant="secondary" className="mr-2" onClick={onCancel}>
                    Cancel
                </Button>
                <LoaderButton
                    variant="primary"
                    loading={setRolesLoading}
                    label="Update"
                    alwaysShowLabel={true}
                    disabled={setRolesLoading}
                    type="submit"
                />
            </div>
        </Modal>
    )
}
