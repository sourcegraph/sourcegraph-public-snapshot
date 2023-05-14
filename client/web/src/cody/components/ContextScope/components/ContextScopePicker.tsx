import { mdiCircleOutline, mdiFileDocumentOutline, mdiGit, mdiChevronDown, mdiEarth, mdiChevronUp } from '@mdi/js'
import classNames from 'classnames'

import { Icon, Menu, MenuButton, MenuList, MenuItem, Position } from '@sourcegraph/wildcard'

import { ContextType, SELECTED } from '../ContextScope'

import styles from './ContextScopeComponents.module.scss'

interface ContextScopePickerProps {
    onSelect?: (itemIndex: ContextType) => void
    selected: typeof SELECTED[keyof typeof SELECTED]
}

export const ContextScopePicker: React.FC<ContextScopePickerProps> = ({ onSelect, selected }) => {
    const handleMenuItemSelect = (itemIndex: ContextType): void => {
        onSelect && onSelect(itemIndex)
    }

    const menuItems = [
        { label: 'Organizations', icon: mdiEarth },
        { label: 'Repositories', icon: mdiGit },
        { label: 'Files', icon: mdiFileDocumentOutline },
        { label: 'None', icon: mdiCircleOutline },
    ]

    return (
        <div className={styles.customMenuWidth}>
            <Menu>
                <MenuButton
                    variant="icon"
                    outline={false}
                    className={classNames('d-flex justify-content-between', styles.customMenuWidth)}
                >
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
