/* eslint-disable jsx-a11y/no-noninteractive-element-interactions */
// NOTE: the eslint-disable above can't be a eslint-disable-next-line because
// JSX syntax doesn't support comments on the line where it's needed.

import React from 'react'

import { WordSensitiveFuzzySearch } from '../../fuzzyFinder/WordSensitiveFuzzySearch'
import { FuzzySearch, FuzzySearchResult, SearchIndexing } from '../../fuzzyFinder/FuzzySearch'

import { FuzzyFinderProps, Indexing, FuzzyFSM } from './FuzzyFinder'
import styles from './FuzzyModal.module.scss'
import { HighlightedLink } from './HighlightedLink'

// Enable this URL query parameter to see debugging information like latency
// numbers and the false-positive ratio for the bloom filter.
const IS_DEBUG = window.location.href.toString().includes('fuzzyFinder=debug')

// Cache for the last fuzzy query. This value is only used to avoid redoing the
// full fuzzy search on every re-render when the user presses the down/up arrow
// keys to move the "focus index".
const lastFuzzySearchResult = new Map<string, FuzzySearchResult>()

// The number of results to jump by on PageUp/PageDown keyboard shortcuts.
const PAGE_DOWN_INCREMENT = 10

export interface FuzzyModalProps
    extends VisibilityProps,
        FuzzyFinderProps,
        MaxResultsProps,
        FocusIndexProps,
        FuzzyFSMProps,
        QueryProps {
    downloadFilenames: () => Promise<string[]>
}
interface VisibilityProps {
    isVisible: boolean
    onClose: () => void
}
interface QueryProps {
    query: string
    setQuery: (query: string) => void
}
interface MaxResultsProps {
    maxResults: number
    increaseMaxResults: () => void
}
interface FocusIndexProps {
    focusIndex: number
    setFocusIndex: (focusIndex: number) => void
}
interface FuzzyFSMProps {
    fsm: FuzzyFSM
    setFsm: (fsm: FuzzyFSM) => void
}

/**
 * Component that interactively displays filenames in the open repository when given fuzzy queries.
 *
 * Similar to "Go to file" in VS Code or the "t" keyboard shortcut on github.com
 */
export const FuzzyModal: React.FunctionComponent<FuzzyModalProps> = props => {
    const fuzzyResult = renderFuzzyResult(props)

    // Sets the new "focus index" so that it's rounded by the number of
    // displayed filenames.  Cycles so that the user can press-hold the down
    // arrow and it goes all the way down and back up to the top result.
    function setRoundedFocusIndex(increment: number): void {
        const newNumber = props.focusIndex + increment
        const index = newNumber % fuzzyResult.resultsCount
        const nextIndex = index < 0 ? fuzzyResult.resultsCount + index : index
        props.setFocusIndex(nextIndex)
        document.querySelector(`#fuzzy-modal-result-${nextIndex}`)?.scrollIntoView(false)
    }

    function onInputKeyDown(event: React.KeyboardEvent): void {
        switch (event.key) {
            case 'Escape':
                props.onClose()
                break
            case 'ArrowDown':
                event.preventDefault() // Don't move the cursor to the end of the input.
                setRoundedFocusIndex(1)
                break
            case 'PageDown':
                setRoundedFocusIndex(PAGE_DOWN_INCREMENT)
                break
            case 'ArrowUp':
                event.preventDefault() // Don't move the cursor to the start of the input.
                setRoundedFocusIndex(-1)
                break
            case 'PageUp':
                setRoundedFocusIndex(-PAGE_DOWN_INCREMENT)
                break
            case 'Enter':
                if (props.focusIndex < fuzzyResult.resultsCount) {
                    const fileAnchor = document.querySelector<HTMLAnchorElement>(
                        `#fuzzy-modal-result-${props.focusIndex} a`
                    )
                    fileAnchor?.click()
                    props.onClose()
                }
                break
            default:
        }
    }

    return (
        // Use 'onMouseDown' instead of 'onClick' to allow selecting the text and mouse up outside the modal
        <div role="navigation" className={styles.modal} onMouseDown={() => props.onClose()}>
            <div role="navigation" className={styles.content} onMouseDown={event => event.stopPropagation()}>
                <div className={styles.header}>
                    <input
                        autoComplete="off"
                        id="fuzzy-modal-input"
                        className={styles.input}
                        value={props.query}
                        onChange={event => {
                            props.setQuery(event.target.value)
                            props.setFocusIndex(0)
                        }}
                        type="text"
                        onKeyDown={onInputKeyDown}
                    />
                </div>
                <div className={styles.body}>{fuzzyResult.element}</div>
                <div className={styles.footer}>
                    <button type="button" className="btn btn-secondary" onClick={() => props.onClose()}>
                        Close
                    </button>
                    {fuzzyFooter(props.fsm, fuzzyResult)}
                </div>
            </div>
        </div>
    )
}

function plural(what: string, count: number, isComplete: boolean): string {
    return count.toLocaleString() + (isComplete ? '' : '+') + ' ' + what + (count === 1 ? '' : 's')
}

function fuzzyFooter(fsm: FuzzyFSM, files: RenderedFuzzyResult): JSX.Element {
    return IS_DEBUG ? (
        <>
            <span>{files.falsePositiveRatio && Math.round(files.falsePositiveRatio * 100)}fp</span>
            <span>{files.elapsedMilliseconds && Math.round(files.elapsedMilliseconds).toLocaleString()}ms</span>
        </>
    ) : (
        <>
            <span>{plural('result', files.resultsCount, files.isComplete)}</span>
            <span>
                {fsm.key === 'indexing' && indexingProgressBar(fsm)}
                {plural('total file', files.totalFileCount, true)}
            </span>
        </>
    )
}

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

interface RenderedFuzzyResult {
    element: JSX.Element
    resultsCount: number
    isComplete: boolean
    totalFileCount: number
    elapsedMilliseconds?: number
    falsePositiveRatio?: number
}

function renderFuzzyResult(props: FuzzyModalProps): RenderedFuzzyResult {
    function empty(element: JSX.Element): RenderedFuzzyResult {
        return {
            element,
            resultsCount: 0,
            isComplete: true,
            totalFileCount: 0,
        }
    }

    function onError(what: string): (error: Error) => void {
        return error => {
            props.setFsm({ key: 'failed', errorMessage: JSON.stringify(error) })
            throw new Error(what)
        }
    }

    const usuallyFast =
        "This step is usually fast unless it's a very large repository. The result is cached so you only have to wait for it once :)"

    switch (props.fsm.key) {
        case 'empty':
            handleEmpty(props).then(() => {}, onError('onEmpty'))
            return empty(<></>)
        case 'downloading':
            return empty(<p>Downloading... {usuallyFast}</p>)
        case 'failed':
            return empty(<p>Error: {props.fsm.errorMessage}</p>)
        case 'indexing': {
            const loader = props.fsm.indexing
            later()
                .then(() => continueIndexing(loader))
                .then(next => props.setFsm(next), onError('onIndexing'))
            return renderFiles(props, props.fsm.indexing.partialFuzzy, props.fsm.indexing)
        }
        case 'ready':
            return renderFiles(props, props.fsm.fuzzy)
        default:
            return empty(<p>ERROR</p>)
    }
}

function renderFiles(props: FuzzyModalProps, search: FuzzySearch, indexing?: SearchIndexing): RenderedFuzzyResult {
    const indexedFileCount = indexing ? indexing.indexedFileCount : ''
    const cacheKey = `${props.query}-${props.maxResults}${indexedFileCount}`
    let fuzzyResult = lastFuzzySearchResult.get(cacheKey)
    if (!fuzzyResult) {
        const start = window.performance.now()
        fuzzyResult = search.search({
            query: props.query,
            maxResults: props.maxResults,
            createUrl: filename => `/${props.repoName}@${props.commitID}/-/blob/${filename}`,
            onClick: () => props.onClose(),
        })
        fuzzyResult.elapsedMilliseconds = window.performance.now() - start
        lastFuzzySearchResult.clear() // Only cache the last query.
        lastFuzzySearchResult.set(cacheKey, fuzzyResult)
    }
    const links = fuzzyResult.links
    if (links.length === 0) {
        return {
            element: <p>No files matching '{props.query}'</p>,
            resultsCount: 0,
            totalFileCount: search.totalFileCount,
            isComplete: fuzzyResult.isComplete,
        }
    }
    const linksToRender = links.slice(0, props.maxResults)
    return {
        element: (
            <ul className={`${styles.results} text-monospace`}>
                {linksToRender.map((file, fileIndex) => (
                    <li
                        id={`fuzzy-modal-result-${fileIndex}`}
                        key={file.text}
                        className={fileIndex === props.focusIndex ? styles.focused : ''}
                    >
                        <HighlightedLink {...file} />
                    </li>
                ))}
                {!fuzzyResult.isComplete && (
                    <li>
                        <button className="btn btn-seconday" type="button" onClick={() => props.increaseMaxResults()}>
                            (...truncated, click to show more results){' '}
                        </button>
                    </li>
                )}
            </ul>
        ),
        resultsCount: linksToRender.length,
        totalFileCount: search.totalFileCount,
        isComplete: fuzzyResult.isComplete,
        elapsedMilliseconds: fuzzyResult.elapsedMilliseconds,
        falsePositiveRatio: fuzzyResult.falsePositiveRatio,
    }
}

function filesCacheKey(props: FuzzyModalProps): string {
    return `/fuzzy-modal.files.${props.repoName}.${props.commitID}`
}

function openCaches(): Promise<Cache> {
    return caches.open('fuzzy-modal')
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

async function loadCachedIndex(props: FuzzyModalProps): Promise<FuzzyFSM | undefined> {
    const cacheAvailable = 'caches' in self
    if (!cacheAvailable) {
        return Promise.resolve(undefined)
    }
    const cacheKey = filesCacheKey(props)
    const cache = await openCaches()
    const cacheRequest = new Request(cacheKey)
    const fromCache = await cache.match(cacheRequest)
    if (!fromCache) {
        return undefined
    }
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    const filenames = JSON.parse(await fromCache.text())
    return handleFilenames(filenames)
}

async function cacheFilenames(props: FuzzyModalProps, filenames: string[]): Promise<void> {
    const cacheAvailable = 'caches' in self
    if (!cacheAvailable) {
        return Promise.resolve()
    }
    const cacheKey = filesCacheKey(props)
    const cache = await openCaches()
    await cache.put(cacheKey, new Response(JSON.stringify(filenames)))
}

async function handleEmpty(props: FuzzyModalProps): Promise<void> {
    const fromCache = await loadCachedIndex(props)
    if (fromCache) {
        props.setFsm(fromCache)
    } else {
        props.setFsm({ key: 'downloading' })
        try {
            const filenames = await props.downloadFilenames()
            props.setFsm(handleFilenames(filenames))
            cacheFilenames(props, filenames).then(
                () => {},
                () => {}
            )
        } catch (error) {
            props.setFsm({
                key: 'failed',
                errorMessage: JSON.stringify(error),
            })
        }
    }
}

function handleFilenames(filenames: string[]): FuzzyFSM {
    const indexing = WordSensitiveFuzzySearch.fromSearchValuesAsync(filenames.map(file => ({ text: file })))
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
