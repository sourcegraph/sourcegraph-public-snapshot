import type { MockedResponse } from '@apollo/client/testing'
import { cleanup, fireEvent } from '@testing-library/react'
import delay from 'delay'
import { escapeRegExp } from 'lodash'
import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'

import { getDocumentNode } from '@sourcegraph/http-client'
import { SymbolKind } from '@sourcegraph/shared/src/graphql-operations'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { stubScrollIntoView } from '@sourcegraph/shared/src/testing/dom-test-helpers'

import type { SymbolsResult } from '../graphql-operations'

import {
    RepoRevisionSidebarSymbols,
    SYMBOLS_QUERY,
    type RepoRevisionSidebarSymbolsProps,
} from './RepoRevisionSidebarSymbols'
import { RenderWithBrandedContextResult, renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

const location = {
    pathname: '/github.com/sourcegraph/sourcegraph@some-branch/-/blob/src/index.js',
}
const route = `${location.pathname}`

const sidebarProps: RepoRevisionSidebarSymbolsProps = {
    repoID: 'repo-id',
    revision: 'some-branch',
    activePath: 'src/index.js',
    onHandleSymbolClick: () => {},
}

const symbolsMock: MockedResponse<SymbolsResult> = {
    request: {
        query: getDocumentNode(SYMBOLS_QUERY),
        variables: {
            query: '',
            first: 100,
            repo: sidebarProps.repoID,
            revision: sidebarProps.revision,
            includePatterns: ['^' + escapeRegExp(sidebarProps.activePath)],
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
    beforeAll(stubScrollIntoView)

    let renderResult: RenderWithBrandedContextResult
    afterEach(cleanup)

    beforeEach(async () => {
        renderResult = renderWithBrandedContext(
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

    it('clicking symbol updates route', async () => {
        expect(renderResult.locationRef.current?.search).toEqual('')

        const symbol = renderResult.getByText('firstSymbol')
        fireEvent.click(symbol)

        // We need to synchronously flush inside the event handler and since this is warning in
        // React 18, we've moved it to a setTimeout. This test needs to wait for this timeout to be
        // flushed
        await delay(0)

        expect(renderResult.locationRef.current?.search).toEqual('?L13:14')
    })
})
