import classNames from 'classnames'
import React, { useState, useEffect, useRef } from 'react'

import { Link } from '../../../components/Link'
import { Keybinding } from '../../../keyboardShortcuts'

import styles from './NavigableList.module.scss'

interface NavigableListItemProps {
    onClick?: () => void
    onFocus?: () => void
    href?: string
    isExternalLink?: boolean
    keybindings?: Keybinding[]
    active?: boolean
    // TODO icon (for symbol type, action item icon)
    icon?: JSX.Element
}

const NavigableListItem: React.FC<NavigableListItemProps> = React.memo(
    ({ onClick, onFocus, href, isExternalLink, keybindings = [], children, active }) => {
        // TODO: refactor to support href and closing command palette + target=_blank
        const Tag = href ? Link : 'button'

        const listItemReference = useRef<HTMLLIElement | null>(null)

        const onClickReference = useRef(onClick)
        onClickReference.current = onClick

        // TODO external vs internal link
        useEffect(() => {
            function handleKeyDown(event: KeyboardEvent): void {
                if (event.key === 'Enter' && active) {
                    event.preventDefault()
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
                listItemReference.current?.scrollIntoView(false)
                onFocusReference.current?.()
            }
        }, [active])

        // Open in new tab if an external link
        const newTabProps =
            href && isExternalLink
                ? {
                      target: '_blank',
                      rel: 'noopener noreferrer',
                  }
                : {}

        return (
            <li tabIndex={-1} ref={listItemReference}>
                <Tag
                    type="button"
                    tabIndex={0}
                    className={classNames(styles.button, { [styles.buttonActive]: active })}
                    onClick={onClick}
                    to={href ?? ''}
                    {...newTabProps}
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
)

interface NavigableListProps<T> {
    items: T[]
    getKey?: (item: T) => string
    children: (item: T, options: { active: boolean }) => JSX.Element | null
}

export function NavigableList<T>({ children, items, getKey }: NavigableListProps<T>): JSX.Element | null {
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

    if (items.length === 0) {
        return null
    }

    return (
        <ul className={styles.list}>
            {items.map((item, index) => (
                <React.Fragment key={getKey?.(item) || index}>
                    {children(item, { active: activeIndex === index })}
                </React.Fragment>
            ))}
        </ul>
    )
}

NavigableList.Item = NavigableListItem
