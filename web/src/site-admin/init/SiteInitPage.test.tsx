import React from 'react'
import { SiteInitPage } from './SiteInitPage'
import { MemoryRouter, Redirect } from 'react-router'
import { createMemoryHistory } from 'history'
import { mount } from 'enzyme'

describe('SiteInitPage', () => {
    const origContext = window.context
    beforeAll(() => {
        window.context = {} as any
    })
    afterAll(() => {
        window.context = origContext
    })

    test('site already initialized', () => {
        const component = mount(
            <MemoryRouter>
                <SiteInitPage
                    isLightTheme={true}
                    needsSiteInit={false}
                    authenticatedUser={null}
                    history={createMemoryHistory()}
                />
            </MemoryRouter>
        )
        const redirect = component.find(Redirect)
        expect(redirect).toBeDefined()
        expect(redirect.props().to).toEqual('/search')
    })

    test('unexpected authed user', () =>
        expect(
            mount(
                <SiteInitPage
                    isLightTheme={true}
                    needsSiteInit={true}
                    authenticatedUser={{ username: 'alice' }}
                    history={createMemoryHistory()}
                />
            ).children()
        ).toMatchSnapshot())

    test('normal', () =>
        expect(
            mount(
                <SiteInitPage
                    isLightTheme={true}
                    needsSiteInit={true}
                    authenticatedUser={null}
                    history={createMemoryHistory()}
                />
            ).children()
        ).toMatchSnapshot())
})
