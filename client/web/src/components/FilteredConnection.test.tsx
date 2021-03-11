import React from 'react'
import { createLocation } from 'history'
import { cleanup, render, screen } from '@testing-library/react'

import { ConnectionNodesForTesting as ConnectionNodes } from './FilteredConnection'

function fakeConnection({ hasNextPage, totalCount }: { hasNextPage: boolean; totalCount: number | null }) {
    return {
        nodes: [{}],
        pageInfo: {
            endCursor: '',
            hasNextPage,
        },
        totalCount,
    }
}

/** A default set of props that are required by the ConnectionNodes component,
    but that are not relevant to any of the things under test here. */
const boringConnectionNodesProps = {
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

describe('ConnectionNodes', () => {
    afterAll(cleanup)

    it('should have a "Show more" button when *not* loading', () => {
        render(
            <ConnectionNodes
                connection={fakeConnection({ hasNextPage: true, totalCount: 2 })}
                loading={false}
                {...boringConnectionNodesProps}
            />
        )
        expect(screen.getByRole('button')).toHaveTextContent('Show more')
        expect(screen.getByText('2 cats total')).toBeVisible()
        expect(screen.getByText('(showing first 1)')).toBeVisible()
    })

    it("should *not* have a 'Show more' button when loading", () => {
        render(
            <ConnectionNodes
                connection={fakeConnection({ hasNextPage: true, totalCount: 2 })}
                loading={true}
                {...boringConnectionNodesProps}
            />
        )
        expect(screen.queryByRole('button')).toBeNull()
        expect(screen.getByText('2 cats total')).toBeVisible()
        expect(screen.getByText('(showing first 1)')).toBeVisible()
        // NOTE: we also expect a LoadingSpinner, but that is not provided by ConnectionNodes.
    })

    it("it doesn't show summary info if totalCount is null", () => {
        render(
            <ConnectionNodes
                connection={fakeConnection({ hasNextPage: true, totalCount: null })}
                loading={true}
                {...boringConnectionNodesProps}
            />
        )
        expect(screen.queryByText('2 cats total')).toBeNull()
        expect(screen.queryByText('(showing first 1)')).toBeNull()
    })
})
