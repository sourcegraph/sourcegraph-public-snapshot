import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../../backend/graphql'
import {
    BatchSpecWorkspaceFields,
    ResolveWorkspacesForBatchSpecResult,
    ResolveWorkspacesForBatchSpecVariables,
} from '../../../../graphql-operations'

export function resolveWorkspacesForBatchSpec(spec: string): Observable<BatchSpecWorkspaceFields[]> {
    return requestGraphQL<ResolveWorkspacesForBatchSpecResult, ResolveWorkspacesForBatchSpecVariables>(
        gql`
            query ResolveWorkspacesForBatchSpec($spec: String!) {
                resolveWorkspacesForBatchSpec(batchSpec: $spec) {
                    ...BatchSpecWorkspaceFields
                }
            }

            fragment BatchSpecWorkspaceFields on BatchSpecWorkspace {
                repository {
                    id
                    name
                    url
                }
                branch {
                    id
                    abbrevName
                    displayName
                    target {
                        oid
                    }
                }
                path
                onlyFetchWorkspace
                steps {
                    command
                    container
                }
            }
        `,
        { spec }
    ).pipe(
        map(dataOrThrowErrors),
        map(result => result.resolveWorkspacesForBatchSpec)
    )
}
