import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { ViewInsightProviderResult, ViewInsightProviderSourceType } from '../types'

export function createExtensionInsight(insight: ViewProviderResult): ViewInsightProviderResult {
    return {
        ...insight,
        // According to our naming convention of insight
        // <type>.<name>.<render view = insight page | directory | home page>
        // You can see insight id generation at extension codebase like here
        // https://github.com/sourcegraph/sourcegraph-search-insights/blob/master/src/search-insights.ts#L86
        id: insight.id.split('.').slice(0, -1).join('.'),
        // Convert error like errors since Firefox and Safari don't support
        // receiving native errors from web worker thread
        view: isErrorLike(insight.view) ? asError(insight.view) : insight.view,
        source: ViewInsightProviderSourceType.Extension,
    }
}
