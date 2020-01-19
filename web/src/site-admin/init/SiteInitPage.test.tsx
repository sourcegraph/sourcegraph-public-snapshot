import React from 'react'
import renderer from 'react-test-renderer'
import { SiteInitPage } from './SiteInitPage'
import * as H from 'history'
import { MemoryRouter, Redirect } from 'react-router'

describe('SiteInitPage', () => {
    const origContext = window.context
    beforeAll(() => {
        window.context = {} as any
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
                    history={H.createMemoryHistory()}
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
                        history={H.createMemoryHistory()}
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
                        history={H.createMemoryHistory()}
                    />
                )
                .toJSON()
        ).toMatchSnapshot())
})
