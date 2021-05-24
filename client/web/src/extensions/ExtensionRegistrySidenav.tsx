import classnames from 'classnames'
import React, { useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

import { EXTENSION_CATEGORIES } from '@sourcegraph/shared/src/schema/extensionSchema'

import { ExtensionCategoryOrAll, ExtensionsEnablement } from './ExtensionRegistry'
import styles from './ExtensionRegistrySidenav.module.scss'

const enablementFilterToLabel: Record<ExtensionsEnablement, string> = {
    all: 'Show all',
    enabled: 'Show enabled extensions',
    disabled: 'Show disabled extensions',
}

interface ExtensionsCategoryFiltersProps {
    selectedCategory: ExtensionCategoryOrAll
    onSelectCategory: (category: ExtensionCategoryOrAll) => void
}
interface ExtensionsEnablementDropdownProps {
    enablementFilter: ExtensionsEnablement
    setEnablementFilter: React.Dispatch<React.SetStateAction<ExtensionsEnablement>>

    showExperimentalExtensions: boolean
    toggleExperimentalExtensions: () => void
}

/**
 * Displays buttons to be rendered alongside the extension registry.
 * Includes category filter buttons and enablement filter dropdown.
 */
export const ExtensionRegistrySidenav: React.FunctionComponent<
    ExtensionsCategoryFiltersProps & ExtensionsEnablementDropdownProps
> = ({
    selectedCategory,
    onSelectCategory,
    enablementFilter,
    setEnablementFilter,
    showExperimentalExtensions,
    toggleExperimentalExtensions,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = (): void => setIsOpen(open => !open)

    const showAll = (): void => setEnablementFilter('all')
    const showEnabled = (): void => setEnablementFilter('enabled')
    const showDisabled = (): void => setEnablementFilter('disabled')

    return (
        <div className={classnames(styles.column, 'mr-4 flex-grow-0 flex-shrink-0')}>
            <div className="d-flex flex-column">
                <h3 className={classnames(styles.header, 'mb-3')}>Categories</h3>

                {['All' as const, ...EXTENSION_CATEGORIES].map(category => (
                    <button
                        type="button"
                        className={classnames(
                            'btn text-left',
                            selectedCategory === category ? 'btn-primary' : styles.inactiveCategory
                        )}
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
                    <DropdownItem onClick={showAll} disabled={enablementFilter === 'all'}>
                        Show all
                    </DropdownItem>
                    <DropdownItem onClick={showEnabled} disabled={enablementFilter === 'enabled'}>
                        Show enabled extensions
                    </DropdownItem>
                    <DropdownItem onClick={showDisabled} disabled={enablementFilter === 'disabled'}>
                        Show disabled extensions
                    </DropdownItem>

                    <DropdownItem divider={true} />

                    <DropdownItem
                        // Hack: clicking <label> inside <DropdownItem> doesn't affect checked state,
                        // so use a <span> for which click events are handled by <DropdownItem>.
                        onClick={toggleExperimentalExtensions}
                    >
                        <div className="d-flex align-items-center">
                            <input
                                type="checkbox"
                                checked={showExperimentalExtensions}
                                onChange={toggleExperimentalExtensions}
                                className=""
                                aria-labelledby="show-experimental-extensions"
                            />
                            <span className="m-0 pl-2" id="show-experimental-extensions">
                                Show experimental extensions
                            </span>
                        </div>
                    </DropdownItem>
                </DropdownMenu>
            </ButtonDropdown>
        </div>
    )
}
