import { omit } from 'lodash'
import fetch, { type RequestInit } from 'node-fetch'
import signale from 'signale'

interface QueryAnnotation {
    created_at: string
    updated_at: string
    name: string
    description: string
    query_id?: string
    id: string
}

interface Query {
    id: string
    breakdowns: string[]
    calculations: object[]
    filters: object[]
    orders: object[]
    limit: number
    time_range: number
}

interface BoardQuery {
    dataset: string
    query?: Query
    query_id: string
    query_annotation_id?: string
    query_style: string
    graph_settings: object
}

interface Board {
    name: string
    description?: string
    style: 'list' | 'visual'
    column_layout: 'multi' | 'single'
    queries: BoardQuery[]
    id: string
}

type NewEntity<T extends object> = Omit<T, 'id' | 'created_at' | 'updated_at'>

interface ErrorResult {
    error: string
}

const HONEYCOMB_API_URL = 'https://api.honeycomb.io/1/'
const WEB_APP_BOARDS_PREFIX = 'web-app:'

const ENV_HEADERS = {
    fromHeaders: {},
    toHeaders: {},
}

function updateGlobalEnvHeaders(fromEnvAPIKey: string, toEnvAPIKey: string): void {
    ENV_HEADERS.fromHeaders = {
        'X-Honeycomb-Team': fromEnvAPIKey,
    }
    ENV_HEADERS.toHeaders = {
        'X-Honeycomb-Team': toEnvAPIKey,
    }
}

interface FetchDataInit extends RequestInit {
    /**
     * Queries that target all the datasets within an environment can be created
     * and read with the same APIs as above, using the special __all__ token in
     * the URL instead of a dataset name. Both creating and reading environment queries is supported.
     *
     * https://docs.honeycomb.io/api/queries/#environment-queries
     */
    datasetIndependentEndpoint?: string
}

async function fetchData<T extends object>(endpoint: string, init?: FetchDataInit): Promise<T> {
    const { datasetIndependentEndpoint, ...requestInit } = init || {}

    const url = new URL(HONEYCOMB_API_URL + endpoint).toString()
    const response = await fetch(url, requestInit)

    if (!response.ok) {
        const errorMessage = await response.text()
        signale.error(errorMessage)
        process.exit(1)
    }

    const data = (await response.json()) as T | ErrorResult

    if ('error' in data) {
        if (datasetIndependentEndpoint) {
            return fetchData<T>(datasetIndependentEndpoint, requestInit)
        }

        signale.error(data.error)
        process.exit(1)
    }

    return data
}

/**
 * Query annotations API documentation:
 * https://docs.honeycomb.io/api/query-annotations/
 *
 * @param id query annotation ID
 * @param queryId query ID
 */
async function cloneQueryAnnotation(id: string, queryId: string): Promise<QueryAnnotation> {
    const queryAnnotation = await fetchData<QueryAnnotation>(`query_annotations/web-app/${id}`, {
        datasetIndependentEndpoint: `query_annotations/__all__/${id}`,
        headers: ENV_HEADERS.fromHeaders,
    })

    const newQueryAnnotation = {
        ...omit(queryAnnotation, 'id', 'updated_at', 'created_at'),
        query_id: queryId,
    } satisfies NewEntity<QueryAnnotation>

    signale.log('--------CREATING QUERY ANNOTATION', JSON.stringify(newQueryAnnotation, null, 2))

    const createdQueryAnnotation = await fetchData<QueryAnnotation>('query_annotations/web-app', {
        method: 'POST',
        body: JSON.stringify(newQueryAnnotation),
        headers: ENV_HEADERS.toHeaders,
    })

    signale.log('--------CREATED QUERY ANNOTATION', createdQueryAnnotation)

    return createdQueryAnnotation
}

interface CloneBoardQueryResult {
    createdQuery: Query
    createdQueryAnnotation?: QueryAnnotation
}

/**
 * Queries API documentation:
 * https://docs.honeycomb.io/api/queries/#create-a-query
 *
 * @param boardQuery existing board query object
 */
async function cloneBoardQuery(boardQuery: BoardQuery): Promise<CloneBoardQueryResult> {
    signale.log('--------QUERY CLONE STARTED', JSON.stringify(boardQuery, null, 2))

    const newQuery = omit(boardQuery.query, 'id')
    signale.log('--------CREATING QUERY', JSON.stringify(newQuery, null, 2))

    const createdQuery = await fetchData<Query>('queries/web-app', {
        method: 'POST',
        body: JSON.stringify(newQuery),
        headers: ENV_HEADERS.toHeaders,
    })

    signale.log('--------CREATED QUERY', createdQuery)

    if (!boardQuery.query_annotation_id) {
        return {
            createdQuery,
        }
    }

    const createdQueryAnnotation = await cloneQueryAnnotation(boardQuery.query_annotation_id, createdQuery.id)

    return {
        createdQuery,
        createdQueryAnnotation,
    }
}

/**
 * Boards API documentation:
 * https://docs.honeycomb.io/api/boards-api/#create-a-board
 *
 * @param board existing board object
 */
async function cloneBoard(board: Board): Promise<void> {
    signale.log('--------BOARD CLONE STARTED', board.name)

    const newBoardQueries = await Promise.all(
        board.queries.map(async boardQuery => {
            const { createdQuery, createdQueryAnnotation } = await cloneBoardQuery(boardQuery)

            const newBoardQuery = {
                ...omit(boardQuery, 'query'),
                query_id: createdQuery.id,
            } satisfies BoardQuery

            if (createdQueryAnnotation) {
                newBoardQuery.query_annotation_id = createdQueryAnnotation.id
            }

            return newBoardQuery
        })
    )

    const boardToCreate = {
        ...omit(board, 'id'),
        name: 'test_' + board.name,
        queries: newBoardQueries,
    } satisfies NewEntity<Board>

    signale.log('--------CREATING BOARD', JSON.stringify(boardToCreate, null, 2))

    const createdBoard = await fetchData<Board>('boards', {
        method: 'POST',
        body: JSON.stringify(boardToCreate),
        headers: ENV_HEADERS.toHeaders,
    })

    signale.log('--------CREATED BOARD', createdBoard)
}

/**
 * Clones all Honeycomb boards with names starting with `WEB_APP_BOARDS_PREFIX` from
 * one environment to another one re-creating all relevant queries and query
 * annotations in the process.
 *
 * You can find relevant API keys here:
 * https://ui.honeycomb.io/sourcegraph/environments
 *
 * @param keys [
 *   fromEnvAPIKey – the API key of the source environment.
 *   toEnvAPIKey – the API key of the target environment.
 * ]
 *
 */
async function cloneBoards(keys: string[]): Promise<void> {
    if (keys.length !== 2) {
        throw new Error('Usage: <fromEnvAPIKey> <toEnvAPIKey>')
    }

    updateGlobalEnvHeaders(...(keys as [string, string]))

    const boards = await fetchData<Board[]>('boards', {
        headers: ENV_HEADERS.fromHeaders,
    })

    signale.log(`--------FOUND ${boards.length} BOARDS`)

    for (const board of boards) {
        if (board.name.startsWith(WEB_APP_BOARDS_PREFIX)) {
            await cloneBoard(board)
        }
    }
}

/**
 * Usage from the observability-server package folder:
 * pnpm honeycomb:clone-boards <fromEnvAPIKey> <toEnvAPIKey>
 */
cloneBoards(process.argv.slice(2)).catch(signale.error)
