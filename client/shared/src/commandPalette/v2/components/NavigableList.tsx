import classNames from 'classnames'
import React, { useState, useEffect } from 'react'

import { Keybinding } from '../../../keyboardShortcuts'

import styles from './NavigableList.module.scss'

interface NavigableListItemProps {
    onClick?: () => void
    onFocus?: () => void
    href?: string
    keybindings?: Keybinding[]
    label: string
    active?: boolean
    // TODO icon (for symbol type, action item icon)
    icon?: JSX.Element
}

const NavigableListItem: React.FC<NavigableListItemProps> = ({
    onClick,
    onFocus,
    href,
    keybindings = [],
    label,
    active,
}) => {
    const Tag = href ? 'a' : 'button'

    useEffect(() => {
        function handleKeyDown(event: KeyboardEvent): void {
            if (event.key === 'Enter' && active) {
                onClick?.()
            }
        }
        document.addEventListener('keydown', handleKeyDown)
        return () => document.removeEventListener('keydown', handleKeyDown)
    }, [onClick, active])

    useEffect(() => {
        if (active) {
            onFocus?.()
        }
    }, [active, onFocus])

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

interface NavigableListProps<T> {
    items: T[]
    children: (item: T, options: { active: boolean }) => JSX.Element
}

export function NavigableList<T>({ children, items }: NavigableListProps<T>): JSX.Element {
    const [activeIndex, setActiveIndex] = useState<number | undefined>()

    useEffect(() => {
        function handleKeyDown(event: KeyboardEvent): void {
            if (event.key === 'ArrowUp') {
                setActiveIndex(activeIndex => ((activeIndex || items.length) - 1) % items.length)
            } else if (event.key === 'ArrowDown') {
                setActiveIndex(activeIndex => ((activeIndex || 0) + 1) % items.length)
            }
        }
        document.addEventListener('keydown', handleKeyDown)
        return () => document.removeEventListener('keydown', handleKeyDown)
    }, [items])

    return (
        <ul className={styles.list}>
            {items.map((item, index) => (
                <React.Fragment key={index}>{children(item, { active: activeIndex === index })}</React.Fragment>
            ))}
        </ul>
    )
}

NavigableList.Item = NavigableListItem
