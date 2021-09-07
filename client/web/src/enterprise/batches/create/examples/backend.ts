import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../../backend/graphql'
import {
    BatchSpecWorkspacesFields,
    ResolveWorkspacesForBatchSpecResult,
    ResolveWorkspacesForBatchSpecVariables,
} from '../../../../graphql-operations'

export function resolveWorkspacesForBatchSpec(spec: string): Observable<BatchSpecWorkspacesFields> {
    return requestGraphQL<ResolveWorkspacesForBatchSpecResult, ResolveWorkspacesForBatchSpecVariables>(
        gql`
            query ResolveWorkspacesForBatchSpec($spec: String!) {
                resolveWorkspacesForBatchSpec(batchSpec: $spec) {
                    ...BatchSpecWorkspacesFields
                }
            }

            fragment BatchSpecWorkspacesFields on BatchSpecWorkspaces {
                rawSpec
                workspaces {
                    ...BatchSpecWorkspaceFields
                }
                allowIgnored
                allowUnsupported
                unsupported {
                    id
                    url
                    name
                }
                ignored {
                    id
                    url
                    name
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
                searchResultPaths
            }
        `,
        { spec }
    ).pipe(
        map(dataOrThrowErrors),
        map(result => result.resolveWorkspacesForBatchSpec)
    )
}
