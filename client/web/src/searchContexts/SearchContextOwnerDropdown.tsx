import classNames from 'classnames'
import React, { useCallback, useState } from 'react'
import { Dropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

import { AuthenticatedUser } from '../auth'

import styles from './SearchContextOwnerDropdown.module.scss'

export type SelectedNamespaceType = 'user' | 'org' | 'no-owner'

export interface SelectedNamespace {
    id: string | null
    type: SelectedNamespaceType
    name: string
}

export interface SearchContextOwnerDropdownProps {
    isDisabled: boolean
    authenticatedUser: AuthenticatedUser
    selectedUserNamespace: SelectedNamespace
    selectedNamespace: SelectedNamespace
    setSelectedNamespace: (selectedNamespace: SelectedNamespace) => void
}

export const SearchContextOwnerDropdown: React.FunctionComponent<SearchContextOwnerDropdownProps> = ({
    isDisabled,
    authenticatedUser,
    selectedNamespace,
    selectedUserNamespace,
    setSelectedNamespace,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(open => !open), [])

    return (
        <Dropdown isOpen={isOpen} toggle={toggleIsOpen}>
            <DropdownToggle
                className={classNames('form-control', styles.searchContextOwnerDropdownToggle)}
                caret={true}
                color="outline-secondary"
                disabled={isDisabled}
                data-tooltip={isDisabled ? "Owner can't be changed." : ''}
            >
                <div>{selectedNamespace.type === 'no-owner' ? 'No owner' : `@${selectedNamespace.name}`}</div>
            </DropdownToggle>
            <DropdownMenu>
                <DropdownItem onClick={() => setSelectedNamespace(selectedUserNamespace)}>
                    @{authenticatedUser.username} <span className="text-muted">(you)</span>
                </DropdownItem>
                {authenticatedUser.organizations.nodes.map(org => (
                    <DropdownItem
                        key={org.name}
                        onClick={() => setSelectedNamespace({ id: org.id, type: 'org', name: org.name })}
                    >
                        @{org.name}
                    </DropdownItem>
                ))}
                {authenticatedUser.siteAdmin && (
                    <>
                        <hr />
                        <DropdownItem onClick={() => setSelectedNamespace({ id: null, type: 'no-owner', name: '' })}>
                            <div>No owner</div>
                            <div className="text-muted">Globally available on Sourcegraph.</div>
                        </DropdownItem>
                    </>
                )}
            </DropdownMenu>
        </Dropdown>
    )
}
