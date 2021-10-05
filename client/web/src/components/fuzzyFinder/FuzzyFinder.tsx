import { ApolloClient, useApolloClient } from '@apollo/client'
import { Shortcut } from '@slimsag/react-shortcuts'
import React, { useState, Dispatch, SetStateAction } from 'react'

import { gql, getDocumentNode } from '@sourcegraph/shared/src/graphql/graphql'

import { FuzzySearch, SearchIndexing } from '../../fuzzyFinder/FuzzySearch'
import { FileNamesResult, FileNamesVariables } from '../../graphql-operations'
import { KEYBOARD_SHORTCUT_CLOSE_FUZZY_FINDER } from '../../keyboardShortcuts/keyboardShortcuts'
import { parseBrowserRepoURL } from '../../util/url'

import { FuzzyModal } from './FuzzyModal'

const DEFAULT_MAX_RESULTS = 100

export interface FuzzyFinderProps {
    setIsVisible: Dispatch<SetStateAction<boolean>>

    /**
     * The maximum number of files a repo can have to use case-insensitive fuzzy finding.
     *
     * Case-insensitive fuzzy finding is more expensive to compute compared to
     * word-sensitive fuzzy finding.  The fuzzy modal will use case-insensitive
     * fuzzy finding when the repo has fewer files than this number, and
     * word-sensitive fuzzy finding otherwise.
     */
    caseInsensitiveFileCountThreshold?: number
}

export const FuzzyFinder: React.FunctionComponent<FuzzyFinderProps> = props => {
    // The state machine of the fuzzy finder. See `FuzzyFSM` for more details
    // about the state transititions.
    const apolloClient = useApolloClient()
    const [fsm, setFsm] = useState<FuzzyFSM>({ key: 'empty' })
    const { repoName = '', commitID = '' } = parseBrowserRepoURL(location.pathname + location.search + location.hash)

    return (
        <>
            <Shortcut
                {...KEYBOARD_SHORTCUT_CLOSE_FUZZY_FINDER.keybindings[0]}
                onMatch={() => props.setIsVisible(false)}
            />
            <FuzzyModal
                repoName={repoName}
                commitID={commitID}
                initialMaxResults={DEFAULT_MAX_RESULTS}
                initialQuery=""
                downloadFilenames={() => downloadFilenamesGQL(repoName, commitID, apolloClient)}
                onClose={() => props.setIsVisible(false)}
                fsm={fsm}
                setFsm={setFsm}
            />
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
    indexing: SearchIndexing
}
export interface Ready {
    key: 'ready'
    fuzzy: FuzzySearch
}
export interface Failed {
    key: 'failed'
    errorMessage: string
}

const FILE_NAMES = gql`
    query FileNames($repository: String!, $commit: String!) {
        repository(name: $repository) {
            commit(rev: $commit) {
                fileNames
            }
        }
    }
`

async function downloadFilenamesGQL(
    repository: string,
    commit: string,
    client: ApolloClient<object>
): Promise<string[]> {
    const response = await client.query<FileNamesResult, FileNamesVariables>({
        query: getDocumentNode(FILE_NAMES),
        variables: { repository, commit },
    })

    const filenames = response.data?.repository?.commit?.fileNames

    if (!filenames) {
        throw new Error(JSON.stringify(response))
    }
    return filenames
}
