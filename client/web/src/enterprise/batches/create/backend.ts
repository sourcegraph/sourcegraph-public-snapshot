import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../backend/graphql'
import { BatchSpecCreateFields, CreateBatchSpecResult, CreateBatchSpecVariables } from '../../../graphql-operations'

export async function createBatchSpec(spec: string): Promise<BatchSpecCreateFields> {
    const result = await requestGraphQL<CreateBatchSpecResult, CreateBatchSpecVariables>(
        gql`
            mutation CreateBatchSpec($spec: String!) {
                createBatchSpecFromRaw(batchSpec: $spec) {
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
    return dataOrThrowErrors(result).createBatchSpecFromRaw
}
