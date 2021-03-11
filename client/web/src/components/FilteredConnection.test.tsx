import React from 'react'
import '@testing-library/jest-dom/extend-expect'
import { cleanup, render, screen } from '@testing-library/react'

import { ConnectionNodesForTesting as ConnectionNodes } from './FilteredConnection'

const emptyLocation = {
    pathname: '',
    search: '',
    hash: '',
    state: {},
    key: '',
}

function fakeConnection(pageInfoProps: { hasNextPage: boolean }) {
    return {
        nodes: [],
        pageInfo: {
            endCursor: '',
            ...pageInfoProps,
        },
        totalCount: 1,
    }
}

/** A default set of props that are required by the ConnectionNodes component,
    but that are not relevant to any of the things under test here. */
const boringConnectionNodesProps = {
    connectionQuery: '',
    first: 0,
    location: emptyLocation,
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

    it('pdubroy it should have a "Show more" button when not loading', () => {
        render(
            <ConnectionNodes
                connection={fakeConnection({ hasNextPage: true })}
                loading={false}
                {...boringConnectionNodesProps}
            />
        )
        expect(screen.getByRole('button')).toHaveTextContent('Show more')
        expect(screen.queryByText('1 cat total')).toBeNull()
        expect(screen.queryByText('(showing first 0)')).toBeNull()
    })

    it('pdubroy it should have no button text when loading', () => {
        render(
            <ConnectionNodes
                connection={fakeConnection({ hasNextPage: true })}
                loading={true}
                {...boringConnectionNodesProps}
            />
        )
        expect(screen.queryByRole('button')).toBeNull()
        expect(screen.getByText('1 cat total')).toBeVisible()
        expect(screen.getByText('(showing first 0)')).toBeVisible()
        // NOTE: we also expect a LoadingSpinner, but that is not provided by ConnectionNodes.
    })
})
