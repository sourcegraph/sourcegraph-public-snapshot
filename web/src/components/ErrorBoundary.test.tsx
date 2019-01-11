import React from 'react'
import renderer from 'react-test-renderer'
import { ErrorBoundary } from './ErrorBoundary'

const ThrowError: React.FunctionComponent = () => {
    throw new Error('x')
}

describe('ErrorBoundary', () => {
    test('passes through if non-error', () =>
        expect(
            renderer
                .create(
                    <ErrorBoundary location={null}>
                        <ThrowError />
                    </ErrorBoundary>
                )
                .toJSON()
        ).toMatchSnapshot())

    test('renders error page if error', () =>
        expect(
            renderer
                .create(
                    <ErrorBoundary location={null}>
                        <span>hello</span>
                    </ErrorBoundary>
                )
                .toJSON()
        ).toMatchSnapshot())
})
