import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import {
    createInvalidGraphQLMutationResponseError,
    dataOrThrowErrors,
    gql,
} from '../../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../../backend/graphql'
import {
    IndexConfigurationResult,
    IndexConfigurationVariables,
    RepositoryIndexConfigurationFields,
    UpdateRepositoryIndexConfigurationResult,
    UpdateRepositoryIndexConfigurationVariables,
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
