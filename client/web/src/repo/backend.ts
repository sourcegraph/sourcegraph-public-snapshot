import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError, memoizeObservable } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { RepoNotFoundError, RepoSeeOtherError } from '@sourcegraph/shared/src/backend/errors'
import { makeRepoURI, RepoRevision } from '@sourcegraph/shared/src/util/url'

import { queryGraphQL, requestGraphQL } from '../backend/graphql'
import {
    ExternalLinkFields,
    RepositoryRedirectResult,
    RepositoryRedirectVariables,
    RepositoryFields,
} from '../graphql-operations'

export const externalLinkFieldsFragment = gql`
    fragment ExternalLinkFields on ExternalLink {
        url
        serviceKind
    }
`

export const repositoryFragment = gql`
    fragment RepositoryFields on Repository {
        id
        name
        url
        externalURLs {
            url
            serviceKind
        }
        description
        viewerCanAdminister
        defaultBranch {
            displayName
            abbrevName
        }
    }
`

/**
 * Fetch the repository.
 */
export const fetchRepository = memoizeObservable(
    (args: { repoName: string }): Observable<RepositoryFields> =>
        requestGraphQL<RepositoryRedirectResult, RepositoryRedirectVariables>(
            gql`
                query RepositoryRedirect($repoName: String!) {
                    repositoryRedirect(name: $repoName) {
                        __typename
                        ... on Repository {
                            ...RepositoryFields
                        }
                        ... on Redirect {
                            url
                        }
                    }
                }
                ${repositoryFragment}
            `,
            args
        ).pipe(
            map(dataOrThrowErrors),
            map(data => {
                if (!data.repositoryRedirect) {
                    throw new RepoNotFoundError(args.repoName)
                }
                if (data.repositoryRedirect.__typename === 'Redirect') {
                    throw new RepoSeeOtherError(data.repositoryRedirect.url)
                }
                return data.repositoryRedirect
            })
        ),
    makeRepoURI
)

export const fetchFileExternalLinks = memoizeObservable(
    (context: RepoRevision & { filePath: string }): Observable<ExternalLinkFields[]> =>
        queryGraphQL(
            gql`
                query FileExternalLinks($repoName: String!, $revision: String!, $filePath: String!) {
                    repository(name: $repoName) {
                        commit(rev: $revision) {
                            file(path: $filePath) {
                                externalURLs {
                                    ...ExternalLinkFields
                                }
                            }
                        }
                    }
                }

                ${externalLinkFieldsFragment}
            `,
            context
        ).pipe(
            map(({ data, errors }) => {
                if (!data?.repository?.commit?.file?.externalURLs) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit.file.externalURLs
            })
        ),
    makeRepoURI
)
