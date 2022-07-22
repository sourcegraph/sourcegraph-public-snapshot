import { useMemo } from 'react'

import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { memoizeObservable } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { makeRepoURI } from '@sourcegraph/shared/src/util/url'
import { useObservable } from '@sourcegraph/wildcard'

import { requestGraphQL } from '../backend/graphql'
import { GitBlameResult, GitBlameVariables } from '../graphql-operations'

const fetchBlame = memoizeObservable(
    ({
        repoName,
        commitID,
        filePath,
    }: {
        repoName: string
        commitID: string
        filePath: string
    }): Observable<any | null> => // GitBlameResult['repository']['commit']['blob']['blame']
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
    const blame = useObservable(
        useMemo(() => (isBlameVisible ? fetchBlame({ repoName, commitID, filePath }) : of(undefined)), [
            isBlameVisible,
            repoName,
            commitID,
            filePath,
        ])
    )

    return blame
}
