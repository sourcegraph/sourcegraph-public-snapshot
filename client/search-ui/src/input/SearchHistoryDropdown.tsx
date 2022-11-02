import React, {
    useState,
    useCallback,
    KeyboardEvent,
    KeyboardEventHandler,
    useRef,
    MouseEventHandler,
    MouseEvent,
} from 'react'

import { mdiHistory } from '@mdi/js'
import classNames from 'classnames'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { Shortcut } from '@sourcegraph/shared/src/react-shortcuts'
import { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'
// eslint-disable-next-line no-restricted-imports
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import {
    Icon,
    Popover,
    PopoverTrigger,
    PopoverContent,
    usePopoverContext,
    Flipping,
    Tooltip,
    PopoverOpenEventReason,
    PopoverOpenEvent,
} from '@sourcegraph/wildcard'

import styles from './SearchHistoryDropdown.module.scss'
import { shortcutDisplayName } from '@sourcegraph/shared/src/keyboardShortcuts'

const buttonContent: React.ReactElement = (
    <span className="text-nowrap">
        <Icon svgPath={mdiHistory} aria-label="Open search history" />
    </span>
)

interface SearchHistoryDropdownProps {
    recentSearches: RecentSearch[]
    onSelect: (search: RecentSearch) => void
    onClose: () => void
}

const recentSearchTooltipContent = `Recent searches ${shortcutDisplayName('Mod+ArrowDown')}`

export const SearchHistoryDropdown: React.FunctionComponent<SearchHistoryDropdownProps> = React.memo(
    ({ recentSearches = [], onSelect, onClose }) => {
        const [isOpen, setIsOpen] = useState(false)

        const handlePopoverToggle = useCallback(
            (event: PopoverOpenEvent): void => {
                const { isOpen, reason } = event

                setIsOpen(isOpen)

                if (reason === PopoverOpenEventReason.Esc) {
                    // In order to wait until Popover will be unmounted and popover internal focus
                    // management returns focus to the target, and then we call onEscapeMenuClose
                    // and put focus to the search box
                    requestAnimationFrame(() => onClose())
                }
            },
            [onClose]
        )

        return (
            <>
                <Popover isOpen={isOpen} onOpenChange={handlePopoverToggle}>
                    <Tooltip content={recentSearchTooltipContent}>
                        <PopoverTrigger
                            type="button"
                            className={classNames(styles.triggerButton, isOpen ? styles.triggerButtonOpen : null)}
                        >
                            {buttonContent}
                        </PopoverTrigger>
                    </Tooltip>
                    <PopoverContent flipping={Flipping.opposite}>
                        <SearchHistoryEntries
                            onSelect={search => {
                                onSelect(search)
                                setIsOpen(false)
                            }}
                            onEsc={() => {
                                setIsOpen(false)
                                onClose()
                            }}
                            recentSearches={recentSearches}
                        />
                    </PopoverContent>
                </Popover>
                <Shortcut onMatch={() => setIsOpen(true)} held={['Mod']} ordered={['ArrowDown']} />
            </>
        )
    }
)

interface SearchHistoryEntriesProps {
    recentSearches: RecentSearch[]
    onSelect: (search: RecentSearch) => void
    onEsc: () => void
}

const SearchHistoryEntries: React.FunctionComponent<SearchHistoryEntriesProps> = ({
    recentSearches,
    onSelect,
    onEsc,
}) => {
    const { isOpen } = usePopoverContext()
    const [selectedIndex, setSelectedIndex] = useState(0)
    const selectedIndexRef = useRef(selectedIndex)
    selectedIndexRef.current = selectedIndex

    const keydownHandler: KeyboardEventHandler<HTMLElement> = useCallback(
        (event: KeyboardEvent) => {
            switch (event.key) {
                case 'ArrowDown':
                    event.preventDefault()
                    setSelectedIndex(index => (index + 1) % recentSearches.length)
                    break
                case 'ArrowUp':
                    event.preventDefault()
                    setSelectedIndex(index => (index - 1 + recentSearches.length) % recentSearches.length)
                    break
                case 'Enter':
                    event.preventDefault()
                    if (recentSearches.length > 0) {
                        onSelect(recentSearches[selectedIndexRef.current])
                    }
                    break
                case 'Esc':
                    event.preventDefault()
                    onEsc()
                    break
            }
        },
        [setSelectedIndex, recentSearches, onSelect]
    )

    const clickHandler: MouseEventHandler<HTMLElement> = useCallback(
        (event: MouseEvent<HTMLElement>) => {
            if (event.button === 0) {
                event.preventDefault()
                const index = (event.target as HTMLElement).closest('li')?.dataset.index
                if (index !== undefined) {
                    onSelect(recentSearches[+index])
                }
            }
        },
        [setSelectedIndex, recentSearches]
    )

    if (!isOpen) {
        return null
    }

    if (recentSearches.length === 0) {
        return <div className="text-muted px-3 py-2">Your recent searches will appear here</div>
    }

    return (
        <ul
            className={styles.list}
            ref={ref => ref?.focus()}
            aria-expanded="true"
            role="listbox"
            tabIndex={0}
            onKeyDown={keydownHandler}
            onClick={clickHandler}
        >
            {recentSearches.map((search, index) => (
                <SearchHistoryEntry key={index} index={index} search={search} selected={index === selectedIndex} />
            ))}
        </ul>
    )
}

const SearchHistoryEntry: React.FunctionComponent<{
    index: number
    search: RecentSearch
    selected: boolean
}> = React.memo(({ index, search, selected }) => (
    <li role="option" data-index={index} aria-selected={selected}>
        <SyntaxHighlightedSearchQuery query={search.query} tabIndex={-1} />
        <span className="ml-1 text-nowrap text-muted">
            <Timestamp date={search.timestamp} />
        </span>
    </li>
))
SearchHistoryEntry.displayName = 'SearchHistoryEntry'
