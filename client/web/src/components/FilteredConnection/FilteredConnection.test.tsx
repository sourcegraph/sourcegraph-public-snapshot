import { cleanup, fireEvent, render, screen, waitFor, act } from '@testing-library/react'
import { createLocation, createMemoryHistory } from 'history'
import { BehaviorSubject } from 'rxjs'
import sinon from 'sinon'

import { ConnectionNodes } from './ConnectionNodes'
import { FilteredConnection } from './FilteredConnection'

function fakeConnection<N>({
    hasNextPage,
    totalCount,
    nodes,
}: {
    hasNextPage: boolean
    totalCount: number | null
    nodes: N[]
}) {
    return {
        nodes,
        pageInfo: {
            endCursor: '',
            hasNextPage,
        },
        totalCount,
    }
}

/** A default set of props that are required by the ConnectionNodes component */
const defaultConnectionNodesProps = {
    connectionQuery: '',
    first: 0,
    location: createLocation('/'),
    noSummaryIfAllNodesVisible: true,
    nodeComponent: () => null,
    nodeComponentProps: {},
    noun: 'cat',
    onShowMore: () => {},
    pluralNoun: 'cats',
    query: '',
}

describe('FilteredConnection', () => {
    afterAll(cleanup)

    describe('useURLQuery', () => {
        it('does not update the history if searchFragment didnt change', async () => {
            const history = createMemoryHistory()
            history.push('/?foo=bar')

            // Hook into history.replace to detect when FilteredConnection updates search
            // fragment
            const oldReplaceHistory = history.replace.bind(history)
            const fakeReplaceHistory = sinon.spy(args => oldReplaceHistory(args))
            history.replace = fakeReplaceHistory

            // This is the fake connection
            const connection = fakeConnection({ hasNextPage: true, totalCount: 2, nodes: [{}] })
            const connectionSubject = new BehaviorSubject(connection)

            render(
                <FilteredConnection
                    {...defaultConnectionNodesProps}
                    location={history.location}
                    history={history}
                    useURLQuery={true}
                    queryConnection={() => connectionSubject}
                />
            )

            act(() => {
                // Emulate polling by re-emiting the connection again.
                // This should not lead to history being updated
                connectionSubject.next(connection)
                connectionSubject.next(connection)
                connectionSubject.next(connection)
            })

            // Click "Show more" button, should cause history to be updated
            fireEvent.click(screen.getByRole('button')!)
            expect(history.location.search).toEqual('?foo=bar&first=40')
            fireEvent.click(screen.getByRole('button')!)
            expect(history.location.search).toEqual('?foo=bar&first=80')

            await waitFor(() => sinon.assert.calledTwice(fakeReplaceHistory))
        })
    })
})

describe('ConnectionNodes', () => {
    afterAll(cleanup)

    it('has a "Show more" button and summary when *not* loading', () => {
        render(
            <ConnectionNodes
                {...defaultConnectionNodesProps}
                connection={fakeConnection({ hasNextPage: true, totalCount: 2, nodes: [{}] })}
                loading={false}
            />
        )
        expect(screen.getByRole('button')).toHaveTextContent('Show more')
        expect(screen.getByText('2 cats total')).toBeVisible()
        expect(screen.getByText('(showing first 1)')).toBeVisible()
    })

    it("*doesn't* have a 'Show more' button or summary when loading", () => {
        render(
            <ConnectionNodes
                {...defaultConnectionNodesProps}
                connection={fakeConnection({ hasNextPage: true, totalCount: 2, nodes: [{}] })}
                loading={true}
            />
        )
        expect(screen.queryByRole('button')).not.toBeInTheDocument()
        expect(screen.queryByText('2 cats total')).not.toBeInTheDocument()
        expect(screen.queryByText('(showing first 1)')).not.toBeInTheDocument()
        // NOTE: we also expect a LoadingSpinner, but that is not provided by ConnectionNodes.
    })

    it("doesn't have a 'Show more' button when noShowMore is true", () => {
        render(
            <ConnectionNodes
                {...defaultConnectionNodesProps}
                connection={fakeConnection({ hasNextPage: true, totalCount: 2, nodes: [{}] })}
                loading={false}
                noShowMore={true}
            />
        )
        expect(screen.queryByRole('button')).not.toBeInTheDocument()
        expect(screen.getByText('2 cats total')).toBeVisible()
        expect(screen.getByText('(showing first 1)')).toBeVisible()
    })

    it("doesn't have a 'Show more' button or a summary if hasNextPage is false ", () => {
        render(
            <ConnectionNodes
                {...defaultConnectionNodesProps}
                connection={fakeConnection({ hasNextPage: false, totalCount: 1, nodes: [{}] })}
                loading={false}
            />
        )
        expect(screen.queryByRole('button')).not.toBeInTheDocument()
        expect(screen.queryByTestId('summary')).not.toBeInTheDocument()
    })

    it('calls the onShowMore callback', async () => {
        const showMoreCallback = sinon.spy(() => undefined)
        render(
            <ConnectionNodes
                {...defaultConnectionNodesProps}
                connection={fakeConnection({ hasNextPage: true, totalCount: 2, nodes: [{}] })}
                loading={false}
                onShowMore={showMoreCallback}
            />
        )
        fireEvent.click(screen.getByRole('button')!)
        await waitFor(() => sinon.assert.calledOnce(showMoreCallback))
    })

    it("doesn't show summary info if totalCount is null", () => {
        render(
            <ConnectionNodes
                {...defaultConnectionNodesProps}
                connection={fakeConnection({ hasNextPage: true, totalCount: null, nodes: [{}] })}
                loading={false}
            />
        )
        expect(screen.queryByTestId('summary')).not.toBeInTheDocument()
    })

    it('shows a summary if noSummaryIfAllNodesVisible is false', () => {
        render(
            <ConnectionNodes
                {...defaultConnectionNodesProps}
                connection={fakeConnection({ hasNextPage: false, totalCount: 1, nodes: [{}] })}
                loading={false}
                noSummaryIfAllNodesVisible={false}
            />
        )
        expect(screen.getByText('1 cat total')).toBeVisible()
        expect(screen.queryByText('(showing first 1)')).not.toBeInTheDocument()

        // Summary should come after the nodes.
        expect(
            screen.getByTestId('summary')!.compareDocumentPosition(screen.getByTestId('filtered-connection-nodes'))
        ).toEqual(Node.DOCUMENT_POSITION_PRECEDING)
    })

    it('shows a summary if nodes.length is 0', () => {
        render(
            <ConnectionNodes
                {...defaultConnectionNodesProps}
                connection={fakeConnection({ hasNextPage: false, totalCount: 1, nodes: [] })}
                loading={false}
            />
        )
        expect(screen.getByText('1 cat total')).toBeVisible()
        expect(screen.queryByText('(showing first 1)')).not.toBeInTheDocument()
    })

    it("shows 'No cats' if totalCount is 0", () => {
        render(
            <ConnectionNodes
                {...defaultConnectionNodesProps}
                connection={fakeConnection({ hasNextPage: false, totalCount: 0, nodes: [] })}
                loading={false}
            />
        )
        expect(screen.getByText('No cats')).toBeVisible()
    })

    it('shows the summary at the top if connectionQuery is specified', () => {
        render(
            <ConnectionNodes
                {...defaultConnectionNodesProps}
                connection={fakeConnection({ hasNextPage: true, totalCount: 2, nodes: [{}] })}
                loading={true}
                connectionQuery="meow?"
            />
        )
        // Summary should come _before_ the nodes.
        expect(
            screen.getByTestId('summary')!.compareDocumentPosition(screen.getByTestId('filtered-connection-nodes'))
        ).toEqual(Node.DOCUMENT_POSITION_FOLLOWING)
    })

    it('shows the summary at the top if connectionQuery is specified', () => {
        render(
            <ConnectionNodes
                {...defaultConnectionNodesProps}
                connection={fakeConnection({ hasNextPage: true, totalCount: 2, nodes: [{}] })}
                loading={true}
                connectionQuery="meow?"
            />
        )
        // Summary should come _before_ the nodes.
        expect(
            screen.getByTestId('summary')!.compareDocumentPosition(screen.getByTestId('filtered-connection-nodes'))
        ).toEqual(Node.DOCUMENT_POSITION_FOLLOWING)
    })
})
