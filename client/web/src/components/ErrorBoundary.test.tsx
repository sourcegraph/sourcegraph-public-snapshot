import * as sentry from '@sentry/browser'
import { render } from '@testing-library/react'
import React from 'react'
import sinon from 'sinon'

import { AbortError } from '@sourcegraph/shared/src/api/util'
import { HTTPStatusError } from '@sourcegraph/shared/src/backend/fetch'

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

const ThrowAbortError: React.FunctionComponent = () => {
    throw new AbortError()
}

const ThrowNotAuthenticatedError: React.FunctionComponent = () => {
    throw new Error('not authenticated')
}

const ThrowHTTPStatusError: React.FunctionComponent<{ status?: number }> = ({ status = 500 }) => {
    const errorResponse = new Response(null, { status })
    throw new HTTPStatusError(errorResponse)
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

    test('Sentry should not capture AbortError', () => {
        const sentryCaptureException = sinon.stub(sentry, 'captureException')

        render(
            <ErrorBoundary location={null}>
                <ThrowAbortError />
            </ErrorBoundary>
        )

        sinon.assert.notCalled(sentryCaptureException)
        sentryCaptureException.restore()
    })

    test('Sentry should not capture ChunkLoadError', () => {
        const sentryCaptureException = sinon.stub(sentry, 'captureException')

        render(
            <ErrorBoundary location={null}>
                <ThrowChunkError />
            </ErrorBoundary>
        )

        sinon.assert.notCalled(sentryCaptureException)
        sentryCaptureException.restore()
    })

    test('Sentry should not capture not authenticated error', () => {
        const sentryCaptureException = sinon.stub(sentry, 'captureException')

        render(
            <ErrorBoundary location={null}>
                <ThrowNotAuthenticatedError />
            </ErrorBoundary>
        )

        sinon.assert.notCalled(sentryCaptureException)
        sentryCaptureException.restore()
    })
})
