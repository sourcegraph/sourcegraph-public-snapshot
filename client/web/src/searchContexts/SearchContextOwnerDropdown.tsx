import classNames from 'classnames'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

import { AuthenticatedUser } from '../auth'

import styles from './SearchContextOwnerDropdown.module.scss'

export type SelectedNamespaceType = 'user' | 'org' | 'global-owner'

export interface SelectedNamespace {
    id: string | null
    type: SelectedNamespaceType
    name: string
}

export interface SearchContextOwnerDropdownProps {
    authenticatedUser: AuthenticatedUser
    selectedUserNamespace: SelectedNamespace
    selectedNamespace: SelectedNamespace
    setSelectedNamespace: (selectedNamespace: SelectedNamespace) => void
}

export const SearchContextOwnerDropdown: React.FunctionComponent<SearchContextOwnerDropdownProps> = ({
    authenticatedUser,
    selectedNamespace,
    selectedUserNamespace,
    setSelectedNamespace,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(open => !open), [])

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen}>
            <DropdownToggle
                className={classNames('btn btn-sm form-control', styles.searchContextOwnerDropdownToggle)}
                caret={true}
                color="outline-secondary"
            >
                {selectedNamespace.type === 'global-owner' ? 'Global owner' : `@${selectedNamespace.name}`}
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
                        <DropdownItem
                            onClick={() => setSelectedNamespace({ id: null, type: 'global-owner', name: '' })}
                        >
                            <div>Global owner</div>
                            <div className="text-muted">Available to everyone.</div>
                        </DropdownItem>
                    </>
                )}
            </DropdownMenu>
        </ButtonDropdown>
    )
}
