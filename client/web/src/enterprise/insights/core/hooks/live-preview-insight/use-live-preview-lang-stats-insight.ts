import { useCallback, useEffect, useState } from 'react'

import { escapeRegExp, partition, sum } from 'lodash'
import { defer, Observable } from 'rxjs'
import { map, retry } from 'rxjs/operators'

import { asError } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../../../../backend/graphql'
import { LangStatsInsightContentResult, LangStatsInsightContentVariables } from '../../../../../graphql-operations'
import { CategoricalChartContent } from '../../backend/code-insights-backend-types'

import { LivePreviewStatus, State } from './types'

interface LangStatsDatum {
    name: string
    totalLines: number
    linkURL: string
    fill: string
}

interface Props {
    skip?: boolean
    repository: string
    otherThreshold: number
    path?: string
}

interface Result<R> {
    state: State<R>
    refetch: () => void
    lazyQuery: () => Promise<R>
}

/**
 * Language stats insight live preview handler. This hook doesn't use code insight GQL API,
 * instead, it internally runs the search GQL API for data fetching and does post-processing
 * (language stats aggregation) on the frontend in browser runtime.
 */
export function useLivePreviewLangStatsInsight(props: Props): Result<CategoricalChartContent<LangStatsDatum>> {
    const { skip = false, path, repository, otherThreshold } = props

    const [updateTrigger, setResetTrigger] = useState(0)
    const [state, setState] = useState<State<CategoricalChartContent<LangStatsDatum>>>({
        status: LivePreviewStatus.Intact,
    })

    useEffect(() => {
        let hasRequestCanceled = false

        if (!skip) {
            setState({ status: LivePreviewStatus.Loading })

            getLangStats({ repository, otherThreshold, path })
                .then(data => !hasRequestCanceled && setState({ status: LivePreviewStatus.Data, data }))
                .catch(
                    error => !hasRequestCanceled && setState({ status: LivePreviewStatus.Error, error: asError(error) })
                )
        } else {
            setState({ status: LivePreviewStatus.Intact })
        }

        return () => {
            hasRequestCanceled = true
        }
    }, [skip, path, repository, otherThreshold, updateTrigger])

    const refetch = useCallback(() => setResetTrigger(count => count + 1), [])
    const lazyQuery = useCallback(() => getLangStats({ repository, otherThreshold, path }), [
        repository,
        otherThreshold,
        path,
    ])

    return {
        state,
        refetch,
        lazyQuery,
    }
}

interface GetInsightContentInputs {
    repository: string
    otherThreshold: number
    path?: string
}

async function getLangStats(inputs: GetInsightContentInputs): Promise<CategoricalChartContent<LangStatsDatum>> {
    const { repository, path, otherThreshold } = inputs

    const pathRegexp = path ? `file:^${escapeRegExp(path)}/` : ''
    const query = `repo:^${escapeRegExp(repository)}$ ${pathRegexp}`

    const stats = await defer(() => fetchLangStatsInsight(query))
        .pipe(
            // The search may time out, but a retry is then likely faster because caches are warm
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
        data,
        getDatumColor: datum => datum.fill,
        getDatumLink: datum => datum.linkURL,
        getDatumName: datum => datum.name,
        getDatumValue: datum => datum.totalLines,
    }
}

function fetchLangStatsInsight(query: string): Observable<LangStatsInsightContentResult> {
    return requestGraphQL<LangStatsInsightContentResult, LangStatsInsightContentVariables>(
        gql`
            query LangStatsInsightContent($query: String!) {
                search(query: $query) {
                    results {
                        limitHit
                    }
                    stats {
                        languages {
                            name
                            totalLines
                        }
                    }
                }
            }
        `,
        { query }
    ).pipe(map(dataOrThrowErrors))
}

async function getLangColor(language: string): Promise<string> {
    const { default: languagesMap } = await import('linguist-languages')

    const isLinguistLanguage = (language: string): language is keyof typeof languagesMap =>
        Object.prototype.hasOwnProperty.call(languagesMap, language)

    if (isLinguistLanguage(language)) {
        return languagesMap[language].color ?? 'gray'
    }

    return 'gray'
}
