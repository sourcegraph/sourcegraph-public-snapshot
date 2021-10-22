import { MockedResponse } from '@apollo/client/testing'
import { cleanup, fireEvent } from '@testing-library/react'
import { escapeRegExp } from 'lodash'
import React from 'react'

import { SymbolKind } from '@sourcegraph/shared/src/graphql-operations'
import { getDocumentNode } from '@sourcegraph/shared/src/graphql/graphql'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithRouter, RenderWithRouterResult } from '@sourcegraph/shared/src/testing/render-with-router'

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
    let renderResult: RenderWithRouterResult
    afterEach(cleanup)

    beforeEach(async () => {
        renderResult = renderWithRouter(
            <MockedTestProvider mocks={[symbolsMock]} addTypename={true}>
                <RepoRevisionSidebarSymbols {...sidebarProps} />
            </MockedTestProvider>,
            { route }
        )
        // NOTE: (@numbers88s)
        // See https://github.com/mui-org/material-ui/issues/15726#issuecomment-876323860
        // Bootstrap's implementation of Tooltip uses PopperJS. The issue is with the underlying
        // implementation of PopperJS calling the document.createRange function when there is no DOM API for it to call.
        // The solution is to upgrade to Jest v26.0.0 (breaking changes no backward compatibility), mock PopperJS
        // or to mock the underlying function that it utilizes. I chose the latter.
        // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
        ;(global as any).document.createRange = () => ({
            setStart: () => {},
            setEnd: () => {},
            commonAncestorContainer: {
                nodeName: 'BODY',
                ownerDocument: document,
            },
        })

        await waitForNextApolloResponse()
    })

    it('renders symbol correctly', () => {
        const symbol = renderResult.getByText('firstSymbol')
        expect(symbol).toBeVisible()
    })

    it('renders summary correctly', () => {
        expect(renderResult.getByText('1 symbol total')).toBeVisible()
    })

    it('clicking symbol updates route', () => {
        expect(renderResult.history.location.search).toEqual('')

        const symbol = renderResult.getByText('firstSymbol')
        fireEvent.click(symbol)

        expect(renderResult.history.location.search).toEqual('?L13:14')
    })
})
