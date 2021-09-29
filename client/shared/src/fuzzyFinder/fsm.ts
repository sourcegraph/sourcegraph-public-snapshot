import { FileNamesResult, FileNamesVariables } from '../graphql-operations'
import { gql } from '../graphql/graphql'
import { PlatformContextProps } from '../platform/context'

import { FuzzySearch, SearchIndexing } from './FuzzySearch'

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

export interface FuzzyFinderProps extends PlatformContextProps<'requestGraphQL' | 'urlToFile'> {
    repoName: string
    commitID: string

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

export async function downloadFilenames(props: FuzzyFinderProps): Promise<string[]> {
    const gqlResult = await props.platformContext
        .requestGraphQL<FileNamesResult, FileNamesVariables>({
            request: gql`
                query FileNames($repository: String!, $commit: String!) {
                    repository(name: $repository) {
                        commit(rev: $commit) {
                            fileNames
                        }
                    }
                }
            `,
            variables: {
                repository: props.repoName,
                commit: props.commitID,
            },
            mightContainPrivateInfo: true,
        })
        .toPromise()
    const filenames = gqlResult.data?.repository?.commit?.fileNames
    if (!filenames) {
        throw new Error(JSON.stringify(gqlResult))
    }
    return filenames
}
