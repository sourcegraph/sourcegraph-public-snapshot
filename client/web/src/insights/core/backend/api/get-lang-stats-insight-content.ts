import linguistLanguages from 'linguist-languages'
import { escapeRegExp, partition, sum } from 'lodash'
import { defer } from 'rxjs'
import { map, retry } from 'rxjs/operators'
import { DirectoryViewContext, PieChartContent } from 'sourcegraph'

import { ViewContexts } from '@sourcegraph/shared/src/api/extension/extensionHostApi'

import { fetchLangStatsInsight } from '../requests/fetch-lang-stats-insight'
import { LangStatsInsightsSettings } from '../types'
import { resolveDocumentURI } from '../utils/resolve-uri'

const isLinguistLanguage = (language: string): language is keyof typeof linguistLanguages =>
    Object.prototype.hasOwnProperty.call(linguistLanguages, language)

interface InsightOptions<D extends keyof ViewContexts> {
    where: D
    context: ViewContexts[D]
}

export async function getLangStatsInsightContent<D extends keyof ViewContexts>(
    insight: LangStatsInsightsSettings,
    options: InsightOptions<D>
): Promise<PieChartContent<any>> {
    const { where, context } = options

    switch (where) {
        case 'directory': {
            const { viewer } = context as DirectoryViewContext
            const { repo, path } = resolveDocumentURI(viewer.directory.uri)

            return getInsightContent({ insight, repo, path })
        }

        case 'homepage':
        case 'insightsPage': {
            return getInsightContent({ insight, repo: insight.repository })
        }
    }

    throw new Error(`This context is not supported for code-stats insight: context: ${where}`)
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
    const query = `repo:^${escapeRegExp(repo)} ${pathRegexp}`

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
