import React, {
    type KeyboardEvent,
    type KeyboardEventHandler,
    type MouseEvent,
    type MouseEventHandler,
    useCallback,
    useRef,
    useState,
} from 'react'

import { mdiClockOutline } from '@mdi/js'
import classNames from 'classnames'

import { pluralize } from '@sourcegraph/common'
import type { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    createRectangle,
    Flipping,
    Icon,
    Popover,
    PopoverContent,
    type PopoverOpenEvent,
    PopoverTrigger,
    Tooltip,
    usePopoverContext,
} from '@sourcegraph/wildcard'

import { Timestamp } from '../../components/Timestamp'
import { SyntaxHighlightedSearchQuery } from '../components'

import styles from './SearchHistoryDropdown.module.scss'

const buttonContent: React.ReactElement = <Icon svgPath={mdiClockOutline} aria-hidden="true" />

interface SearchHistoryDropdownProps extends TelemetryProps, TelemetryV2Props {
    className?: string
    recentSearches: RecentSearch[]
    onSelect: (search: RecentSearch) => void
}

// Adds padding to the popover content to add some space between the trigger
// button and the content
const popoverPadding = createRectangle(0, 0, 0, 2)

export const SearchHistoryDropdown: React.FunctionComponent<SearchHistoryDropdownProps> = React.memo(
    function SearchHistoryDropdown({ recentSearches = [], onSelect, className, telemetryService, telemetryRecorder }) {
        const [isOpen, setIsOpen] = useState(false)

        const handlePopoverToggle = useCallback(
            (event: PopoverOpenEvent): void => {
                setIsOpen(event.isOpen)
                telemetryService.log(event.isOpen ? 'RecentSearchesListOpened' : 'RecentSearchesListDismissed')
                if (event.isOpen) {
                    telemetryRecorder.recordEvent('search.historyDropdown.recentSearchesList', 'open')
                } else {
                    telemetryRecorder.recordEvent('search.historyDropdown.recentSearchesList', 'dismiss')
                }
            },
            [telemetryService, telemetryRecorder, setIsOpen]
        )

        const onSelectInternal = useCallback(
            (search: RecentSearch) => {
                telemetryService.log('RecentSearchSelected')
                telemetryRecorder.recordEvent('search.historyDropdown.search', 'select')
                onSelect(search)
                setIsOpen(false)
            },
            [telemetryService, telemetryRecorder, onSelect, setIsOpen]
        )

        return (
            <>
                <Popover isOpen={isOpen} onOpenChange={handlePopoverToggle}>
                    <Tooltip content="Recent searches">
                        <PopoverTrigger
                            type="button"
                            className={classNames(styles.triggerButton, isOpen ? styles.open : null, className)}
                            aria-label="Open search history"
                        >
                            {buttonContent}
                        </PopoverTrigger>
                    </Tooltip>
                    <PopoverContent flipping={Flipping.opposite} targetPadding={popoverPadding}>
                        <SearchHistoryEntries onSelect={onSelectInternal} recentSearches={recentSearches} />
                    </PopoverContent>
                </Popover>
            </>
        )
    }
)

interface SearchHistoryEntriesProps {
    recentSearches: RecentSearch[]
    onSelect: (search: RecentSearch) => void
}

const SearchHistoryEntries: React.FunctionComponent<SearchHistoryEntriesProps> = ({ recentSearches, onSelect }) => {
    const { isOpen } = usePopoverContext()
    const [selectedIndex, setSelectedIndex] = useState(0)
    const selectedIndexRef = useRef(selectedIndex)
    selectedIndexRef.current = selectedIndex

    const keydownHandler: KeyboardEventHandler<HTMLElement> = useCallback(
        (event: KeyboardEvent) => {
            switch (event.key) {
                case 'ArrowDown': {
                    event.preventDefault()
                    setSelectedIndex(index => index + (index + 1 < recentSearches.length ? 1 : 0))
                    break
                }
                case 'ArrowUp': {
                    event.preventDefault()
                    setSelectedIndex(index => index - (index - 1 > -1 ? 1 : 0))
                    break
                }
                case 'Enter': {
                    event.preventDefault()
                    if (recentSearches.length > 0) {
                        onSelect(recentSearches[selectedIndexRef.current])
                    }
                    break
                }
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
        [recentSearches, onSelect]
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
                <SearchHistoryEntry
                    key={`${search.timestamp}-${search.query}`}
                    index={index}
                    search={search}
                    selected={index === selectedIndex}
                />
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
            <span className="sr-only">,</span>
            {search.resultCount !== undefined && (
                <span>
                    {`${search.resultCount}${search.limitHit ? '+' : ''} ${pluralize('result', search.resultCount)}`} •{' '}
                </span>
            )}
            <Timestamp date={search.timestamp} />
        </span>
    </li>
))
SearchHistoryEntry.displayName = 'SearchHistoryEntry'
