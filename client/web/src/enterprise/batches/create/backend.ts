import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../backend/graphql'
import {
    BatchSpecCreateFields,
    CreateBatchSpecResult,
    CreateBatchSpecVariables,
    Scalars,
} from '../../../graphql-operations'

export async function createBatchSpec(spec: Scalars['ID']): Promise<BatchSpecCreateFields> {
    const result = await requestGraphQL<CreateBatchSpecResult, CreateBatchSpecVariables>(
        gql`
            mutation CreateBatchSpec($id: ID!) {
                executeBatchSpec(batchSpec: $id) {
                    ...BatchSpecCreateFields
                }
            }

            fragment BatchSpecCreateFields on BatchSpec {
                id
                namespace {
                    url
                }
            }
        `,
        { spec }
    ).toPromise()
    return dataOrThrowErrors(result).executeBatchSpec
}
