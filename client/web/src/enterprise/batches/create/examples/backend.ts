import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../../backend/graphql'
import {
    BatchSpecWorkspacesFields,
    CreateBatchSpecFromRawResult,
    CreateBatchSpecFromRawVariables,
} from '../../../../graphql-operations'

export function createBatchSpecFromRaw(spec: string): Observable<BatchSpecWorkspacesFields> {
    return requestGraphQL<CreateBatchSpecFromRawResult, CreateBatchSpecFromRawVariables>(
        gql`
            mutation CreateBatchSpecFromRaw($spec: String!) {
                createBatchSpecFromRaw(batchSpec: $spec) {
                    ...BatchSpecWorkspacesFields
                }
            }

            fragment BatchSpecWorkspacesFields on BatchSpec {
                originalInput
                workspaceResolution {
                    workspaces {
                        nodes {
                            ...BatchSpecWorkspaceFields
                        }
                    }
                    allowIgnored
                    allowUnsupported
                    unsupported {
                        nodes {
                            id
                            url
                            name
                        }
                    }
                }
            }

            fragment BatchSpecWorkspaceFields on BatchSpecWorkspace {
                repository {
                    id
                    name
                    url
                }
                ignored
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
                    run
                    container
                }
                searchResultPaths
            }
        `,
        { spec }
    ).pipe(
        map(dataOrThrowErrors),
        map(result => result.createBatchSpecFromRaw)
    )
}
