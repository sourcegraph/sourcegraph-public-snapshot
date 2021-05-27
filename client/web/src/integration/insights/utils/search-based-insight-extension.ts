
import { View } from 'sourcegraph';

import { INSIGHT_VIEW_TYPES_MIGRATION } from './insight-mock-data';

/**
 * Generates simplify version of search insight extension for testing purpose.
 * */
export function generateSearchInsightExtensionBundle(data?: Record<string, View>): string {
    const injectedDataString = JSON.stringify(data ?? {})

    /**
     * If test didn't provide any data for particular insight with id we use
     * 'types migration' insight's data as a fallback.
     * */
    const injectedDefaultViewString = JSON.stringify(INSIGHT_VIEW_TYPES_MIGRATION)

    return `
        var sourcegraph = require('sourcegraph')
        var insightViewStore = JSON.parse('${injectedDataString}')
        var defaultView = JSON.parse('${injectedDefaultViewString}')

        exports.activate = (context) => {

            function handleInsights(config) {
                const insights = Object.entries(config).filter(([key]) => key.startsWith('searchInsights.insight.'))

                for (var insight of insights) {
                    const [id, settings] = insight;

                    var provideView = () => insightViewStore[id] ?? defaultView

                    context.subscriptions.add(sourcegraph.app.registerViewProvider(id + '.insightsPage', {
                        where: 'insightsPage',
                        provideView,
                    }))
                }
            }

            sourcegraph.configuration.subscribe(() => {
                var config = sourcegraph.configuration.get().value

                handleInsights(config)
            })
        }
    `
}
