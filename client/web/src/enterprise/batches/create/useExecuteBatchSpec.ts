import { useCallback, useState } from 'react'
import { useHistory } from 'react-router'

import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { useMutation } from '@sourcegraph/shared/src/graphql/graphql'

import { ExecuteBatchSpecResult, ExecuteBatchSpecVariables } from '../../../graphql-operations'

import { EXECUTE_BATCH_SPEC } from './backend'

interface UseExecuteBatchSpecResult {
    /** Method to invoke the GraphQL mutation to execute the current batch spec. */
    executeBatchSpec: () => void
    /** Whether or not an execution request is currently in flight. */
    isLoading: boolean
    /** Any error from `executeBatchSpec`. */
    error?: Error
}

/**
 * Custom hook for "CreateOrEdit" page which wraps `useMutation` for `EXECUTE_BATCH_SPEC`.
 *
 * @param batchSpecID The current batch spec ID.
 */
export const useExecuteBatchSpec = (batchSpecID?: Scalars['ID']): UseExecuteBatchSpecResult => {
    const [submitBatchSpec, { loading }] = useMutation<ExecuteBatchSpecResult, ExecuteBatchSpecVariables>(
        EXECUTE_BATCH_SPEC
    )

    const [executionError, setExecutionError] = useState<Error>()

    const history = useHistory()
    const executeBatchSpec = useCallback(() => {
        if (!batchSpecID) {
            return
        }

        submitBatchSpec({ variables: { batchSpec: batchSpecID } })
            .then(({ data }) => {
                if (data?.executeBatchSpec) {
                    history.push(`/batch-changes/executions/${data.executeBatchSpec.id}`)
                }
            })
            .catch(setExecutionError)
    }, [submitBatchSpec, history, batchSpecID])

    return {
        executeBatchSpec,
        isLoading: loading,
        error: executionError,
    }
}
