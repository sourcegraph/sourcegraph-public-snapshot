import React from 'react'

import { renderWithRouter } from '../testing/render-with-router'

import { ButtonLink } from './LinkOrButton'

describe('LinkOrButton', () => {
    test('render a link when "to" is set', () => {
        const component = renderWithRouter(<ButtonLink to="http://example.com">foo</ButtonLink>)
        expect(component.asFragment()).toMatchSnapshot()
    })

    test('render a button when "to" is undefined', () => {
        const component = renderWithRouter(<ButtonLink to={undefined}>foo</ButtonLink>)
        expect(component.asFragment()).toMatchSnapshot()
    })
})
