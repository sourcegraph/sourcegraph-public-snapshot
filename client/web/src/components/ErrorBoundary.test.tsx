import React from 'react'

import { describe, expect, test, vi } from 'vitest'

import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { ErrorBoundary } from './ErrorBoundary'

vi.mock('mdi-react/AlertCircleIcon', () => () => 'AlertCircleIcon')
vi.mock('mdi-react/ReloadIcon', () => () => 'ReloadIcon')

const ThrowError: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => {
    throw new Error('x')
}

/** Throws an error that resembles the  error when a dynamic import(...) fails.  */
const ThrowDynamicImportError: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => {
    throw new TypeError('Failed to fetch dynamically imported module: https://example.com/x.js')
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
})
