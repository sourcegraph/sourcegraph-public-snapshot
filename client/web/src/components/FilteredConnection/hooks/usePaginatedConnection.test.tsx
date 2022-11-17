import { MockedResponse } from '@apollo/client/testing'
import { fireEvent } from '@testing-library/react'

import { dataOrThrowErrors, getDocumentNode, gql } from '@sourcegraph/http-client'
import { renderWithBrandedContext, RenderWithBrandedContextResult } from '@sourcegraph/shared/src/testing'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { Text } from '@sourcegraph/wildcard'

import {
    TestPaginatedConnectionQueryFields,
    TestPaginatedConnectionQueryResult,
    TestPaginatedConnectionQueryVariables,
} from '../../../graphql-operations'

import { usePaginatedConnection } from './usePaginatedConnection'

const TEST_PAGINATED_CONNECTION_QUERY = gql`
    query TestPaginatedConnectionQuery($first: Int, $last: Int, $after: String, $before: String) {
        savedSearchesByNamespace(
            namespaceType: "Org"
            namespaceId: "1"
            first: $first
            last: $last
            after: $after
            before: $before
        ) {
            nodes {
                ...TestPaginatedConnectionQueryFields
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }

    fragment TestPaginatedConnectionQueryFields on SavedSearch {
        id
        description
    }
`

const PAGE_SIZE = 3

const TestComponent = () => {
    const {
        connection,
        nextPage,
        hasNextPage,
        previousPage,
        hasPreviousPage,
        firstPage,
        lastPage,
    } = usePaginatedConnection<
        TestPaginatedConnectionQueryResult,
        TestPaginatedConnectionQueryVariables,
        TestPaginatedConnectionQueryFields
    >({
        query: TEST_PAGINATED_CONNECTION_QUERY,
        variables: {},
        getConnection: result => {
            const data = dataOrThrowErrors(result)
            return data.savedSearchesByNamespace
        },
        options: {
            useURL: true,
            pageSize: PAGE_SIZE,
        },
    })

    return (
        <>
            <ul>
                {connection?.nodes.map((node, index) => (
                    <li key={index.toString()}>{node.description}</li>
                ))}
            </ul>
            {connection?.totalCount && <Text>Total count: {connection.totalCount}</Text>}
            <button type="button" onClick={firstPage}>
                First page
            </button>
            {connection?.pageInfo?.hasNextPage && (
                <button type="button" onClick={nextPage}>
                    Next page
                </button>
            )}
            {connection?.pageInfo?.hasPreviousPage && (
                <button type="button" onClick={previousPage}>
                    Previous page
                </button>
            )}
            <button type="button" onClick={lastPage}>
                Last page
            </button>
        </>
    )
}

const generateMockRequest = ({
    after,
    first,
    last,
    before,
}: {
    after: string | null
    first: number | null
    before: string | null
    last: number | null
}): MockedResponse<TestPaginatedConnectionQueryResult>['request'] => ({
    query: getDocumentNode(TEST_PAGINATED_CONNECTION_QUERY),
    variables: {
        after,
        first,
        last,
        before,
    },
})

const generateMockResult = ({
    startCursor,
    hasPreviousPage,
    endCursor,
    hasNextPage,
    nodes,
    totalCount,
}: {
    startCursor: string | null
    hasPreviousPage: boolean
    endCursor: string | null
    hasNextPage: boolean
    nodes: TestPaginatedConnectionQueryFields[]
    totalCount: number
}): MockedResponse<TestPaginatedConnectionQueryResult>['result'] => ({
    data: {
        savedSearchesByNamespace: {
            nodes,
            pageInfo: {
                startCursor,
                endCursor,
                hasNextPage,
                hasPreviousPage,
            },
            totalCount,
        },
    },
})

const goToNextPage = async (renderResult: RenderWithBrandedContextResult) => {
    const fetchMoreButton = renderResult.getByText('Next page')
    fireEvent.click(fetchMoreButton)
    // Skip loading state
    await waitForNextApolloResponse()
}

const mockResultNodes: TestPaginatedConnectionQueryFields[] = [
    { __typename: 'SavedSearch', id: '1', description: 'result 1' },
    { __typename: 'SavedSearch', id: '2', description: 'result 2' },
    { __typename: 'SavedSearch', id: '3', description: 'result 3' },
    { __typename: 'SavedSearch', id: '4', description: 'result 4' },
    { __typename: 'SavedSearch', id: '5', description: 'result 5' },
    { __typename: 'SavedSearch', id: '6', description: 'result 6' },
    { __typename: 'SavedSearch', id: '7', description: 'result 7' },
    { __typename: 'SavedSearch', id: '8', description: 'result 8' },
    { __typename: 'SavedSearch', id: '9', description: 'result 9' },
    { __typename: 'SavedSearch', id: '10', description: 'result 10' },
]

const getCursorForId = (id: string): string => `cursor:${id}`

const generateMockCursorResponsesForEveryPage = (
    nodes: TestPaginatedConnectionQueryFields[]
): MockedResponse<TestPaginatedConnectionQueryResult>[] => {
    const responses: MockedResponse<TestPaginatedConnectionQueryResult>[] = []

    const totalPages = Math.ceil(nodes.length / PAGE_SIZE)
    for (let pageIndex = 0; pageIndex < totalPages; pageIndex++) {
        const nodesOnPage = nodes.slice(pageIndex * PAGE_SIZE, (pageIndex + 1) * PAGE_SIZE)

        const after = pageIndex === 0 ? null : getCursorForId(nodes[pageIndex * PAGE_SIZE - 1].id)
        const before = pageIndex === totalPages - 1 ? null : getCursorForId(nodes[(pageIndex + 1) * PAGE_SIZE].id)

        // Forward request
        responses.push({
            request: generateMockRequest({ after, first: PAGE_SIZE, last: null, before: null }),
            result: generateMockResult({
                nodes: nodesOnPage,
                totalCount: nodes.length,
                startCursor: nodesOnPage.length > 0 ? getCursorForId(nodesOnPage[0].id) : null,
                endCursor: nodesOnPage.length > 0 ? getCursorForId(nodesOnPage[nodesOnPage.length - 1].id) : null,
                hasNextPage: pageIndex < totalPages - 1,
                hasPreviousPage: pageIndex > 0,
            }),
        })

        // Backward request
        responses.push({
            request: generateMockRequest({ before, last: PAGE_SIZE, after: null, first: null }),
            result: generateMockResult({
                nodes: nodesOnPage,
                totalCount: nodes.length,
                startCursor: nodesOnPage.length > 0 ? getCursorForId(nodesOnPage[0].id) : null,
                endCursor: nodesOnPage.length > 0 ? getCursorForId(nodesOnPage[nodesOnPage.length - 1].id) : null,
                hasNextPage: pageIndex < totalPages - 1,
                hasPreviousPage: pageIndex > 0,
            }),
        })
    }

    return responses
}

describe('usePaginatedConnection', () => {
    const renderWithMocks = async (mocks: MockedResponse<TestPaginatedConnectionQueryResult>[], route = '/') => {
        const renderResult = renderWithBrandedContext(
            <MockedTestProvider mocks={mocks}>
                <TestComponent />
            </MockedTestProvider>,
            { route }
        )

        // Skip loading state
        await waitForNextApolloResponse()

        return renderResult
    }

    describe('Cursor based pagination', () => {
        const cursorMocks = generateMockCursorResponsesForEveryPage(mockResultNodes)

        it('renders the first page', async () => {
            const queries = await renderWithMocks(cursorMocks)

            expect(queries.getAllByRole('listitem').length).toBe(3)
            expect(queries.getByText('result 1')).toBeVisible()
            expect(queries.getByText('result 2')).toBeVisible()
            expect(queries.getByText('result 3')).toBeVisible()
            expect(queries.getByText('Total count: 10')).toBeVisible()

            expect(queries.getByText('First page')).toBeVisible()
            expect(queries.getByText('Previous page')).not.toBeInTheDocument()
            expect(queries.getByText('Next page')).toBeVisible()
            expect(queries.getByText('Last page')).toBeVisible()
        })

        it('renders the next page', async () => {
            const queries = await renderWithMocks(cursorMocks)

            await goToNextPage(queries)

            expect(queries.getAllByRole('listitem').length).toBe(3)
            expect(queries.getByText('result 4')).toBeVisible()
            expect(queries.getByText('result 5')).toBeVisible()
            expect(queries.getByText('result 6')).toBeVisible()
            expect(queries.getByText('Total count: 10')).toBeVisible()

            console.log(queries.debug())

            expect(queries.getByText('First page')).toBeVisible()
            expect(queries.getByText('Previous page')).toBeVisible()
            expect(queries.getByText('Next page')).toBeVisible()
            expect(queries.getByText('Last page')).toBeVisible()
        })

        // it('fetches next page of results correctly', async () => {
        //     const queries = await renderWithMocks(cursorMocks)

        //     // Fetch first page
        //     await fetchNextPage(queries)

        //     // Both pages are now displayed
        //     expect(queries.getAllByRole('listitem').length).toBe(2)
        //     expect(queries.getByText('repo-A')).toBeVisible()
        //     expect(queries.getByText('repo-B')).toBeVisible()
        //     expect(queries.getByText('Total count: 4')).toBeVisible()
        //     expect(queries.getByText('Fetch more')).toBeVisible()

        //     // URL updates to match visible results
        //     expect(queries.history.location.search).toBe('?visible=2')
        // })

        // it('fetches final page of results correctly', async () => {
        //     const queries = await renderWithMocks(cursorMocks)

        //     // Fetch all pages
        //     await fetchNextPage(queries)
        //     await fetchNextPage(queries)
        //     await fetchNextPage(queries)

        //     // All pages of results are displayed
        //     expect(queries.getAllByRole('listitem').length).toBe(4)
        //     expect(queries.getByText('repo-A')).toBeVisible()
        //     expect(queries.getByText('repo-B')).toBeVisible()
        //     expect(queries.getByText('repo-C')).toBeVisible()
        //     expect(queries.getByText('repo-D')).toBeVisible()
        //     expect(queries.getByText('Total count: 4')).toBeVisible()

        //     // Fetch more button is now no longer visible
        //     expect(queries.queryByText('Fetch more')).not.toBeInTheDocument()

        //     // URL updates to match visible results
        //     expect(queries.history.location.search).toBe('?visible=4')
        // })

        // it('fetches correct amount of results when navigating directly with a URL', async () => {
        //     // We need to add an extra mock here, as we will derive a different `first` variable from `visible` in the URL.
        //     const mockFromVisible: MockedResponse<TestPaginatedConnectionQueryResult> = {
        //         request: generateMockRequest({ first: 3 }),
        //         result: generateMockResult({
        //             nodes: [mockResultNodes[0], mockResultNodes[1], mockResultNodes[2]],
        //             hasNextPage: true,
        //             endCursor: '3',
        //             totalCount: 4,
        //         }),
        //     }

        //     const queries = await renderWithMocks([...cursorMocks, mockFromVisible], '/?visible=3')

        //     // Renders 3 results without having to manually fetch
        //     expect(queries.getAllByRole('listitem').length).toBe(3)
        //     expect(queries.getByText('repo-A')).toBeVisible()
        //     expect(queries.getByText('repo-B')).toBeVisible()
        //     expect(queries.getByText('repo-C')).toBeVisible()
        //     expect(queries.getByText('Total count: 4')).toBeVisible()

        //     // Fetching next page should work as usual
        //     await fetchNextPage(queries)
        //     expect(queries.getAllByRole('listitem').length).toBe(4)
        //     expect(queries.getByText('repo-C')).toBeVisible()
        //     expect(queries.getByText('repo-D')).toBeVisible()
        //     expect(queries.getByText('Total count: 4')).toBeVisible()

        //     // URL should be overidden
        //     expect(queries.history.location.search).toBe('?visible=4')
        // })
    })
})
