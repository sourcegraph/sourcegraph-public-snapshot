import { ApolloError } from '@apollo/client'
import Dialog from '@reach/dialog'
import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useState, useEffect } from 'react'

import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { CaseInsensitiveFuzzySearch } from '../../fuzzyFinder/CaseInsensitiveFuzzySearch'
import { FuzzySearch, FuzzySearchResult, SearchIndexing, SearchValue } from '../../fuzzyFinder/FuzzySearch'
import { WordSensitiveFuzzySearch } from '../../fuzzyFinder/WordSensitiveFuzzySearch'
import { parseBrowserRepoURL } from '../../util/url'

import { Indexing, FuzzyFSM } from './FuzzyFinder'
import styles from './FuzzyModal.module.scss'
import { HighlightedLink } from './HighlightedLink'

// The default value of 80k filenames is picked from the following observations:
// - case-insensitive search is slow but works in the torvalds/linux repo (72k files)
// - case-insensitive search is almost unusable in the chromium/chromium repo (360k files)
const DEFAULT_CASE_INSENSITIVE_FILE_COUNT_THRESHOLD = 80000

const FUZZY_MODAL_TITLE = 'fuzzy-modal-title'
const FUZZY_MODAL_RESULTS = 'fuzzy-modal-results'

// Cache for the last fuzzy query. This value is only used to avoid redoing the
// full fuzzy search on every re-render when the user presses the down/up arrow
// keys to move the "focus index".
const lastFuzzySearchResult = new Map<string, FuzzySearchResult>()

// The number of results to jump by on PageUp/PageDown keyboard shortcuts.
const PAGE_DOWN_INCREMENT = 10

export interface FuzzyModalProps {
    repoName: string
    commitID: string
    initialMaxResults: number
    initialQuery: string
    downloadFilenames: string[]
    isLoading: boolean
    isError: ApolloError | undefined
    onClose: () => void
    fsm: FuzzyFSM
    setFsm: (fsm: FuzzyFSM) => void
}

/**
 * Component that interactively displays filenames in the open repository when given fuzzy queries.
 *
 * Similar to "Go to file" in VS Code or the "t" keyboard shortcut on github.com
 */
export const FuzzyModal: React.FunctionComponent<FuzzyModalProps> = props => {
    // NOTE: the query is cached in local storage to mimic the file pickers in
    // IntelliJ (by default) and VS Code (when "Workbench > Quick Open >
    // Preserve Input" is enabled).
    const [query, setQuery] = useLocalStorage(`fuzzy-modal.query.${props.repoName}`, props.initialQuery)

    // The "focus index" is the index of the file result that the user has
    // select with up/down arrow keys. The focused item is highlighted and the
    // window.location is moved to that URL when the user presses the enter key.
    const [focusIndex, setFocusIndex] = useState(0)

    // The maximum number of results to display in the fuzzy finder. For large
    // repositories, a generic query like "src" may return thousands of results
    // making DOM rendering slow.  The user can increase this number by clicking
    // on a button at the bottom of the result list.
    const [maxResults, setMaxResults] = useState(props.initialMaxResults)

    const [resultsCount, setResultsCount] = useState(0)
    const [isComplete, setIsComplete] = useState<boolean>(false)
    const [totalFileCount, setTotalFileCount] = useState(0)
    const [fuzzyResultElement, setFuzzyResultElement] = useState<JSX.Element>()

    useEffect(() => {
        function handleEmpty(props: FuzzyModalProps): void {
            props.setFsm(handleFilenames(props.downloadFilenames))
            cleanLegacyCacheStorage()
        }

        function onError(what: string): (error: Error) => void {
            return error => {
                props.setFsm({ key: 'failed', errorMessage: JSON.stringify(error) })
                throw new Error(what)
            }
        }

        function empty(element: JSX.Element): void {
            setFuzzyResultElement(element)
            setResultsCount(0)
            setIsComplete(true)
            setTotalFileCount(0)
        }

        function renderFiles(search: FuzzySearch, indexing?: SearchIndexing): void {
            // Parse the URL here instead of accepting it as a React prop because the
            // URL can change based on shortcuts like `y` that won't trigger a re-render
            // in React. By parsing the URL here, we avoid the risk of rendering links to a revision that
            // doesn't match the active revision in the browser's address bar.
            const repoUrl = parseBrowserRepoURL(location.pathname + location.search + location.hash)
            const indexedFileCount = indexing ? indexing.indexedFileCount : ''
            const cacheKey = `${query}-${maxResults}${indexedFileCount}-${repoUrl.revision || ''}`
            let fuzzyResult = lastFuzzySearchResult.get(cacheKey)
            if (!fuzzyResult) {
                const start = window.performance.now()
                fuzzyResult = search.search({
                    query,
                    maxResults,
                    createUrl: filename =>
                        toPrettyBlobURL({
                            filePath: filename,
                            revision: repoUrl.revision,
                            repoName: props.repoName,
                            commitID: props.commitID,
                        }),
                    onClick: () => props.onClose(),
                })
                fuzzyResult.elapsedMilliseconds = window.performance.now() - start
                lastFuzzySearchResult.clear() // Only cache the last query.
                lastFuzzySearchResult.set(cacheKey, fuzzyResult)
            }
            const links = fuzzyResult.links
            if (links.length === 0) {
                setFuzzyResultElement(<p>No files matching '{query}'</p>)
                setResultsCount(0)
                setTotalFileCount(search.totalFileCount)
                return setIsComplete(fuzzyResult.isComplete)
            }

            const linksToRender = links.slice(0, maxResults)
            const element = (
                <ul
                    id={FUZZY_MODAL_RESULTS}
                    className={styles.results}
                    role="listbox"
                    aria-label="Fuzzy finder results"
                >
                    {linksToRender.map((file, fileIndex) => (
                        <li
                            id={fuzzyResultId(fileIndex)}
                            key={file.text}
                            role="option"
                            aria-selected={fileIndex === focusIndex}
                            className={classNames('p-1', fileIndex === focusIndex && styles.focused)}
                        >
                            <HighlightedLink {...file} />
                        </li>
                    ))}
                </ul>
            )
            setFuzzyResultElement(element)
            setResultsCount(linksToRender.length)
            setTotalFileCount(search.totalFileCount)
            return setIsComplete(fuzzyResult.isComplete)
        }

        if (props.isLoading) {
            return empty(<p>Downloading...</p>)
        }

        if (props.isError) {
            return empty(<p>Error: {JSON.stringify(props.isError)}</p>)
        }

        switch (props.fsm.key) {
            case 'empty':
                handleEmpty(props)
                return empty(<></>)
            case 'downloading':
                return empty(<p>Downloading...</p>)
            case 'failed':
                return empty(<p>Error: {props.fsm.errorMessage}</p>)
            case 'indexing': {
                const loader = props.fsm.indexing
                later()
                    .then(() => continueIndexing(loader))
                    .then(next => props.setFsm(next), onError('onIndexing'))

                return renderFiles(props.fsm.indexing.partialFuzzy, props.fsm.indexing)
            }
            case 'ready':
                return renderFiles(props.fsm.fuzzy)
            default:
                return empty(<p>ERROR</p>)
        }
    }, [props, focusIndex, maxResults, query])

    // Sets the new "focus index" so that it's rounded by the number of
    // displayed filenames.  Cycles so that the user can press-hold the down
    // arrow and it goes all the way down and back up to the top result.
    function setRoundedFocusIndex(increment: number): void {
        const newNumber = focusIndex + increment
        const index = newNumber % resultsCount
        const nextIndex = index < 0 ? resultsCount + index : index
        setFocusIndex(nextIndex)
        document.querySelector(`#fuzzy-modal-result-${nextIndex}`)?.scrollIntoView(false)
    }

    function onInputKeyDown(event: React.KeyboardEvent): void {
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
                if (focusIndex < resultsCount) {
                    const fileAnchor = document.querySelector<HTMLAnchorElement>(`#fuzzy-modal-result-${focusIndex} a`)
                    fileAnchor?.click()
                    setQuery('')
                    props.onClose()
                }
                break
            case event.key === 'p' && (event.metaKey || event.ctrlKey):
                event.preventDefault()
            default:
        }
    }

    return (
        <Dialog
            className={classNames(styles.modal, 'modal-body p-4 rounded border')}
            onDismiss={() => props.onClose()}
            aria-labelledby={FUZZY_MODAL_TITLE}
        >
            <div className={styles.content}>
                <div className={styles.header}>
                    <h3 className="mb-0" id={FUZZY_MODAL_TITLE}>
                        Find file
                    </h3>
                    <button type="button" className="btn btn-icon" onClick={() => props.onClose()} aria-label="Close">
                        <CloseIcon className={classNames('icon-inline', styles.closeIcon)} />
                    </button>
                </div>
                <input
                    autoComplete="off"
                    spellCheck="false"
                    role="combobox"
                    autoFocus={true}
                    aria-autocomplete="list"
                    aria-controls={FUZZY_MODAL_RESULTS}
                    aria-owns={FUZZY_MODAL_RESULTS}
                    aria-expanded={props.fsm.key !== 'downloading'}
                    aria-activedescendant={fuzzyResultId(focusIndex)}
                    id="fuzzy-modal-input"
                    className={classNames('form-control py-1', styles.input)}
                    placeholder="Enter a partial file path or name"
                    value={query}
                    onChange={({ target: { value } }) => {
                        setQuery(value)
                        setFocusIndex(0)
                    }}
                    type="text"
                    onKeyDown={onInputKeyDown}
                />
                <div className={styles.summary}>
                    <FuzzyResultsSummary
                        fsm={props.fsm}
                        resultsCount={resultsCount}
                        isComplete={isComplete}
                        totalFileCount={totalFileCount}
                    />
                </div>
                {fuzzyResultElement}
                {!isComplete && (
                    <button
                        className={classNames('btn btn-secondary', styles.showMore)}
                        type="button"
                        onClick={() => setMaxResults(maxResults + props.initialMaxResults)}
                    >
                        Show more
                    </button>
                )}
            </div>
        </Dialog>
    )
}

function plural(what: string, count: number, isComplete: boolean): string {
    return `${count.toLocaleString()}${isComplete ? '' : '+'} ${pluralize(what, count)}`
}
interface FuzzyResultsSummaryProps {
    fsm: FuzzyFSM
    resultsCount: number
    isComplete: boolean
    totalFileCount: number
}

const FuzzyResultsSummary: React.FunctionComponent<FuzzyResultsSummaryProps> = ({
    fsm,
    resultsCount,
    isComplete,
    totalFileCount,
}) => (
    <>
        <span className={styles.resultCount}>
            {plural('result', resultsCount, isComplete)} - {fsm.key === 'indexing' && indexingProgressBar(fsm)}{' '}
            {plural('total file', totalFileCount, true)}
        </span>
        <i className="text-muted">
            <kbd>↑</kbd> and <kbd>↓</kbd> arrow keys browse. <kbd>⏎</kbd> selects.
        </i>
    </>
)

function indexingProgressBar(indexing: Indexing): JSX.Element {
    const indexedFiles = indexing.indexing.indexedFileCount
    const totalFiles = indexing.indexing.totalFileCount
    const percentage = Math.round((indexedFiles / totalFiles) * 100)
    return (
        <progress value={indexedFiles} max={totalFiles}>
            {percentage}%
        </progress>
    )
}

async function later(): Promise<void> {
    return new Promise(resolve => setTimeout(() => resolve(), 0))
}

async function continueIndexing(indexing: SearchIndexing): Promise<FuzzyFSM> {
    const next = await indexing.continue()
    if (next.key === 'indexing') {
        return { key: 'indexing', indexing: next }
    }
    return {
        key: 'ready',
        fuzzy: next.value,
    }
}

function handleFilenames(filenames: string[]): FuzzyFSM {
    const values: SearchValue[] = filenames.map(file => ({ text: file }))
    if (filenames.length < DEFAULT_CASE_INSENSITIVE_FILE_COUNT_THRESHOLD) {
        return {
            key: 'ready',
            fuzzy: new CaseInsensitiveFuzzySearch(values),
        }
    }
    const indexing = WordSensitiveFuzzySearch.fromSearchValuesAsync(values)
    if (indexing.key === 'ready') {
        return {
            key: 'ready',
            fuzzy: indexing.value,
        }
    }
    return {
        key: 'indexing',
        indexing,
    }
}

/**
 * Removes unused cache storage from the initial implementation of the fuzzy finder.
 *
 * This method can be removed in the future. The cache storage was no longer
 * needed after we landed an optimization in the backend that made it faster to
 * download filenames.
 */
function cleanLegacyCacheStorage(): void {
    const cacheAvailable = 'caches' in self
    if (!cacheAvailable) {
        return
    }

    caches.delete('fuzzy-modal').then(
        () => {},
        () => {}
    )
}

function fuzzyResultId(id: number): string {
    return `fuzzy-modal-result-${id}`
}
