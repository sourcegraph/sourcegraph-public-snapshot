/* eslint-disable @typescript-eslint/no-unsafe-return */
/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-unsafe-call */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable jsx-a11y/no-noninteractive-element-interactions */
/* eslint-disable jsx-a11y/click-events-have-key-events */
import React from 'react'

import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../backend/graphql'

import { BloomFilterFuzzySearch } from './BloomFilterFuzzySearch'
import { FuzzySearch, FuzzySearchResult } from './FuzzySearch'
import { HighlightedText, HighlightedTextProps } from './HighlightedText'
import { useEphemeralState, useLocalStorage, State } from './useLocalStorage'

const DEFAULT_MAX_RESULTS = 100

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
    // NOTE(olafur): the query is cached in local storage to mimic IntelliJ.
    // It' quite annoying in VS Code when it doesn't cache the "Go to symbol in
    // workspace" query. For example, I can't count the times I have typed a
    // query like "FilePro", browsed the results and want to update the query to
    // become "FileProvider" but VS Code has forgotten the original query so I
    // have to type it out from scratch again.
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
        const max = files.results.length
        const index = newNumber % max
        const nextIndex = index < 0 ? max + index : index
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
                if (focusIndex.value < files.results.length) {
                    const url = files.results[focusIndex.value].url
                    if (url) {
                        window.location.href = url
                    }
                }
                break
            default:
        }
    }

    return (
        <div role="navigation" className="fuzzy-modal" onClick={() => props.onClose()}>
            <div role="navigation" className="fuzzy-modal-content" onClick={event => event.stopPropagation()}>
                <div className="fuzzy-modal-header">
                    <div className="fuzzy-modal-cursor">
                        <input
                            autoComplete="off"
                            id="fuzzy-modal-input"
                            className="fuzzy-modal-input"
                            value={query.value}
                            onChange={event => {
                                query.set(event.target.value)
                                focusIndex.set(0)
                            }}
                            type="text"
                            onKeyDown={onInputKeyDown}
                        />
                        <i />
                    </div>
                </div>
                <div className="fuzzy-modal-body">{files.element}</div>
                <div className="fuzzy-modal-footer">
                    <button type="button" className="btn btn-secondary" onClick={() => props.onClose()}>
                        Close
                    </button>
                </div>
            </div>
        </div>
    )
}

/**
 * The fuzzy finder modal is implemented as a state machine with the following transitions:
 *
 * ```
 * Empty ──> Loading ──> Indexing ──> Ready
 *            ╰─────────────────────> Failed
 * ```
 *
 * - Empty: start state.
 * - Loading: downloading filenames from the remote server.
 * - Indexing: processing the downloaded filenames. This step is usually
 * instant, unless the repo is HUGE.
 * - Ready: user can fuzzy find filenames.
 * - Failed: error, the user can't fuzzy find files.
 */
type Loaded = Empty | Loading | Indexing | Ready | Failed
interface Empty {
    key: 'empty'
}
interface Loading {
    key: 'loading'
}
interface Indexing {
    key: 'indexing'
    filenames: string[]
}
interface Ready {
    key: 'ready'
    fuzzy: FuzzySearch
}
interface Failed {
    key: 'failed'
    errorMessage: string
}

interface RenderedFiles {
    element: JSX.Element
    results: HighlightedTextProps[]
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
            results: [],
        }
    }

    function onError(error: any): void {
        loaded.set({ key: 'failed', errorMessage: JSON.stringify(error) })
    }

    const usuallyFast =
        "This step is usually fast unless it's a very large repository. The result is cached so you only have to wait for it once :)"

    switch (loaded.value.key) {
        case 'empty':
            handleEmpty(props, loaded).then(() => {}, onError)
            return empty(<></>)
        case 'loading':
            return empty(<p>Downloading... {usuallyFast}</p>)
        case 'failed':
            return empty(<p>Error: {loaded.value.errorMessage}</p>)
        case 'indexing':
            handleIndexing(props, loaded.value.filenames).then(next => loaded.set(next), onError)
            return empty(<p>Indexing... {usuallyFast}</p>)
        case 'ready':
            return renderReady(query, focusIndex, maxResults, loaded.value.fuzzy)
        default:
            return empty(<p>ERROR</p>)
    }
}
function renderReady(
    query: State<string>,
    focusIndex: State<number>,
    maxResults: State<number>,
    fuzzy: FuzzySearch
): RenderedFiles {
    const cacheKey = `${query.value}-${maxResults.value}`
    let fuzzyResult = lastFuzzySearchResult.get(cacheKey)
    if (!fuzzyResult) {
        fuzzyResult = fuzzy.search({
            value: query.value,
            maxResults: maxResults.value,
        })
        lastFuzzySearchResult.clear() // Only cache the last query.
        lastFuzzySearchResult.set(cacheKey, fuzzyResult)
    }
    const matchingFiles = fuzzyResult.values

    if (matchingFiles.length === 0) {
        return { element: <p>No files matching '{query.value}'</p>, results: [] }
    }
    const filesToRender = matchingFiles.slice(0, maxResults.value)
    return {
        element: (
            <ul className="fuzzy-modal-results">
                {filesToRender.map((file, fileIndex) => (
                    <li
                        id={`fuzzy-modal-result-${fileIndex}`}
                        key={file.text}
                        className={fileIndex === focusIndex.value ? 'fuzzy-modal-focused' : ''}
                    >
                        <HighlightedText value={file} />
                    </li>
                ))}
                {!fuzzyResult.isComplete && (
                    <li>
                        <button
                            type="button"
                            onClick={() => {
                                console.log('EXPAND')
                                maxResults.set(maxResults.value * 2)
                            }}
                        >
                            (...truncated, click to show more results){' '}
                        </button>
                    </li>
                )}
            </ul>
        ),
        results: filesToRender,
    }
}

function filesCacheKey(props: FuzzyModalProps): string {
    return `/fuzzy-modal.files.${props.repoName}.${props.commitID}`
}

function openCaches(): Promise<Cache> {
    return caches.open('fuzzy-modal')
}

async function handleIndexing(props: FuzzyModalProps, files: string[]): Promise<Ready> {
    const result = await new Promise<Ready>(resolve =>
        setTimeout(
            () =>
                resolve({
                    key: 'ready',
                    fuzzy: BloomFilterFuzzySearch.fromSearchValues(
                        files.map(file => ({ value: file, url: `/${props.repoName}@${props.commitID}/-/blob/${file}` }))
                    ),
                }),
            0
        )
    )
    const cache = await openCaches()
    const text = serializeIndex(result)
    if (text) {
        await cache.put(new Request(filesCacheKey(props)), text)
    }
    return result
}

async function deserializeIndex(ready: Response): Promise<Ready> {
    return {
        key: 'ready',
        fuzzy: BloomFilterFuzzySearch.fromSerializedString(await ready.text()),
    }
}

function serializeIndex(ready: Ready): Response | undefined {
    const serializable = ready.fuzzy.serialize()
    return serializable ? new Response(JSON.stringify(serializable)) : undefined
}

async function handleEmpty(props: FuzzyModalProps, files: State<Loaded>): Promise<void> {
    const cache = await openCaches()
    const cacheKey = filesCacheKey(props)
    const cacheRequest = new Request(cacheKey)
    const fromCache = await cache.match(cacheRequest)
    if (fromCache) {
        files.set(await deserializeIndex(fromCache))
    } else {
        files.set({ key: 'loading' })
        try {
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
            const response = next.data?.repository?.commit?.tree?.files?.map((file: any) => file.path)
            if (response) {
                files.set({
                    key: 'indexing',
                    filenames: response,
                })
                await cache.put(cacheRequest, new Response(JSON.stringify(response)))
            } else {
                files.set({
                    key: 'failed',
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
