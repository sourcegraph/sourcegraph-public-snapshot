
/**
 * Generates simplify version of search insight extension for testing purpose.
 * */
export function generateSearchInsightExtensionBundle(): string {

    return `
        var sourcegraph = require('sourcegraph')

        exports.activate = (context) => {

            var provideView = () => ({
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
            })

            context.subscriptions.add(sourcegraph.app.registerViewProvider('searchInsight.insight.tsMigration.insightsPage', {
                where: 'insightsPage',
                provideView,
            }))
        }
    `
}
