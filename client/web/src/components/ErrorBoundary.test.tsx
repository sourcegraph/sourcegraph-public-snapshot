import * as sentry from '@sentry/browser'
import { render } from '@testing-library/react'
import React from 'react'
import renderer from 'react-test-renderer'
import sinon from 'sinon'

import { HTTPStatusError } from '@sourcegraph/shared/src/backend/fetch'

import { ErrorBoundary } from './ErrorBoundary'

jest.mock('mdi-react/ErrorIcon', () => 'ErrorIcon')
jest.mock('mdi-react/ReloadIcon', () => 'ReloadIcon')

const ThrowError: React.FunctionComponent = () => {
    throw new Error('x')
}

/** Throws an error that resembles the Webpack error when chunk loading fails.  */
const ThrowChunkError: React.FunctionComponent = () => {
    throw new Error('Loading chunk 123 failed.')
}

const ThrowHTTPStatusError: React.FunctionComponent<{ status?: number }> = ({ status = 500 }) => {
    const errorResponse = new Response(null, { status })
    throw new HTTPStatusError(errorResponse)
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

    test('renders reload page if chunk error', () =>
        expect(
            renderer
                .create(
                    <ErrorBoundary location={null}>
                        <ThrowChunkError />
                    </ErrorBoundary>
                )
                .toJSON()
        ).toMatchSnapshot())

    test('Sentry should capture HttpStatusError except for Server response errors (5xx)', () => {
        const sentryCaptureException = sinon.stub(sentry, 'captureException')

        render(
            <ErrorBoundary location={null}>
                <ThrowHTTPStatusError status={500} />
            </ErrorBoundary>
        )

        sinon.assert.notCalled(sentryCaptureException)
        sentryCaptureException.reset()

        render(
            <ErrorBoundary location={null}>
                <ThrowHTTPStatusError status={400} />
            </ErrorBoundary>
        )

        sinon.assert.calledOnce(sentryCaptureException)

        const [capturedError] = sentryCaptureException.getCall(0).args
        expect(capturedError).toBeInstanceOf(HTTPStatusError)
        expect(capturedError).toHaveProperty('status', 400)

        sentryCaptureException.restore()
    })
})
