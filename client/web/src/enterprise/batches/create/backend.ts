import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../backend/graphql'
import {
    BatchSpecExecutionFields,
    CreateBatchSpecExecutionResult,
    CreateBatchSpecExecutionVariables,
} from '../../../graphql-operations'

export async function createBatchSpecExecution(spec: string): Promise<BatchSpecExecutionFields> {
    const result = await requestGraphQL<CreateBatchSpecExecutionResult, CreateBatchSpecExecutionVariables>(
        gql`
            mutation CreateBatchSpecExecution($spec: String!) {
                createBatchSpecExecution(spec: $spec) {
                    ...BatchSpecExecutionFields
                }
            }

            fragment BatchSpecExecutionFields on BatchSpecExecution {
                id
            }
        `,
        { spec }
    ).toPromise()
    return dataOrThrowErrors(result).createBatchSpecExecution
}
