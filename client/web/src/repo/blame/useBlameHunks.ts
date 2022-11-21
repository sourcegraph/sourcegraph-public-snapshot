import { useMemo } from 'react'

import { fetchEventSource } from '@microsoft/fetch-event-source'
import { formatDistanceStrict } from 'date-fns'
import { truncate } from 'lodash'
import { Observable, of, BehaviorSubject } from 'rxjs'
import { map } from 'rxjs/operators'

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

export type BlameHunk = NonNullable<
    NonNullable<NonNullable<GitBlameResult['repository']>['commit']>['blob']
>['blame'][number] & { displayInfo: BlameHunkDisplayInfo }

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
        name: string
        email: string
        date: string
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
    }): Observable<Omit<BlameHunk, 'displayInfo'>[] | undefined> => {
        return new Observable<Omit<BlameHunk, 'displayInfo'>[] | undefined>(subscriber => {
            // const hunks = new BehaviorSubject<Omit<BlameHunk, 'displayInfo'>[] | undefined>(undefined)
            let assembledHunks: Omit<BlameHunk, 'displayInfo'>[] = []
            const repoAndRevisionPath = `/.api/blame/${repoName}${revision ? `@${revision}` : ''}`

            fetchEventSource(`${repoAndRevisionPath}/stream/${filePath}`, {
                method: 'GET',
                headers: {
                    'X-Requested-With': 'Sourcegraph',
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
                                    date: hunk.author.date,
                                    person: {
                                        email: hunk.author.email,
                                        displayName: hunk.author.name,
                                        user: null,
                                    },
                                },
                                commit: {
                                    url: `${repoAndRevisionPath}/-/commit/${hunk.commitID}`,
                                },
                            })
                        }
                        subscriber.next(assembledHunks)
                        assembledHunks = [...assembledHunks]
                    }
                },
                onerror(event) {
                    console.error(event)
                },
            }).then(
                () => subscriber.complete(),
                error => subscriber.error(error)
            )
            // return hunks
        })
    },
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

    console.log(enableStreamingGitBlame)

    const [isBlameVisible] = useBlameVisibility()
    const shouldFetchBlame = enableCodeMirror && isBlameVisible && status !== 'initial'

    const hunks = useObservable(
        useMemo(
            () =>
                shouldFetchBlame
                    ? enableStreamingGitBlame
                        ? fetchBlameViaStreaming({ revision, repoName, filePath })
                        : fetchBlameViaGraphQL({ revision, repoName, filePath })
                    : of(undefined),
            [shouldFetchBlame, enableStreamingGitBlame, revision, repoName, filePath]
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
