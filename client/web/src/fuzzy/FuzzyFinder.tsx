import { Shortcut } from '@slimsag/react-shortcuts'
import React, { useState } from 'react'

import { gql } from '@sourcegraph/shared/src/graphql/graphql'
import { useLocalStorage } from '@sourcegraph/shared/src/util/useLocalStorage'

import { requestGraphQL } from '../backend/graphql'
import {
    KEYBOARD_SHORTCUT_CLOSE_FUZZY_FILES,
    KEYBOARD_SHORTCUT_FUZZY_FILES,
} from '../keyboardShortcuts/keyboardShortcuts'

import { FuzzyModal } from './FuzzyModal'
import { FuzzySearch, SearchIndexing } from './FuzzySearch'

const DEFAULT_MAX_RESULTS = 100

export interface FuzzyFinderProps {
    repoName: string
    commitID: string
}

export const FuzzyFinder: React.FunctionComponent<FuzzyFinderProps> = props => {
    const [isVisible, setIsVisible] = useState(false)
    // NOTE: the query is cached in local storage to mimic the file pickers in
    // IntelliJ (by default) and VS Code (when "Workbench > Quick Open >
    // Preserve Input" is enabled).
    const [query, setQuery] = useLocalStorage(`fuzzy-modal.query.${props.repoName}`, '')

    // The "focus index" is the index of the file result that the user has
    // select with up/down arrow keys. The focused item is highlighted and the
    // window.location is moved to that URL when the user presses the enter key.
    const [focusIndex, setFocusIndex] = useState(0)

    // The maximum number of results to display in the fuzzy finder. For large
    // repositories, a generic query like "src" may return thousands of results
    // making DOM rendering slow.  The user can increase this number by clicking
    // on a button at the bottom of the result list.
    const [maxResults, setMaxResults] = useState(DEFAULT_MAX_RESULTS)

    // The state machine of the fuzzy finder. See `FuzzyFSM` for more details
    // about the state transititions.
    const [fsm, setFsm] = useState<FuzzyFSM>({ key: 'empty' })

    return (
        <>
            <Shortcut
                {...KEYBOARD_SHORTCUT_FUZZY_FILES.keybindings[0]}
                onMatch={() => {
                    setIsVisible(true)
                    const input = document.querySelector<HTMLInputElement>('#fuzzy-modal-input')
                    input?.focus()
                    input?.select()
                }}
            />
            <Shortcut {...KEYBOARD_SHORTCUT_CLOSE_FUZZY_FILES.keybindings[0]} onMatch={() => setIsVisible(false)} />
            {isVisible && (
                <FuzzyModal
                    {...props}
                    isVisible={isVisible}
                    onClose={() => setIsVisible(false)}
                    query={query}
                    setQuery={setQuery}
                    focusIndex={focusIndex}
                    setFocusIndex={setFocusIndex}
                    maxResults={maxResults}
                    increaseMaxResults={() => setMaxResults(maxResults + DEFAULT_MAX_RESULTS)}
                    fsm={fsm}
                    setFsm={setFsm}
                    downloadFilenames={() => downloadFilenames(props)}
                />
            )}
        </>
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
export type FuzzyFSM = Empty | Downloading | Indexing | Ready | Failed
export interface Empty {
    key: 'empty'
}
export interface Downloading {
    key: 'downloading'
}
export interface Indexing {
    key: 'indexing'
    loader: SearchIndexing
    totalFileCount: number
}
export interface Ready {
    key: 'ready'
    fuzzy: FuzzySearch
    totalFileCount: number
}
export interface Failed {
    key: 'failed'
    errorMessage: string
}

async function downloadFilenames(props: FuzzyFinderProps): Promise<string[]> {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const gqlResult: any = await requestGraphQL(
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
    const filenames = gqlResult.data?.repository?.commit?.tree?.files?.map((file: any) => file.path) as
        | string[]
        | undefined
    if (!filenames) {
        throw new Error(JSON.stringify(gqlResult))
    }
    return filenames
}
