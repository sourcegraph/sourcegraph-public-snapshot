import React, { useMemo } from 'react'

import MenuDownIcon from 'mdi-react/MenuDownIcon'

import { Namespace } from '@sourcegraph/shared/src/schema'
import { Menu, MenuButton, MenuDivider, MenuItem, MenuList, Icon } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'

import styles from './SearchContextOwnerDropdown.module.scss'

export type SelectedNamespaceType = 'user' | 'org' | 'global-owner'

export interface SelectedNamespace {
    id: string | null
    type: SelectedNamespaceType
    name: string
}

export function getSelectedNamespace(namespace: Namespace | null): SelectedNamespace {
    if (!namespace) {
        return { id: null, type: 'global-owner', name: '' }
    }
    return {
        id: namespace.id,
        type: namespace.__typename === 'User' ? 'user' : 'org',
        name: namespace.namespaceName,
    }
}

export function getSelectedNamespaceFromUser(authenticatedUser: AuthenticatedUser): SelectedNamespace {
    return {
        id: authenticatedUser.id,
        type: 'user',
        name: authenticatedUser.username,
    }
}

export interface SearchContextOwnerDropdownProps {
    isDisabled: boolean
    authenticatedUser: AuthenticatedUser
    selectedNamespace: SelectedNamespace
    setSelectedNamespace: (selectedNamespace: SelectedNamespace) => void
}

export const SearchContextOwnerDropdown: React.FunctionComponent<
    React.PropsWithChildren<SearchContextOwnerDropdownProps>
> = ({ isDisabled, authenticatedUser, selectedNamespace, setSelectedNamespace }) => {
    const selectedUserNamespace = useMemo(() => getSelectedNamespaceFromUser(authenticatedUser), [authenticatedUser])
    return (
        <Menu>
            <MenuButton
                className={styles.searchContextOwnerDropdownToggle}
                outline={true}
                variant="secondary"
                disabled={isDisabled}
                data-tooltip={isDisabled ? "Owner can't be changed." : ''}
            >
                {selectedNamespace.type === 'global-owner' ? 'Global' : `@${selectedNamespace.name}`}{' '}
                <Icon as={MenuDownIcon} />
            </MenuButton>
            <MenuList className={styles.menuList}>
                <MenuItem onSelect={() => setSelectedNamespace(selectedUserNamespace)}>
                    @{authenticatedUser.username} <span className="text-muted">(you)</span>
                </MenuItem>
                {authenticatedUser.organizations.nodes.map(org => (
                    <MenuItem
                        key={org.name}
                        onSelect={() => setSelectedNamespace({ id: org.id, type: 'org', name: org.name })}
                    >
                        @{org.name}
                    </MenuItem>
                ))}
                {authenticatedUser.siteAdmin && (
                    <>
                        <MenuDivider />
                        <MenuItem onSelect={() => setSelectedNamespace({ id: null, type: 'global-owner', name: '' })}>
                            <div>Global owner</div>
                            <div className="text-muted">Available to everyone.</div>
                        </MenuItem>
                    </>
                )}
            </MenuList>
        </Menu>
    )
}
