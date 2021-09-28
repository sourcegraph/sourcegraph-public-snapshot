import React from 'react'

import { Keybinding } from '../../../keyboardShortcuts'

interface CommandPaletteResultItemProps {
    onClick: () => void
    href?: string
    keybindings?: Keybinding[]
    label: string
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
            <Tag type="button" onClick={onClick} href={href}>
                {label}

                {keybindings.map(({ ordered, held }, index) => (
                    <span key={index}>
                        {[...ordered, ...(held || [])].map(key => (
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
} = ({ children }) => <ul>{children}</ul>

CommandPaletteResultList.Item = CommandPaletteResultListItem
