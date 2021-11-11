import { render } from '@testing-library/react'
import React from 'react'

import { ErrorBoundary } from './ErrorBoundary'

jest.mock('mdi-react/AlertCircleIcon', () => 'AlertCircleIcon')
jest.mock('mdi-react/ReloadIcon', () => 'ReloadIcon')

const ThrowError: React.FunctionComponent = () => {
    throw new Error('x')
}

/** Throws an error that resembles the Webpack error when chunk loading fails.  */
const ThrowChunkError: React.FunctionComponent = () => {
    const ChunkError = new Error('Loading chunk 123 failed.')
    ChunkError.name = 'ChunkLoadError'
    throw ChunkError
}

describe('ErrorBoundary', () => {
    test('passes through if non-error', () =>
        expect(
            render(
                <ErrorBoundary location={null}>
                    <ThrowError />
                </ErrorBoundary>
            ).asFragment()
        ).toMatchSnapshot())

    test('renders error page if error', () =>
        expect(
            render(
                <ErrorBoundary location={null}>
                    <span>hello</span>
                </ErrorBoundary>
            ).asFragment()
        ).toMatchSnapshot())

    test('renders reload page if chunk error', () =>
        expect(
            render(
                <ErrorBoundary location={null}>
                    <ThrowChunkError />
                </ErrorBoundary>
            ).asFragment()
        ).toMatchSnapshot())
})
