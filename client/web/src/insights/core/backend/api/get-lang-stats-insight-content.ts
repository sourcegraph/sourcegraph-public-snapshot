import linguistLanguages from 'linguist-languages'
import { escapeRegExp, partition, sum } from 'lodash'
import { defer } from 'rxjs'
import { map, retry } from 'rxjs/operators'
import { PieChartContent } from 'sourcegraph';

import { fetchLangStatsInsight } from '../requests/fetch-lang-stats-insight'
import { LangStatsInsightsSettings } from '../types'

const isLinguistLanguage = (language: string): language is keyof typeof linguistLanguages =>
    Object.prototype.hasOwnProperty.call(linguistLanguages, language)

/**
 * This logic is a simplified copy of fetch logic of lang-stats-based code insight extension.
 * In order to have live preview for creation UI we had to copy this logic from
 * extension.
 *
 * */
export async function getLangStatsInsightContent(settings: LangStatsInsightsSettings): Promise<PieChartContent<any>> {
    const { repository, threshold = 0.03 } = settings

    const query = `repo:^${escapeRegExp(repository)}`
    const stats = await defer(() => fetchLangStatsInsight(query))
        .pipe(
            // The search may timeout, but a retry is then likely faster because caches are warm
            retry(3),
            map(data => data.search!.stats)
        )
        .toPromise()

    if (stats.languages.length === 0) {
        throw new Error("We couldn't find the language statistics, try changing the repository.")
    }

    const totalLines = sum(stats.languages.map(language => language.totalLines))
    const linkURL = new URL('/stats', window.location.origin)

    linkURL.searchParams.set('q', query)

    const [notOther, other] = partition(stats.languages, language => language.totalLines / totalLines >= threshold)
    return {
        chart: 'pie' as const,
        pies: [
            {
                data: [
                    ...notOther,
                    {
                        name: 'Other',
                        totalLines: sum(other.map(language => language.totalLines)),
                    },
                ].map(language => ({
                    ...language,
                    fill: (isLinguistLanguage(language.name) && linguistLanguages[language.name].color) || 'gray',
                    linkURL: linkURL.href,
                })),
                dataKey: 'totalLines',
                nameKey: 'name',
                fillKey: 'fill',
                linkURLKey: 'linkURL',
            },
        ],
    }
}
