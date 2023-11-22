import type { MockedResponse } from '@apollo/client/testing'

import type { EvaluateFeatureFlagResult, EvaluateFeatureFlagVariables } from '../graphql-operations'

import type { FeatureFlagName } from './featureFlags'
import { EVALUATE_FEATURE_FLAG_QUERY } from './useFeatureFlag'

type FlagMock = MockedResponse<EvaluateFeatureFlagResult, EvaluateFeatureFlagVariables>

export const createFlagMock = (flagName: FeatureFlagName, valueOrError: boolean | Error): FlagMock => ({
    request: {
        query: EVALUATE_FEATURE_FLAG_QUERY,
        variables: {
            flagName,
        },
    },
    ...(typeof valueOrError === 'boolean' && {
        result: {
            data: {
                evaluateFeatureFlag: valueOrError,
            },
        },
    }),
    ...(typeof valueOrError === 'object' && {
        error: valueOrError,
    }),
})
