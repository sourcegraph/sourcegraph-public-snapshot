import React from 'react'

import { Keybinding } from '../../../keyboardShortcuts'

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
} = ({ children }) => {
    console.log('LIST')
    // TODO: keyboard navigation
    return <ul>{children}</ul>
}

CommandPaletteResultList.Item = CommandPaletteResultListItem
