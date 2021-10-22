import React from 'react'
import { MemoryRouter, Redirect } from 'react-router'
import renderer from 'react-test-renderer'

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
        const component = renderer.create(
            <MemoryRouter>
                <SiteInitPage
                    isLightTheme={true}
                    needsSiteInit={false}
                    authenticatedUser={null}
                    context={{ authProviders: [], sourcegraphDotComMode: false }}
                    featureFlags={EMPTY_FEATURE_FLAGS}
                />
            </MemoryRouter>
        )
        const redirect = component.root.findByType(Redirect)
        expect(redirect).toBeDefined()
        expect(redirect.props.to).toEqual('/search')
    })

    test('unexpected authed user', () =>
        expect(
            renderer
                .create(
                    <SiteInitPage
                        isLightTheme={true}
                        needsSiteInit={true}
                        authenticatedUser={{ username: 'alice' }}
                        context={{ authProviders: [], sourcegraphDotComMode: false }}
                        featureFlags={EMPTY_FEATURE_FLAGS}
                    />
                )
                .toJSON()
        ).toMatchSnapshot())

    test('normal', () =>
        expect(
            renderer
                .create(
                    <SiteInitPage
                        isLightTheme={true}
                        needsSiteInit={true}
                        authenticatedUser={null}
                        context={{ authProviders: [], sourcegraphDotComMode: false }}
                        featureFlags={EMPTY_FEATURE_FLAGS}
                    />
                )
                .toJSON()
        ).toMatchSnapshot())
})
