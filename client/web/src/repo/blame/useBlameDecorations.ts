import { useMemo } from 'react'

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

export type BlameHunk = NonNullable<
    NonNullable<NonNullable<GitBlameResult['repository']>['commit']>['blob']
>['blame'][number]

const fetchBlame = memoizeObservable(
    ({
        repoName,
        commitID,
        filePath,
    }: {
        repoName: string
        commitID: string
        filePath: string
    }): Observable<BlameHunk[] | undefined> =>
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
            map(({ repository }) => repository?.commit?.blob?.blame)
        ),
    makeRepoURI
)

export const useBlameDecorations = (args?: {
    repoName: string
    commitID: string
    filePath: string
}): BlameHunk[] | undefined => {
    const { repoName, commitID, filePath } = args ?? {}
    const extensionsAsCoreFeatures = useExperimentalFeatures(features => features.extensionsAsCoreFeatures)
    const [isBlameVisible] = useBlameVisibility()
    const hunks = useObservable(
        useMemo(
            () =>
                extensionsAsCoreFeatures && commitID && repoName && filePath && isBlameVisible
                    ? fetchBlame({ commitID, repoName, filePath })
                    : of(undefined),
            [extensionsAsCoreFeatures, isBlameVisible, commitID, repoName, filePath]
        )
    )

    return hunks
}
