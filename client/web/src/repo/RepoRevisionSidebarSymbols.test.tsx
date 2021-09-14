import { MockedResponse } from '@apollo/client/testing'
import { cleanup, fireEvent, render, RenderResult } from '@testing-library/react'
import { escapeRegExp } from 'lodash'
import React from 'react'
import { MemoryRouter } from 'react-router'

import { SymbolKind } from '@sourcegraph/shared/src/graphql-operations'
import { getDocumentNode } from '@sourcegraph/shared/src/graphql/graphql'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithRouter } from '@sourcegraph/shared/src/testing/render-with-router'

import { SymbolsResult } from '../graphql-operations'

import {
    RepoRevisionSidebarSymbols,
    RepoRevisionSidebarSymbolsProps,
    SYMBOLS_QUERY,
} from './RepoRevisionSidebarSymbols'

const location = {
    pathname: '/github.com/sourcegraph/sourcegraph@some-branch/-/blob/src/index.js',
}
const route = `${location.pathname}`

const sidebarProps: RepoRevisionSidebarSymbolsProps = {
    repoID: 'repo-id',
    revision: 'some-branch',
    activePath: 'src/index.js',
}

const symbolsMock: MockedResponse<SymbolsResult> = {
    request: {
        query: getDocumentNode(SYMBOLS_QUERY),
        variables: {
            query: '',
            first: 100,
            repo: sidebarProps.repoID,
            revision: sidebarProps.revision,
            includePatterns: [escapeRegExp(sidebarProps.activePath)],
        },
    },
    result: {
        data: {
            node: {
                __typename: 'Repository',
                commit: {
                    symbols: {
                        __typename: 'SymbolConnection',
                        nodes: [
                            {
                                __typename: 'Symbol',
                                kind: SymbolKind.CONSTANT,
                                language: 'TypeScript',
                                name: 'firstSymbol',
                                url: `${location.pathname}?L13:14`,
                                containerName: null,
                                location: {
                                    resource: {
                                        path: 'src/index.js',
                                    },
                                    range: null,
                                },
                            },
                        ],
                        pageInfo: {
                            hasNextPage: false,
                        },
                    },
                },
            },
        },
    },
}

describe('RepoRevisionSidebarSymbols', () => {
    let renderResult: RenderResult
    afterEach(cleanup)

    beforeEach(async () => {
        renderResult = render(
            <MemoryRouter initialEntries={[route]}>
                <MockedTestProvider mocks={[symbolsMock]} addTypename={true}>
                    <RepoRevisionSidebarSymbols {...sidebarProps} />
                </MockedTestProvider>
            </MemoryRouter>
        )
        await waitForNextApolloResponse()
    })

    it('renders all symbols correctly', () => {
        renderResult.debug()
        const symbol = renderResult.getByText('firstSymbol')
        expect(symbol).toBeVisible()
        expect(symbol.parentElement).toHaveTextContent('firstSymbolsrc/index.js')
    })
})
