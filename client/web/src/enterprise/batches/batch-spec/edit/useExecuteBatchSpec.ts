import { useCallback, useEffect, useState } from 'react'

import { useNavigate } from 'react-router-dom'

import { logger } from '@sourcegraph/common'
import { useMutation } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'

import { useFeatureFlag } from '../../../../featureFlags/useFeatureFlag'
import { ExecuteBatchSpecResult, ExecuteBatchSpecVariables } from '../../../../graphql-operations'
import { EXECUTE_BATCH_SPEC } from '../../create/backend'

interface UseExecuteBatchSpecResult {
    /** Method to invoke the GraphQL mutation to execute the current batch spec. */
    executeBatchSpec: () => void
    /** Whether or not an execution request is currently in flight. */
    isLoading: boolean
    /** Any error from `executeBatchSpec`. */
    error?: Error
}

/**
 * Custom hook for edit page which wraps `useMutation` for `EXECUTE_BATCH_SPEC`.
 *
 * @param batchSpecID The current batch spec ID.
 */
export const useExecuteBatchSpec = (
    batchSpecID?: Scalars['ID'],
    noCache?: boolean,
    experimentalV2Execution?: boolean
): UseExecuteBatchSpecResult => {
    const [submitBatchSpec, { loading }] = useMutation<ExecuteBatchSpecResult, ExecuteBatchSpecVariables>(
        EXECUTE_BATCH_SPEC
    )
    const [useExperimentalExecution, , featureFlagError] = useFeatureFlag('native-ssbc-execution', false)

    useEffect(() => {
        if (featureFlagError) {
            logger.error('failed to get feature flag', featureFlagError)
        }
    }, [featureFlagError])

    const [executionError, setExecutionError] = useState<Error>()

    const navigate = useNavigate()
    const executeBatchSpec = useCallback(() => {
        if (!batchSpecID) {
            return
        }

        submitBatchSpec({
            variables: {
                batchSpec: batchSpecID,
                noCache: noCache === undefined ? null : noCache,
                useExperimentalExecution:
                    experimentalV2Execution !== undefined ? experimentalV2Execution : useExperimentalExecution,
            },
        })
            .then(({ data }) => {
                if (data?.executeBatchSpec) {
                    navigate(
                        `${data.executeBatchSpec.namespace.url}/batch-changes/${data.executeBatchSpec.description.name}/executions/${data.executeBatchSpec.id}`,
                        { replace: true }
                    )
                }
            })
            .catch(setExecutionError)
    }, [submitBatchSpec, noCache, experimentalV2Execution, navigate, batchSpecID, useExperimentalExecution])

    return {
        executeBatchSpec,
        isLoading: loading,
        error: executionError,
    }
}
