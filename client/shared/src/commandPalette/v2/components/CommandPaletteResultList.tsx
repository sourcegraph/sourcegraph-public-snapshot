import classNames from 'classnames'
import React, { useState, useEffect } from 'react'

import { Keybinding } from '../../../keyboardShortcuts'

import styles from './CommandPaletteResultList.module.scss'

interface CommandPaletteResultItemProps {
    onClick: () => void
    href?: string
    keybindings?: Keybinding[]
    label: string
    active?: boolean
    // TODO icon (for symbol type, action item icon)
    icon?: JSX.Element
}

const CommandPaletteResultListItem: React.FC<CommandPaletteResultItemProps> = ({
    onClick,
    href,
    keybindings = [],
    label,
    active,
}) => {
    const Tag = href ? 'a' : 'button'

    useEffect(() => {
        function handleKeyDown(event: KeyboardEvent): void {
            if (event.key === 'Enter' && active) {
                onClick()
            }
        }
        document.addEventListener('keydown', handleKeyDown)
        return () => document.removeEventListener('keydown', handleKeyDown)
    }, [onClick, active])

    return (
        <li tabIndex={-1}>
            <Tag
                type="button"
                tabIndex={0}
                className={classNames(styles.button, { [styles.buttonActive]: active })}
                onClick={onClick}
                href={href}
            >
                {label}

                {keybindings.map(({ ordered, held }, index) => (
                    <span tabIndex={-1} key={index} className={styles.keybindings}>
                        {[held || [], ...ordered].map((key, index) => (
                            <kbd key={index}>{key}</kbd>
                        ))}
                    </span>
                ))}
            </Tag>
        </li>
    )
}

interface CommandPaletteResultListProps<T> {
    items: T[]
    children: (item: T, options: { active: boolean }) => JSX.Element
}

export function CommandPaletteResultList<T>({ children, items }: CommandPaletteResultListProps<T>): JSX.Element {
    const [selected, setSelected] = useState<number | undefined>()

    useEffect(() => {
        function handleKeyDown(event: KeyboardEvent): void {
            if (event.key === 'ArrowUp') {
                setSelected(selected => ((selected || items.length) - 1) % items.length)
            } else if (event.key === 'ArrowDown') {
                setSelected(selected => ((selected || 0) + 1) % items.length)
            }
        }
        document.addEventListener('keydown', handleKeyDown)
        return () => document.removeEventListener('keydown', handleKeyDown)
    }, [items])

    return (
        <ul className={styles.list}>
            {items.map((item, index) => (
                <React.Fragment key={index}>{children(item, { active: selected === index })}</React.Fragment>
            ))}
        </ul>
    )
}

CommandPaletteResultList.Item = CommandPaletteResultListItem
