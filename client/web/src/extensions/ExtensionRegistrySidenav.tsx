import React from 'react'

import classNames from 'classnames'
import MenuDownIcon from 'mdi-react/MenuDownIcon'

import { EXTENSION_CATEGORIES } from '@sourcegraph/shared/src/schema/extensionSchema'
import {
    Button,
    Link,
    Menu,
    MenuButton,
    MenuDivider,
    MenuItem,
    MenuList,
    Icon,
    H3,
    H4,
    Checkbox,
} from '@sourcegraph/wildcard'

import { SidebarGroup, SidebarGroupHeader } from '../components/Sidebar'

import { ExtensionCategoryOrAll, ExtensionsEnablement } from './ExtensionRegistry'
import { extensionBannerIconURL } from './icons'

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
    React.PropsWithChildren<ExtensionsCategoryFiltersProps & ExtensionsEnablementDropdownProps>
> = ({
    selectedCategory,
    onSelectCategory,
    enablementFilter,
    setEnablementFilter,
    showExperimentalExtensions,
    toggleExperimentalExtensions,
}) => {
    const showAll = (): void => setEnablementFilter('all')
    const showEnabled = (): void => setEnablementFilter('enabled')
    const showDisabled = (): void => setEnablementFilter('disabled')

    return (
        <div className={classNames(styles.column, 'mr-4 flex-grow-0 flex-shrink-0')}>
            <SidebarGroup>
                <SidebarGroupHeader label="Categories" />
                {['All' as const, ...EXTENSION_CATEGORIES].map(category => (
                    <Button
                        className="text-left sidebar__link--inactive d-flex w-100"
                        variant={selectedCategory === category ? 'primary' : undefined}
                        data-test-extension-category={category}
                        key={category}
                        onClick={() => onSelectCategory(category)}
                    >
                        {category}
                    </Button>
                ))}
            </SidebarGroup>

            <hr className={classNames('my-3', styles.divider)} />

            <Menu>
                <MenuButton size="sm" variant="secondary" outline={true}>
                    {enablementFilterToLabel[enablementFilter]} <Icon role="img" as={MenuDownIcon} aria-hidden={true} />
                </MenuButton>
                <MenuList>
                    <MenuItem onSelect={showAll} disabled={enablementFilter === 'all'}>
                        Show all
                    </MenuItem>
                    <MenuItem onSelect={showEnabled} disabled={enablementFilter === 'enabled'}>
                        Show enabled extensions
                    </MenuItem>
                    <MenuItem onSelect={showDisabled} disabled={enablementFilter === 'disabled'}>
                        Show disabled extensions
                    </MenuItem>

                    <MenuDivider />

                    <MenuItem onSelect={toggleExperimentalExtensions}>
                        <Checkbox
                            id="show-experimental-extensions"
                            checked={showExperimentalExtensions}
                            label="Show experimental extensions"
                        />
                    </MenuItem>
                </MenuList>
            </Menu>

            <ExtensionSidenavBanner />
        </div>
    )
}

const ExtensionSidenavBanner: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <div className={classNames(styles.banner, 'mx-2')}>
        <img className={classNames(styles.bannerIcon, 'mb-2')} src={extensionBannerIconURL} alt="" />
        {/* Override h4 font-weight */}
        <H4 as={H3} className="mt-2 font-weight-bold">
            Create custom extensions!
        </H4>
        <small>
            You can improve your workflow by creating custom extensions. See{' '}
            <Link
                to="https://docs.sourcegraph.com/extensions/authoring"
                // eslint-disable-next-line react/jsx-no-target-blank
                target="_blank"
                rel="noreferrer"
            >
                Sourcegraph Docs
            </Link>{' '}
            for details about writing and publishing.
        </small>
    </div>
)
