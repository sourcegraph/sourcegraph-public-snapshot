import React from 'react'
import { ErrorBoundary } from './ErrorBoundary'
import { mount } from 'enzyme'

jest.mock('mdi-react/ErrorIcon', () => 'ErrorIcon')
jest.mock('mdi-react/ReloadIcon', () => 'ReloadIcon')

const ThrowError: React.FunctionComponent = () => {
    throw new Error('x')
}

/** Throws an error that resembles the Webpack error when chunk loading fails.  */
const ThrowChunkError: React.FunctionComponent = () => {
    throw new Error('Loading chunk 123 failed.')
}

describe('ErrorBoundary', () => {
    test('passes through if non-error', () =>
        expect(
            mount(
                <ErrorBoundary location={null}>
                    <ThrowError />
                </ErrorBoundary>
            ).children()
        ).toMatchSnapshot())

    test('renders error page if error', () =>
        expect(
            mount(
                <ErrorBoundary location={null}>
                    <span>hello</span>
                </ErrorBoundary>
            ).children()
        ).toMatchSnapshot())

    test('renders reload page if chunk error', () =>
        expect(
            mount(
                <ErrorBoundary location={null}>
                    <ThrowChunkError />
                </ErrorBoundary>
            ).children()
        ).toMatchSnapshot())
})
