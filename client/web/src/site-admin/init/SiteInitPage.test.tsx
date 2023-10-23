import { afterAll, beforeAll, describe, expect, test } from '@jest/globals'

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
                />
            ).asFragment()
        ).toMatchSnapshot())
})
