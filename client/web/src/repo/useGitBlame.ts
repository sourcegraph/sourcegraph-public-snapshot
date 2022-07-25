import { useMemo } from 'react'

import formatDistanceStrict from 'date-fns/formatDistanceStrict'
import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'
import { Range } from '@sourcegraph/extension-api-classes'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../backend/graphql'
import { GitBlameResult, GitBlameVariables } from '../graphql-operations'

type BlameHunk = NonNullable<NonNullable<NonNullable<GitBlameResult['repository']>['commit']>['blob']>['blame'][number]

const getDecorationFromHunk = (hunk: BlameHunk, decoratedLine: number, now: number): TextDocumentDecoration => {
    const { displayName, username, dateString, linkURL, hoverMessage } = getDisplayInfoFromHunk({
        hunk,
        now,
    })

    return {
        range: new Range(decoratedLine, 0, decoratedLine, 0),
        isWholeLine: true,
        after: {
            light: {
                color: 'rgba(0, 0, 25, 0.55)',
                backgroundColor: 'rgba(193, 217, 255, 0.65)',
            },
            dark: {
                color: 'rgba(235, 235, 255, 0.55)',
                backgroundColor: 'rgba(15, 43, 89, 0.65)',
            },
            contentText: `${dateString} • ${username}${displayName} [${truncate(hunk.message, 45)}]`,
            hoverMessage,
            linkURL,
        },
    }
}

function truncate(string_: string, max: number, omission = '…'): string {
    if (string_.length <= max) {
        return string_
    }
    return `${string_.slice(0, max)}${omission}`
}

/**
 * Get display info shared between status bar items and text document decorations.
 */
const getDisplayInfoFromHunk = ({
    hunk: { author, commit, message },
    now,
}: {
    hunk: BlameHunk
    now: number
}): { displayName: string; username: string; dateString: string; linkURL: string; hoverMessage: string } => {
    const displayName = truncate(author.person.displayName, 25)
    const username = author.person.user ? `(${author.person.user.username}) ` : ''
    const dateString = formatDistanceStrict(new Date(author.date), now, { addSuffix: true })
    const linkURL = new URL(commit.url, 'https://sourcegraph.com').href
    const hoverMessage = `${author.person.email} • ${truncate(message, 1000)}`

    return {
        displayName,
        username,
        dateString,
        linkURL,
        hoverMessage,
    }
}

const fetchBlame = memoizeObservable(
    ({
        repoName,
        commitID,
        filePath,
    }: {
        repoName: string
        commitID: string
        filePath: string
    }): Observable<BlameHunk[]> =>
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
            { repo: repoName, rev: commitID, path: filePath }
        ).pipe(
            map(dataOrThrowErrors),
            map(({ repository }) => {
                console.log(repository)
                if (!repository?.commit?.blob) {
                    throw new Error('no blame data is available (repository, commit, or path not found)')
                }

                return repository.commit.blob.blame
            })
        ),
    makeRepoURI
)

const getBlameDecorations = (hunks: BlameHunk[]): TextDocumentDecoration[] => {
    const now = Date.now()

    return hunks.map(hunk => getDecorationFromHunk(hunk, hunk.startLine - 1, now))
}

export const useGitBlame = ({
    repoName,
    commitID,
    filePath,
}: {
    repoName: string
    commitID: string
    filePath: string
}): unknown => {
    const [isBlameVisible] = useTemporarySetting('git.showBlame', false)
    const hunks = useObservable(
        useMemo(() => (isBlameVisible || true ? fetchBlame({ repoName, commitID, filePath }) : of(undefined)), [
            isBlameVisible,
            repoName,
            commitID,
            filePath,
        ])
    )

    return hunks ? getBlameDecorations(hunks) : undefined
}
