import { useMemo } from 'react'

import { fetchEventSource } from '@microsoft/fetch-event-source'
import { formatDistanceStrict } from 'date-fns'
import { truncate } from 'lodash'
import { Observable, of } from 'rxjs'
import { map, throttleTime } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../../backend/graphql'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import {
    FirstCommitDateResult,
    FirstCommitDateVariables,
    GitBlameResult,
    GitBlameVariables,
} from '../../graphql-operations'

import { useBlameVisibility } from './useBlameVisibility'

interface BlameHunkDisplayInfo {
    displayName: string
    username: string
    dateString: string
    timestampString: string
    linkURL: string
    message: string
    commitDate: Date
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
        sourcegraphURL,
    }: {
        repoName: string
        revision: string
        filePath: string
        sourcegraphURL: string
    }): Observable<{ current: BlameHunk[] | undefined; firstCommitDate: Date | undefined }> =>
        requestGraphQL<GitBlameResult, GitBlameVariables>(
            gql`
                query GitBlame($repo: String!, $rev: String!, $path: String!) {
                    repository(name: $repo) {
                        firstEverCommit {
                            author {
                                date
                            }
                        }
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
            map(({ repository }) => {
                const hunks = repository?.commit?.blob?.blame
                const firstCommitDate = repository?.firstEverCommit?.author?.date
                if (hunks) {
                    return {
                        current: hunks.map(blame => addDisplayInfoForHunk(blame, sourcegraphURL)),
                        firstCommitDate: firstCommitDate ? new Date(firstCommitDate) : undefined,
                    }
                }

                return { current: undefined, firstCommitDate: undefined }
            })
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
        sourcegraphURL,
    }: {
        repoName: string
        revision: string
        filePath: string
        sourcegraphURL: string
    }): Observable<{ current: BlameHunk[] | undefined; firstCommitDate: Date | undefined }> =>
        new Observable<{ current: BlameHunk[] | undefined; firstCommitDate: Date | undefined }>(subscriber => {
            ;(async function () {
                // TODO: Fetch the first commit in parallel to the blame hunks.
                const firstCommitDate = await fetchFirstCommitDate(repoName)

                const assembledHunks: BlameHunk[] = []
                const repoAndRevisionPath = `/${repoName}${revision ? `@${revision}` : ''}`
                await fetchEventSource(`/.api/blame${repoAndRevisionPath}/stream/${filePath}`, {
                    method: 'GET',
                    headers: {
                        'X-Requested-With': 'Sourcegraph',
                        'X-Sourcegraph-Should-Trace':
                            new URLSearchParams(window.location.search).get('trace') || 'false',
                    },
                    onmessage(event) {
                        if (event.event === 'hunk') {
                            const rawHunks: RawStreamHunk[] = JSON.parse(event.data)
                            for (const rawHunk of rawHunks) {
                                const hunk: Omit<BlameHunk, 'displayInfo'> = {
                                    startLine: rawHunk.startLine,
                                    endLine: rawHunk.endLine,
                                    message: rawHunk.message,
                                    rev: rawHunk.commitID,
                                    author: {
                                        date: rawHunk.author.Date,
                                        person: {
                                            email: rawHunk.author.Email,
                                            displayName: rawHunk.author.Name,
                                            user: null,
                                        },
                                    },
                                    commit: {
                                        url: rawHunk.commit.url,
                                        parents: rawHunk.commit.parents
                                            ? rawHunk.commit.parents.map(oid => ({ oid }))
                                            : [],
                                    },
                                }
                                assembledHunks.push(addDisplayInfoForHunk(hunk, sourcegraphURL))
                            }
                            subscriber.next({ current: assembledHunks, firstCommitDate })
                        }
                    },
                    onerror(event) {
                        // eslint-disable-next-line no-console
                        console.error(event)
                    },
                })
                subscriber.complete()
            })().catch(error => subscriber.error(error))
        })
            // Throttle the results to avoid re-rendering the blame sidebar for every hunk
            .pipe(throttleTime(1000, undefined, { leading: true, trailing: true })),
    makeRepoURI
)

async function fetchFirstCommitDate(repoName: string): Promise<Date | undefined> {
    return requestGraphQL<FirstCommitDateResult, FirstCommitDateVariables>(
        gql`
            query FirstCommitDate($repo: String!) {
                repository(name: $repo) {
                    firstEverCommit {
                        author {
                            date
                        }
                    }
                }
            }
        `,
        { repo: repoName }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(({ repository }) => repository?.firstEverCommit?.author?.date),
            map(date => (date ? new Date(date) : undefined))
        )
        .toPromise()
}

/**
 * Get display info shared between status bar items and text document decorations.
 */
const addDisplayInfoForHunk = (hunk: Omit<BlameHunk, 'displayInfo'>, sourcegraphURL: string): BlameHunk => {
    const now = Date.now()
    const { author, commit, message } = hunk

    const displayName = truncate(author.person.displayName, { length: 25 })
    const username = author.person.user ? `(${author.person.user.username}) ` : ''
    const commitDate = new Date(author.date)
    const dateString = formatDateForBlame(commitDate, now)
    const timestampString = commitDate.toLocaleString()
    const linkURL = new URL(commit.url, sourcegraphURL).href
    const content = truncate(message, { length: 45 })

    ;(hunk as BlameHunk).displayInfo = {
        displayName,
        username,
        commitDate,
        dateString,
        timestampString,
        linkURL,
        message: content,
    }
    return hunk as BlameHunk
}

/**
 * For performance reasons, the hunks array can be mutated in place. To still be
 * able to propagate updates accordingly, this is wrapped in a ref object that
 * can be recreated whenever we emit new values.
 */
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
): { current: BlameHunk[] | undefined; firstCommitDate: Date | undefined } => {
    const [enableStreamingGitBlame, status] = useFeatureFlag('enable-streaming-git-blame')

    const [isBlameVisible] = useBlameVisibility()
    const shouldFetchBlame = isBlameVisible && status !== 'initial'

    const hunks = useObservable(
        useMemo(
            () =>
                shouldFetchBlame
                    ? enableCodeMirror && enableStreamingGitBlame
                        ? fetchBlameViaStreaming({ revision, repoName, filePath, sourcegraphURL })
                        : fetchBlameViaGraphQL({ revision, repoName, filePath, sourcegraphURL })
                    : of({ current: undefined, firstCommitDate: undefined }),
            [shouldFetchBlame, enableCodeMirror, enableStreamingGitBlame, revision, repoName, filePath, sourcegraphURL]
        )
    )

    return hunks || { current: undefined, firstCommitDate: undefined }
}

const ONE_MONTH = 30 * 24 * 60 * 60 * 1000
function formatDateForBlame(commitDate: Date, now: number): string {
    if (now - commitDate.getTime() < ONE_MONTH) {
        return formatDistanceStrict(commitDate, now, { addSuffix: true })
    }
    if (commitDate.getFullYear() === new Date(now).getFullYear()) {
        return commitDate.toLocaleString('default', { month: 'short', day: 'numeric' })
    }
    return commitDate.toLocaleString('default', { year: 'numeric', month: 'short', day: 'numeric' })
}
