import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { ContributableViewContainer } from '../../protocol'
import { RegisteredViewProvider, ViewContexts, ViewProviderResult } from '../extensionHostApi'

import { callViewProvidersInParallel } from './callViewProvidersInParallel'
import { proxySubscribable, ProxySubscribable } from './common'

type InsightsPageViewContextType = typeof ContributableViewContainer.InsightsPage
type InsightsPageContext = ViewContexts[InsightsPageViewContextType]

/**
 * Returns insights result list for the insights page.
 *
 * @param context - Insights page context meta data.
 * @param providers - List of all insights providers.
 * @param insightIds - list of insights ids to resolve from the providers.
 */
export function getInsightsViews(
    context: InsightsPageContext,
    providers: Observable<readonly RegisteredViewProvider<InsightsPageViewContextType>[]>,
    insightIds?: string[]
): ProxySubscribable<ViewProviderResult[]> {
    const insightsIdSet = new Set(insightIds)
    const dashboardInsights = providers.pipe(
        map(providers =>
            providers.filter(provider => {
                // If insight ids was specified we should resolve only
                // insights from this list
                if (insightIds) {
                    return insightsIdSet.has(provider.id)
                }

                // Otherwise we are in all insights mode and have to resolve
                // all insights that we have.
                return true
            })
        )
    )

    return proxySubscribable(callViewProvidersInParallel(context, dashboardInsights))
}
