import { Duration, formatISO, isAfter, startOfDay, sub } from 'date-fns'
import escapeRegExp from 'lodash/escapeRegExp'
import { defer } from 'rxjs'
import { retry } from 'rxjs/operators'

import { InsightContentType } from '../../../../types/insight/common'
import { GetSearchInsightContentInput, InsightSeriesContent } from '../../../code-insights-backend-types'

import { fetchRawSearchInsightResults, fetchSearchInsightCommits } from './utils/fetch-search-insight'
import { queryHasCountFilter } from './utils/query-has-count-filter'

interface RepoCommit {
    date: Date
    commit: string
    repo: string
}

interface InsightSeriesData {
    date: number
    [seriesName: string]: number
}

export async function getSearchInsightContent(
    input: GetSearchInsightContentInput
): Promise<InsightSeriesContent<InsightSeriesData>> {
    return getInsightContent(input)
}

export async function getInsightContent(
    inputs: GetSearchInsightContentInput
): Promise<InsightSeriesContent<InsightSeriesData>> {
    const { series, step, repositories } = inputs
    const dates = getDaysToQuery(step, 7)

    // Get commits to search for each day for each repository.
    const repoCommits = (
        await Promise.all(
            repositories.map(async repo =>
                (await determineCommitsToSearch(dates, repo)).map(commit => ({ repo, ...commit }))
            )
        )
    )
        .flat()
        // For commit which we couldn't find we should not run search API request.
        // Instead of it we will use just EMPTY_DATA_POINT_VALUE
        .filter(commitData => commitData.commit !== null) as RepoCommit[]

    const searchQueries = series.flatMap(({ query, name, id }) =>
        repoCommits.map(({ date, repo, commit }) => ({
            seriesId: id,
            name,
            date,
            repo,
            commit,
            query: [`repo:^${escapeRegExp(repo)}$@${commit}`, `${getQueryWithCount(query)}`].filter(Boolean).join(' '),
        }))
    )

    if (searchQueries.length === 0) {
        throw new Error('Data for these repositories not found')
    }

    const rawSearchResults = await defer(() => fetchRawSearchInsightResults(searchQueries.map(search => search.query)))
        // The bulk search may time out, but a retry is then likely faster
        // because caches are warm
        .pipe(retry(3))
        .toPromise()

    const searchResults = Object.entries(rawSearchResults).map(([field, result]) => {
        const index = +field.slice('search'.length)
        const query = searchQueries[index]

        return { ...query, result }
    })

    // Generate series map with points map for each series. All points initially
    // have null value
    const seriesData = generateInitialDataSeries(series, dates)

    for (const { seriesId, date, result } of searchResults) {
        const countForRepo = result.results.matchCount
        const point = seriesData[seriesId][date.getTime()]

        if (point.value === null) {
            point.value = countForRepo
        } else {
            point.value += countForRepo
        }
    }

    return {
        type: InsightContentType.Series,
        content: {
            series: series.map(series => ({
                id: series.id,
                name: series.name,
                color: series.stroke,
                data: Object.values(seriesData[series.id]),
                getXValue: datum => new Date(datum.date),
                getYValue: datum => datum.value,
                getLinkURL: datum => {
                    const date = datum.date
                    // Link to diff search that explains what new cases were added between two data points
                    const url = new URL('/search', window.location.origin)
                    // Use formatISO instead of toISOString(), because toISOString() always outputs UTC.
                    // They mark the same point in time, but using the user's timezone makes the date string
                    // easier to read (else the date component may be off by one day)
                    const after = formatISO(sub(date, step))
                    const before = formatISO(date)
                    const repoFilter = `repo:^(${repositories.map(escapeRegExp).join('|')})$`
                    const diffQuery = `${repoFilter} type:diff after:${after} before:${before} ${series.query}`

                    url.searchParams.set('q', diffQuery)

                    return url.href
                },
            })),
        },
    }
}

interface SearchCommit {
    date: Date
    commit: string | null
}

async function determineCommitsToSearch(dates: Date[], repo: string): Promise<SearchCommit[]> {
    const commitQueries = dates.map(date => {
        const before = formatISO(date)
        return `repo:^${escapeRegExp(repo)}$ type:commit before:${before} count:1`
    })

    const commitResults = await fetchSearchInsightCommits(commitQueries).toPromise()

    return Object.entries(commitResults).map(([name, search], index) => {
        const index_ = +name.slice('search'.length)
        const date = dates[index_]

        if (index_ !== index) {
            throw new Error(`Expected field ${name} to be at index ${index_} of object keys`)
        }

        const firstCommit = search.results.results[0]
        if (search.results.results.length === 0 || firstCommit?.__typename !== 'CommitSearchResult') {
            console.warn(`No result for ${commitQueries[index_]}`)

            return { commit: null, date }
        }

        const commit = firstCommit.commit

        // Sanity check
        const commitDate = commit.committer && new Date(commit.committer.date)

        if (!commitDate) {
            throw new Error(`Expected commit to have committer: \`${commit.oid}\``)
        }

        if (isAfter(commitDate, date)) {
            throw new Error(
                `Expected commit \`${commit.oid}\` to be before ${formatISO(date)}, but was after: ${formatISO(
                    commitDate
                )}.\nSearch query: ${commitQueries[index_]}`
            )
        }

        return { commit: commit.oid, date }
    })
}

function getDaysToQuery(step: Duration, numberOfPoints: number): Date[] {
    // Date.now used here for testing purpose we can mock now
    // method in test and avoid flaky test by that.
    const now = startOfDay(new Date(Date.now()))
    const dates: Date[] = []

    for (let index = 0, date = now; index < numberOfPoints; index++) {
        dates.unshift(date)
        date = sub(date, step)
    }

    return dates
}

function getQueryWithCount(query: string): string {
    return queryHasCountFilter(query)
        ? query
        : // Increase the number to the maximum value to get all the data we can have
          `${query} count:99999`
}

type SeriesId = string

interface SeriesData {
    [dateTime: number]: {
        date: Date
        value: number | null
    }
}

interface InitialSeriesData {
    [id: SeriesId]: SeriesData
}

/**
 * Generates initial series points values for each X (time) point in the dataset.
 * So
 */
function generateInitialDataSeries(series: { id: string }[], dates: Date[]): InitialSeriesData {
    const store: Record<SeriesId, SeriesData> = {}

    for (const line of series) {
        const { id: seriesId } = line
        const seriesData: SeriesData = {}

        for (const date of dates) {
            seriesData[date.getTime()] = { date, value: null }
        }

        store[seriesId] = seriesData
    }

    return store
}
