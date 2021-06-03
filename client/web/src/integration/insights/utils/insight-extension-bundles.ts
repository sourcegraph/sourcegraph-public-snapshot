import { View } from 'sourcegraph'

import { ErrorLike } from '@sourcegraph/shared/src/util/errors'

import { InsightTypePrefix } from '../../../insights/core/types'

/**
 * Generates a simplified version of search insight extension for testing purpose.
 * Full version of search insight extension you can find be link below
 * https://github.com/sourcegraph/sourcegraph-search-insights/blob/master/src/search-insights.ts
 * */
export function getSearchInsightExtensionBundle(data?: Record<string, View | undefined | ErrorLike>): string {
    return getUniversalInsightExtensionBundle(InsightTypePrefix.search, data ?? {})
}

/**
 * Generates simplify version of code stats insight extension for testing purpose.
 * Full version of code stats insight extension you find by link below
 * https://github.com/sourcegraph/sourcegraph-code-stats-insights/blob/master/src/code-stats-insights.ts
 * */
export function getCodeStatsInsightExtensionBundle(data?: Record<string, View | undefined | ErrorLike>): string {
    return getUniversalInsightExtensionBundle(InsightTypePrefix.langStats, data ?? {})
}

/**
 * Generates common insight extension mock implementation of real insight extensions.
 * Testing extension bundle listen setting cascade, filters setting and finds insights
 * configs by type param and provides mock data for insights by id from data param.
 * */
function getUniversalInsightExtensionBundle(
    type: InsightTypePrefix,
    data: Record<string, View | undefined | ErrorLike>
): string {
    const injectedDataString = JSON.stringify(data ?? {})

    return `
        var sourcegraph = require('sourcegraph')

        function activate(context) {
            var insightViewStore = JSON.parse('${injectedDataString}')
            var subscriptions = []

            function handleInsights(config) {
                const insights = Object.entries(config).filter(([key]) => key.startsWith('${type}.'))

                for (var insight of insights) {
                    const [id, settings] = insight;

                    var provideView =  () => insightViewStore[id]

                    subscriptions.push(sourcegraph.app.registerViewProvider(id + '.insightsPage', {
                        where: 'insightsPage',
                        provideView,
                    }))
                }
            }

            sourcegraph.configuration.subscribe(() => {
                var config = sourcegraph.configuration.get().value

                subscriptions.forEach(sub => sub.unsubscribe())
                subscriptions = []

                handleInsights(config)
            })
        }

        exports.activate = activate
    `
}
