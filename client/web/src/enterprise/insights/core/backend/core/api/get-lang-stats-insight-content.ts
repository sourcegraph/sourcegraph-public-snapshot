import { escapeRegExp, partition, sum } from 'lodash'
import { defer } from 'rxjs'
import { map, retry } from 'rxjs/operators'
import { PieChartContent } from 'sourcegraph'

import { LangStatsInsightsSettings } from '../../code-insights-backend-types'
import { fetchLangStatsInsight } from '../requests/fetch-lang-stats-insight'

const getLangColor = async (language: string): Promise<string> => {
    const { default: languagesMap } = await import('linguist-languages')

    const isLinguistLanguage = (language: string): language is keyof typeof languagesMap =>
        Object.prototype.hasOwnProperty.call(languagesMap, language)

    if (isLinguistLanguage(language)) {
        return languagesMap[language].color ?? 'gray'
    }

    return 'gray'
}

export async function getLangStatsInsightContent(insight: LangStatsInsightsSettings): Promise<PieChartContent<any>> {
    return getInsightContent({ insight, repo: insight.repository })
}

interface GetInsightContentInputs {
    insight: LangStatsInsightsSettings
    repo: string
    path?: string
}

async function getInsightContent(inputs: GetInsightContentInputs): Promise<PieChartContent<any>> {
    const {
        insight: { otherThreshold },
        repo,
        path,
    } = inputs

    const pathRegexp = path ? `file:^${escapeRegExp(path)}/` : ''
    const query = `repo:^${escapeRegExp(repo)}$ ${pathRegexp}`

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

    const [notOther, other] = partition(stats.languages, language => language.totalLines / totalLines >= otherThreshold)
    const data = await Promise.all(
        [...notOther, { name: 'Other', totalLines: sum(other.map(language => language.totalLines)) }].map(
            async language => ({
                ...language,
                fill: await getLangColor(language.name),
                linkURL: linkURL.href,
            })
        )
    )

    return {
        chart: 'pie' as const,
        pies: [
            {
                data,
                dataKey: 'totalLines',
                nameKey: 'name',
                fillKey: 'fill',
                linkURLKey: 'linkURL',
            },
        ],
    }
}
