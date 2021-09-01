import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../../backend/graphql'
import {
    BatchSpecMatchingRepositoryFields,
    ResolveRepositoriesForBatchSpecResult,
    ResolveRepositoriesForBatchSpecVariables,
} from '../../../../graphql-operations'

export function resolveRepositoriesForBatchSpec(spec: string): Observable<BatchSpecMatchingRepositoryFields[]> {
    return requestGraphQL<ResolveRepositoriesForBatchSpecResult, ResolveRepositoriesForBatchSpecVariables>(
        gql`
            query ResolveRepositoriesForBatchSpec($spec: String!) {
                resolveRepositoriesForBatchSpec(batchSpec: $spec) {
                    ...BatchSpecMatchingRepositoryFields
                }
            }

            fragment BatchSpecMatchingRepositoryFields on BatchSpecMatchingRepository {
                repository {
                    id
                    name
                    url
                }
                path
            }
        `,
        { spec }
    ).pipe(
        map(dataOrThrowErrors),
        map(result => result.resolveRepositoriesForBatchSpec)
    )
}
