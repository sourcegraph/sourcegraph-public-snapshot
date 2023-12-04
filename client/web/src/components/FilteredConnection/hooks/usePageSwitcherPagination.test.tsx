import type { MockedResponse } from '@apollo/client/testing'
import { fireEvent } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { dataOrThrowErrors, getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { Text } from '@sourcegraph/wildcard'
import { type RenderWithBrandedContextResult, renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { usePageSwitcherPagination } from './usePageSwitcherPagination'

type TestPageSwitcherPaginationQueryFields = any
type TestPageSwitcherPaginationQueryResult = any
type TestPageSwitcherPaginationQueryVariables = any
const TEST_PAGINATED_CONNECTION_QUERY = `
    query TestPageSwitcherPaginationQuery($first: Int, $last: Int, $after: String, $before: String) {
        savedSearchesByNamespace(
            namespaceType: "Org"
            namespaceId: "1"
            first: $first
            last: $last
            after: $after
            before: $before
        ) {
            nodes {
                ...TestPageSwitcherPaginationQueryFields
            }
            totalCount
            pageInfo {
                endCursor
                startCursor
                hasNextPage
                hasPreviousPage
            }
        }
    }

    fragment TestPageSwitcherPaginationQueryFields on SavedSearch {
        id
        description
    }
`

const PAGE_SIZE = 3

const TestComponent = ({ useURL }: { useURL: boolean }) => {
    const { connection, loading, goToNextPage, goToPreviousPage, goToFirstPage, goToLastPage } =
        usePageSwitcherPagination<
            TestPageSwitcherPaginationQueryResult,
            TestPageSwitcherPaginationQueryVariables,
            TestPageSwitcherPaginationQueryFields
        >({
            query: TEST_PAGINATED_CONNECTION_QUERY,
            variables: {},
            getConnection: result => {
                const data = dataOrThrowErrors(result)
                return data.savedSearchesByNamespace
            },
            options: {
                useURL,
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
            {loading ? <Text>Loading...</Text> : null}
            {connection?.totalCount && <Text>Total count: {connection.totalCount}</Text>}
            <button type="button" onClick={goToFirstPage}>
                First page
            </button>
            {connection?.pageInfo?.hasNextPage && (
                <button type="button" onClick={goToNextPage}>
                    Next page
                </button>
            )}
            {connection?.pageInfo?.hasPreviousPage && (
                <button type="button" onClick={goToPreviousPage}>
                    Previous page
                </button>
            )}
            <button type="button" onClick={goToLastPage}>
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
}): MockedResponse<TestPageSwitcherPaginationQueryResult>['request'] => ({
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
    nodes: TestPageSwitcherPaginationQueryFields[]
    totalCount: number
}): MockedResponse<TestPageSwitcherPaginationQueryResult>['result'] => ({
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

const goToFirstPage = async (renderResult: RenderWithBrandedContextResult) => {
    fireEvent.click(renderResult.getByText('First page'))
    await waitForNextApolloResponse()
}
const goToNextPage = async (renderResult: RenderWithBrandedContextResult) => {
    fireEvent.click(renderResult.getByText('Next page'))
    await waitForNextApolloResponse()
}
const goToPreviousPage = async (renderResult: RenderWithBrandedContextResult) => {
    fireEvent.click(renderResult.getByText('Previous page'))
    await waitForNextApolloResponse()
}
const goToLastPage = async (renderResult: RenderWithBrandedContextResult) => {
    fireEvent.click(renderResult.getByText('Last page'))
    await waitForNextApolloResponse()
}

const mockResultNodes: TestPageSwitcherPaginationQueryFields[] = [
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

const getCursorForId = (id: string): string => `cursor_${id}`

const generateMockCursorResponsesForEveryPage = (
    nodes: TestPageSwitcherPaginationQueryFields[]
): MockedResponse<TestPageSwitcherPaginationQueryResult>[] => {
    const responses: MockedResponse<TestPageSwitcherPaginationQueryResult>[] = []

    const totalPages = Math.ceil(nodes.length / PAGE_SIZE)

    // Forward pagination
    for (let pageIndex = 0; pageIndex < totalPages; pageIndex++) {
        const nodesOnPage = nodes.slice(pageIndex * PAGE_SIZE, (pageIndex + 1) * PAGE_SIZE)
        const after = pageIndex === 0 ? null : getCursorForId(nodes[pageIndex * PAGE_SIZE - 1].id)
        responses.push({
            request: generateMockRequest({ after, first: PAGE_SIZE, last: null, before: null }),
            result: generateMockResult({
                nodes: nodesOnPage,
                totalCount: nodes.length,
                startCursor: nodesOnPage.length > 0 ? getCursorForId(nodesOnPage[0].id) : null,
                endCursor: nodesOnPage.length > 0 ? getCursorForId(nodesOnPage.at(-1).id) : null,
                hasNextPage: pageIndex < totalPages - 1,
                hasPreviousPage: pageIndex > 0,
            }),
        })
    }

    // Backward pagination
    const reverseNodes = [...nodes].reverse()
    for (let pageIndex = 0; pageIndex < totalPages; pageIndex++) {
        const nodesOnPage = reverseNodes.slice(pageIndex * PAGE_SIZE, (pageIndex + 1) * PAGE_SIZE).reverse()
        const before = pageIndex === 0 ? null : getCursorForId(reverseNodes[pageIndex * PAGE_SIZE - 1].id)
        responses.push({
            request: generateMockRequest({ before, last: PAGE_SIZE, after: null, first: null }),
            result: generateMockResult({
                nodes: nodesOnPage,
                totalCount: reverseNodes.length,
                startCursor: nodesOnPage.length > 0 ? getCursorForId(nodesOnPage[0].id) : null,
                endCursor: nodesOnPage.length > 0 ? getCursorForId(nodesOnPage.at(-1).id) : null,
                hasNextPage: pageIndex > 0,
                hasPreviousPage: pageIndex < totalPages - 1,
            }),
        })
    }

    return responses
}

describe('usePageSwitcherPagination', () => {
    const renderWithMocks = async (
        mocks: MockedResponse<TestPageSwitcherPaginationQueryResult>[],
        useURL: boolean = true,
        initialRoute = '/'
    ) => {
        const renderResult = renderWithBrandedContext(
            <MockedTestProvider mocks={mocks}>
                <TestComponent useURL={useURL} />
            </MockedTestProvider>,
            { route: initialRoute }
        )

        // Skip loading state
        await waitForNextApolloResponse()

        return renderResult
    }

    const cursorMocks = generateMockCursorResponsesForEveryPage(mockResultNodes)

    it('renders the first page', async () => {
        const page = await renderWithMocks(cursorMocks)

        expect(page.getAllByRole('listitem').length).toBe(3)
        expect(page.getAllByRole('listitem')[0]).toHaveTextContent('result 1')
        expect(page.getAllByRole('listitem')[1]).toHaveTextContent('result 2')
        expect(page.getAllByRole('listitem')[2]).toHaveTextContent('result 3')
        expect(page.getByText('Total count: 10')).toBeVisible()

        expect(page.getByText('First page')).toBeVisible()
        expect(() => page.getByText('Previous page')).toThrowError(/Unable to find an element/)
        expect(page.getByText('Next page')).toBeVisible()
        expect(page.getByText('Last page')).toBeVisible()

        expect(page.locationRef.current?.search).toBe('')
    })

    it('supports forward pagination', async () => {
        const page = await renderWithMocks(cursorMocks)

        await goToNextPage(page)

        expect(page.getAllByRole('listitem').length).toBe(3)
        expect(page.getAllByRole('listitem')[0]).toHaveTextContent('result 4')
        expect(page.getAllByRole('listitem')[1]).toHaveTextContent('result 5')
        expect(page.getAllByRole('listitem')[2]).toHaveTextContent('result 6')
        expect(page.getByText('Total count: 10')).toBeVisible()

        expect(page.getByText('First page')).toBeVisible()
        expect(page.getByText('Previous page')).toBeVisible()
        expect(page.getByText('Next page')).toBeVisible()
        expect(page.getByText('Last page')).toBeVisible()

        expect(page.locationRef.current?.search).toBe(`?after=${getCursorForId('3')}`)

        await goToNextPage(page)

        expect(page.getAllByRole('listitem').length).toBe(3)
        expect(page.getAllByRole('listitem')[0]).toHaveTextContent('result 7')
        expect(page.getAllByRole('listitem')[1]).toHaveTextContent('result 8')
        expect(page.getAllByRole('listitem')[2]).toHaveTextContent('result 9')
        expect(page.getByText('Total count: 10')).toBeVisible()

        expect(page.getByText('First page')).toBeVisible()
        expect(page.getByText('Previous page')).toBeVisible()
        expect(page.getByText('Next page')).toBeVisible()
        expect(page.getByText('Last page')).toBeVisible()

        expect(page.locationRef.current?.search).toBe(`?after=${getCursorForId('6')}`)

        await goToNextPage(page)

        expect(page.getAllByRole('listitem').length).toBe(1)
        expect(page.getAllByRole('listitem')[0]).toHaveTextContent('result 10')
        expect(page.getByText('Total count: 10')).toBeVisible()

        expect(page.getByText('First page')).toBeVisible()
        expect(page.getByText('Previous page')).toBeVisible()
        expect(() => page.getByText('Next page')).toThrowError(/Unable to find an element/)
        expect(page.getByText('Last page')).toBeVisible()

        expect(page.locationRef.current?.search).toBe(`?after=${getCursorForId('9')}`)
    })

    it('supports restoration from forward pagination URL', async () => {
        const page = await renderWithMocks(cursorMocks, true, `/?after=${getCursorForId('6')}`)

        expect(page.getAllByRole('listitem').length).toBe(3)
        expect(page.getAllByRole('listitem')[0]).toHaveTextContent('result 7')
        expect(page.getAllByRole('listitem')[1]).toHaveTextContent('result 8')
        expect(page.getAllByRole('listitem')[2]).toHaveTextContent('result 9')
        expect(page.getByText('Total count: 10')).toBeVisible()

        expect(page.getByText('First page')).toBeVisible()
        expect(page.getByText('Previous page')).toBeVisible()
        expect(page.getByText('Next page')).toBeVisible()
        expect(page.getByText('Last page')).toBeVisible()
    })

    it('supports jumping to the first page', async () => {
        const page = await renderWithMocks(cursorMocks, true, '/?last=3')

        await goToFirstPage(page)

        expect(page.getAllByRole('listitem').length).toBe(3)
        expect(page.getAllByRole('listitem')[0]).toHaveTextContent('result 1')
        expect(page.getAllByRole('listitem')[1]).toHaveTextContent('result 2')
        expect(page.getAllByRole('listitem')[2]).toHaveTextContent('result 3')
        expect(page.getByText('Total count: 10')).toBeVisible()

        expect(page.getByText('First page')).toBeVisible()
        expect(() => page.getByText('Previous page')).toThrowError(/Unable to find an element/)
        expect(page.getByText('Next page')).toBeVisible()
        expect(page.getByText('Last page')).toBeVisible()

        expect(page.locationRef.current?.search).toBe('')
    })

    it('supports jumping to the last page', async () => {
        const page = await renderWithMocks(cursorMocks)

        await goToLastPage(page)

        expect(page.getAllByRole('listitem').length).toBe(3)
        expect(page.getAllByRole('listitem')[0]).toHaveTextContent('result 8')
        expect(page.getAllByRole('listitem')[1]).toHaveTextContent('result 9')
        expect(page.getAllByRole('listitem')[2]).toHaveTextContent('result 10')
        expect(page.getByText('Total count: 10')).toBeVisible()

        expect(page.getByText('First page')).toBeVisible()
        expect(page.getByText('Previous page')).toBeVisible()
        expect(() => page.getByText('Next page')).toThrowError(/Unable to find an element/)
        expect(page.getByText('Last page')).toBeVisible()

        expect(page.locationRef.current?.search).toBe('?last=3')
    })

    it('supports restoration from last page URL', async () => {
        const page = await renderWithMocks(cursorMocks, true, '/?last=3')

        expect(page.getAllByRole('listitem').length).toBe(3)
        expect(page.getAllByRole('listitem')[0]).toHaveTextContent('result 8')
        expect(page.getAllByRole('listitem')[1]).toHaveTextContent('result 9')
        expect(page.getAllByRole('listitem')[2]).toHaveTextContent('result 10')
        expect(page.getByText('Total count: 10')).toBeVisible()

        expect(page.getByText('First page')).toBeVisible()
        expect(page.getByText('Previous page')).toBeVisible()
        expect(() => page.getByText('Next page')).toThrowError(/Unable to find an element/)
        expect(page.getByText('Last page')).toBeVisible()
    })

    it('supports backward pagination', async () => {
        const page = await renderWithMocks(cursorMocks)

        await goToLastPage(page)
        await goToPreviousPage(page)

        expect(page.getAllByRole('listitem').length).toBe(3)
        expect(page.getAllByRole('listitem')[0]).toHaveTextContent('result 5')
        expect(page.getAllByRole('listitem')[1]).toHaveTextContent('result 6')
        expect(page.getAllByRole('listitem')[2]).toHaveTextContent('result 7')
        expect(page.getByText('Total count: 10')).toBeVisible()

        expect(page.getByText('First page')).toBeVisible()
        expect(page.getByText('Previous page')).toBeVisible()
        expect(page.getByText('Next page')).toBeVisible()
        expect(page.getByText('Last page')).toBeVisible()

        expect(page.locationRef.current?.search).toBe(`?before=${getCursorForId('8')}`)

        await goToPreviousPage(page)

        expect(page.getAllByRole('listitem').length).toBe(3)
        expect(page.getAllByRole('listitem')[0]).toHaveTextContent('result 2')
        expect(page.getAllByRole('listitem')[1]).toHaveTextContent('result 3')
        expect(page.getAllByRole('listitem')[2]).toHaveTextContent('result 4')
        expect(page.getByText('Total count: 10')).toBeVisible()

        expect(page.getByText('First page')).toBeVisible()
        expect(page.getByText('Previous page')).toBeVisible()
        expect(page.getByText('Next page')).toBeVisible()
        expect(page.getByText('Last page')).toBeVisible()

        expect(page.locationRef.current?.search).toBe(`?before=${getCursorForId('5')}`)

        await goToPreviousPage(page)

        expect(page.getAllByRole('listitem').length).toBe(1)
        expect(page.getAllByRole('listitem')[0]).toHaveTextContent('result 1')
        expect(page.getByText('Total count: 10')).toBeVisible()

        expect(page.getByText('First page')).toBeVisible()
        expect(() => page.getByText('Previous page')).toThrowError(/Unable to find an element/)
        expect(page.getByText('Next page')).toBeVisible()
        expect(page.getByText('Last page')).toBeVisible()

        expect(page.locationRef.current?.search).toBe(`?before=${getCursorForId('2')}`)
    })

    it('supports restoration from backward pagination URL', async () => {
        const page = await renderWithMocks(cursorMocks, true, `?before=${getCursorForId('5')}`)

        expect(page.getAllByRole('listitem').length).toBe(3)
        expect(page.getAllByRole('listitem')[0]).toHaveTextContent('result 2')
        expect(page.getAllByRole('listitem')[1]).toHaveTextContent('result 3')
        expect(page.getAllByRole('listitem')[2]).toHaveTextContent('result 4')
        expect(page.getByText('Total count: 10')).toBeVisible()

        expect(page.getByText('First page')).toBeVisible()
        expect(page.getByText('Previous page')).toBeVisible()
        expect(page.getByText('Next page')).toBeVisible()
        expect(page.getByText('Last page')).toBeVisible()
    })

    it('does not change the URL when useURL is disabled', async () => {
        const page = await renderWithMocks(cursorMocks, false, `/?after=${getCursorForId('6')}`)

        expect(page.getAllByRole('listitem').length).toBe(3)
        expect(page.getAllByRole('listitem')[0]).toHaveTextContent('result 1')
        expect(page.getAllByRole('listitem')[1]).toHaveTextContent('result 2')
        expect(page.getAllByRole('listitem')[2]).toHaveTextContent('result 3')
        expect(page.getByText('Total count: 10')).toBeVisible()

        expect(page.getByText('First page')).toBeVisible()
        expect(() => page.getByText('Previous page')).toThrowError(/Unable to find an element/)
        expect(page.getByText('Next page')).toBeVisible()
        expect(page.getByText('Last page')).toBeVisible()

        await goToLastPage(page)

        expect(page.locationRef.current?.search).toBe(`?after=${getCursorForId('6')}`)
    })
})
