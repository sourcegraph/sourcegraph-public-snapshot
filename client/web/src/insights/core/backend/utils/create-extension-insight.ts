import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'

import { ViewInsightProviderResult, ViewInsightProviderSourceType } from '../types'

export function createExtensionInsight(insight: ViewProviderResult): ViewInsightProviderResult {
    return {
        ...insight,
        source: ViewInsightProviderSourceType.Extension,
    }
}
