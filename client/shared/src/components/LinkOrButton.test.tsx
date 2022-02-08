import React from 'react'

import { renderWithBrandedContext } from '../testing'

import { ButtonLink } from './LinkOrButton'

describe('LinkOrButton', () => {
    test('render a link when "to" is set', () => {
        const component = renderWithBrandedContext(<ButtonLink to="http://example.com">foo</ButtonLink>)
        expect(component.asFragment()).toMatchSnapshot()
    })

    test('render a button when "to" is undefined', () => {
        const component = renderWithBrandedContext(<ButtonLink to={undefined}>foo</ButtonLink>)
        expect(component.asFragment()).toMatchSnapshot()
    })
})
