import { useMemo } from 'react'

import { fetchEventSource } from '@microsoft/fetch-event-source'
import { formatDistanceStrict } from 'date-fns'
import { truncate } from 'lodash'
import { asyncScheduler, Observable, of } from 'rxjs'
import { map, throttleTime } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../backend/graphql'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import { GitBlameResult, GitBlameVariables } from '../../graphql-operations'

import { useBlameVisibility } from './useBlameVisibility'

interface BlameHunkDisplayInfo {
    displayName: string
    username: string
    dateString: string
    timestampString: string
    linkURL: string
    message: string
}

export interface BlameHunk {
    startLine: number
    endLine: number
    message: string
    rev: string
    author: {
        date: string
        person: {
            email: string
            displayName: string
            user:
                | undefined
                | null
                | {
                      username: string
                  }
        }
    }
    commit: {
        url: string
        parents: {
            oid: string
        }[]
    }
    displayInfo: BlameHunkDisplayInfo
}

const fetchBlameViaGraphQL = memoizeObservable(
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
                                        parents {
                                            oid
                                        }
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

interface RawStreamHunk {
    author: {
        Name: string
        Email: string
        Date: string
    }
    commit: {
        parents: string[]
        url: string
    }
    commitID: string
    endLine: number
    startLine: number
    filename: string
    message: string
}

const fetchBlameViaStreaming = memoizeObservable(
    ({
        repoName,
        revision,
        filePath,
    }: {
        repoName: string
        revision: string
        filePath: string
    }): Observable<Omit<BlameHunk, 'displayInfo'>[] | undefined> =>
        new Observable<Omit<BlameHunk, 'displayInfo'>[] | undefined>(subscriber => {
            let assembledHunks: Omit<BlameHunk, 'displayInfo'>[] = []
            const repoAndRevisionPath = `/${repoName}${revision ? `@${revision}` : ''}`
            fetchEventSource(`/.api/blame${repoAndRevisionPath}/stream/${filePath}`, {
                method: 'GET',
                headers: {
                    'X-Requested-With': 'Sourcegraph',
                    'X-Sourcegraph-Should-Trace': new URLSearchParams(window.location.search).get('trace') || 'false',
                },
                onmessage(event) {
                    if (event.event === 'hunk') {
                        const rawHunks: RawStreamHunk[] = JSON.parse(event.data)
                        for (const hunk of rawHunks) {
                            assembledHunks.push({
                                startLine: hunk.startLine,
                                endLine: hunk.endLine,
                                message: hunk.message,
                                rev: hunk.commitID,
                                author: {
                                    date: hunk.author.Date,
                                    person: {
                                        email: hunk.author.Email,
                                        displayName: hunk.author.Name,
                                        user: null,
                                    },
                                },
                                commit: {
                                    url: hunk.commit.url,
                                    parents: hunk.commit.parents ? hunk.commit.parents.map(oid => ({ oid })) : [],
                                },
                            })
                        }
                        subscriber.next(assembledHunks)
                        assembledHunks = [...assembledHunks]
                    }
                },
                onerror(event) {
                    // eslint-disable-next-line no-console
                    console.error(event)
                },
            }).then(
                () => subscriber.complete(),
                error => subscriber.error(error)
            )
            // Throttle the results to avoid re-rendering the blame sidebar for every hunk
        }).pipe(throttleTime(1000, asyncScheduler, { leading: true, trailing: true })),
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
        enableCodeMirror,
    }: {
        repoName: string
        revision: string
        filePath: string
        enableCodeMirror: boolean
    },
    sourcegraphURL: string
): BlameHunk[] | undefined => {
    const [enableStreamingGitBlame, status] = useFeatureFlag('enable-streaming-git-blame')

    const [isBlameVisible] = useBlameVisibility()
    const shouldFetchBlame = isBlameVisible && status !== 'initial'

    const hunks = useObservable(
        useMemo(
            () =>
                shouldFetchBlame
                    ? enableCodeMirror && enableStreamingGitBlame
                        ? fetchBlameViaStreaming({ revision, repoName, filePath })
                        : fetchBlameViaGraphQL({ revision, repoName, filePath })
                    : of(undefined),
            [shouldFetchBlame, enableCodeMirror, enableStreamingGitBlame, revision, repoName, filePath]
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
