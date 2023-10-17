import React from 'react'

import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { ErrorBoundary } from './ErrorBoundary'

jest.mock('mdi-react/AlertCircleIcon', () => 'AlertCircleIcon')
jest.mock('mdi-react/ReloadIcon', () => 'ReloadIcon')

const ThrowError: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => {
    throw new Error('x')
}

/** Throws an error that resembles the  error when a dynamic import(...) fails.  */
const ThrowDynamicImportError: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => {
    throw new TypeError('Failed to fetch dynamically imported module: https://example.com/x.js')
}

/** Throws an error that resembles the Webpack error when chunk loading fails.  */
const ThrowChunkError: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => {
    const ChunkError = new Error('Loading chunk 123 failed.')
    ChunkError.name = 'ChunkLoadError'
    throw ChunkError
}

describe('ErrorBoundary', () => {
    test('passes through if non-error', () =>
        expect(
            renderWithBrandedContext(
                <ErrorBoundary location={null}>
                    <ThrowError />
                </ErrorBoundary>
            ).asFragment()
        ).toMatchSnapshot())

    test('renders error page if error', () =>
        expect(
            renderWithBrandedContext(
                <ErrorBoundary location={null}>
                    <span>hello</span>
                </ErrorBoundary>
            ).asFragment()
        ).toMatchSnapshot())

    test('renders reload page if dynamic import error', () =>
        expect(
            renderWithBrandedContext(
                <ErrorBoundary location={null}>
                    <ThrowDynamicImportError />
                </ErrorBoundary>
            ).asFragment()
        ).toMatchSnapshot())

    test('renders reload page if webpack chunk error', () =>
        expect(
            renderWithBrandedContext(
                <ErrorBoundary location={null}>
                    <ThrowChunkError />
                </ErrorBoundary>
            ).asFragment()
        ).toMatchSnapshot())
})
