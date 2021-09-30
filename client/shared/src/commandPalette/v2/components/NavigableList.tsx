import classNames from 'classnames'
import React, { useState, useEffect, useRef } from 'react'

import { Keybinding } from '../../../keyboardShortcuts'

import styles from './NavigableList.module.scss'

interface NavigableListItemProps {
    onClick?: () => void
    onFocus?: () => void
    href?: string
    keybindings?: Keybinding[]
    active?: boolean
    // TODO icon (for symbol type, action item icon)
    icon?: JSX.Element
}

const NavigableListItem: React.FC<NavigableListItemProps> = ({
    onClick,
    onFocus,
    href,
    keybindings = [],
    children,
    active,
}) => {
    const Tag = href ? 'a' : 'button'

    const onClickReference = useRef(onClick)
    onClickReference.current = onClick

    useEffect(() => {
        function handleKeyDown(event: KeyboardEvent): void {
            if (event.key === 'Enter' && active) {
                onClickReference.current?.()
            }
        }
        document.addEventListener('keydown', handleKeyDown)
        return () => document.removeEventListener('keydown', handleKeyDown)
    }, [active])

    // TODO hack, find better way to do this.
    // Prevent infinite calls of onFocus when an item is active.
    const onFocusReference = useRef(onFocus)
    onFocusReference.current = onFocus

    useEffect(() => {
        if (active) {
            onFocusReference.current?.()
        }
    }, [active])

    return (
        <li tabIndex={-1}>
            <Tag
                type="button"
                tabIndex={0}
                className={classNames(styles.button, { [styles.buttonActive]: active })}
                onClick={onClick}
                href={href}
            >
                {children}

                <span className={styles.keybindings}>
                    {keybindings.map(({ ordered, held }, index) => (
                        <span tabIndex={-1} key={index} className={styles.keybinding}>
                            {[...(held || []), ...ordered].map((key, index) => (
                                <kbd key={index}>{key}</kbd>
                            ))}
                        </span>
                    ))}
                </span>
            </Tag>
        </li>
    )
}

interface NavigableListProps<T> {
    items: T[]
    children: (item: T, options: { active: boolean }) => JSX.Element
}

export function NavigableList<T>({ children, items }: NavigableListProps<T>): JSX.Element {
    const [activeIndex, setActiveIndex] = useState<number>(0)

    useEffect(() => {
        function handleKeyDown(event: KeyboardEvent): void {
            if (event.key === 'ArrowUp') {
                setActiveIndex(activeIndex => ((activeIndex || items.length) - 1) % items.length)
                // Prevent moving cursor for input
                event.preventDefault()
            } else if (event.key === 'ArrowDown') {
                setActiveIndex(activeIndex => (activeIndex + 1) % items.length)
                // Prevent moving cursor for input
                event.preventDefault()
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
