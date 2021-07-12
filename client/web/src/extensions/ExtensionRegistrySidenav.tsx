import classnames from 'classnames'
import React, { useState } from 'react'
import { ButtonDropdown, DropdownItem, DropdownMenu, DropdownToggle } from 'reactstrap'

import { EXTENSION_CATEGORIES } from '@sourcegraph/shared/src/schema/extensionSchema'

import { SidebarGroup, SidebarGroupHeader, SidebarGroupItems } from '../components/Sidebar'

import { ExtensionCategoryOrAll, ExtensionsEnablement } from './ExtensionRegistry'
import styles from './ExtensionRegistrySidenav.module.scss'
import { extensionBannerIconURL } from './icons'

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
            <SidebarGroup>
                <SidebarGroupHeader label="Categories" />
                <SidebarGroupItems>
                    {['All' as const, ...EXTENSION_CATEGORIES].map(category => (
                        <button
                            type="button"
                            className={classnames(
                                'btn text-left sidebar__link--inactive d-flex sidebar-nav-link w-100',
                                selectedCategory === category && 'btn-primary'
                            )}
                            data-test-extension-category={category}
                            key={category}
                            onClick={() => onSelectCategory(category)}
                        >
                            {category}
                        </button>
                    ))}
                </SidebarGroupItems>
            </SidebarGroup>

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

            <ExtensionSidenavBanner />
        </div>
    )
}

const ExtensionSidenavBanner: React.FunctionComponent = () => (
    <div className={classnames(styles.banner, 'mx-2')}>
        <img className={classnames(styles.bannerIcon, 'mb-2')} src={extensionBannerIconURL} alt="" />
        {/* Override .theme-redesign h4 font-weight */}
        <h4 className="mt-2 font-weight-bold">Create custom extensions!</h4>
        <small>
            You can improve your workflow by creating custom extensions. See{' '}
            <a
                href="https://docs.sourcegraph.com/extensions/authoring"
                // eslint-disable-next-line react/jsx-no-target-blank
                target="_blank"
                rel="noreferrer"
            >
                Sourcegraph Docs
            </a>{' '}
            for details about writing and publishing.
        </small>
    </div>
)
