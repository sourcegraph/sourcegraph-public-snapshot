/* eslint-disable jsx-a11y/no-noninteractive-element-interactions */
import React from 'react'

import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../backend/graphql'

import { BloomFilterFuzzySearch, Indexing as FuzzyIndexing } from './BloomFilterFuzzySearch'
import styles from './FuzzyModal.module.scss'
import { FuzzySearch, FuzzySearchResult } from './FuzzySearch'
import { HighlightedText } from './HighlightedText'
import { useEphemeralState, useLocalStorage, State } from './useLocalStorage'

const DEFAULT_MAX_RESULTS = 100

const IS_DEBUG = window.location.href.toString().includes('debug=true')

export interface FuzzyModalProps {
    isVisible: boolean
    onClose(): void
    repoName: string
    commitID: string
}

/**
 * React component that interactively displays filenames in the open repository when given fuzzy queries.
 *
 * Similar to "Go to file" in VS Code or the "t" keyboard shortcut on github.com
 */
export const FuzzyModal: React.FunctionComponent<FuzzyModalProps> = props => {
    // NOTE: the query is cached in local storage to mimic IntelliJ.  It' quite
    // annoying in VS Code when it doesn't cache the "Go to symbol in workspace"
    // query. For example, I can't count the times I have typed a query like
    // "FilePro", browsed the results and want to update the query to become
    // "FileProvider" but VS Code has forgotten the original query so I have to
    // type it out from scratch again.
    const query = useLocalStorage(`fuzzy-modal.query.${props.repoName}`, '')

    // The "focus index" is the index of the file result that the user has
    // select with up/down arrow keys. The focused item is highlighted and the
    // window.location is moved to that URL when the user presses the enter key.
    const focusIndex = useEphemeralState(0)

    const maxResults = useEphemeralState(DEFAULT_MAX_RESULTS)

    const loaded = useEphemeralState<Loaded>({ key: 'empty' })

    if (!props.isVisible) {
        return null
    }

    const files = renderFiles(props, query, focusIndex, maxResults, loaded)

    // Sets the new "focus index" so that it's rounded by the number of
    // displayed filenames.  Cycles so that the user can press-hold the down
    // arrow and it goes all the way down and back up to the top result.
    function setRoundedFocusIndex(newNumber: number): void {
        const index = newNumber % files.resultsCount
        const nextIndex = index < 0 ? files.resultsCount + index : index
        focusIndex.set(nextIndex)
        document.querySelector(`#fuzzy-modal-result-${nextIndex}`)?.scrollIntoView(false)
    }

    function onInputKeyDown(event: React.KeyboardEvent): void {
        switch (event.key) {
            case 'Escape':
                props.onClose()
                break
            case 'ArrowDown':
                event.preventDefault() // Don't move the cursor to the end of the input.
                setRoundedFocusIndex(focusIndex.value + 1)
                break
            case 'PageDown':
                setRoundedFocusIndex(focusIndex.value + 10)
                break
            case 'ArrowUp':
                event.preventDefault() // Don't move the cursor to the start of the input.
                setRoundedFocusIndex(focusIndex.value - 1)
                break
            case 'PageUp':
                setRoundedFocusIndex(focusIndex.value - 10)
                break
            case 'Enter':
                if (focusIndex.value < files.resultsCount) {
                    const fileAnchor = document.querySelector<HTMLAnchorElement>(
                        `#fuzzy-modal-result-${focusIndex.value} a`
                    )
                    fileAnchor?.click()
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
                        value={query.value}
                        onChange={event => {
                            query.set(event.target.value)
                            focusIndex.set(0)
                        }}
                        type="text"
                        onKeyDown={onInputKeyDown}
                    />
                </div>
                <div className={styles.body}>{files.element}</div>
                <div className={styles.footer}>
                    <button type="button" className="btn btn-secondary" onClick={() => props.onClose()}>
                        Close
                    </button>
                    {fuzzyFooter(loaded.value, files)}
                </div>
            </div>
        </div>
    )
}

function plural(what: string, count: number, isComplete: boolean): string {
    return count.toLocaleString() + (isComplete ? '' : '+') + ' ' + what + (count === 1 ? '' : 's')
}

function fuzzyFooter(loaded: Loaded, files: RenderedFiles): JSX.Element {
    return IS_DEBUG ? (
        <>
            <span>{files.falsePositiveRatio && Math.round(files.falsePositiveRatio * 100)}fp</span>
            <span>{files.elapsedMilliseconds && Math.round(files.elapsedMilliseconds).toLocaleString()}ms</span>
        </>
    ) : (
        <>
            <span>{plural('result', files.resultsCount, files.isComplete)}</span>
            <span>
                {loaded.key === 'indexing' && indexingProgressBar(loaded)}
                {plural('total file', files.totalFileCount, true)}
            </span>
        </>
    )
}

function indexingProgressBar(indexing: Indexing): JSX.Element {
    const indexedFiles = indexing.loader.indexedFileCount
    const totalFiles = indexing.loader.totalFileCount
    const percentage = Math.round((indexedFiles / totalFiles) * 100)
    return (
        <progress value={indexedFiles} max={totalFiles}>
            {percentage}%
        </progress>
    )
}

/**
 * The fuzzy finder modal is implemented as a state machine with the following transitions:
 *
 * ```
 *   ╭────[cached]───────────────────────╮  ╭──╮
 *   │                                   v  │  v
 * Empty ─[uncached]───> Downloading ──> Indexing ──> Ready
 *                       ╰──────────────────────> Failed
 * ```
 *
 * - Empty: start state.
 * - Downloading: downloading filenames from the remote server. The filenames
 *                are cached using the browser's CacheStorage, if available.
 * - Indexing: processing the downloaded filenames. This step is usually
 *             instant, unless the repo is very large (>100k source files).
 *             In the torvalds/linux repo (~70k files), this step takes <1s
 *             on my computer but the chromium/chromium repo (~360k files)
 *             it takes ~3-5 seconds. This step is async so that the user can
 *             query against partially indexed results.
 * - Ready: all filenames have been indexed.
 * - Failed: something unexpected happened, the user can't fuzzy find files.
 */
type Loaded = Empty | Downloading | Indexing | Ready | Failed
interface Empty {
    key: 'empty'
}
interface Downloading {
    key: 'downloading'
}
interface Indexing {
    key: 'indexing'
    loader: FuzzyIndexing
    totalFileCount: number
}
interface Ready {
    key: 'ready'
    fuzzy: FuzzySearch
    totalFileCount: number
}
interface Failed {
    key: 'failed'
    errorMessage: string
}

interface RenderedFiles {
    element: JSX.Element
    resultsCount: number
    isComplete: boolean
    totalFileCount: number
    elapsedMilliseconds?: number
    falsePositiveRatio?: number
}

// Cache for the last fuzzy query. This value is only used to avoid redoing the
// full fuzzy search on every re-render when the user presses the down/up arrow
// keys to move the "focus index".
const lastFuzzySearchResult = new Map<string, FuzzySearchResult>()

function renderFiles(
    props: FuzzyModalProps,
    query: State<string>,
    focusIndex: State<number>,
    maxResults: State<number>,
    loaded: State<Loaded>
): RenderedFiles {
    function empty(element: JSX.Element): RenderedFiles {
        return {
            element,
            resultsCount: 0,
            isComplete: true,
            totalFileCount: 0,
        }
    }

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    function onError(what: string): (error: any) => void {
        return error => {
            // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
            error.what = what
            loaded.set({ key: 'failed', errorMessage: JSON.stringify(error) })
            throw new Error(what)
        }
    }

    const usuallyFast =
        "This step is usually fast unless it's a very large repository. The result is cached so you only have to wait for it once :)"

    switch (loaded.value.key) {
        case 'empty':
            handleEmpty(props, loaded).then(() => {}, onError('onEmpty'))
            return empty(<></>)
        case 'downloading':
            return empty(<p>Downloading... {usuallyFast}</p>)
        case 'failed':
            return empty(<p>Error: {loaded.value.errorMessage}</p>)
        case 'indexing': {
            const loader = loaded.value.loader
            later()
                .then(() => continueIndexing(props, loader))
                .then(next => loaded.set(next), onError('onIndexing'))
            return renderReady(props, query, focusIndex, maxResults, loaded.value.loader.partialValue, loaded.value)
        }
        case 'ready':
            return renderReady(props, query, focusIndex, maxResults, loaded.value.fuzzy, loaded.value)
        default:
            return empty(<p>ERROR</p>)
    }
}

function renderReady(
    props: FuzzyModalProps,
    query: State<string>,
    focusIndex: State<number>,
    maxResults: State<number>,
    fuzzy: FuzzySearch,
    ready: Ready | Indexing
): RenderedFiles {
    const indexedFileCount = ready.key === 'indexing' ? ready.loader.indexedFileCount : ''
    const cacheKey = `${query.value}-${maxResults.value}${indexedFileCount}`
    let fuzzyResult = lastFuzzySearchResult.get(cacheKey)
    if (!fuzzyResult) {
        const start = window.performance.now()
        fuzzyResult = fuzzy.search({
            value: query.value,
            maxResults: maxResults.value,
            createUrl: filename => `/${props.repoName}@${props.commitID}/-/blob/${filename}`,
        })
        fuzzyResult.elapsedMilliseconds = window.performance.now() - start
        lastFuzzySearchResult.clear() // Only cache the last query.
        lastFuzzySearchResult.set(cacheKey, fuzzyResult)
    }
    const matchingFiles = fuzzyResult.values
    if (matchingFiles.length === 0) {
        return {
            element: <p>No files matching '{query.value}'</p>,
            resultsCount: 0,
            totalFileCount: fuzzy.totalFileCount,
            isComplete: fuzzyResult.isComplete,
        }
    }
    const filesToRender = matchingFiles.slice(0, maxResults.value)
    return {
        element: (
            <ul className={`${styles.results} text-monospace`}>
                {filesToRender.map((file, fileIndex) => (
                    <li
                        id={`fuzzy-modal-result-${fileIndex}`}
                        key={file.text}
                        className={fileIndex === focusIndex.value ? styles.focused : ''}
                    >
                        <HighlightedText {...file} />
                    </li>
                ))}
                {!fuzzyResult.isComplete && (
                    <li>
                        <button
                            className="btn btn-seconday"
                            type="button"
                            onClick={() => {
                                maxResults.set(maxResults.value + DEFAULT_MAX_RESULTS)
                            }}
                        >
                            (...truncated, click to show more results){' '}
                        </button>
                    </li>
                )}
            </ul>
        ),
        resultsCount: filesToRender.length,
        totalFileCount: fuzzy.totalFileCount,
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

async function continueIndexing(props: FuzzyModalProps, indexing: FuzzyIndexing): Promise<Loaded> {
    const next = await indexing.continue()
    if (next.key === 'indexing') {
        return { key: 'indexing', loader: next, totalFileCount: next.totalFileCount }
    }
    return {
        key: 'ready',
        fuzzy: next.value,
        totalFileCount: next.value.totalFileCount,
    }
}

async function loadCachedIndex(props: FuzzyModalProps): Promise<Loaded | undefined> {
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
    return handleFilenames(props, filenames)
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

async function handleEmpty(props: FuzzyModalProps, files: State<Loaded>): Promise<void> {
    const fromCache = await loadCachedIndex(props)
    if (fromCache) {
        files.set(fromCache)
    } else {
        files.set({ key: 'downloading' })
        try {
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            const next: any = await requestGraphQL(
                gql`
                    query Files($repository: String!, $commit: String!) {
                        repository(name: $repository) {
                            commit(rev: $commit) {
                                tree(recursive: true) {
                                    files(first: 1000000, recursive: true) {
                                        path
                                    }
                                }
                            }
                        }
                    }
                `,
                {
                    repository: props.repoName,
                    commit: props.commitID,
                }
            ).toPromise()
            // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-call, @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-return
            const filenames = next.data?.repository?.commit?.tree?.files?.map((file: any) => file.path) as
                | string[]
                | undefined
            if (filenames) {
                files.set(handleFilenames(props, filenames))
                cacheFilenames(props, filenames).then(
                    () => {},
                    () => {}
                )
            } else {
                files.set({
                    key: 'failed',
                    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
                    errorMessage: JSON.stringify(next.data),
                })
            }
        } catch (error) {
            files.set({
                key: 'failed',
                errorMessage: JSON.stringify(error),
            })
        }
    }
}
function handleFilenames(props: FuzzyModalProps, filenames: string[]): Loaded {
    const loader = BloomFilterFuzzySearch.fromSearchValuesAsync(filenames.map(file => ({ text: file })))
    if (loader.key === 'ready') {
        return {
            key: 'ready',
            fuzzy: loader.value,
            totalFileCount: loader.value.totalFileCount,
        }
    }
    return {
        key: 'indexing',
        loader,
        totalFileCount: loader.totalFileCount,
    }
}
