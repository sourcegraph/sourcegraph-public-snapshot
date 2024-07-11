import React, {
    useState,
    useEffect,
    useRef,
    useMemo,
    type KeyboardEvent,
    useLayoutEffect,
    useCallback,
    type SetStateAction,
    type Dispatch,
} from 'react'

import { mdiClose } from '@mdi/js'
import type { TabsProps } from '@reach/tabs'
import classNames from 'classnames'
import type * as H from 'history'
import { escapeRegExp } from 'lodash'

import { pluralize } from '@sourcegraph/common'
import { KEYBOARD_SHORTCUTS } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import {
    Button,
    Modal,
    Icon,
    Text,
    Input,
    useSessionStorage,
    Code,
    Link,
    Tabs,
    Tab,
    Select,
    H3,
    TabPanels,
    TabPanel,
    TabList,
    Badge,
    LoadingSpinner,
} from '@sourcegraph/wildcard'

import { AggregateFuzzySearch } from '../../fuzzyFinder/AggregateFuzzySearch'
import type { FuzzySearch, FuzzySearchResult } from '../../fuzzyFinder/FuzzySearch'
import type { SearchValueRankingCache } from '../../fuzzyFinder/SearchValueRankingCache'
import { mergedHandler } from '../../fuzzyFinder/WordSensitiveFuzzySearch'
import { Keybindings } from '../KeyboardShortcutsHelp/KeyboardShortcutsHelp'

import { parseFuzzyFileQuery } from './FuzzyFiles'
import { fuzzyErrors, type FuzzyState, type FuzzyTabs, type FuzzyTabKey, type FuzzyScope } from './FuzzyTabs'
import { HighlightedLink, type HighlightedLinkProps, linkStyle } from './HighlightedLink'

import styles from './FuzzyModal.module.scss'

const FUZZY_MODAL_RESULTS = 'fuzzy-modal-results'

// The number of results to jump by on PageUp/PageDown keyboard shortcuts.
const PAGE_DOWN_INCREMENT = 10

export interface FuzzyModalProps extends FuzzyState {
    initialMaxResults: number
    initialQuery: string
    onClose: () => void
    onClickResult: () => void
    onClickGoToResultsPage: () => void
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
    fsmGeneration: number
    result: FuzzySearchResult
    resultCount: number
    isComplete: boolean
    totalFileCount: number
}
interface QueryResult extends RenderProps {
    jsxElement: JSX.Element
}

function newFuzzySearch(query: string, activeTab: FuzzyTabKey, scope: FuzzyScope, tabs: FuzzyTabs): FuzzySearch {
    const searches: FuzzySearch[] = []
    for (const tab of tabs.fsms) {
        if (!tab.isActive(activeTab, scope)) {
            continue
        }
        tab.onQuery?.(query) // trigger downloads
        const fsm = tab.fsm()
        switch (fsm.key) {
            case 'downloading': {
                if (fsm.downloading?.partialFuzzy) {
                    searches.push(fsm.downloading.partialFuzzy)
                }
                break
            }
            case 'indexing': {
                searches.push(fsm.indexing.partialFuzzy)
                break
            }
            case 'ready': {
                searches.push(fsm.fuzzy)
            }
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
    scope: FuzzyScope,
    maxResults: number,
    tabs: FuzzyTabs,
    fsmGeneration: number,
    cache: SearchValueRankingCache
): RenderProps {
    const search = newFuzzySearch(query, activeTab, scope, tabs)
    const start = window.performance.now()
    const result = search.search({ query, maxResults, cache })
    result.elapsedMilliseconds = window.performance.now() - start
    return {
        result,
        query,
        fsmGeneration,
        resultCount: Math.min(maxResults, result.links.length),
        isComplete: result.isComplete,
        totalFileCount: search.totalFileCount,
    }
}

interface ResultProps {
    fileIndex: number
    file: HighlightedLinkProps
    isSelected: boolean
    onClickItem: () => void
}

const Result: React.FC<ResultProps> = ({ fileIndex, file, isSelected, onClickItem }) => {
    const ref = useRef<HTMLLIElement>(null)

    useEffect(() => {
        if (isSelected && ref.current) {
            ref.current.scrollIntoView({ block: 'nearest' })
        }
    }, [isSelected])

    return (
        <li
            // This ID is required to make the `Enter` shortcut work.
            id={fuzzyResultId(fileIndex)}
            ref={ref}
            role="option"
            aria-selected={isSelected}
            className={classNames(
                'd-flex align-items-center py-1 px-3 rounded-0',
                styles.resultItem,
                isSelected && styles.focused
            )}
        >
            <HighlightedLink {...file} onClick={mergedHandler(file.onClick, onClickItem)} />
        </li>
    )
}

function renderFuzzyResults(
    props: RenderProps,
    focusIndex: number,
    maxResults: number,
    initialMaxResults: number,
    setMaxResults: Dispatch<SetStateAction<number>>,
    onClickItem: () => void
): QueryResult {
    if (props.result.links.length === 0) {
        return {
            ...props,
            jsxElement: (
                // See original comment on FuzzyState.fsmGeneration for details
                // why we include this arbitrary number here. It's an arbitrary
                // decision to place the number here, as long as the number is
                // recorded as a dependency to `renderFuzzyResults` then it
                // should work OK.
                <Text data-fsm-generation={props.fsmGeneration} className="p-3 text-center text-muted">
                    No matches
                </Text>
            ),
        }
    }

    const linksToRender = props.result.links.slice(0, props.resultCount)
    const element = (
        <ul id={FUZZY_MODAL_RESULTS} role="listbox" aria-label="Fuzzy finder results" className="py-1 px-0 mb-0">
            {linksToRender.map((file, fileIndex) => (
                <Result
                    key={file.url || file.text}
                    file={file}
                    fileIndex={fileIndex}
                    isSelected={focusIndex === fileIndex}
                    onClickItem={onClickItem}
                />
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
        fsmGeneration: 0,
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
        initialMaxResults,
        onClose,
        onClickResult,
        onClickGoToResultsPage,
        fsmGeneration,
        rankingCache,
        query,
        setQuery,
        tabs,
        isScopeToggleDisabled,
        setScope,
        scope,
        activeTab,
        setActiveTab,
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
    const [maxResults, setMaxResults] = useState(initialMaxResults)

    // Stage 1: compute fuzzy results. Most importantly, this stage does not
    // depend on `focusIndex` so that we avoid re-running the fuzzy finder
    // whenever the user presses up/down to cycle through the results.
    const fuzzySearchResult = useMemo<RenderProps>(
        () => fuzzySearch(query, activeTab, scope, maxResults, tabs, fsmGeneration, rankingCache),
        [fsmGeneration, maxResults, query, activeTab, scope, tabs, rankingCache]
    )

    // Stage 2: render results from the fuzzy matcher.
    const queryResult = useMemo<QueryResult>(() => {
        const fsmErrors = fuzzyErrors(tabs, activeTab, scope)
        if (fsmErrors.length > 0) {
            return emptyResults(<Text>Error: {JSON.stringify(fsmErrors)}</Text>)
        }
        return renderFuzzyResults(
            fuzzySearchResult,
            focusIndex,
            maxResults,
            initialMaxResults,
            setMaxResults,
            onClickResult
        )
    }, [
        activeTab,
        scope,
        fuzzySearchResult,
        focusIndex,
        maxResults,
        initialMaxResults,
        setMaxResults,
        onClickResult,
        tabs,
    ])

    // Sets the new "focus index" so that it's rounded by the number of
    // displayed filenames.  Cycles so that the user can press-hold the down
    // arrow and it goes all the way down and back up to the top result.
    const setRoundedFocusIndex = useCallback(
        (increment: number): void => {
            const newNumber = focusIndex + increment
            const index = newNumber % queryResult.resultCount
            const nextIndex = index < 0 ? queryResult.resultCount + index : index
            setFocusIndex(nextIndex)
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
                case event.key === 'Escape': {
                    onClose()
                    break
                }
                case event.key === 'g' && event.ctrlKey: {
                    // common Emacs binding to close things
                    onClose()
                    break
                }
                case event.key === 'n' && event.ctrlKey: {
                    event.preventDefault()
                    setRoundedFocusIndex(1)
                    break
                }
                case event.key === 'p' && event.ctrlKey: {
                    event.preventDefault()
                    setRoundedFocusIndex(-1)
                    break
                }
                case event.key === 'ArrowDown': {
                    event.preventDefault() // Don't move the cursor to the end of the input.
                    setRoundedFocusIndex(1)
                    break
                }
                case event.key === 'PageDown': {
                    setRoundedFocusIndex(PAGE_DOWN_INCREMENT)
                    break
                }
                case event.key === 'ArrowUp': {
                    event.preventDefault() // Don't move the cursor to the start of input.
                    setRoundedFocusIndex(-1)
                    break
                }
                case event.key === 'PageUp': {
                    setRoundedFocusIndex(-PAGE_DOWN_INCREMENT)
                    break
                }
                case event.key === 'Enter': {
                    if (focusIndex < queryResult.resultCount) {
                        const fileAnchor = document.querySelector<HTMLAnchorElement>(
                            `#fuzzy-modal-result-${focusIndex} .${linkStyle}`
                        )
                        fileAnchor?.click()
                    }
                    break
                }
                default:
            }
        },
        [onClose, queryResult, focusIndex, setRoundedFocusIndex]
    )

    const showTabs = !tabs.isOnlyFilesEnabled()
    const WrapperComponent = showTabs ? Tabs : 'div'
    const wrapperComponentProps: TabsProps | {} = showTabs
        ? {
              size: 'large',
              index: tabs.activeIndex(activeTab),
              onChange: (index: number) => setActiveTab(tabs.focusTab(index)),
          }
        : {}

    return (
        <Modal
            position="center"
            className={styles.modal}
            onDismiss={() => onClose()}
            aria-label={tabs.underlying[activeTab].title}
        >
            <WrapperComponent
                className="d-flex align-items-stretch flex-column h-100 bg-transparent"
                {...wrapperComponentProps}
            >
                <div
                    className={classNames(
                        'd-flex justify-space-between align-items-center pt-2 pb-0 px-3',
                        styles.header
                    )}
                    data-testid="fuzzy-modal-header"
                >
                    {showTabs ? (
                        <TabList className={styles.tabList}>
                            {tabs.entries().map(([key, tab]) => (
                                <Tab key={key} className={styles.tab} data-testid={key}>
                                    {tab.title}
                                    <span className={styles.shortcut}>
                                        {tab?.plaintextShortcut && ' ' + tab.plaintextShortcut}
                                    </span>
                                </Tab>
                            ))}
                        </TabList>
                    ) : (
                        <H3>Find files</H3>
                    )}
                    <Badge
                        variant="info"
                        href="https://community.sourcegraph.com/t/experimental-fuzzy-finder-in-code-search/230"
                        tooltip="Provide feedback on this experimental feature"
                        className={styles.experimentalBadge}
                    >
                        Experimental
                    </Badge>
                    <Button variant="icon" onClick={onClose} aria-label="Close" className={styles.closeButton}>
                        <Icon aria-hidden={true} svgPath={mdiClose} />
                    </Button>
                </div>
                <hr className="mt-0 mb-2 w-100" />
                <Input
                    id="fuzzy-modal-input"
                    autoComplete="off"
                    spellCheck="false"
                    role="combobox"
                    autoFocus={true}
                    aria-autocomplete="list"
                    aria-controls={FUZZY_MODAL_RESULTS}
                    aria-owns={FUZZY_MODAL_RESULTS}
                    aria-expanded={tabs.isDownloading(activeTab, scope)}
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
                    <FuzzyResultsSummary activeTab={activeTab} scope={scope} tabs={tabs} queryResult={queryResult} />
                    {!tabs.isOnlyFilesEnabled() && (
                        <span className={classNames(styles.fuzzyScopeSelector)}>
                            <ScopeSelect
                                activeTab={activeTab}
                                scope={scope}
                                isScopeToggleDisabled={isScopeToggleDisabled}
                                setScope={setScope}
                            />
                        </span>
                    )}
                </div>
                <hr className="my-0 w-100" />
                {showTabs ? (
                    <TabPanels className="flex-1 overflow-auto">
                        {tabs.entries().map(([key]) => (
                            <TabPanel key={key}>{activeTab === key && queryResult.jsxElement}</TabPanel>
                        ))}
                    </TabPanels>
                ) : (
                    <div className="flex-1 overflow-auto">{queryResult.jsxElement}</div>
                )}
                <hr className="my-0 w-100" />
                <div className="d-flex align-items-center w-100 p-3">
                    <SearchQueryLink {...props} onClickItem={onClickGoToResultsPage} />
                    <span className="ml-auto mr-2">
                        <ArrowKeyExplanation />
                    </span>
                </div>
            </WrapperComponent>
        </Modal>
    )
}

function plural(what: string, count: number, isComplete: boolean): string {
    return `${count.toLocaleString()}${isComplete ? '' : '+'} ${pluralize(what, count)}`
}

const ArrowKeyExplanation: React.FunctionComponent = () => (
    <span className={styles.keyboardExplanation}>
        Press <kbd>↑</kbd>
        <kbd>↓</kbd> to navigate through results
    </span>
)

interface ScopeSelectProps {
    activeTab: FuzzyTabKey
    scope: FuzzyScope
    setScope: Dispatch<SetStateAction<FuzzyScope>>
    isScopeToggleDisabled: boolean
}

const ToggleShortcut: React.FunctionComponent<{ activeTab: FuzzyTabKey }> = ({ activeTab }) => {
    switch (activeTab) {
        case 'all': {
            return <Keybindings uppercaseOrdered={true} keybindings={KEYBOARD_SHORTCUTS.fuzzyFinder.keybindings} />
        }
        case 'files': {
            return <Keybindings uppercaseOrdered={true} keybindings={KEYBOARD_SHORTCUTS.fuzzyFinderFiles.keybindings} />
        }
        case 'symbols': {
            return (
                <Keybindings uppercaseOrdered={true} keybindings={KEYBOARD_SHORTCUTS.fuzzyFinderSymbols.keybindings} />
            )
        }
        default: {
            return <></>
        }
    }
}

const ScopeSelect: React.FunctionComponent<ScopeSelectProps> = ({
    activeTab,
    scope,
    setScope,
    isScopeToggleDisabled,
}) => (
    <Select
        label=""
        isCustomStyle={true}
        id="fuzzy-scope"
        value={scope}
        selectSize="sm"
        className={styles.fuzzyScopeSelector}
        disabled={isScopeToggleDisabled}
        onChange={value => {
            switch (value.target.value) {
                case 'everywhere':
                case 'repository': {
                    setScope(value.target.value)
                    focusFuzzyInput()
                }
            }
        }}
    >
        <option value="everywhere">
            <ToggleShortcut activeTab={activeTab} /> Searching everywhere
        </option>
        <option value="repository">
            <ToggleShortcut activeTab={activeTab} /> Searching in this repository
        </option>
    </Select>
)

const SearchQueryLink: React.FunctionComponent<FuzzyState & { onClickItem: () => void }> = props => {
    const { onClickItem, scope } = props
    const searchQueryLink = useCallback(
        (query: string): JSX.Element => {
            const searchParams = new URLSearchParams()
            searchParams.set('q', query)
            const url = `/search?${searchParams.toString()}`
            return (
                <Code>
                    <Link to={url} onClick={onClickItem}>
                        {query}
                    </Link>{' '}
                </Code>
            )
        },
        [onClickItem]
    )
    const isScopeEverywhere = scope === 'everywhere'
    switch (props.activeTab) {
        case 'symbols': {
            return searchQueryLink(`type:symbol ${props.query}${isScopeEverywhere ? '' : repoFilter(props)}`)
        }
        case 'files': {
            return searchQueryLink(
                `type:path ${parseFuzzyFileQuery(props.query).filename}${isScopeEverywhere ? '' : repoFilter(props)}`
            )
        }
        case 'repos': {
            return searchQueryLink(`type:repo ${props.query}`)
        }
        case 'all': {
            return searchQueryLink(`${props.query}${isScopeEverywhere ? '' : repoFilter(props)}`)
        }
        default: {
            return <></>
        }
    }
}

function repoFilter(state: FuzzyState): string {
    const isGlobal = !state.repoRevision.repositoryName
    const revision = state.repoRevision.revision ? `@${state.repoRevision.revision}` : ''
    return isGlobal ? '' : ` repo:^${escapeRegExp(state.repoRevision.repositoryName)}$${revision}`
}

interface FuzzyResultsSummaryProps {
    activeTab: FuzzyTabKey
    scope: FuzzyScope
    tabs: FuzzyTabs
    queryResult: QueryResult
}

const FuzzyResultsSummary: React.FunctionComponent<React.PropsWithChildren<FuzzyResultsSummaryProps>> = ({
    activeTab,
    scope,
    tabs,
    queryResult,
}) => {
    let indexedFiles = 0
    let totalFiles = 0
    const downloadingTabs: string[] = []
    for (const tab of tabs.fsms) {
        if (!tab.isActive(activeTab, scope)) {
            continue
        }
        const fsm = tab.fsm()
        if (fsm.key === 'downloading') {
            downloadingTabs.push(tab.key)
        }
        if (fsm.key === 'indexing') {
            indexedFiles += fsm.indexing.indexedFileCount
            totalFiles += fsm.indexing.totalFileCount
        }
    }
    return (
        <span data-testid="fuzzy-modal-summary" className={styles.resultCount}>
            {plural('result', queryResult.resultCount, queryResult.isComplete)} out of{' '}
            {plural('total', queryResult.totalFileCount, true)}
            <ProgressBar value={indexedFiles} max={totalFiles} />
            {downloadingTabs.length > 0 && <LoadingSpinner />}
        </span>
    )
}

interface ProgressBarProps {
    value: number
    max: number
}

const ProgressBar: React.FunctionComponent<ProgressBarProps> = ({ value, max }) => {
    if (max === 0) {
        return <></>
    }
    const percentage = Math.round((value / max) * 100)
    return (
        <progress value={value} max={max}>
            {percentage}%
        </progress>
    )
}

function fuzzyResultId(id: number): string {
    return `fuzzy-modal-result-${id}`
}

function focusFuzzyInput(): void {
    // Redirect the focus to the fuzzy search bar
    const input = document.querySelector<HTMLInputElement>('#fuzzy-modal-input')
    input?.focus()
}
