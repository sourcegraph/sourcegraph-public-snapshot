import React, { useMemo } from 'react'

import { mdiMenuDown } from '@mdi/js'

import type { SearchContextFields } from '@sourcegraph/shared/src/graphql-operations'
import { Menu, MenuButton, MenuDivider, MenuItem, MenuList, Icon, Tooltip } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'

import styles from './SearchContextOwnerDropdown.module.scss'

export type SelectedNamespaceType = 'user' | 'org' | 'global-owner'

export interface SelectedNamespace {
    id: string | null
    type: SelectedNamespaceType
    name: string
}

export function getSelectedNamespace(namespace: SearchContextFields['namespace']): SelectedNamespace {
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
            <Tooltip content={isDisabled ? "Owner can't be changed." : ''}>
                <MenuButton
                    className={styles.searchContextOwnerDropdownToggle}
                    outline={true}
                    variant="secondary"
                    disabled={isDisabled}
                >
                    {selectedNamespace.type === 'global-owner' ? 'Global' : `@${selectedNamespace.name}`}{' '}
                    <Icon aria-hidden={true} svgPath={mdiMenuDown} />
                </MenuButton>
            </Tooltip>
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
