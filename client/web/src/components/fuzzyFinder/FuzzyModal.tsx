import React, { useState, useEffect, useMemo, KeyboardEvent, useLayoutEffect, useCallback } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'
import * as H from 'history'

import { pluralize } from '@sourcegraph/common'
import { Button, Modal, Icon, H3, Text, Input, useSessionStorage } from '@sourcegraph/wildcard'

import { AggregateFuzzySearch } from '../../fuzzyFinder/AggregateFuzzySearch'
import { FuzzySearch, FuzzySearchResult } from '../../fuzzyFinder/FuzzySearch'
import { mergedHandler } from '../../fuzzyFinder/WordSensitiveFuzzySearch'

import { FuzzyRepoRevision } from './FuzzyRepoRevision'
import { fuzzyErrors, FuzzyState, fuzzyIsActive, FuzzyTabs, FuzzyTabKey } from './FuzzyTabs'
import { HighlightedLink, linkStyle } from './HighlightedLink'

import styles from './FuzzyModal.module.scss'

const FUZZY_MODAL_RESULTS = 'fuzzy-modal-results'

// The number of results to jump by on PageUp/PageDown keyboard shortcuts.
const PAGE_DOWN_INCREMENT = 10

export interface FuzzyModalProps extends FuzzyState {
    initialMaxResults: number
    initialQuery: string
    onClose: () => void
    tabs: FuzzyTabs
    location: H.Location
}

function cleanupOldLocalStorage(): void {
    for (let index = 0; index < localStorage.length; index++) {
        const key = localStorage.key(index)
        if (key?.startsWith('fuzzy-modal.')) {
            localStorage.removeItem(key)
        }
    }
}

interface RenderProps {
    query: string
    result: FuzzySearchResult
    resultCount: number
    isComplete: boolean
    totalFileCount: number
}
interface QueryResult extends RenderProps {
    jsxElement: JSX.Element
}

function newFuzzySearch(activeTab: FuzzyTabKey, repoRevision: FuzzyRepoRevision, tabs: FuzzyTabs): FuzzySearch {
    const searches: FuzzySearch[] = []
    for (const [key, tab] of tabs.entries()) {
        if (!fuzzyIsActive(activeTab, repoRevision, key)) {
            continue
        }
        if (!tab.fsm) {
            continue
        }
        switch (tab.fsm.key) {
            case 'indexing':
                searches.push(tab.fsm.indexing.partialFuzzy)
                break
            case 'ready':
                searches.push(tab.fsm.fuzzy)
        }
    }
    if (searches.length === 1) {
        return searches[0]
    }
    return new AggregateFuzzySearch(searches)
}

function fuzzySearch(
    query: string,
    activeTab: FuzzyTabKey,
    repoRevision: FuzzyRepoRevision,
    maxResults: number,
    tabs: FuzzyTabs
): RenderProps {
    const search = newFuzzySearch(activeTab, repoRevision, tabs)
    const start = window.performance.now()
    const result = search.search({ query, maxResults })
    result.elapsedMilliseconds = window.performance.now() - start
    return {
        result,
        query,
        resultCount: Math.min(maxResults, result.links.length),
        isComplete: result.isComplete,
        totalFileCount: search.totalFileCount,
    }
}

function renderFuzzyResults(props: RenderProps, focusIndex: number, onClickItem: () => void): QueryResult {
    if (props.result.links.length === 0) {
        return {
            ...props,
            jsxElement: <Text>No matches for '{props.query}'</Text>,
        }
    }

    const linksToRender = props.result.links.slice(0, props.resultCount)
    const element = (
        <ul id={FUZZY_MODAL_RESULTS} className={styles.results} role="listbox" aria-label="Fuzzy finder results">
            {linksToRender.map((file, fileIndex) => (
                <li
                    id={fuzzyResultId(fileIndex)}
                    key={file.url || file.text}
                    role="option"
                    aria-selected={fileIndex === focusIndex}
                    className={classNames('p-1', fileIndex === focusIndex && styles.focused)}
                >
                    <HighlightedLink {...file} onClick={mergedHandler(file.onClick, onClickItem)} />
                </li>
            ))}
        </ul>
    )
    return {
        ...props,
        jsxElement: element,
    }
}

function emptyResults(element: JSX.Element): QueryResult {
    return {
        query: '',
        result: { isComplete: true, links: [] },
        resultCount: 0,
        isComplete: true,
        totalFileCount: 0,
        jsxElement: element,
    }
}

/**
 * Component that interactively displays filenames in the open repository when given fuzzy queries.
 *
 * Similar to "Go to file" in VS Code or the "t" keyboard shortcut on github.com
 */
export const FuzzyModal: React.FunctionComponent<React.PropsWithChildren<FuzzyModalProps>> = props => {
    const {
        onClickItem,
        query,
        setQuery,
        activeTab,
        setActiveTab,
        repoRevision,
        repoRevision: { repositoryName, revision },
    } = props
    // The "focus index" is the index of the file result that the user has
    // select with up/down arrow keys. The focused item is highlighted and the
    // window.location is moved to that URL when the user presses the enter key.
    const [focusIndex, setFocusIndex] = useSessionStorage(`fuzzy-modal.focus-index.${repositoryName}.${revision}`, 0)

    // Old versions of the fuzzy finder used local storage for the query and
    // focus index.  This logic attempts to remove old keys from localStorage
    // since we only use session storage now.
    useEffect(() => cleanupOldLocalStorage(), [])

    // The maximum number of results to display in the fuzzy finder. For large
    // repositories, a generic query like "src" may return thousands of results
    // making DOM rendering slow.  The user can increase this number by clicking
    // on a button at the bottom of the result list.
    const [maxResults, setMaxResults] = useState(props.initialMaxResults)

    // Stage 1: compute fuzzy results. Most importantly, this stage does not
    // depend on `focusIndex` so that we avoid re-running the fuzzy finder
    // whenever the user presses up/down to cycle through the results.
    const fuzzySearchResult = useMemo<RenderProps>(
        () => fuzzySearch(query, activeTab, repoRevision, maxResults, props.tabs),
        [maxResults, query, activeTab, repoRevision, props.tabs]
    )

    // Stage 2: render results from the fuzzy matcher.
    const queryResult = useMemo<QueryResult>(() => {
        if (props.tabs.isDownloading()) {
            return emptyResults(<Text>Downloading...</Text>)
        }
        const fsmErrors = fuzzyErrors(props.tabs, activeTab, repoRevision)
        if (fsmErrors.length > 0) {
            return emptyResults(<Text>Error: {JSON.stringify(fsmErrors)}</Text>)
        }
        return renderFuzzyResults(fuzzySearchResult, focusIndex, onClickItem)
    }, [activeTab, repoRevision, fuzzySearchResult, focusIndex, onClickItem, props.tabs])

    // Sets the new "focus index" so that it's rounded by the number of
    // displayed filenames.  Cycles so that the user can press-hold the down
    // arrow and it goes all the way down and back up to the top result.
    const setRoundedFocusIndex = useCallback(
        (increment: number): void => {
            const newNumber = focusIndex + increment
            const index = newNumber % queryResult.resultCount
            const nextIndex = index < 0 ? queryResult.resultCount + index : index
            setFocusIndex(nextIndex)
            document.querySelector(`#fuzzy-modal-result-${nextIndex}`)?.scrollIntoView(false)
        },
        [focusIndex, setFocusIndex, queryResult]
    )

    useLayoutEffect(() => {
        const input = document.querySelector<HTMLInputElement>('#fuzzy-modal-input')
        if (!input) {
            return
        }
        input.select()
        setFocusIndex(0)
    }, [activeTab, setFocusIndex])

    const onInputKeyDown = useCallback(
        (event: KeyboardEvent<HTMLInputElement>): void => {
            switch (true) {
                case event.key === 'Escape':
                    props.onClose()
                    break
                case event.key === 'ArrowDown':
                    event.preventDefault() // Don't move the cursor to the end of the input.
                    setRoundedFocusIndex(1)
                    break
                case event.key === 'PageDown':
                    setRoundedFocusIndex(PAGE_DOWN_INCREMENT)
                    break
                case event.key === 'ArrowUp':
                    event.preventDefault() // Don't move the cursor to the start of input.
                    setRoundedFocusIndex(-1)
                    break
                case event.key === 'PageUp':
                    setRoundedFocusIndex(-PAGE_DOWN_INCREMENT)
                    break
                case event.key === 'Enter':
                    if (focusIndex < queryResult.resultCount) {
                        const fileAnchor = document.querySelector<HTMLAnchorElement>(
                            `#fuzzy-modal-result-${focusIndex} .${linkStyle}`
                        )
                        fileAnchor?.click()
                    }
                    break
                case event.key === 'Tab':
                    if (props.tabs.isOnlyFilesEnabled()) {
                        break
                    }
                    event.preventDefault()
                    setActiveTab(props.tabs.focusTabWithIncrement(activeTab, event.shiftKey ? -1 : 1))
                default:
            }
        },
        [activeTab, setActiveTab, props, queryResult, focusIndex, setRoundedFocusIndex]
    )

    return (
        <Modal position="center" className={styles.modal} onDismiss={() => props.onClose()} aria-label="Fuzzy finder">
            <div className={styles.content}>
                <div className={styles.header} data-testid="fuzzy-modal-header">
                    {props.tabs.entries().map(([key, tab]) => (
                        <div
                            key={tab.title}
                            className={classNames(
                                'mb-0',
                                styles.tab,
                                styles.tab,
                                !props.tabs.isOnlyFilesEnabled() ? styles.tabCenteredText : '',
                                !props.tabs.isOnlyFilesEnabled() && activeTab === key ? styles.activeTab : ''
                            )}
                        >
                            <H3
                                role="link"
                                onClick={() => {
                                    const input = document.querySelector<HTMLInputElement>('#fuzzy-modal-input')
                                    if (input) {
                                        const nextActiveTab = props.tabs.focusNamedTab(key)
                                        if (nextActiveTab) {
                                            setActiveTab(nextActiveTab)
                                            input.focus()
                                        }
                                    }
                                }}
                            >
                                {props.tabs.isOnlyFilesEnabled() ? 'Find files' : tab.title}
                                {!props.tabs.isOnlyFilesEnabled() && tab?.shortcut}
                            </H3>
                        </div>
                    ))}
                    <Button variant="icon" onClick={() => props.onClose()} aria-label="Close">
                        <Icon className={styles.closeIcon} aria-hidden={true} svgPath={mdiClose} />
                    </Button>
                </div>
                <Input
                    id="fuzzy-modal-input"
                    autoComplete="off"
                    spellCheck="false"
                    role="combobox"
                    autoFocus={true}
                    aria-autocomplete="list"
                    aria-controls={FUZZY_MODAL_RESULTS}
                    aria-owns={FUZZY_MODAL_RESULTS}
                    aria-expanded={props.tabs.isDownloading()}
                    aria-activedescendant={fuzzyResultId(focusIndex)}
                    onFocus={input => input.target.select()}
                    className={styles.input}
                    placeholder="Enter a fuzzy query"
                    value={query}
                    onChange={event => {
                        setQuery(event.target.value)
                        setFocusIndex(0)
                    }}
                    onKeyDown={onInputKeyDown}
                />
                <div className={styles.summary}>
                    <FuzzyResultsSummary tabs={props.tabs} queryResult={queryResult} />
                </div>
                {queryResult.jsxElement}
                {!queryResult.isComplete && (
                    <Button
                        className={styles.showMore}
                        onClick={() => setMaxResults(maxResults + props.initialMaxResults)}
                        variant="secondary"
                    >
                        Show more
                    </Button>
                )}
            </div>
        </Modal>
    )
}

function plural(what: string, count: number, isComplete: boolean): string {
    return `${count.toLocaleString()}${isComplete ? '' : '+'} ${pluralize(what, count)}`
}
interface FuzzyResultsSummaryProps {
    tabs: FuzzyTabs
    queryResult: QueryResult
}

const FuzzyResultsSummary: React.FunctionComponent<React.PropsWithChildren<FuzzyResultsSummaryProps>> = ({
    tabs,
    queryResult,
}) => (
    <>
        <span data-testid="fuzzy-modal-summary" className={styles.resultCount}>
            {plural('result', queryResult.resultCount, queryResult.isComplete)} - {indexingProgressBar(tabs)}{' '}
            {plural('total', queryResult.totalFileCount, true)}
        </span>
        <i className="text-muted">
            <kbd>↑</kbd> and <kbd>↓</kbd> arrow keys browse. <kbd>⏎</kbd> selects.
        </i>
    </>
)

function indexingProgressBar(tabs: FuzzyTabs): JSX.Element {
    let indexedFiles = 0
    let totalFiles = 0
    for (const [, tab] of tabs.entries()) {
        if (!tab.fsm) {
            continue
        }
        if (tab.fsm.key === 'indexing') {
            indexedFiles += tab.fsm.indexing.indexedFileCount
            totalFiles += tab.fsm.indexing.totalFileCount
        }
    }
    if (indexedFiles === 0 && totalFiles === 0) {
        return <></>
    }
    const percentage = Math.round((indexedFiles / totalFiles) * 100)
    return (
        <progress value={indexedFiles} max={totalFiles}>
            {percentage}%
        </progress>
    )
}

function fuzzyResultId(id: number): string {
    return `fuzzy-modal-result-${id}`
}
