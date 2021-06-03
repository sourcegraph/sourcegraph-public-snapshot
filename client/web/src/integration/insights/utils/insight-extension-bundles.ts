import { View, Unsubscribable, ProviderResult } from 'sourcegraph'

import { ErrorLike } from '@sourcegraph/shared/src/util/errors'

import { InsightTypePrefix } from '../../../insights/core/types'

/**
 * Generates a simplified version of search insight extension for testing purpose.
 * Full version of search insight extension you can find be link below
 * https://github.com/sourcegraph/sourcegraph-search-insights/blob/master/src/search-insights.ts
 * */
export function getSearchInsightExtensionBundle(views?: Record<string, View | undefined | ErrorLike>): string {
    return getUniversalInsightExtensionBundle(InsightTypePrefix.search, views ?? {})
}

/**
 * Generates a simplified version of code stats insight extension for testing purposes.
 * Full version of code stats insight extension you find by link below
 * https://github.com/sourcegraph/sourcegraph-code-stats-insights/blob/master/src/code-stats-insights.ts
 * */
export function getCodeStatsInsightExtensionBundle(views?: Record<string, View | undefined | ErrorLike>): string {
    return getUniversalInsightExtensionBundle(InsightTypePrefix.langStats, views ?? {})
}

/**
 * Generates common insight extension mock implementation of real insight extensions.
 * Testing extension bundle listen setting cascade, filters setting and finds insights
 * configs by type param and provides mock data for insights by id from data param.
 * */
function getUniversalInsightExtensionBundle(
    type: InsightTypePrefix,
    views: Record<string, View | undefined | ErrorLike>
): string {
    /**
     * Note that $TYPE and $VIEWS are placeholders which will be replaced by
     * insight data on function serialization step below.
     */
    function extensionBundle(): void {
        // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
        const sourcegraph = require('sourcegraph') as typeof import('sourcegraph')

        function activate(): void {
            // prettier-ignore
            const views: Record<string, ProviderResult<View>> = {/* $VIEWS */}
            let subscriptions: Unsubscribable[] = []

            function handleInsights(config: Record<string, unknown>): void {
                const insights = Object.entries(config).filter(([key]) => key.startsWith('$TYPE'))

                for (const insight of insights) {
                    const [id] = insight

                    // eslint-disable-next-line unicorn/consistent-function-scoping,@typescript-eslint/explicit-function-return-type
                    const provideView = () => views[id]

                    subscriptions.push(
                        sourcegraph.app.registerViewProvider(id + '.insightsPage', {
                            where: 'insightsPage',
                            provideView,
                        })
                    )
                }
            }

            sourcegraph.configuration.subscribe(() => {
                const config = sourcegraph.configuration.get().value

                for (const subscription of subscriptions) {
                    subscription.unsubscribe()
                }

                subscriptions = []

                handleInsights(config)
            })
        }

        exports.activate = activate
    }

    return `(${extensionBundle
        .toString()
        .replace("'$TYPE'", `'${type}'`)
        .replace('{ /* $VIEWS */}', JSON.stringify(views))})()`
}
