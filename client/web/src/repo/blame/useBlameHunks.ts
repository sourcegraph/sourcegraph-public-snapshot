import { useMemo } from 'react'

import { formatDistanceStrict } from 'date-fns'
import { truncate } from 'lodash'
import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../backend/graphql'
import { GitBlameResult, GitBlameVariables } from '../../graphql-operations'
import { useExperimentalFeatures } from '../../stores'

import { useBlameVisibility } from './useBlameVisibility'

interface BlameHunkDisplayInfo {
    displayName: string
    username: string
    dateString: string
    timestampString: string
    linkURL: string
    message: string
}

export type BlameHunk = NonNullable<
    NonNullable<NonNullable<GitBlameResult['repository']>['commit']>['blob']
>['blame'][number] & { displayInfo: BlameHunkDisplayInfo }

const fetchBlame = memoizeObservable(
    ({
        repoName,
        revision,
        filePath,
    }: {
        repoName: string
        revision: string
        filePath: string
    }): Observable<Omit<BlameHunk, 'displayInfo'>[] | undefined> =>
        requestGraphQL<GitBlameResult, GitBlameVariables>(
            gql`
                query GitBlame($repo: String!, $rev: String!, $path: String!) {
                    repository(name: $repo) {
                        commit(rev: $rev) {
                            blob(path: $path) {
                                blame(startLine: 0, endLine: 0) {
                                    startLine
                                    endLine
                                    author {
                                        person {
                                            email
                                            displayName
                                            user {
                                                username
                                            }
                                        }
                                        date
                                    }
                                    message
                                    rev
                                    commit {
                                        url
                                    }
                                }
                            }
                        }
                    }
                }
            `,
            { repo: repoName, rev: revision, path: filePath }
        ).pipe(
            map(dataOrThrowErrors),
            map(({ repository }) => repository?.commit?.blob?.blame)
        ),
    makeRepoURI
)

/**
 * Get display info shared between status bar items and text document decorations.
 */
const getDisplayInfoFromHunk = (
    { author, commit, message }: Omit<BlameHunk, 'displayInfo'>,
    sourcegraphURL: string,
    now: number
): BlameHunkDisplayInfo => {
    const displayName = truncate(author.person.displayName, { length: 25 })
    const username = author.person.user ? `(${author.person.user.username}) ` : ''
    const dateString = formatDistanceStrict(new Date(author.date), now, { addSuffix: true })
    const timestampString = new Date(author.date).toLocaleString()
    const linkURL = new URL(commit.url, sourcegraphURL).href
    const content = `${dateString} • ${username}${displayName} [${truncate(message, { length: 45 })}]`

    return {
        displayName,
        username,
        dateString,
        timestampString,
        linkURL,
        message: content,
    }
}

export const useBlameHunks = (
    {
        repoName,
        revision,
        filePath,
    }: {
        repoName: string
        revision: string
        filePath: string
    },
    sourcegraphURL: string
): BlameHunk[] | undefined => {
    const extensionsAsCoreFeatures = useExperimentalFeatures(features => features.extensionsAsCoreFeatures)
    const [isBlameVisible] = useBlameVisibility()
    const hunks = useObservable(
        useMemo(
            () =>
                extensionsAsCoreFeatures && isBlameVisible
                    ? fetchBlame({ revision, repoName, filePath })
                    : of(undefined),
            [extensionsAsCoreFeatures, isBlameVisible, revision, repoName, filePath]
        )
    )

    const hunksWithDisplayInfo = useMemo(() => {
        const now = Date.now()
        return hunks?.map(hunk => ({
            ...hunk,
            displayInfo: getDisplayInfoFromHunk(hunk, sourcegraphURL, now),
        }))
    }, [hunks, sourcegraphURL])

    return hunksWithDisplayInfo
}
