import { EventStreamContentType, fetchEventSource } from '@microsoft/fetch-event-source'
import { formatDistanceStrict } from 'date-fns'
import { truncate } from 'lodash'
import { Observable } from 'rxjs'
import { throttleTime, map, scan } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'
import { makeRepoGitURI } from '@sourcegraph/shared/src/util/url'

import type { ExternalServiceKind } from '../../graphql-operations'

export interface BlameHunkDisplayInfo {
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
    filename: string
    author: {
        date: string
        person: {
            email: string
            displayName: string
            avatarURL: string | null
            user:
                | undefined
                | null
                | {
                      username: string
                      displayName: string | null
                      avatarURL: string | null
                  }
        }
    }
    commit: {
        url: string
        previous: {
            rev: string
            filename: string
        } | null
    }
    displayInfo: BlameHunkDisplayInfo
}

interface RawStreamHunk {
    author: {
        Name: string
        Email: string
        Date: string
    }
    commit: {
        previous: {
            commitID: string
            filename: string
        } | null
        url: string
    }
    commitID: string
    endLine: number
    startLine: number
    filename: string
    message: string
    user?: {
        username: string
        displayName: string | null
        avatarURL: string | null
    }
}

export interface BlameHunkData {
    current: BlameHunk[] | undefined
    externalURLs: { url: string; serviceKind: ExternalServiceKind | null }[] | undefined
}

function rawHunkToBlameHunk(rawHunk: RawStreamHunk): BlameHunk {
    // TODO: do we actually need all this massaging? It would probably
    // be less confusing just to use the API types directly.
    const commitDate = new Date(rawHunk.author.Date)
    return {
        startLine: rawHunk.startLine,
        endLine: rawHunk.endLine,
        message: rawHunk.message,
        rev: rawHunk.commitID,
        filename: rawHunk.filename,
        author: {
            date: rawHunk.author.Date,
            person: {
                email: rawHunk.author?.Email,
                displayName: rawHunk.author.Name,
                avatarURL: rawHunk.user?.avatarURL ?? null,
                user: rawHunk.user ?? null,
            },
        },
        commit: {
            url: rawHunk.commit.url,
            previous: rawHunk.commit.previous
                ? {
                      rev: rawHunk.commit.previous.commitID,
                      filename: rawHunk.commit.previous.filename,
                  }
                : null,
        },
        displayInfo: {
            displayName: truncate(rawHunk.author.Name, { length: 25 }),
            username: rawHunk.user ? `(${rawHunk.user.username}) ` : '',
            commitDate,
            dateString: formatDateForBlame(commitDate, Date.now()),
            timestampString: commitDate.toLocaleString(),
            linkURL: `/${rawHunk.commit.url}`,
            message: truncate(rawHunk.message, { length: 45 }),
        },
    }
}

// fetchBlameHunksMemoized is a thin wrapper around fetchBlameHunks that memoizes
// the result
export const fetchBlameHunksMemoized = memoizeObservable(
    ({ repoName, revision, filePath }: { repoName: string; revision: string; filePath: string }) =>
        fetchBlameHunks(repoName, revision, filePath),
    makeRepoGitURI
)

// fetchBlameHunks returns an observable of all the hunks that have been streamed so far
function fetchBlameHunks(repoName: string, revision: string, filePath: string): Observable<BlameHunk[]> {
    return fetchRawBlameHunks(repoName, revision, filePath).pipe(
        // Map RawHunk to BlameHunk
        map(rawHunks => rawHunks.map(rawHunkToBlameHunk)),
        // Aggregate into an array, emitting an event for each update
        scan((aggregated, newHunks) => [...aggregated, ...newHunks]),
        // Limit frequency of updates to avoid a rerender on every event
        throttleTime(500, undefined, { leading: true, trailing: true })
    )
}

// fetchRawBlameHunks creates an observable that closely represents the event stream.
// There is one event on the observable for each "hunk" event on the stream, and the
// events are not aggregated (each event will probably only contain one RawStreamHunk).
function fetchRawBlameHunks(repoName: string, revision: string, filePath: string): Observable<RawStreamHunk[]> {
    const repoAndRevisionPath = `/${repoName}${revision ? `@${revision}` : ''}`
    return new Observable<RawStreamHunk[]>(subscriber => {
        fetchEventSource(`/.api/blame${repoAndRevisionPath}/-/stream/${filePath}`, {
            method: 'GET',
            headers: {
                ...window.context.xhrHeaders,
                'X-Sourcegraph-Should-Trace': new URLSearchParams(window.location.search).get('trace') || 'false',
            },
            async onopen(response) {
                if (response.ok && response.headers.get('content-type') === EventStreamContentType) {
                    return
                }
                throw new Error('request for blame data failed: ' + (await response.text()))
            },
            onmessage(event) {
                if (event.event === 'hunk') {
                    const rawHunks: RawStreamHunk[] = JSON.parse(event.data)
                    subscriber.next(rawHunks)
                }
            },
            onerror(err) {
                throw err
            },
        }).then(
            () => subscriber.complete(),
            error => subscriber.error(error)
        )
    })
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
