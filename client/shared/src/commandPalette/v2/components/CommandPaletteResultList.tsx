import React from 'react'

import { Keybinding } from '../../../keyboardShortcuts'

import styles from './CommandPaletteResultList.module.scss'

interface CommandPaletteResultItemProps {
    onClick: () => void
    href?: string
    keybindings?: Keybinding[]
    label: string
    // TODO icon (for symbol type, action item icon)
    icon?: JSX.Element
}

const CommandPaletteResultListItem: React.FC<CommandPaletteResultItemProps> = ({
    onClick,
    href,
    keybindings = [],
    label,
}) => {
    const Tag = href ? 'a' : 'button'

    return (
        <li>
            <Tag type="button" className={styles.Button} onClick={onClick} href={href}>
                {label}

                {keybindings.map(({ ordered, held }, index) => (
                    <span key={index} className={styles.Keybindings}>
                        {[held || [], ...ordered].map(key => (
                            <kbd key={key}>{key}</kbd>
                        ))}
                    </span>
                ))}
            </Tag>
        </li>
    )
}

export const CommandPaletteResultList: React.FC & {
    Item: typeof CommandPaletteResultListItem
} = ({ children }) => <ul className={styles.List}>{children}</ul>

CommandPaletteResultList.Item = CommandPaletteResultListItem
