import { createMemoryHistory } from 'history'
import React from 'react'

import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { EMPTY_FEATURE_FLAGS } from '../../featureFlags/featureFlags'

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
        const history = createMemoryHistory({ initialEntries: ['/'] })
        renderWithBrandedContext(
            <SiteInitPage
                isLightTheme={true}
                needsSiteInit={false}
                authenticatedUser={null}
                context={{ authProviders: [], sourcegraphDotComMode: false }}
                featureFlags={EMPTY_FEATURE_FLAGS}
            />,
            { history }
        )
        expect(history.location.pathname).toEqual('/search')
    })

    test('unexpected authed user', () =>
        expect(
            renderWithBrandedContext(
                <SiteInitPage
                    isLightTheme={true}
                    needsSiteInit={true}
                    authenticatedUser={{ username: 'alice' }}
                    context={{ authProviders: [], sourcegraphDotComMode: false }}
                    featureFlags={EMPTY_FEATURE_FLAGS}
                />
            ).asFragment()
        ).toMatchSnapshot())

    test('normal', () =>
        expect(
            renderWithBrandedContext(
                <SiteInitPage
                    isLightTheme={true}
                    needsSiteInit={true}
                    authenticatedUser={null}
                    context={{ authProviders: [], sourcegraphDotComMode: false }}
                    featureFlags={EMPTY_FEATURE_FLAGS}
                />
            ).asFragment()
        ).toMatchSnapshot())
})
