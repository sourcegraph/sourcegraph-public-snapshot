import classnames from 'classnames'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

import { EXTENSION_CATEGORIES, ExtensionCategory } from '@sourcegraph/shared/src/schema/extensionSchema'

import { ExtensionsEnablement } from './ExtensionRegistry'
import styles from './ExtensionRegistrySidenav.module.scss'

const enablementFilterToLabel: Record<ExtensionsEnablement, string> = {
    all: 'Show all',
    enabled: 'Show enabled extensions',
    disabled: 'Show disabled extensions',
}

interface ExtensionsCategoryFiltersProps {
    selectedCategory: ExtensionCategory | 'All'
    onSelectCategory: (category: ExtensionCategory | 'All') => void
}
interface ExtensionsEnablementDropdownProps {
    enablementFilter: ExtensionsEnablement
    setEnablementFilter: React.Dispatch<React.SetStateAction<ExtensionsEnablement>>
}

/**
 * Displays buttons to be rendered alongside the extension registry.
 * Includes category filter buttons and enablement filter dropdown.
 */
export const ExtensionRegistrySidenav: React.FunctionComponent<
    ExtensionsCategoryFiltersProps & ExtensionsEnablementDropdownProps
> = ({ selectedCategory, onSelectCategory, enablementFilter, setEnablementFilter }) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(open => !open), [])

    return (
        <div className={classnames(styles.column, 'mr-4 flex-grow-0 flex-shrink-0')}>
            <div className="d-flex flex-column">
                <h3 className={classnames(styles.header, 'mb-3')}>Categories</h3>

                {['All' as const, ...EXTENSION_CATEGORIES].map(category => (
                    <button
                        type="button"
                        className={classnames('btn text-left', {
                            'btn-primary': selectedCategory === category,
                        })}
                        data-test-extension-category={category}
                        key={category}
                        onClick={() => onSelectCategory(category)}
                    >
                        {category}
                    </button>
                ))}
            </div>

            <hr className={classnames('my-3', styles.divider)} />

            <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen} className="ml-2">
                <DropdownToggle className="btn-sm" caret={true} color="outline-secondary">
                    {enablementFilterToLabel[enablementFilter]}
                </DropdownToggle>
                <DropdownMenu>
                    <DropdownItem
                        // eslint-disable-next-line react/jsx-no-bind
                        onClick={() => setEnablementFilter('all')}
                        disabled={enablementFilter === 'all'}
                    >
                        Show all
                    </DropdownItem>
                    <DropdownItem
                        // eslint-disable-next-line react/jsx-no-bind
                        onClick={() => setEnablementFilter('enabled')}
                        disabled={enablementFilter === 'enabled'}
                    >
                        Show enabled extensions
                    </DropdownItem>
                    <DropdownItem
                        // eslint-disable-next-line react/jsx-no-bind
                        onClick={() => setEnablementFilter('disabled')}
                        disabled={enablementFilter === 'disabled'}
                    >
                        Show disabled extensions
                    </DropdownItem>
                </DropdownMenu>
            </ButtonDropdown>
        </div>
    )
}
