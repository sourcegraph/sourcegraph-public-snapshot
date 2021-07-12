import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../backend/graphql'
import {
    BatchSpecExecutionCreateFields,
    CreateBatchSpecExecutionResult,
    CreateBatchSpecExecutionVariables,
} from '../../../graphql-operations'

export async function createBatchSpecExecution(spec: string): Promise<BatchSpecExecutionCreateFields> {
    const result = await requestGraphQL<CreateBatchSpecExecutionResult, CreateBatchSpecExecutionVariables>(
        gql`
            mutation CreateBatchSpecExecution($spec: String!) {
                createBatchSpecExecution(spec: $spec) {
                    ...BatchSpecExecutionCreateFields
                }
            }

            fragment BatchSpecExecutionCreateFields on BatchSpecExecution {
                id
                namespace {
                    url
                }
            }
        `,
        { spec }
    ).toPromise()
    return dataOrThrowErrors(result).createBatchSpecExecution
}
