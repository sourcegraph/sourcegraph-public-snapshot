import { afterAll, beforeAll, describe, expect, test } from 'vitest'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { SiteInitPage } from './SiteInitPage'

describe('SiteInitPage', () => {
    const origContext = window.context
    beforeAll(() => {
        window.context = {
            authProviders: [],
        } as any
    })
    afterAll(() => {
        window.context = origContext
    })

    test('site already initialized', () => {
        const result = renderWithBrandedContext(
            <SiteInitPage
                needsSiteInit={false}
                authenticatedUser={null}
                context={{
                    authPasswordPolicy: {},
                    authMinPasswordLength: 12,
                }}
                telemetryRecorder={noOpTelemetryRecorder}
            />,
            { route: '/init', path: '/init', extraRoutes: [{ path: '/search', element: null }] }
        )

        expect(result.locationRef.current?.pathname).toEqual('/search')
    })

    test('unexpected authed user', () =>
        expect(
            renderWithBrandedContext(
                <SiteInitPage
                    needsSiteInit={true}
                    authenticatedUser={{ username: 'alice' }}
                    context={{
                        authPasswordPolicy: {},
                        authMinPasswordLength: 12,
                    }}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            ).asFragment()
        ).toMatchSnapshot())

    test('normal', () =>
        expect(
            renderWithBrandedContext(
                <SiteInitPage
                    needsSiteInit={true}
                    authenticatedUser={null}
                    context={{
                        authPasswordPolicy: {},
                        authMinPasswordLength: 12,
                    }}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            ).asFragment()
        ).toMatchSnapshot())
})
