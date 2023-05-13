import { useState, useEffect } from 'react'

import { mdiCircleOutline, mdiFileDocumentOutline, mdiGit, mdiChevronDown, mdiEarth } from '@mdi/js'

import { Icon, Menu, MenuButton, MenuList, MenuItem } from '@sourcegraph/wildcard'

import styles from './ContextScopeComponents.module.scss'

interface ContextScopePickerProps {
    onSelect?: (itemIndex: number) => void
    selected?: typeof SELECTED[keyof typeof SELECTED]
}

export const SELECTED = {
    NONE: 0,
    FILES: 1,
    REPOSITORIES: 2,
    ORGANIZATIONS: 3,
} as const

export type SelectedType = typeof SELECTED[keyof typeof SELECTED]

export const ContextScopePicker: React.FC<ContextScopePickerProps> = ({ onSelect, selected = SELECTED.NONE }) => {
    const [selectedItem, setSelectedItem] = useState<SelectedType>(selected)

    const handleMenuItemSelect = (itemIndex: SelectedType): void => {
        setSelectedItem(itemIndex)
        onSelect && onSelect(itemIndex)
    }

    useEffect(() => {
        setSelectedItem(selected)
    }, [selected])

    const menuItems = [
        { label: 'None', icon: mdiCircleOutline },
        { label: 'Files', icon: mdiFileDocumentOutline },
        { label: 'Repositories', icon: mdiGit },
        { label: 'Organizations', icon: mdiEarth },
    ]

    return (
        <Menu>
            <MenuButton variant="icon" outline={false} className={styles.customMenuButton}>
                <div>
                    <Icon aria-hidden={true} svgPath={menuItems[selectedItem].icon} /> {menuItems[selectedItem].label}
                </div>

                <Icon aria-hidden={true} svgPath={mdiChevronDown} />
            </MenuButton>

            <MenuList>
                {Object.entries(SELECTED).map(([key, value]) => (
                    <MenuItem
                        key={value}
                        onSelect={() => handleMenuItemSelect(value)}
                        className={selectedItem === value ? styles.menuSelectedItem : ''}
                    >
                        <Icon aria-hidden={true} svgPath={menuItems[value].icon} /> {menuItems[value].label}
                    </MenuItem>
                ))}
            </MenuList>
        </Menu>
    )
}
