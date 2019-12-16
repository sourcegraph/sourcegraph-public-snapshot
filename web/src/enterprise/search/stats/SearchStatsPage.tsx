import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PieChart, Pie, Tooltip, ResponsiveContainer, PieLabelRenderProps, Cell, TooltipFormatter } from 'recharts'
import ChartLineIcon from 'mdi-react/ChartLineIcon'
import React, { useCallback, useState, useMemo } from 'react'
import H from 'history'
import { Link } from 'react-router-dom'
import { numberWithCommas, pluralize } from '../../../../../shared/src/util/strings'
import { buildSearchURLQuery } from '../../../../../shared/src/util/url'
import { Form } from '../../../components/Form'
import { useObservable } from '../../../util/useObservable'
import { querySearchResultsStats } from './backend'

const OTHER_LANGUAGE = 'Other'
const UNKNOWN_LANGUAGE = 'Unknown'

/**
 * Return a copy of the stats with all languages that make up less than `minFraction` of the total
 * grouped together as "" (which is displayed as "Other").
 */
export const summarizeSearchResultsStatsLanguages = (
    languages: GQL.ISearchResultsStats['languages'],
    minFraction: number
): GQL.ISearchResultsStats['languages'] => {
    const totalLines = languages.reduce((sum, l) => sum + l.totalLines, 0)
    const minLines = minFraction * totalLines
    const languagesAboveMin = languages.filter(l => l.totalLines >= minLines)
    const otherLines = totalLines - languagesAboveMin.reduce((sum, l) => sum + l.totalLines, 0)
    return [
        ...languagesAboveMin,
        { __typename: 'LanguageStatistics', name: OTHER_LANGUAGE, totalBytes: 0, totalLines: otherLines },
    ]
}

/** Nice-looking colors for the pie chart that have good contrast in both light and dark themes. */
const COLORS = ['#278389', '#f16321', '#753fff', '#0091ea', '#00c853', '#ffab00', '#ff3d00', '#ff7700']

const OTHER_COLOR = '#999999'
const UNKNOWN_COLOR = '#777777'

interface Props {
    location: H.Location
    history: H.History

    /** Mockable in tests. */
    _querySearchResultsStats?: typeof querySearchResultsStats
}

/**
 * Shows statistics about the results for a search query.
 */
export const SearchStatsPage: React.FunctionComponent<Props> = ({
    location,
    history,
    _querySearchResultsStats = querySearchResultsStats,
}) => {
    const query = new URLSearchParams(location.search).get('q') || ''
    const [uncommittedQuery, setUncommittedQuery] = useState(query)
    const onUncommittedQueryChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(e => {
        setUncommittedQuery(e.currentTarget.value)
    }, [])
    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        e => {
            e.preventDefault()
            history.push({ ...location, search: new URLSearchParams({ q: uncommittedQuery }).toString() })
        },
        [history, location, uncommittedQuery]
    )

    // TODO(sqs): reuse the user's current patternType
    const stats = useObservable(
        useMemo(
            () => _querySearchResultsStats(query + ' count:99999999'), // add large count: to ensure we get all results
            [query, _querySearchResultsStats]
        )
    )
    const chartData = stats
        ? summarizeSearchResultsStatsLanguages(stats.languages, 0.02).map((l, i) => ({
              ...l,
              name: l.name || UNKNOWN_LANGUAGE,
              color: COLORS[i % COLORS.length],
          }))
        : undefined
    const totalLines = stats ? stats.languages.reduce((sum, l) => sum + l.totalLines, 0) : undefined

    const urlToSearchWithExtraQuery = useCallback(
        (extraQuery: string) =>
            `/search?${buildSearchURLQuery(`${query} ${extraQuery}`, GQL.SearchPatternType.literal)}`,
        [query]
    )

    const percent = useCallback(
        (lines: number) =>
            totalLines !== undefined && totalLines !== 0 ? `${Math.round((100 * lines) / totalLines)}%` : '',
        [totalLines]
    )
    const labelRenderer = useCallback(
        (props: PieLabelRenderProps): string =>
            `${props.name || UNKNOWN_LANGUAGE} ${
                props.percent !== undefined ? `(${Math.round(100 * props.percent)}%)` : ''
            }`,
        []
    )
    const tooltipFormatter = useCallback<TooltipFormatter>(
        value => (typeof value === 'number' ? `${numberWithCommas(value)} ${pluralize('line', value)}` : ''),
        []
    )

    return (
        <div className="search-stats-page container mt-4">
            <header className="d-flex align-items-center justify-content-between mb-3">
                <h2 className="d-flex align-items-center mb-0">
                    <ChartLineIcon className="icon-inline mr-2" /> Code statistics
                </h2>
            </header>
            <Form onSubmit={onSubmit} className="form">
                <div className="form-group d-flex align-items-stretch">
                    <input
                        id="stats-page__query"
                        className="form-control flex-1 e2e-stats-query"
                        type="search"
                        placeholder="Enter a Sourcegraph search query"
                        value={uncommittedQuery}
                        onChange={onUncommittedQueryChange}
                        autoCapitalize="off"
                        spellCheck={false}
                        autoCorrect="off"
                        autoComplete="off"
                    />
                    {uncommittedQuery !== query && (
                        <button type="submit" className="btn btn-primary ml-2 e2e-stats-query-update">
                            Update
                        </button>
                    )}
                </div>
            </Form>
            <hr className="my-3" />
            {stats === undefined || totalLines === undefined || chartData === undefined ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <div className="card mb-3">
                    <h4 className="card-header">Languages</h4>
                    {stats.languages.length > 0 ? (
                        <div className="d-flex">
                            <div className="flex-0 border-right">
                                <table className="search-stats-page__table table mb-0 border-top-0">
                                    <thead>
                                        <tr className="small">
                                            <th className="border-top-0">
                                                <span className="sr-only">Language</span>
                                            </th>
                                            <th className="border-top-0">Lines</th>
                                            <th className="border-top-0">
                                                <span className="sr-only">Percent</span>
                                            </th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {stats.languages.map(({ name, totalLines: lines }, i) => (
                                            <tr key={name || i}>
                                                <td>
                                                    {name ? (
                                                        <Link
                                                            to={urlToSearchWithExtraQuery(`lang:${name.toLowerCase()}`)}
                                                        >
                                                            {name}
                                                        </Link>
                                                    ) : (
                                                        UNKNOWN_LANGUAGE
                                                    )}
                                                </td>
                                                <td>{numberWithCommas(lines)}</td>
                                                <td>{percent(lines)}</td>
                                            </tr>
                                        ))}
                                    </tbody>
                                </table>
                            </div>
                            <ResponsiveContainer className="flex-1 mx-3" minHeight={600} aspect={1}>
                                <PieChart>
                                    <Pie
                                        dataKey="totalLines"
                                        isAnimationActive={false}
                                        data={chartData}
                                        label={labelRenderer}
                                    >
                                        {chartData.map((entry, i) => (
                                            <Cell
                                                key={entry.name}
                                                fill={
                                                    entry.name === UNKNOWN_LANGUAGE
                                                        ? UNKNOWN_COLOR
                                                        : entry.name === OTHER_LANGUAGE
                                                        ? OTHER_COLOR
                                                        : COLORS[i % COLORS.length]
                                                }
                                            />
                                        ))}
                                    </Pie>
                                    <Tooltip animationDuration={0} formatter={tooltipFormatter} />
                                </PieChart>
                            </ResponsiveContainer>
                        </div>
                    ) : (
                        <div className="card-body text-muted">No language statistics available.</div>
                    )}
                </div>
            )}
        </div>
    )
}
