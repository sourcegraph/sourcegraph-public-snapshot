import { mdiCircleOutline, mdiFileDocumentOutline, mdiGit, mdiChevronDown, mdiEarth, mdiChevronUp } from '@mdi/js'

import { Icon, Menu, MenuButton, MenuList, MenuItem, Position } from '@sourcegraph/wildcard'

import { SelectedType, SELECTED } from '../ContextScope'

import styles from './ContextScopeComponents.module.scss'

interface ContextScopePickerProps {
    onSelect?: (itemIndex: SelectedType) => void
    selected: typeof SELECTED[keyof typeof SELECTED]
}

export const ContextScopePicker: React.FC<ContextScopePickerProps> = ({ onSelect, selected }) => {
    const handleMenuItemSelect = (itemIndex: SelectedType): void => {
        onSelect && onSelect(itemIndex)
    }

    const menuItems = [
        { label: 'Organizations', icon: mdiEarth },
        { label: 'Repositories', icon: mdiGit },
        { label: 'Files', icon: mdiFileDocumentOutline },
        { label: 'None', icon: mdiCircleOutline },
    ]

    return (
        <div style={{ width: 165 }}>
            <Menu>
                <MenuButton variant="icon" outline={false} className={styles.customMenuButton}>
                    <div>
                        <Icon aria-hidden={true} svgPath={menuItems[selected].icon} /> {menuItems[selected].label}
                    </div>

                    <Icon aria-hidden={true} svgPath={mdiChevronUp} />
                </MenuButton>

                <MenuList position={Position.topStart}>
                    {Object.entries(SELECTED).map(([key, value]) => (
                        <MenuItem
                            key={value}
                            onSelect={() => handleMenuItemSelect(value)}
                            className={selected === value ? styles.menuSelectedItem : ''}
                        >
                            <Icon aria-hidden={true} svgPath={menuItems[value].icon} /> {menuItems[value].label}
                        </MenuItem>
                    ))}
                </MenuList>
            </Menu>
        </div>
    )
}
