/**
 * Generates simplify version of search insight extension for testing purpose.
 * */
import { View } from 'sourcegraph';

const DEFAULT_INSIGHT_DATA: View = {
    'title': 'Migration to new GraphQL TS types',
    'content': [{
        'chart': 'line',
        'data': [{
            'date': 1600203600000,
            'Imports of old GQL.* types': 188,
            'Imports of new graphql-operations types': 203
        }, {
            'date': 1603832400000,
            'Imports of old GQL.* types': 178,
            'Imports of new graphql-operations types': 234
        }, {
            'date': 1607461200000,
            'Imports of old GQL.* types': 162,
            'Imports of new graphql-operations types': 282
        }, {
            'date': 1611090000000,
            'Imports of old GQL.* types': 139,
            'Imports of new graphql-operations types': 340
        }, {
            'date': 1614718800000,
            'Imports of old GQL.* types': 139,
            'Imports of new graphql-operations types': 354
        }, {
            'date': 1618347600000,
            'Imports of old GQL.* types': 139,
            'Imports of new graphql-operations types': 369
        }, {'date': 1621976400000, 'Imports of old GQL.* types': 131, 'Imports of new graphql-operations types': 427}],
        'series': [{
            'dataKey': 'Imports of old GQL.* types',
            'name': 'Imports of old GQL.* types',
            'stroke': 'var(--oc-red-7)',
        }, {
            'dataKey': 'Imports of new graphql-operations types',
            'name': 'Imports of new graphql-operations types',
            'stroke': 'var(--oc-blue-7)',
        }],
        'xAxis': {'dataKey': 'date', 'type': 'number', 'scale': 'time'}
    }]
}

export function generateSearchInsightExtensionBundle(data?: Record<string, View>): string {
    const injectedDataString = JSON.stringify(data ?? {})
    const injectedDefaultViewString = JSON.stringify(DEFAULT_INSIGHT_DATA)

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
