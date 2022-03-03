import { useContext, useMemo } from 'react'

import { useObservable } from '@sourcegraph/wildcard'

import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'

export interface UseUiFeatures {
    licensed: boolean
}

export function useUiFeatures(): UseUiFeatures {
    const { isCodeInsightsLicensed } = useContext(CodeInsightsBackendContext)

    const licensed = !!useObservable(useMemo(() => isCodeInsightsLicensed(), [isCodeInsightsLicensed]))

    return {
        licensed,
    }
}
