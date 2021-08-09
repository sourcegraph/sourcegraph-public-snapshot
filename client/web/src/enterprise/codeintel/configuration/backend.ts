import { Observable } from 'rxjs'
import { map, mapTo } from 'rxjs/operators'

import {
    createInvalidGraphQLMutationResponseError,
    dataOrThrowErrors,
    gql,
} from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../backend/graphql'
import {
    IndexConfigurationResult,
    IndexConfigurationVariables,
    RepositoryIndexConfigurationFields,
    UpdateRepositoryIndexConfigurationResult,
    UpdateRepositoryIndexConfigurationVariables,
    QueueAutoIndexJobForRepoResult,
    QueueAutoIndexJobForRepoVariables,
} from '../../../graphql-operations'

export function getConfiguration({ id }: { id: string }): Observable<RepositoryIndexConfigurationFields | null> {
    const query = gql`
        query IndexConfiguration($id: ID!) {
            node(id: $id) {
                ...RepositoryIndexConfigurationFields
            }
        }

        fragment RepositoryIndexConfigurationFields on Repository {
            __typename
            indexConfiguration {
                configuration
                inferredConfiguration
            }
        }
    `

    return requestGraphQL<IndexConfigurationResult, IndexConfigurationVariables>(query, { id }).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('No such Repository')
            }
            return node
        })
    )
}

export function updateConfiguration({ id, content }: { id: string; content: string }): Observable<void> {
    const query = gql`
        mutation UpdateRepositoryIndexConfiguration($id: ID!, $content: String!) {
            updateRepositoryIndexConfiguration(repository: $id, configuration: $content) {
                alwaysNil
            }
        }
    `

    return requestGraphQL<UpdateRepositoryIndexConfigurationResult, UpdateRepositoryIndexConfigurationVariables>(
        query,
        {
            id,
            content,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.updateRepositoryIndexConfiguration) {
                throw createInvalidGraphQLMutationResponseError('UpdateRepositoryIndexConfiguration')
            }
        })
    )
}

export function enqueueIndexJob(id: string, revision: string): Observable<void> {
    const query = gql`
        mutation QueueAutoIndexJobForRepo($id: ID!, $rev: String) {
            queueAutoIndexJobForRepo(repository: $id, rev: $rev) {
                alwaysNil
            }
        }
    `

    return requestGraphQL<QueueAutoIndexJobForRepoResult, QueueAutoIndexJobForRepoVariables>(query, {
        id,
        rev: revision,
    }).pipe(map(dataOrThrowErrors), mapTo(undefined))
}
